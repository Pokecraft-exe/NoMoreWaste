package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// tourneesHandler covers both "rechercher/modifier/creer une distribution"
// and the equivalent collecte tournees: a tournee is a run of stops, and
// type_tournee (collecte | distribution | mixte) tells them apart.
func tourneesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerTournee(w, r, token)
	case http.MethodGet:
		rechercherTournee(w, r)
	case http.MethodPatch:
		modifierTournee(w, r, token)
	case http.MethodDelete:
		supprimerTournee(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerTournee(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can create a tournee")
		return
	}

	dateTournee := r.FormValue("date_tournee")
	typeTournee := r.FormValue("type_tournee")
	if dateTournee == "" || typeTournee == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "date_tournee and type_tournee are required")
		return
	}

	var siteID *int
	if v := r.FormValue("site_id"); v != "" {
		id, convErr := strconv.Atoi(v)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "site_id must be a valid integer")
			return
		}
		siteID = &id
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
		`insert into tournee (site_id, date_tournee, type_tournee, vehicule, commentaire, created_by)
		values ($1, $2, $3, nullif($4, ''), nullif($5, ''), $6) returning tournee_id`,
		siteID, dateTournee, typeTournee, r.FormValue("vehicule"), r.FormValue("commentaire"), token.CompteID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid type_tournee or failed to create tournee")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Tournee created successfully", "tournee_id": newID})
}

func rechercherTournee(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if idParam := r.URL.Query().Get("id"); idParam != "" {
		tourneeID, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}

		var dateTournee, typeTournee, statut string
		var vehicule, commentaire *string
		err = conn.QueryRow(
			ctx,
			"select date_tournee, type_tournee, statut, vehicule, commentaire from tournee where tournee_id=$1",
			tourneeID,
		).Scan(&dateTournee, &typeTournee, &statut, &vehicule, &commentaire)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Tournee not found")
			return
		}

		etapeRows, err := conn.Query(
			ctx,
			`select tournee_etape_id, ordre, collecte_id, partenaire_id, adresse, ville, heure_prevue, heure_reelle
			from tournee_etape where tournee_id=$1 order by ordre asc`,
			tourneeID,
		)
		if err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the tournee steps")
			return
		}
		defer etapeRows.Close()

		etapes := make([]map[string]any, 0)
		for etapeRows.Next() {
			var etapeID, ordre int
			var collecteID, partenaireID *int
			var adresse, ville, heurePrevue, heureReelle *string
			if err := etapeRows.Scan(&etapeID, &ordre, &collecteID, &partenaireID, &adresse, &ville, &heurePrevue, &heureReelle); err != nil {
				writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the tournee steps")
				return
			}
			etapes = append(etapes, map[string]any{
				"tournee_etape_id": etapeID,
				"ordre":            ordre,
				"collecte_id":      collecteID,
				"partenaire_id":    partenaireID,
				"adresse":          adresse,
				"ville":            ville,
				"heure_prevue":     heurePrevue,
				"heure_reelle":     heureReelle,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"tournee_id":   tourneeID,
			"date_tournee": dateTournee,
			"type_tournee": typeTournee,
			"statut":       statut,
			"vehicule":     vehicule,
			"commentaire":  commentaire,
			"etapes":       etapes,
		})
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	query := "select tournee_id, date_tournee, type_tournee, statut, vehicule, commentaire from tournee"
	conditions := make([]string, 0)
	args := []any{}
	argPos := 1

	if typeTournee := r.URL.Query().Get("type_tournee"); typeTournee != "" {
		conditions = append(conditions, "type_tournee = $"+strconv.Itoa(argPos))
		args = append(args, typeTournee)
		argPos++
	}
	for i, c := range conditions {
		if i == 0 {
			query += " where " + c
		} else {
			query += " and " + c
		}
	}
	query += " order by date_tournee asc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query tournees")
		return
	}
	defer rows.Close()

	tournees := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var dateTournee, typeTournee, statut string
		var vehicule, commentaire *string
		if err := rows.Scan(&id, &dateTournee, &typeTournee, &statut, &vehicule, &commentaire); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan tournees")
			return
		}
		tournees = append(tournees, map[string]any{
			"tournee_id":   id,
			"date_tournee": dateTournee,
			"type_tournee": typeTournee,
			"statut":       statut,
			"vehicule":     vehicule,
			"commentaire":  commentaire,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(tournees), "tournees": tournees})
}

func modifierTournee(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can modify a tournee")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	tourneeID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid tournee id")
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

	if v := r.FormValue("date_tournee"); v != "" {
		addField("date_tournee", v)
	}
	if v := r.FormValue("statut"); v != "" {
		addField("statut", v)
	}
	if v := r.FormValue("vehicule"); v != "" {
		addField("vehicule", v)
	}
	if v := r.FormValue("commentaire"); v != "" {
		addField("commentaire", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, tourneeID)
	tag, err := conn.Exec(ctx, "update tournee set "+strings.Join(updates, ", ")+" where tournee_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update tournee")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Tournee not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Tournee updated successfully"})
}

func supprimerTournee(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can cancel a tournee")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	tourneeID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid tournee id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "update tournee set statut='annulee' where tournee_id=$1", tourneeID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to cancel tournee")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Tournee not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Tournee cancelled successfully"})
}

// tourneeEtapesHandler adds a stop (arret) to a tournee - a collecte pickup
// or a distribution drop-off at a partenaire.
func tourneeEtapesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can manage tournee steps")
		return
	}

	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	tourneeIDParam := r.FormValue("tournee_id")
	ordreParam := r.FormValue("ordre")
	if tourneeIDParam == "" || ordreParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "tournee_id and ordre are required")
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
		`insert into tournee_etape (tournee_id, ordre, collecte_id, partenaire_id, adresse, ville, heure_prevue, commentaire)
		values ($1, $2, nullif($3, '')::int, nullif($4, '')::int, nullif($5, ''), nullif($6, ''), nullif($7, '')::time, nullif($8, ''))
		returning tournee_etape_id`,
		tourneeIDParam, ordreParam, r.FormValue("collecte_id"), r.FormValue("partenaire_id"), r.FormValue("adresse"), r.FormValue("ville"), r.FormValue("heure_prevue"), r.FormValue("commentaire"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to add tournee step (check the ordre is unique for this tournee)")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Step added successfully", "tournee_etape_id": newID})
}
