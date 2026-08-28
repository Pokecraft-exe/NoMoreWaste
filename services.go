package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// servicesHandler manages the service catalog ("gestion des services :
// propositions, plannings, inscriptions" - this is the "propositions" part).
func servicesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can propose a new service")
			return
		}
		creerService(w, r)
	case http.MethodGet:
		listerServices(w, r)
	case http.MethodPatch:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can modify a service")
			return
		}
		modifierService(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerService(w http.ResponseWriter, r *http.Request) {
	libelle := r.FormValue("libelle")
	categorieService := r.FormValue("categorie_service")
	description := r.FormValue("description")
	if libelle == "" || categorieService == "" || description == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "libelle, categorie_service and description are required")
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
		"insert into service (libelle, categorie_service, description, tarif) values ($1, $2, $3, coalesce(nullif($4,'')::numeric, 0)) returning service_id",
		libelle, categorieService, description, r.FormValue("tarif"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid categorie_service, tarif, or a service with this libelle already exists")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Service created successfully", "service_id": newID})
}

func listerServices(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := "select service_id, libelle, categorie_service, description, tarif, actif from service where actif='1'"
	args := []any{}
	if categorie := r.URL.Query().Get("categorie_service"); categorie != "" {
		query += " and categorie_service = $1"
		args = append(args, categorie)
	}
	query += " order by libelle asc"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query services")
		return
	}
	defer rows.Close()

	services := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var libelle, categorieService, description, actif string
		var tarif float64
		if err := rows.Scan(&id, &libelle, &categorieService, &description, &tarif, &actif); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan services")
			return
		}
		services = append(services, map[string]any{
			"service_id":        id,
			"libelle":           libelle,
			"categorie_service": categorieService,
			"description":       description,
			"tarif":             tarif,
			"actif":             actif == "1",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(services), "services": services})
}

func modifierService(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	serviceID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid service id")
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

	if v := r.FormValue("description"); v != "" {
		addField("description", v)
	}
	if v := r.FormValue("tarif"); v != "" {
		addField("tarif", v)
	}
	if v := r.FormValue("actif"); v != "" {
		addField("actif", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, serviceID)
	tag, err := conn.Exec(ctx, "update service set "+strings.Join(updates, ", ")+" where service_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update service")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Service not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Service updated successfully"})
}

// planningServiceHandler implements "Gerer le planning" / "Ajouter une
// date": staff schedules a service slot (date, heures, capacite).
func planningServiceHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can schedule a service slot")
			return
		}
		ajouterDatePlanning(w, r, token)
	case http.MethodGet:
		listerPlanningService(w, r)
	case http.MethodPatch:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can modify the planning")
			return
		}
		modifierPlanningService(w, r)
	case http.MethodDelete:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can cancel a planning slot")
			return
		}
		annulerPlanningService(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func ajouterDatePlanning(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	serviceID := r.FormValue("service_id")
	dateService := r.FormValue("date_service")
	heureDebut := r.FormValue("heure_debut")
	heureFin := r.FormValue("heure_fin")
	if serviceID == "" || dateService == "" || heureDebut == "" || heureFin == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "service_id, date_service, heure_debut and heure_fin are required")
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
		`insert into planning_service (service_id, site_id, date_service, heure_debut, heure_fin, capacite, commentaire, created_by)
		values ($1, nullif($2,'')::int, $3, $4, $5, coalesce(nullif($6,'')::int, 0), nullif($7,''), $8)
		returning planning_service_id`,
		serviceID, r.FormValue("site_id"), dateService, heureDebut, heureFin, r.FormValue("capacite"), r.FormValue("commentaire"), token.CompteID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to schedule the service slot")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Planning slot added successfully", "planning_service_id": newID})
}

