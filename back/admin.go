package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// adminComptesHandler is administrateur-only. It is the sole way to grant
// the "responsable" or "administrateur" type_utilisateur - those are never
// reachable through /oauth/v3/inscription (public registration only offers
// visiteur, adherent, commercant) nor through the benevole candidature flow.
func adminComptesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	// Creating an account or changing its type_utilisateur (which can grant
	// responsable/administrateur) stays administrateur-only. Looking accounts
	// up is a routine back-office task (e.g. resolving an email to a
	// compte_id when creating a collecte or a candidature) so any staff
	// member - responsable included - can do it.
	if r.Method == http.MethodGet {
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can look up accounts")
			return
		}
	} else if !isAdministrateur(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only administrateurs can manage accounts")
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerCompteAdmin(w, r)
	case http.MethodGet:
		listerComptes(w, r)
	case http.MethodPatch:
		changerTypeUtilisateur(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerCompteAdmin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Malformed request body.")
		return
	}

	email := strings.TrimSpace(r.PostForm.Get("email"))
	secret := r.PostForm.Get("mot_de_passe")
	nom := r.PostForm.Get("nom")
	prenom := r.PostForm.Get("prenom")
	telephone := r.PostForm.Get("telephone")
	adresse := r.PostForm.Get("adresse")
	codePostal := r.PostForm.Get("code_postal")
	ville := r.PostForm.Get("ville")
	userType := strings.ToLower(strings.TrimSpace(r.PostForm.Get("type_utilisateur")))
	if userType == "" {
		userType = UserTypeVisiteur
	}

	if email == "" || secret == "" || nom == "" || prenom == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required fields: email, mot_de_passe, nom, prenom")
		return
	}
	if !isEmailIdentifier(email) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "email must be a valid email address")
		return
	}
	if !isValidUserType(userType) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Unknown type_utilisateur: "+userType)
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_ADMIN_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	hashedPassword := hashPassword(secret)
	if hashedPassword == "" {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to hash password")
		return
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer tx.Rollback(ctx)

	var personneID int
	err = tx.QueryRow(
		ctx,
		"insert into personne (nom, prenom, email, telephone, adresse, code_postal, ville) values ($1, $2, $3, nullif($4,''), nullif($5,''), nullif($6,''), nullif($7,'')) returning personne_id",
		nom, prenom, email, telephone, adresse, codePostal, ville,
	).Scan(&personneID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "A personne with this email already exists.")
		return
	}

	var compteID int
	if err := tx.QueryRow(ctx, "insert into compte (personne_id, mot_de_passe, type_utilisateur) values ($1, $2, $3) returning compte_id", personneID, hashedPassword, userType).Scan(&compteID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the account")
		return
	}

	if userType == UserTypeAdherent {
		var adherentID int
		if err := tx.QueryRow(ctx, "insert into adherent (personne_id) values ($1) returning adherent_id", personneID).Scan(&adherentID); err == nil {
			_, _ = tx.Exec(
				ctx,
				`insert into adhesion_association (adherent_id, forfait_id, date_debut, date_fin)
				select $1, forfait_id, current_date, (current_date + interval '1 year')::date from forfait where libelle='Adhesion standard'`,
				adherentID,
			)
		}
	} else if userType == UserTypeCommercant {
		raisonSociale := strings.TrimSpace(prenom + " " + nom + " #" + strconv.Itoa(personneID))
		_, _ = tx.Exec(
			ctx,
			"insert into commercant (personne_id, raison_sociale, email, telephone, adresse, code_postal, ville) values ($1, $2, $3, nullif($4,''), coalesce(nullif($5,''),'-'), coalesce(nullif($6,''),'00000'), coalesce(nullif($7,''),'-'))",
			personneID, raisonSociale, email, telephone, adresse, codePostal, ville,
		)
	}

	if err := tx.Commit(ctx); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the account")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Account created successfully", "compte_id": compteID})
}

func listerComptes(w http.ResponseWriter, r *http.Request) {
	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_ADMIN_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := `select compte.compte_id, personne.nom, personne.prenom, personne.email, compte.type_utilisateur, compte.actif
		from compte inner join personne on personne.personne_id = compte.personne_id`
	args := []any{}
	argPos := 1
	conditions := []string{}
	if q := r.URL.Query().Get("q"); q != "" {
		conditions = append(conditions, "(personne.email ilike $"+strconv.Itoa(argPos)+" or personne.nom ilike $"+strconv.Itoa(argPos)+")")
		args = append(args, "%"+q+"%")
		argPos++
	}
	if userType := r.URL.Query().Get("type_utilisateur"); userType != "" {
		conditions = append(conditions, "compte.type_utilisateur = $"+strconv.Itoa(argPos))
		args = append(args, userType)
		argPos++
	}
	if len(conditions) > 0 {
		query += " where " + strings.Join(conditions, " and ")
	}
	query += " order by compte.compte_id offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query comptes")
		return
	}
	defer rows.Close()

	comptes := make([]map[string]any, 0)
	for rows.Next() {
		var compteID int
		var nom, prenom, email, userType, actif string
		if err := rows.Scan(&compteID, &nom, &prenom, &email, &userType, &actif); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan comptes")
			return
		}
		comptes = append(comptes, map[string]any{
			"compte_id":        compteID,
			"nom":              nom,
			"prenom":           prenom,
			"email":            email,
			"type_utilisateur": userType,
			"actif":            actif == "1",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(comptes), "comptes": comptes})
}

func changerTypeUtilisateur(w http.ResponseWriter, r *http.Request) {
	compteIDParam := r.FormValue("compte_id")
	userType := r.FormValue("type_utilisateur")
	if compteIDParam == "" || userType == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "compte_id and type_utilisateur are required")
		return
	}
	if !isValidUserType(userType) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Unknown type_utilisateur: "+userType)
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_ADMIN_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	actif := "1"
	if v := r.FormValue("actif"); v != "" {
		actif = v
	}

	tag, err := conn.Exec(ctx, "update compte set type_utilisateur = $1, actif = $2 where compte_id = $3", userType, actif, compteIDParam)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the account")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Account not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Account type updated successfully"})
}
