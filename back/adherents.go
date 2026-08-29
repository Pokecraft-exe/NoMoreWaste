package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// adherentsHandler implements "rechercher un adherant" (GET) and "modifier
// l'adherant" (PATCH). Both are back-office, staff-only actions.
func adherentsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can manage adherents")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rechercherAdherent(w, r)
	case http.MethodPatch:
		modifierAdherent(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func rechercherAdherent(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if idParam := r.URL.Query().Get("id"); idParam != "" {
		adherentID, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}

		var nom, prenom, email, statut, dateInscription string
		err = conn.QueryRow(
			ctx,
			`select personne.nom, personne.prenom, personne.email, adherent.statut, adherent.date_inscription
			from adherent inner join personne on personne.personne_id = adherent.personne_id
			where adherent.adherent_id = $1`,
			adherentID,
		).Scan(&nom, &prenom, &email, &statut, &dateInscription)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Adherent not found")
			return
		}

		response := map[string]any{
			"adherent_id":      adherentID,
			"nom":              nom,
			"prenom":           prenom,
			"email":            email,
			"statut":           statut,
			"date_inscription": dateInscription,
		}

		var adhesionID int
		var dateFin *string
		if conn.QueryRow(
			ctx,
			"select adhesion_association_id, date_fin::text from adhesion_association where adherent_id=$1 order by date_fin desc nulls last limit 1",
			adherentID,
		).Scan(&adhesionID, &dateFin) == nil {
			response["adhesion_association_id"] = adhesionID
			response["adhesion_date_fin"] = dateFin
		}

		writeJSON(w, http.StatusOK, response)
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// expirant_sous=<jours> restricts the list to adherents whose current
	// (most recent) adhesion ends within that many days from today - used by
	// the back-office dashboard's "Adhesions expirant sous 30 jours" widget.
	expirantSous := r.URL.Query().Get("expirant_sous")

	query := `select adherent.adherent_id, personne.nom, personne.prenom, personne.email, adherent.statut, adherent.date_inscription`
	from_ := " from adherent inner join personne on personne.personne_id = adherent.personne_id"
	if expirantSous != "" {
		query += ", a.date_fin::text"
		from_ += ` inner join lateral (
			select max(date_fin) as date_fin from adhesion_association where adhesion_association.adherent_id = adherent.adherent_id
		) a on true`
	}
	query += from_

	conditions := make([]string, 0)
	args := []any{}
	argPos := 1
	if q := r.URL.Query().Get("q"); q != "" {
		conditions = append(conditions, "(personne.nom ilike $"+strconv.Itoa(argPos)+" or personne.prenom ilike $"+strconv.Itoa(argPos)+" or personne.email ilike $"+strconv.Itoa(argPos)+")")
		args = append(args, "%"+q+"%")
		argPos++
	}
	if expirantSous != "" {
		jours, convErr := strconv.Atoi(expirantSous)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "expirant_sous must be a valid integer")
			return
		}
		conditions = append(conditions, "a.date_fin is not null and a.date_fin >= current_date and a.date_fin <= current_date + $"+strconv.Itoa(argPos)+" * interval '1 day'")
		args = append(args, jours)
		argPos++
	}
	for i, c := range conditions {
		if i == 0 {
			query += " where " + c
		} else {
			query += " and " + c
		}
	}
	if expirantSous != "" {
		query += " order by a.date_fin asc"
	} else {
		query += " order by adherent.adherent_id"
	}
	query += " offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query adherents")
		return
	}
	defer rows.Close()

	adherents := make([]map[string]any, 0)
	for rows.Next() {
		var adherentID int
		var nom, prenom, email, statut, dateInscription string
		var dateFin *string
		var scanErr error
		if expirantSous != "" {
			scanErr = rows.Scan(&adherentID, &nom, &prenom, &email, &statut, &dateInscription, &dateFin)
		} else {
			scanErr = rows.Scan(&adherentID, &nom, &prenom, &email, &statut, &dateInscription)
		}
		if scanErr != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan adherents")
			return
		}
		entry := map[string]any{
			"adherent_id":      adherentID,
			"nom":              nom,
			"prenom":           prenom,
			"email":            email,
			"statut":           statut,
			"date_inscription": dateInscription,
		}
		if dateFin != nil {
			entry["adhesion_date_fin"] = *dateFin
		}
		adherents = append(adherents, entry)
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(adherents), "adherents": adherents})
}

func modifierAdherent(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	adherentID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid adherent id")
		return
	}

	statut := r.FormValue("statut")
	if statut == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "statut is required (actif, suspendu, radie or en_attente)")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "update adherent set statut=$1 where adherent_id=$2", statut, adherentID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid statut or failed to update adherent")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Adherent not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Adherent updated successfully"})
	_ = token.CompteID
}

// rappelAdhesionHandler implements "Envoyer un rappel de renouvellement":
// staff schedules a reminder for an adhesion that is about to expire.
func rappelAdhesionHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can send renewal reminders")
		return
	}

	switch r.Method {
	case http.MethodPut:
		envoyerRappelAdhesion(w, r)
	case http.MethodGet:
		listerRappelsAdhesion(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func envoyerRappelAdhesion(w http.ResponseWriter, r *http.Request) {
	adhesionIDParam := r.FormValue("adhesion_association_id")
	canal := r.FormValue("canal")
	message := r.FormValue("message")
	if adhesionIDParam == "" || canal == "" || message == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "adhesion_association_id, canal and message are required")
		return
	}
	adhesionID, err := strconv.Atoi(adhesionIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "adhesion_association_id must be a valid integer")
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
		"insert into rappel_adhesion (adhesion_association_id, canal, message, statut) values ($1, $2, $3, 'a_envoyer') returning rappel_adhesion_id",
		adhesionID, canal, message,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid adhesion_association_id or canal")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Reminder scheduled successfully", "rappel_adhesion_id": newID})
}

func listerRappelsAdhesion(w http.ResponseWriter, r *http.Request) {
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

	rows, err := conn.Query(
		ctx,
		`select rappel_adhesion.rappel_adhesion_id, rappel_adhesion.adhesion_association_id, rappel_adhesion.date_rappel,
			rappel_adhesion.canal, rappel_adhesion.statut, rappel_adhesion.message
		from rappel_adhesion order by rappel_adhesion.date_rappel desc offset $1 limit $2`,
		from, size,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query reminders")
		return
	}
	defer rows.Close()

	rappels := make([]map[string]any, 0)
	for rows.Next() {
		var rappelID, adhesionID int
		var dateRappel, canal, statut, message string
		if err := rows.Scan(&rappelID, &adhesionID, &dateRappel, &canal, &statut, &message); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan reminders")
			return
		}
		rappels = append(rappels, map[string]any{
			"rappel_adhesion_id":      rappelID,
			"adhesion_association_id": adhesionID,
			"date_rappel":             dateRappel,
			"canal":                   canal,
			"statut":                  statut,
			"message":                 message,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(rappels), "rappels": rappels})
}
