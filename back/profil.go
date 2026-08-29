package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// profilHandler covers "Consulter son profil" (GET) and "Modifier ses
// informations" (PATCH) from the use case diagram. Every authenticated
// account, whatever its type_utilisateur, can read and edit its own profil.
func profilHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodGet:
		consulterProfil(w, r, token)
	case http.MethodPatch:
		modifierProfil(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func consulterProfil(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var nom, prenom, email, pays string
	var telephone, adresse, complementAdresse, codePostal, ville *string
	var dateCreation string
	err = conn.QueryRow(
		ctx,
		`select personne.nom, personne.prenom, personne.email, personne.telephone, personne.adresse,
			personne.complement_adresse, personne.code_postal, personne.ville, personne.pays, personne.date_creation::text
		from compte
		inner join personne on personne.personne_id = compte.personne_id
		where compte.compte_id = $1`,
		token.CompteID,
	).Scan(&nom, &prenom, &email, &telephone, &adresse, &complementAdresse, &codePostal, &ville, &pays, &dateCreation)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Profil not found")
		return
	}

	var adhesionStatut *string
	var adhesionDateFin *string
	_ = conn.QueryRow(
		ctx,
		`select adhesion_association.statut, adhesion_association.date_fin::text
		from adhesion_association
		inner join adherent on adherent.adherent_id = adhesion_association.adherent_id
		inner join compte on compte.personne_id = adherent.personne_id
		where compte.compte_id = $1
		order by adhesion_association.date_fin desc nulls last
		limit 1`,
		token.CompteID,
	).Scan(&adhesionStatut, &adhesionDateFin)

	writeJSON(w, http.StatusOK, map[string]any{
		"compte_id":          token.CompteID,
		"type_utilisateur":   token.UserType,
		"nom":                nom,
		"prenom":             prenom,
		"email":              email,
		"telephone":          telephone,
		"adresse":            adresse,
		"complement_adresse": complementAdresse,
		"code_postal":        codePostal,
		"ville":              ville,
		"pays":               pays,
		"date_creation":      dateCreation,
		"adhesion_statut":    adhesionStatut,
		"adhesion_date_fin":  adhesionDateFin,
	})
}

func modifierProfil(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	updates := make([]string, 0)
	args := make([]any, 0)
	argIndex := 1

	addField := func(column, value string) {
		updates = append(updates, column+"=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	if v := r.FormValue("telephone"); v != "" {
		if !telephoneRegexp.MatchString(v) {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "telephone must be in international format, e.g. +33612345678")
			return
		}
		addField("telephone", v)
	}
	if v := r.FormValue("adresse"); v != "" {
		addField("adresse", v)
	}
	if v := r.FormValue("complement_adresse"); v != "" {
		addField("complement_adresse", v)
	}
	if v := r.FormValue("code_postal"); v != "" {
		if !codePostalRegexp.MatchString(v) {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "code_postal must be exactly 5 digits")
			return
		}
		addField("code_postal", v)
	}
	if v := r.FormValue("ville"); v != "" {
		addField("ville", v)
	}
	if v := r.FormValue("pays"); v != "" {
		addField("pays", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	query := "update personne set " + strings.Join(updates, ", ") +
		" where personne_id = (select personne_id from compte where compte_id = $" + strconv.Itoa(argIndex) + ")"
	args = append(args, token.CompteID)

	if _, err := conn.Exec(ctx, query, args...); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update profil")
		return
	}

	newPassword := r.FormValue("nouveau_mot_de_passe")
	if newPassword != "" {
		if newPassword != r.FormValue("confirmation_mot_de_passe") {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Passwords do not match.")
			return
		}
		hashed := hashPassword(newPassword)
		if hashed == "" {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to hash password")
			return
		}
		if _, err := conn.Exec(ctx, "update compte set mot_de_passe=$1 where compte_id=$2", hashed, token.CompteID); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update password")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Profil updated successfully"})
}

// renouvelerAdhesionHandler implements "Renouveler son adhesion": only an
// adherent can renew, and only their own adhesion.
func renouvelerAdhesionHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	if !isAdherent(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only adherents can renew their own adhesion")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var adherentID int
	err = conn.QueryRow(
		ctx,
		"select adherent.adherent_id from adherent inner join compte on compte.personne_id = adherent.personne_id where compte.compte_id = $1",
		token.CompteID,
	).Scan(&adherentID)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "No adherent profil found for this account")
		return
	}

	forfaitLibelle := r.FormValue("forfait")
	if forfaitLibelle == "" {
		forfaitLibelle = "Adhesion standard"
	}

	var existingDateFin *string
	_ = conn.QueryRow(
		ctx,
		"select max(date_fin)::text from adhesion_association where adherent_id=$1",
		adherentID,
	).Scan(&existingDateFin)

	var newID int
	var dateFin string
	err = conn.QueryRow(
		ctx,
		`insert into adhesion_association (adherent_id, forfait_id, date_debut, date_fin, created_by)
		select $1, forfait_id, current_date, greatest(current_date, coalesce($3::date, current_date)) + interval '1 year', $4
		from forfait where libelle = $2
		on conflict (adherent_id, forfait_id, date_debut) do update set date_fin = excluded.date_fin, created_by = excluded.created_by
		returning adhesion_association_id, date_fin::text`,
		adherentID, forfaitLibelle, existingDateFin, token.CompteID,
	).Scan(&newID, &dateFin)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Unknown forfait, or failed to renew adhesion")
		return
	}

	_, _ = conn.Exec(ctx, "update adherent set statut='actif' where adherent_id=$1", adherentID)

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":                 "Adhesion renewed successfully",
		"adhesion_association_id": newID,
		"date_fin":                dateFin,
	})
}
