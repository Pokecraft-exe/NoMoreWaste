package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// candidaturesBenevoleHandler implements "gerer le suivi des benevoles,
// depuis leur candidature jusqu'a leur affectation a un service donne" -
// this is the candidature step. Any authenticated account can apply; only
// staff can review and decide.
func candidaturesBenevoleHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		deposerCandidatureBenevole(w, r, token)
	case http.MethodGet:
		listerCandidaturesBenevole(w, r, token)
	case http.MethodPatch:
		deciderCandidatureBenevole(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func deposerCandidatureBenevole(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	// "personne_id" is a historical name kept for compatibility: staff
	// creating a candidature on someone's behalf (back-office) sends that
	// person's compte_id there; a benevole applying for themselves omits it
	// and their own token is used instead.
	compteID := token.CompteID
	if v := r.FormValue("personne_id"); v != "" && isStaff(token) {
		if parsed, convErr := strconv.Atoi(v); convErr == nil {
			compteID = parsed
		}
	}

	var personneID int
	err = conn.QueryRow(ctx, "select personne_id from compte where compte_id = $1", compteID).Scan(&personneID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to resolve the candidate's personne")
		return
	}

	var newID int
	err = conn.QueryRow(
		ctx,
		"insert into candidature_benevole (personne_id, motivation, disponibilite, commentaire) values ($1, nullif($2,''), nullif($3,''), nullif($4,'')) returning candidature_benevole_id",
		personneID, r.FormValue("motivation"), r.FormValue("disponibilite"), r.FormValue("commentaire"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to record the candidature")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Candidature received successfully", "candidature_benevole_id": newID})
}

func listerCandidaturesBenevole(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can review candidatures")
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := `select candidature_benevole.candidature_benevole_id, personne.nom, personne.prenom, personne.email,
		candidature_benevole.statut, candidature_benevole.date_candidature::text, candidature_benevole.motivation,
		candidature_benevole.disponibilite, candidature_benevole.commentaire
		from candidature_benevole inner join personne on personne.personne_id = candidature_benevole.personne_id`
	args := []any{}
	argPos := 1
	if statut := r.URL.Query().Get("statut"); statut != "" {
		query += " where candidature_benevole.statut = $" + strconv.Itoa(argPos)
		args = append(args, statut)
		argPos++
	}
	query += " order by candidature_benevole.date_candidature desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query candidatures")
		return
	}
	defer rows.Close()

	candidatures := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var nom, prenom, email, statut, dateCandidature string
		var motivation, disponibilite, commentaire *string
		if err := rows.Scan(&id, &nom, &prenom, &email, &statut, &dateCandidature, &motivation, &disponibilite, &commentaire); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan candidatures")
			return
		}
		candidatures = append(candidatures, map[string]any{
			"candidature_benevole_id": id,
			"nom":                     nom,
			"prenom":                  prenom,
			"email":                   email,
			"statut":                  statut,
			"date_candidature":        dateCandidature,
			"motivation":              motivation,
			"disponibilite":           disponibilite,
			"commentaire":             commentaire,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(candidatures), "candidatures": candidatures})
}

// deciderCandidatureBenevole validates or refuses a candidature. Validating
// it is the only way a compte's type_utilisateur ever becomes "benevole" -
// it also creates the benevole record.
func deciderCandidatureBenevole(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can decide on a candidature")
		return
	}

	idParam := r.URL.Query().Get("id")
	statut := r.FormValue("statut")
	if idParam == "" || statut == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id and statut are required (en_etude, validee, refusee or archivee)")
		return
	}
	candidatureID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid candidature id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(
		ctx,
		"update candidature_benevole set statut=$1, commentaire=coalesce(nullif($2,''), commentaire), traite_par=$3, date_decision=now() where candidature_benevole_id=$4",
		statut, r.FormValue("commentaire"), token.CompteID, candidatureID,
	)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid statut or failed to update candidature")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Candidature not found")
		return
	}

	if statut != "validee" {
		writeJSON(w, http.StatusOK, map[string]any{"message": "Candidature updated successfully"})
		return
	}

	var personneID int
	var candidatureDisponibilite *string
	if err := conn.QueryRow(ctx, "select personne_id, disponibilite from candidature_benevole where candidature_benevole_id=$1", candidatureID).Scan(&personneID, &candidatureDisponibilite); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to resolve the candidate's personne")
		return
	}

	var benevoleID int
	err = conn.QueryRow(
		ctx,
		`insert into benevole (personne_id, statut, disponibilite, commentaire) values ($1, 'actif', $2, nullif($3,''))
		on conflict (personne_id) do update set statut='actif', disponibilite=coalesce(excluded.disponibilite, benevole.disponibilite)
		returning benevole_id`,
		personneID, candidatureDisponibilite, r.FormValue("commentaire"),
	).Scan(&benevoleID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the benevole record")
		return
	}

	// The compte's type_utilisateur lives on the auth database (owned by the
	// oauth service); it is only ever set to "benevole" here, once staff has
	// validated the candidature.
	authConn, err := pgx.Connect(ctx, DATABASE_AUTH_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the auth database")
		return
	}
	defer authConn.Close(ctx)

	_, err = authConn.Exec(ctx, "update compte set type_utilisateur='benevole' where personne_id=$1", personneID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to promote the account to benevole")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Candidature validated, benevole created", "benevole_id": benevoleID})
}

