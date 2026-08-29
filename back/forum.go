package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// forumHandler implements "Accéder à forum de conseil" (GET) and "Créer un
// forum" (PUT). Every authenticated account can participate.
func forumHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerThread(w, r, token)
	case http.MethodGet:
		if r.URL.Query().Get("id") != "" {
			consulterThread(w, r)
		} else {
			listerThreads(w, r)
		}
	case http.MethodPatch:
		modifierThread(w, r, token)
	case http.MethodDelete:
		supprimerThread(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	titre := r.FormValue("titre")
	message := r.FormValue("message")
	if titre == "" || message == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "titre and message are required")
		return
	}

	// A staff member creating a thread from the back-office can post it on
	// behalf of another account by passing its compte_id.
	compteID := token.CompteID
	if v := r.FormValue("compte_id"); v != "" && isStaff(token) {
		if parsed, convErr := strconv.Atoi(v); convErr == nil {
			compteID = parsed
		}
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
		"insert into forum_thread (compte_id, titre, message) values ($1, $2, $3) returning forum_thread_id",
		compteID, titre, message,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the thread")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Thread created successfully", "forum_thread_id": newID})
}

func listerThreads(w http.ResponseWriter, r *http.Request) {
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

	query := `select forum_thread.forum_thread_id, forum_thread.compte_id, personne.nom, personne.prenom, forum_thread.titre, forum_thread.message,
		forum_thread.vues, forum_thread.date_creation::text
		from forum_thread
		inner join compte on compte.compte_id = forum_thread.compte_id
		inner join personne on personne.personne_id = compte.personne_id`
	conditions := make([]string, 0)
	args := []any{}
	argPos := 1
	if q := r.URL.Query().Get("q"); q != "" {
		conditions = append(conditions, "forum_thread.titre ilike $"+strconv.Itoa(argPos))
		args = append(args, "%"+q+"%")
		argPos++
	}
	if auteurParam := r.URL.Query().Get("auteur_id"); auteurParam != "" {
		if auteurID, convErr := strconv.Atoi(auteurParam); convErr == nil {
			conditions = append(conditions, "forum_thread.compte_id = $"+strconv.Itoa(argPos))
			args = append(args, auteurID)
			argPos++
		}
	}
	for i, c := range conditions {
		if i == 0 {
			query += " where " + c
		} else {
			query += " and " + c
		}
	}
	query += " order by forum_thread.date_creation desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query threads")
		return
	}
	defer rows.Close()

	threads := make([]map[string]any, 0)
	for rows.Next() {
		var id, auteurID, vues int
		var nom, prenom, titre, message, dateCreation string
		if err := rows.Scan(&id, &auteurID, &nom, &prenom, &titre, &message, &vues, &dateCreation); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan threads")
			return
		}
		threads = append(threads, map[string]any{
			"forum_thread_id": id,
			"auteur_id":       auteurID,
			"auteur":          prenom + " " + nom,
			"titre":           titre,
			"message":         message,
			"vues":            vues,
			"date_creation":   dateCreation,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(threads), "threads": threads})
}

func consulterThread(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var auteurID, vues int
	var nom, prenom, titre, message, dateCreation string
	err = conn.QueryRow(
		ctx,
		`select forum_thread.compte_id, personne.nom, personne.prenom, forum_thread.titre, forum_thread.message,
			forum_thread.vues, forum_thread.date_creation::text
		from forum_thread
		inner join compte on compte.compte_id = forum_thread.compte_id
		inner join personne on personne.personne_id = compte.personne_id
		where forum_thread_id = $1`,
		threadID,
	).Scan(&auteurID, &nom, &prenom, &titre, &message, &vues, &dateCreation)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Thread not found")
		return
	}

	_, _ = conn.Exec(ctx, "update forum_thread set vues = vues + 1 where forum_thread_id = $1", threadID)

	writeJSON(w, http.StatusOK, map[string]any{
		"forum_thread_id": threadID,
		"auteur_id":       auteurID,
		"auteur":          prenom + " " + nom,
		"titre":           titre,
		"message":         message,
		"vues":            vues + 1,
		"date_creation":   dateCreation,
	})
}

func modifierThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	threadID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid thread id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var authorID int
	if err := conn.QueryRow(ctx, "select compte_id from forum_thread where forum_thread_id=$1", threadID).Scan(&authorID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Thread not found")
		return
	}
	if authorID != token.CompteID && !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You don't have the right to modify this thread")
		return
	}

	titre := r.FormValue("titre")
	message := r.FormValue("message")
	if titre == "" && message == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "titre or message is required")
		return
	}

	if _, err := conn.Exec(
		ctx,
		"update forum_thread set titre=coalesce(nullif($1,''),titre), message=coalesce(nullif($2,''),message) where forum_thread_id=$3",
		titre, message, threadID,
	); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the thread")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Thread updated successfully"})
}

func supprimerThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	threadID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid thread id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var authorID int
	if err := conn.QueryRow(ctx, "select compte_id from forum_thread where forum_thread_id=$1", threadID).Scan(&authorID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Thread not found")
		return
	}
	if authorID != token.CompteID && !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You don't have the right to delete this thread")
		return
	}

	if _, err := conn.Exec(ctx, "delete from forum_thread where forum_thread_id=$1", threadID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the thread")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// forumMessagesHandler implements replying inside a thread.
func forumMessagesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		repondreThread(w, r, token)
	case http.MethodGet:
		listerMessagesThread(w, r)
	case http.MethodDelete:
		supprimerMessageThread(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func repondreThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	threadID := r.FormValue("forum_thread_id")
	message := r.FormValue("message")
	if threadID == "" || message == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "forum_thread_id and message are required")
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
		"insert into forum_message (forum_thread_id, compte_id, message, parent_id) values ($1, $2, $3, nullif($4,'')::int) returning forum_message_id",
		threadID, token.CompteID, message, r.FormValue("parent_id"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid forum_thread_id or parent_id")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Reply posted successfully", "forum_message_id": newID})
}

func listerMessagesThread(w http.ResponseWriter, r *http.Request) {
	threadID := r.URL.Query().Get("forum_thread_id")
	if threadID == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'forum_thread_id' parameter")
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
		`select forum_message.forum_message_id, forum_message.compte_id, personne.nom, personne.prenom,
			forum_message.message, forum_message.parent_id, forum_message.date_envoi::text
		from forum_message
		inner join compte on compte.compte_id = forum_message.compte_id
		inner join personne on personne.personne_id = compte.personne_id
		where forum_message.forum_thread_id = $1 order by forum_message.date_envoi asc`,
		threadID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query messages")
		return
	}
	defer rows.Close()

	messages := make([]map[string]any, 0)
	for rows.Next() {
		var id, auteurID int
		var nom, prenom, message, dateEnvoi string
		var parentID *int
		if err := rows.Scan(&id, &auteurID, &nom, &prenom, &message, &parentID, &dateEnvoi); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan messages")
			return
		}
		messages = append(messages, map[string]any{
			"forum_message_id": id,
			"auteur_id":        auteurID,
			"auteur":           prenom + " " + nom,
			"message":          message,
			"parent_id":        parentID,
			"date_envoi":       dateEnvoi,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"forum_thread_id": threadID, "total": len(messages), "messages": messages})
}

func supprimerMessageThread(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	messageID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid message id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var authorID int
	if err := conn.QueryRow(ctx, "select compte_id from forum_message where forum_message_id=$1", messageID).Scan(&authorID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Message not found")
		return
	}
	if authorID != token.CompteID && !isAdministrateur(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You don't have the right to delete this message")
		return
	}

	if _, err := conn.Exec(ctx, "delete from forum_message where forum_message_id=$1", messageID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the message")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