func listerPlanningService(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := `select planning_service.planning_service_id, service.libelle, planning_service.date_service,
		planning_service.heure_debut, planning_service.heure_fin, planning_service.capacite, planning_service.statut
		from planning_service inner join service on service.service_id = planning_service.service_id`
	conditions := make([]string, 0)
	args := []any{}
	argPos := 1
	if serviceID := r.URL.Query().Get("service_id"); serviceID != "" {
		conditions = append(conditions, "planning_service.service_id = $"+strconv.Itoa(argPos))
		args = append(args, serviceID)
		argPos++
	}
	if r.URL.Query().Get("upcoming") == "1" {
		conditions = append(conditions, "planning_service.date_service >= current_date")
	}
	for i, c := range conditions {
		if i == 0 {
			query += " where " + c
		} else {
			query += " and " + c
		}
	}
	query += " order by planning_service.date_service asc"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query the planning")
		return
	}
	defer rows.Close()

	planning := make([]map[string]any, 0)
	for rows.Next() {
		var id, capacite int
		var libelle, dateService, heureDebut, heureFin, statut string
		if err := rows.Scan(&id, &libelle, &dateService, &heureDebut, &heureFin, &capacite, &statut); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan the planning")
			return
		}
		planning = append(planning, map[string]any{
			"planning_service_id": id,
			"service":             libelle,
			"date_service":        dateService,
			"heure_debut":         heureDebut,
			"heure_fin":           heureFin,
			"capacite":            capacite,
			"statut":              statut,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(planning), "planning": planning})
}

func modifierPlanningService(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	planningID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid planning_service id")
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
	if v := r.FormValue("capacite"); v != "" {
		addField("capacite", v)
	}
	if v := r.FormValue("commentaire"); v != "" {
		addField("commentaire", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, planningID)
	tag, err := conn.Exec(ctx, "update planning_service set "+strings.Join(updates, ", ")+" where planning_service_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the planning slot")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Planning slot not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Planning slot updated successfully"})
}

func annulerPlanningService(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "update planning_service set statut='annule' where planning_service_id=$1", idParam)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to cancel the planning slot")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Planning slot not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Planning slot cancelled successfully"})
}

// inscriptionsServiceHandler lets an adherent register for a service slot.
func inscriptionsServiceHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		sInscrireService(w, r, token)
	case http.MethodGet:
		listerInscriptionsService(w, r, token)
	case http.MethodDelete:
		annulerInscriptionService(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func ownAdherentID(ctx context.Context, conn *pgx.Conn, compteID int) (int, error) {
	var adherentID int
	err := conn.QueryRow(
		ctx,
		"select adherent.adherent_id from adherent inner join compte on compte.personne_id = adherent.personne_id where compte.compte_id = $1",
		compteID,
	).Scan(&adherentID)
	return adherentID, err
}

func sInscrireService(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isAdherent(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only adherents can register for a service")
		return
	}
	planningServiceID := r.FormValue("planning_service_id")
	if planningServiceID == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required field: planning_service_id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	adherentID, err := ownAdherentID(ctx, conn, token.CompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No adherent profil found for this account")
		return
	}

	var newID int
	err = conn.QueryRow(
		ctx,
		"insert into inscription_service (planning_service_id, adherent_id) values ($1, $2) on conflict (planning_service_id, adherent_id) do nothing returning inscription_service_id",
		planningServiceID, adherentID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Already registered, or invalid planning_service_id")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Registered successfully", "inscription_service_id": newID})
}

func listerInscriptionsService(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	planningServiceID := r.URL.Query().Get("planning_service_id")
	if planningServiceID == "" || !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can list a slot's registrations")
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
		`select inscription_service.inscription_service_id, personne.nom, personne.prenom, inscription_service.statut
		from inscription_service
		inner join adherent on adherent.adherent_id = inscription_service.adherent_id
		inner join personne on personne.personne_id = adherent.personne_id
		where inscription_service.planning_service_id = $1`,
		planningServiceID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query registrations")
		return
	}
	defer rows.Close()

	inscriptions := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var nom, prenom, statut string
		if err := rows.Scan(&id, &nom, &prenom, &statut); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan registrations")
			return
		}
		inscriptions = append(inscriptions, map[string]any{"inscription_service_id": id, "nom": nom, "prenom": prenom, "statut": statut})
	}

	writeJSON(w, http.StatusOK, map[string]any{"planning_service_id": planningServiceID, "total": len(inscriptions), "inscriptions": inscriptions})
}

func annulerInscriptionService(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isAdherent(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only adherents can cancel their own registration")
		return
	}
	planningServiceID := r.URL.Query().Get("planning_service_id")
	if planningServiceID == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'planning_service_id' parameter")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	adherentID, err := ownAdherentID(ctx, conn, token.CompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No adherent profil found for this account")
		return
	}

	_, err = conn.Exec(ctx, "update inscription_service set statut='annule' where planning_service_id=$1 and adherent_id=$2", planningServiceID, adherentID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to cancel the registration")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