// benevolesHandler lets staff search and edit confirmed benevoles.
func benevolesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can manage benevoles")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rechercherBenevole(w, r)
	case http.MethodPatch:
		modifierBenevole(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func rechercherBenevole(w http.ResponseWriter, r *http.Request) {
	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := `select benevole.benevole_id, personne.nom, personne.prenom, personne.email, benevole.statut,
		benevole.disponibilite, benevole.date_inscription
		from benevole inner join personne on personne.personne_id = benevole.personne_id`
	args := []any{}
	argPos := 1
	if q := r.URL.Query().Get("q"); q != "" {
		query += " where personne.nom ilike $" + strconv.Itoa(argPos) + " or personne.prenom ilike $" + strconv.Itoa(argPos)
		args = append(args, "%"+q+"%")
		argPos++
	}
	query += " order by benevole.benevole_id offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query benevoles")
		return
	}
	defer rows.Close()

	benevoles := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var nom, prenom, email, statut, dateInscription string
		var disponibilite *string
		if err := rows.Scan(&id, &nom, &prenom, &email, &statut, &disponibilite, &dateInscription); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan benevoles")
			return
		}
		benevoles = append(benevoles, map[string]any{
			"benevole_id":      id,
			"nom":              nom,
			"prenom":           prenom,
			"email":            email,
			"statut":           statut,
			"disponibilite":    disponibilite,
			"date_inscription": dateInscription,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(benevoles), "benevoles": benevoles})
}

func modifierBenevole(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	benevoleID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid benevole id")
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

	if v := r.FormValue("statut"); v != "" {
		addField("statut", v)
	}
	if v := r.FormValue("disponibilite"); v != "" {
		addField("disponibilite", v)
	}
	if v := r.FormValue("commentaire"); v != "" {
		addField("commentaire", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, benevoleID)
	tag, err := conn.Exec(ctx, "update benevole set "+strings.Join(updates, ", ")+" where benevole_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update benevole")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Benevole not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Benevole updated successfully"})
}

// benevoleCompetencesHandler records the skills of a benevole (chauffeur,
// cuisinier, plombier, ...).
func benevoleCompetencesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodGet:
		listerCompetences(w, r)
	case http.MethodPut:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can assign competences")
			return
		}
		assignerCompetence(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func listerCompetences(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if benevoleIDParam := r.URL.Query().Get("benevole_id"); benevoleIDParam != "" {
		rows, err := conn.Query(
			ctx,
			`select competence.competence_id, competence.libelle, benevole_competence.niveau
			from benevole_competence
			inner join competence on competence.competence_id = benevole_competence.competence_id
			where benevole_competence.benevole_id = $1`,
			benevoleIDParam,
		)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query competences")
			return
		}
		defer rows.Close()

		competences := make([]map[string]any, 0)
		for rows.Next() {
			var id, niveau int
			var libelle string
			if err := rows.Scan(&id, &libelle, &niveau); err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan competences")
				return
			}
			competences = append(competences, map[string]any{"competence_id": id, "libelle": libelle, "niveau": niveau})
		}
		writeJSON(w, http.StatusOK, map[string]any{"benevole_id": benevoleIDParam, "competences": competences})
		return
	}

	rows, err := conn.Query(ctx, "select competence_id, libelle, description from competence order by libelle asc")
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the competence catalog")
		return
	}
	defer rows.Close()

	catalog := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var libelle string
		var description *string
		if err := rows.Scan(&id, &libelle, &description); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the competence catalog")
			return
		}
		catalog = append(catalog, map[string]any{"competence_id": id, "libelle": libelle, "description": description})
	}
	writeJSON(w, http.StatusOK, map[string]any{"total": len(catalog), "competences": catalog})
}

