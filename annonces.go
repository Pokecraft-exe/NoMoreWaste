package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// annoncesHandler implements "Accéder aux annonces PAP" (GET, filterable by
// catégorie: covoiturage, reparation, gardiennage, echange) and "Créer une
// annonce PAP" (PUT) - the peer-to-peer service exchange from the use case
// diagram (partage de vehicules, reparation, gardiennage, echange de
// services entre particuliers).
func annoncesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerAnnonce(w, r, token)
	case http.MethodGet:
		getAnnonce(w, r)
	case http.MethodPatch:
		modifierAnnonce(w, r, token)
	case http.MethodDelete:
		supprimerAnnonce(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	categorie := r.FormValue("categorie")
	titre := r.FormValue("titre")
	description := r.FormValue("description")
	if categorie == "" || titre == "" || description == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "categorie, titre and description are required")
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
		"insert into annonce_echange (compte_id, categorie, titre, description, prix) values ($1, $2, $3, $4, nullif($5,'')::numeric) returning annonce_echange_id",
		token.CompteID, categorie, titre, description, r.FormValue("prix"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to create the annonce, unknown categorie")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Annonce created successfully", "annonce_echange_id": newID})
}

func getAnnonce(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if idParam := r.URL.Query().Get("id"); idParam != "" {
		id, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}

		var auteurID int
		var nom, prenom, categorie, titre, description, statut, datePublication string
		var prix *float64
		err = conn.QueryRow(
			ctx,
			`select annonce_echange.compte_id, personne.nom, personne.prenom, categorie, titre, description, prix, statut, date_publication::text
			from annonce_echange
			inner join compte on compte.compte_id = annonce_echange.compte_id
			inner join personne on personne.personne_id = compte.personne_id
			where annonce_echange_id=$1`,
			id,
		).Scan(&auteurID, &nom, &prenom, &categorie, &titre, &description, &prix, &statut, &datePublication)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Annonce not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"annonce_echange_id": id,
			"auteur_id":          auteurID,
			"auteur":             prenom + " " + nom,
			"categorie":          categorie,
			"titre":              titre,
			"description":        description,
			"prix":               prix,
			"statut":             statut,
			"date_publication":   datePublication,
		})
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	query := "select annonce_echange_id, categorie, titre, description, prix, statut, date_publication::text from annonce_echange"
	conditions := make([]string, 0)
	args := []any{}
	argPos := 1
	if categorie := r.URL.Query().Get("categorie"); categorie != "" {
		conditions = append(conditions, "categorie = $"+strconv.Itoa(argPos))
		args = append(args, categorie)
		argPos++
	}
	conditions = append(conditions, "statut != 'cloturee'")
	for i, c := range conditions {
		if i == 0 {
			query += " where " + c
		} else {
			query += " and " + c
		}
	}
	query += " order by date_publication desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query annonces")
		return
	}
	defer rows.Close()

	annonces := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var categorie, titre, description, statut, datePublication string
		var prix *float64
		if err := rows.Scan(&id, &categorie, &titre, &description, &prix, &statut, &datePublication); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan annonces")
			return
		}
		annonces = append(annonces, map[string]any{
			"annonce_echange_id": id,
			"categorie":          categorie,
			"titre":              titre,
			"description":        description,
			"prix":               prix,
			"statut":             statut,
			"date_publication":   datePublication,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(annonces), "annonces": annonces})
}

func modifierAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	annonceID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid annonce id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var auteurID int
	if err := conn.QueryRow(ctx, "select compte_id from annonce_echange where annonce_echange_id=$1", annonceID).Scan(&auteurID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Annonce not found")
		return
	}
	if auteurID != token.CompteID && !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You don't have the right to modify this annonce")
		return
	}

	statut := r.FormValue("statut")
	if statut == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "statut is required (ouverte, en_cours or cloturee)")
		return
	}

	if _, err := conn.Exec(ctx, "update annonce_echange set statut=$1 where annonce_echange_id=$2", statut, annonceID); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid statut")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Annonce updated successfully"})
}

func supprimerAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	annonceID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid annonce id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var auteurID int
	if err := conn.QueryRow(ctx, "select compte_id from annonce_echange where annonce_echange_id=$1", annonceID).Scan(&auteurID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Annonce not found")
		return
	}
	if auteurID != token.CompteID && !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You don't have the right to delete this annonce")
		return
	}

	if _, err := conn.Exec(ctx, "delete from annonce_echange where annonce_echange_id=$1", annonceID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the annonce")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// messagesAnnonceHandler implements "Répondre à une annonce PAP".
func messagesAnnonceHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		repondreAnnonce(w, r, token)
	case http.MethodGet:
		listerMessagesAnnonce(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func repondreAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	annonceID := r.FormValue("annonce_echange_id")
	message := r.FormValue("message")
	if annonceID == "" || message == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "annonce_echange_id and message are required")
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
		"insert into message_annonce_echange (annonce_echange_id, expediteur_id, message, prix_propose) values ($1, $2, $3, nullif($4,'')::numeric) returning message_annonce_echange_id",
		annonceID, token.CompteID, message, r.FormValue("prix_propose"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid annonce_echange_id")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Reply sent successfully", "message_annonce_echange_id": newID})
}

func canReadAnnonceMessages(ctx context.Context, conn *pgx.Conn, token *IntrospectionPayload, annonceID int) bool {
	var authorID int
	if err := conn.QueryRow(ctx, "select compte_id from annonce_echange where annonce_echange_id=$1", annonceID).Scan(&authorID); err != nil {
		return false
	}
	if authorID == token.CompteID {
		return true
	}
	var count int
	_ = conn.QueryRow(ctx, "select count(*) from message_annonce_echange where annonce_echange_id=$1 and expediteur_id=$2", annonceID, token.CompteID).Scan(&count)
	return count > 0
}

func listerMessagesAnnonce(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	annonceID, err := strconv.Atoi(r.URL.Query().Get("annonce_echange_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "annonce_echange_id must be a valid integer")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !canReadAnnonceMessages(ctx, conn, token, annonceID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You are not part of this conversation")
		return
	}

	rows, err := conn.Query(
		ctx,
		`select message_annonce_echange.message_annonce_echange_id, expediteur_id, personne.nom, personne.prenom,
			message, prix_propose, date_envoi::text
		from message_annonce_echange
		inner join compte on compte.compte_id = message_annonce_echange.expediteur_id
		inner join personne on personne.personne_id = compte.personne_id
		where annonce_echange_id=$1 order by date_envoi asc`,
		annonceID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query messages")
		return
	}
	defer rows.Close()

	messages := make([]map[string]any, 0)
	for rows.Next() {
		var id, expediteurID int
		var nom, prenom, message, dateEnvoi string
		var prixPropose *float64
		if err := rows.Scan(&id, &expediteurID, &nom, &prenom, &message, &prixPropose, &dateEnvoi); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan messages")
			return
		}
		messages = append(messages, map[string]any{
			"message_annonce_echange_id": id,
			"expediteur_id":              expediteurID,
			"expediteur":                 prenom + " " + nom,
			"message":                    message,
			"prix_propose":               prixPropose,
			"date_envoi":                 dateEnvoi,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"annonce_echange_id": annonceID, "total": len(messages), "messages": messages})
}
