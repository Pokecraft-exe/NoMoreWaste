package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func ticketsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		creerTicket(w, r)
	case http.MethodGet:
		rechercherTicket(w, r)
	case http.MethodPatch:
		repondreTicket(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerTicket(w http.ResponseWriter, r *http.Request) {
	sujet := r.FormValue("sujet")
	message := r.FormValue("message")
	if sujet == "" || message == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "sujet and message are required")
		return
	}

	token := tryOptionalAuth(r)

	var auteurID any
	var contactNom, contactEmail string
	if token.Active {
		auteurID = token.CompteID
	} else {
		contactNom = r.FormValue("contact_nom")
		contactEmail = r.FormValue("contact_email")
		if contactNom == "" || contactEmail == "" {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "contact_nom and contact_email are required when not authenticated")
			return
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
		"insert into ticket (auteur_id, contact_nom, contact_email, sujet, message) values ($1, nullif($2,''), nullif($3,''), $4, $5) returning ticket_id",
		auteurID, contactNom, contactEmail, sujet, message,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the ticket")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Ticket created successfully", "ticket_id": newID})
}

func rechercherTicket(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can search tickets")
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

	query := "select ticket_id, auteur_id, contact_nom, contact_email, sujet, message, statut, reponse, date_creation::text from ticket"
	args := []any{}
	argPos := 1
	if statut := r.URL.Query().Get("statut"); statut != "" {
		query += " where statut = $" + strconv.Itoa(argPos)
		args = append(args, statut)
		argPos++
	}
	query += " order by date_creation desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query tickets")
		return
	}
	defer rows.Close()

	tickets := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var auteurID *int
		var contactNom, contactEmail, reponse *string
		var sujet, message, statut, dateCreation string
		if err := rows.Scan(&id, &auteurID, &contactNom, &contactEmail, &sujet, &message, &statut, &reponse, &dateCreation); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan tickets")
			return
		}
		tickets = append(tickets, map[string]any{
			"ticket_id": id, "auteur_id": auteurID, "contact_nom": contactNom, "contact_email": contactEmail,
			"sujet": sujet, "message": message, "statut": statut, "reponse": reponse, "date_creation": dateCreation,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(tickets), "tickets": tickets})
}

func repondreTicket(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can answer a ticket")
		return
	}

	idParam := r.URL.Query().Get("id")
	reponse := r.FormValue("reponse")
	if idParam == "" || reponse == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id and reponse are required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var auteurID *int
	_ = conn.QueryRow(ctx, "select auteur_id from ticket where ticket_id=$1", idParam).Scan(&auteurID)

	tag, err := conn.Exec(
		ctx,
		"update ticket set statut='traite', reponse=$1, traite_par=$2, date_traitement=now() where ticket_id=$3",
		reponse, token.CompteID, idParam,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to answer the ticket")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Ticket not found")
		return
	}

	if auteurID != nil {
		notifierCompte(ctx, conn, *auteurID, "ticket_traite", "Votre ticket a ete traite : "+reponse, "")
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Ticket answered successfully"})
}