func assignerCompetence(w http.ResponseWriter, r *http.Request) {
	benevoleID := r.FormValue("benevole_id")
	competenceID := r.FormValue("competence_id")
	if benevoleID == "" || competenceID == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "benevole_id and competence_id are required")
		return
	}
	niveau := r.FormValue("niveau")
	if niveau == "" {
		niveau = "1"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(
		ctx,
		`insert into benevole_competence (benevole_id, competence_id, niveau) values ($1, $2, $3)
		on conflict (benevole_id, competence_id) do update set niveau = excluded.niveau`,
		benevoleID, competenceID, niveau,
	)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid benevole_id/competence_id/niveau")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// affectationsBenevoleHandler implements "Ajouter un benevole" / "affecter
// le benevole" on the planning: assigning a benevole to a tournee or to a
// service planning slot.
func affectationsBenevoleHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can assign a benevole")
			return
		}
		affecterBenevole(w, r)
	case http.MethodGet:
		listerAffectationsBenevole(w, r, token)
	case http.MethodPatch:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can update an affectation")
			return
		}
		modifierAffectationBenevole(w, r)
	case http.MethodDelete:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can remove an affectation")
			return
		}
		supprimerAffectationBenevole(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func affecterBenevole(w http.ResponseWriter, r *http.Request) {
	benevoleID := r.FormValue("benevole_id")
	roleMission := r.FormValue("role_mission")
	tourneeID := r.FormValue("tournee_id")
	planningServiceID := r.FormValue("planning_service_id")
	if benevoleID == "" || roleMission == "" || (tourneeID == "" && planningServiceID == "") {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "benevole_id, role_mission and either tournee_id or planning_service_id are required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var newID int
	err = conn.QueryRow(
		ctx,
		`insert into affectation_benevole (benevole_id, tournee_id, planning_service_id, role_mission, commentaire)
		values ($1, nullif($2,'')::int, nullif($3,'')::int, $4, nullif($5,''))
		returning affectation_benevole_id`,
		benevoleID, tourneeID, planningServiceID, roleMission, r.FormValue("commentaire"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to create the affectation")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Benevole assigned successfully", "affectation_benevole_id": newID})
}

func listerAffectationsBenevole(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var benevoleID string
	if isStaff(token) {
		benevoleID = r.URL.Query().Get("benevole_id")
		if benevoleID == "" {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'benevole_id' parameter")
			return
		}
	} else if isBenevole(token) {
		ownID, ownErr := ownBenevoleID(ctx, conn, token.CompteID)
		if ownErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "No benevole profil found for this account")
			return
		}
		benevoleID = strconv.Itoa(ownID)
	} else {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff or the benevole themselves can list affectations")
		return
	}

	rows, err := conn.Query(
		ctx,
		`select affectation_benevole_id, tournee_id, planning_service_id, role_mission, statut, date_affectation
		from affectation_benevole where benevole_id = $1 order by date_affectation desc`,
		benevoleID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query affectations")
		return
	}
	defer rows.Close()

	affectations := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var tourneeID, planningServiceID *int
		var roleMission, statut, dateAffectation string
		if err := rows.Scan(&id, &tourneeID, &planningServiceID, &roleMission, &statut, &dateAffectation); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan affectations")
			return
		}
		affectations = append(affectations, map[string]any{
			"affectation_benevole_id": id,
			"tournee_id":              tourneeID,
			"planning_service_id":     planningServiceID,
			"role_mission":            roleMission,
			"statut":                  statut,
			"date_affectation":        dateAffectation,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"benevole_id": benevoleID, "total": len(affectations), "affectations": affectations})
}

func modifierAffectationBenevole(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	statut := r.FormValue("statut")
	if idParam == "" || statut == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id and statut are required")
		return
	}
	affectationID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid affectation id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "update affectation_benevole set statut=$1 where affectation_benevole_id=$2", statut, affectationID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid statut or failed to update affectation")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Affectation not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Affectation updated successfully"})
}

func supprimerAffectationBenevole(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	affectationID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid affectation id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "delete from affectation_benevole where affectation_benevole_id=$1", affectationID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete affectation")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Affectation not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Affectation removed successfully"})
}
