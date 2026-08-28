package main

// This file holds what used to be the standalone oauth service
// (back/oauth): account registration, login (token), and token
// introspection. It now runs in the same process as the rest of the API, on
// the same port, so tryAuth() (shared.go) calls validateToken() directly
// instead of making an HTTP round-trip to /oauth/v3/introspect - that
// endpoint is kept only so external/spec-compliant OAuth2 clients can still
// call it.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// readBasicCredentials decodes a "Basic <base64(identifier:secret)>" header
// into its two parts.
func readBasicCredentials(authHeader string) (string, string, error) {
	if authHeader == "" {
		return "", "", http.ErrNoCookie
	}

	authParts := strings.SplitN(authHeader, " ", 2)
	if len(authParts) != 2 || authParts[0] != "Basic" {
		return "", "", http.ErrNoCookie
	}

	basicDecoded, err := base64.StdEncoding.DecodeString(authParts[1])
	if err != nil {
		return "", "", err
	}

	creds := strings.SplitN(string(basicDecoded), ":", 2)
	if len(creds) != 2 {
		return "", "", http.ErrNoCookie
	}

	return creds[0], creds[1], nil
}

func isEmailIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}
	_, err := mail.ParseAddress(identifier)
	return err == nil
}

func generateTokenValue() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// validateToken looks up a bearer token and returns the IntrospectionPayload
// tryAuth() and /oauth/v3/introspect both rely on. Active is only true for a
// token that exists, is not expired/revoked, and whose account is actif.
func validateToken(ctx context.Context, tokenValue string) *IntrospectionPayload {
	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		return &IntrospectionPayload{}
	}
	defer conn.Close(ctx)

	var tokenID int64
	var tokenActive, tokenRevoked string
	err = conn.QueryRow(ctx, "select token_id, active, revoked from token where token_value=$1", tokenValue).Scan(&tokenID, &tokenActive, &tokenRevoked)
	if err != nil || tokenActive == "0" || tokenRevoked == "1" {
		return &IntrospectionPayload{}
	}

	var compteID int
	var email, userType, actif string
	err = conn.QueryRow(
		ctx,
		`select compte.compte_id, personne.email, compte.type_utilisateur, compte.actif
		from token
		inner join compte on compte.compte_id = token.compte_id
		inner join personne on personne.personne_id = compte.personne_id
		where token.token_id = $1`,
		tokenID,
	).Scan(&compteID, &email, &userType, &actif)
	if err != nil || actif != "1" {
		return &IntrospectionPayload{}
	}

	return &IntrospectionPayload{Active: true, UserType: userType, CompteID: compteID, Username: email}
}

// token issues an access token. The caller authenticates with HTTP Basic
// auth using their personne.email and compte.mot_de_passe.
func token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	email, secret, err := readBasicCredentials(r.Header.Get("Authorization"))
	if err != nil || !isEmailIdentifier(email) {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	var compteID int
	var hashedSecret, actif string
	err = conn.QueryRow(
		ctx,
		"select compte.compte_id, compte.mot_de_passe, compte.actif from compte inner join personne on personne.personne_id = compte.personne_id where personne.email = $1",
		email,
	).Scan(&compteID, &hashedSecret, &actif)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	if actif != "1" {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "This account has been deactivated.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secret)); err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	accessToken, err := generateTokenValue()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = conn.Exec(
		ctx,
		"insert into token (compte_id, token_value, date_expiration, active, revoked) values ($1, $2, now() + interval '1 hour', '1', '0')",
		compteID,
		accessToken,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = conn.Exec(ctx, "update compte set dernier_login = now() where compte_id = $1", compteID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        "",
	})
}

// typeUtilisateur lets a client check its own type_utilisateur without going
// through the full introspection flow. This replaces the old /roles endpoint
// now that there is no role table to enumerate.
func typeUtilisateur(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	email, secret, err := readBasicCredentials(r.Header.Get("Authorization"))
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	var compteID int
	var hashedSecret, userType string
	err = conn.QueryRow(
		ctx,
		"select compte.compte_id, compte.mot_de_passe, compte.type_utilisateur from compte inner join personne on personne.personne_id = compte.personne_id where personne.email = $1",
		email,
	).Scan(&compteID, &hashedSecret, &userType)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(secret)); err != nil {
		writeAPIError(w, http.StatusUnauthorized, "invalid_client", "The requested service needs credentials, but the ones provided were invalid.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"compte_id":        strconv.Itoa(compteID),
		"type_utilisateur": userType,
	})
}

// introspect is the standard OAuth2 token introspection endpoint (RFC 7662).
// Nothing in this service calls it internally any more (tryAuth uses
// validateToken directly) - it is kept for external/spec-compliant callers.
// The scope it returns IS the account's type_utilisateur (a single value,
// never a list of roles).
func introspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		writeAPIError(w, http.StatusBadRequest, "invalid_client", "The URI does not support the requested content type.")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		writeAPIError(w, http.StatusBadRequest, "invalid_client", "The client failed to authenticate with the authorization server.")
		return
	}

	payload := validateToken(r.Context(), strings.TrimPrefix(authHeader, "Bearer "))
	if !payload.Active {
		writeJSON(w, http.StatusOK, map[string]any{
			"active": false, "scope": nil, "client_id": nil, "username": nil,
			"iat": nil, "exp": nil, "sub": nil, "aud": nil, "iss": nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"active":     true,
		"scope":      payload.UserType,
		"client_id":  payload.CompteID,
		"username":   payload.Username,
		"token_type": "Bearer",
		"iat":        nil,
		"exp":        nil,
		"sub":        nil,
		"aud":        nil,
		"iss":        nil,
	})
}

// createAccount registers a new personne + compte. The use case diagram only
// exposes self-registration "comme Commercant / Adherant / Visiteur" - so
// account_type is restricted to those three. Becoming benevole requires a
// candidature reviewed by staff (see benevoles.go), and
// responsable/administrateur accounts can only be created by an
// administrateur (see admin.go).
func createAccount(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body.")
		return
	}

	email := strings.TrimSpace(r.PostForm.Get("email"))
	secret := r.PostForm.Get("mot_de_passe")
	confirm := r.PostForm.Get("mot_de_passe_confirmation")
	nom := r.PostForm.Get("nom")
	prenom := r.PostForm.Get("prenom")
	telephone := r.PostForm.Get("telephone")
	adresse := r.PostForm.Get("adresse")
	codePostal := r.PostForm.Get("code_postal")
	ville := r.PostForm.Get("ville")
	pays := "France"

	if email == "" || secret == "" || nom == "" || prenom == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: email, mot_de_passe, nom, prenom")
		return
	}
	if !isEmailIdentifier(email) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "email must be a valid email address")
		return
	}
	if secret != confirm {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Passwords do not match.")
		return
	}
	if telephone != "" && !telephoneRegexp.MatchString(telephone) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "telephone must be in international format, e.g. +33612345678")
		return
	}
	if codePostal != "" && !codePostalRegexp.MatchString(codePostal) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "code_postal must be exactly 5 digits")
		return
	}

	accountType := strings.ToLower(strings.TrimSpace(r.PostForm.Get("type_utilisateur")))
	if accountType == "" {
		accountType = UserTypeVisiteur
	}
	if !isSelfServiceUserType(accountType) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "type_utilisateur must be one of: visiteur, adherent, commercant")
		return
	}

	conn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Close(ctx)

	var count int64
	err = conn.QueryRow(ctx, "select count(*) from personne where email=$1", email).Scan(&count)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "A personne with this email already exists.")
		return
	}

	hashedPassword := hashPassword(secret)
	if hashedPassword == "" {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	var personneID int
	err = tx.QueryRow(
		ctx,
		"insert into personne (nom, prenom, email, telephone, adresse, code_postal, ville, pays) values ($1, $2, $3, nullif($4,''), nullif($5,''), nullif($6,''), nullif($7,''), $8) returning personne_id",
		nom, prenom, email, telephone, adresse, codePostal, ville, pays,
	).Scan(&personneID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var compteID int
	err = tx.QueryRow(
		ctx,
		"insert into compte (personne_id, mot_de_passe, type_utilisateur) values ($1, $2, $3) returning compte_id",
		personneID, hashedPassword, accountType,
	).Scan(&compteID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"message":          "Account created successfully",
		"personne_id":      personneID,
		"compte_id":        compteID,
		"type_utilisateur": accountType,
	}

	switch accountType {
	case UserTypeAdherent:
		var adherentID int
		err = tx.QueryRow(ctx, "insert into adherent (personne_id) values ($1) returning adherent_id", personneID).Scan(&adherentID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec(
			ctx,
			`insert into adhesion_association (adherent_id, forfait_id, date_debut, date_fin)
			select $1, forfait_id, current_date, (current_date + interval '1 year')::date from forfait where libelle='Adhesion standard'`,
			adherentID,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		response["adherent_id"] = adherentID

	case UserTypeCommercant:
		var commercantID int
		raisonSociale := strings.TrimSpace(prenom + " " + nom + " #" + strconv.Itoa(personneID))
		err = tx.QueryRow(
			ctx,
			"insert into commercant (personne_id, raison_sociale, email, telephone, adresse, code_postal, ville, pays) values ($1, $2, $3, nullif($4,''), coalesce(nullif($5,''),'-'), coalesce(nullif($6,''),'00000'), coalesce(nullif($7,''),'-'), $8) returning commercant_id",
			personneID, raisonSociale, email, telephone, adresse, codePostal, ville, pays,
		).Scan(&commercantID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = tx.Exec(
			ctx,
			`insert into adhesion_commercant (commercant_id, forfait_id, date_debut, date_fin)
			select $1, forfait_id, current_date, (current_date + interval '1 year')::date from forfait where libelle='Adhesion commercant'`,
			commercantID,
		)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		response["commercant_id"] = commercantID
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}
