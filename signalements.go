package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func signalementsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerSignalement(w, r, token)
	case http.MethodGet:
		if r.URL.Query().Get("id") != "" {
			consulterSignalement(w, r, token)
		} else {
			rechercherSignalement(w, r, token)
		}
	case http.MethodPatch:
		resoudreSignalementHandler(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerSignalement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	typeSignalement := r.FormValue("type_signalement")
	motif := r.FormValue("motif")
	if typeSignalement == "" || motif == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "type_signalement and motif are required")
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
	switch typeSignalement {
	case "forum":
		forumThreadID := r.FormValue("forum_thread_id")
		if forumThreadID == "" {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "forum_thread_id is required for type_signalement=forum")
			return
		}
		err = conn.QueryRow(
			ctx,
			"insert into signalement (type_signalement, forum_thread_id, signale_par, motif) values ('forum', $1, $2, $3) returning signalement_id",
			forumThreadID, token.CompteID, motif,
		).Scan(&newID)
	case "forum_message":
		forumThreadID := r.FormValue("forum_thread_id")
		forumMessageID := r.FormValue("forum_message_id")
		if forumThreadID == "" || forumMessageID == "" {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "forum_thread_id and forum_message_id are required for type_signalement=forum_message")
			return
		}
		err = conn.QueryRow(
			ctx,
			"insert into signalement (type_signalement, forum_thread_id, forum_message_id, signale_par, motif) values ('forum_message', $1, $2, $3, $4) returning signalement_id",
			forumThreadID, forumMessageID, token.CompteID, motif,
		).Scan(&newID)
	case "annonce_message":
		annonceMessageID := r.FormValue("annonce_message_id")
		if annonceMessageID == "" {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "annonce_message_id is required for type_signalement=annonce_message")
			return
		}
		err = conn.QueryRow(
			ctx,
			"insert into signalement (type_signalement, annonce_message_id, signale_par, motif) values ('annonce_message', $1, $2, $3) returning signalement_id",
			annonceMessageID, token.CompteID, motif,
		).Scan(&newID)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "type_signalement must be forum, forum_message or annonce_message")
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to create the signalement")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Signalement submitted successfully", "signalement_id": newID})
}

func rechercherSignalement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can search signalements")
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

	query := `select signalement.signalement_id, signalement.type_signalement, signalement.signale_par, personne.nom, personne.prenom,
		signalement.motif, signalement.statut, signalement.date_signalement::text
		from signalement
		inner join compte on compte.compte_id = signalement.signale_par
		inner join personne on personne.personne_id = compte.personne_id`
	args := []any{}
	argPos := 1
	if statut := r.URL.Query().Get("statut"); statut != "" {
		query += " where signalement.statut = $" + strconv.Itoa(argPos)
		args = append(args, statut)
		argPos++
	}
	query += " order by signalement.date_signalement desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query signalements")
		return
	}
	defer rows.Close()

	signalements := make([]map[string]any, 0)
	for rows.Next() {
		var id, signalePar int
		var typeSignalement, nom, prenom, motif, statut, dateSignalement string
		if err := rows.Scan(&id, &typeSignalement, &signalePar, &nom, &prenom, &motif, &statut, &dateSignalement); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan signalements")
			return
		}
		signalements = append(signalements, map[string]any{
			"signalement_id": id, "type_signalement": typeSignalement, "signale_par": signalePar,
			"signale_par_nom": prenom + " " + nom, "motif": motif, "statut": statut, "date_signalement": dateSignalement,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(signalements), "signalements": signalements})
}

func consulterSignalement(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can consult a signalement")
		return
	}

	signalementID, err := strconv.Atoi(r.URL.Query().Get("id"))
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

	var typeSignalement, motif, statut, dateSignalement string
	var forumThreadID, forumMessageID, annonceMessageID *int
	var signalePar int
	var commentaire *string
	err = conn.QueryRow(
		ctx,
		`select type_signalement, forum_thread_id, forum_message_id, annonce_message_id, signale_par, motif, statut, commentaire, date_signalement::text
		from signalement where signalement_id = $1`,
		signalementID,
	).Scan(&typeSignalement, &forumThreadID, &forumMessageID, &annonceMessageID, &signalePar, &motif, &statut, &commentaire, &dateSignalement)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Signalement not found")
		return
	}

	response := map[string]any{
		"signalement_id": signalementID, "type_signalement": typeSignalement, "signale_par": signalePar,
		"motif": motif, "statut": statut, "commentaire": commentaire, "date_signalement": dateSignalement,
	}

	if statut == "traite" {
		var auteurID int
		var message, dateMessage string
		err = conn.QueryRow(
			ctx,
			"select auteur_id, message, date_message::text from message_archive where signalement_id = $1",
			signalementID,
		).Scan(&auteurID, &message, &dateMessage)
		if err == nil {
			response["message_signale"] = map[string]any{"auteur_id": auteurID, "message": message, "date": dateMessage}
		}
		writeJSON(w, http.StatusOK, response)
		return
	}

	messages := make([]map[string]any, 0)
	if typeSignalement == "annonce_message" {
		rows, err := conn.Query(
			ctx,
			`select message_annonce_echange.message_annonce_echange_id, message_annonce_echange.expediteur_id, personne.nom, personne.prenom,
				message_annonce_echange.message, message_annonce_echange.date_envoi::text
			from message_annonce_echange
			inner join compte on compte.compte_id = message_annonce_echange.expediteur_id
			inner join personne on personne.personne_id = compte.personne_id
			where annonce_echange_id = (select annonce_echange_id from message_annonce_echange where message_annonce_echange_id = $1)
			order by date_envoi`,
			*annonceMessageID,
		)
		if err == nil {
			for rows.Next() {
				var messageID, auteurID int
				var nom, prenom, message, date string
				if rows.Scan(&messageID, &auteurID, &nom, &prenom, &message, &date) == nil {
					messages = append(messages, map[string]any{
						"auteur_id": auteurID, "auteur": prenom + " " + nom, "message": message, "date": date,
						"signale": messageID == *annonceMessageID,
					})
				}
			}
			rows.Close()
		}
	} else if forumThreadID != nil {
		var auteurID int
		var nom, prenom, message, date string
		if conn.QueryRow(
			ctx,
			`select forum_thread.compte_id, personne.nom, personne.prenom, forum_thread.message, forum_thread.date_creation::text
			from forum_thread
			inner join compte on compte.compte_id = forum_thread.compte_id
			inner join personne on personne.personne_id = compte.personne_id
			where forum_thread_id = $1`,
			*forumThreadID,
		).Scan(&auteurID, &nom, &prenom, &message, &date) == nil {
			messages = append(messages, map[string]any{
				"auteur_id": auteurID, "auteur": prenom + " " + nom, "message": message, "date": date,
				"signale": typeSignalement == "forum",
			})
		}

		rows, err := conn.Query(
			ctx,
			`select forum_message.forum_message_id, forum_message.compte_id, personne.nom, personne.prenom,
				forum_message.message, forum_message.date_envoi::text
			from forum_message
			inner join compte on compte.compte_id = forum_message.compte_id
			inner join personne on personne.personne_id = compte.personne_id
			where forum_thread_id = $1 order by date_envoi`,
			*forumThreadID,
		)
		if err == nil {
			for rows.Next() {
				var messageID, auteurID int
				var nom, prenom, message, date string
				if rows.Scan(&messageID, &auteurID, &nom, &prenom, &message, &date) == nil {
					messages = append(messages, map[string]any{
						"auteur_id": auteurID, "auteur": prenom + " " + nom, "message": message, "date": date,
						"signale": forumMessageID != nil && messageID == *forumMessageID,
					})
				}
			}
			rows.Close()
		}
	}

	response["discussion"] = messages
	writeJSON(w, http.StatusOK, response)
}

func resoudreSignalementHandler(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can resolve a signalement")
		return
	}

	idParam := r.URL.Query().Get("id")
	commentaire := r.FormValue("commentaire")
	if idParam == "" || commentaire == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "id and commentaire are required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var signalementID int
	var typeSignalement string
	var forumThreadID, forumMessageID, annonceMessageID *int
	var signalePar int
	err = conn.QueryRow(
		ctx,
		"select signalement_id, type_signalement, forum_thread_id, forum_message_id, annonce_message_id, signale_par from signalement where signalement_id=$1 and statut='ouvert'",
		idParam,
	).Scan(&signalementID, &typeSignalement, &forumThreadID, &forumMessageID, &annonceMessageID, &signalePar)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Signalement not found, or already resolved")
		return
	}

	var flaggedAuteurID int
	var flaggedMessage, flaggedDate string
	haveFlaggedMessage := false
	switch {
	case typeSignalement == "annonce_message" && annonceMessageID != nil:
		haveFlaggedMessage = conn.QueryRow(ctx, "select expediteur_id, message, date_envoi::text from message_annonce_echange where message_annonce_echange_id=$1", *annonceMessageID).
			Scan(&flaggedAuteurID, &flaggedMessage, &flaggedDate) == nil
	case typeSignalement == "forum_message" && forumMessageID != nil:
		haveFlaggedMessage = conn.QueryRow(ctx, "select compte_id, message, date_envoi::text from forum_message where forum_message_id=$1", *forumMessageID).
			Scan(&flaggedAuteurID, &flaggedMessage, &flaggedDate) == nil
	case forumThreadID != nil:
		haveFlaggedMessage = conn.QueryRow(ctx, "select compte_id, message, date_creation::text from forum_thread where forum_thread_id=$1", *forumThreadID).
			Scan(&flaggedAuteurID, &flaggedMessage, &flaggedDate) == nil
	}

	tx, txErr := conn.Begin(ctx)
	if txErr != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to resolve the signalement")
		return
	}
	defer tx.Rollback(ctx)

	tag, execErr := tx.Exec(
		ctx,
		"update signalement set statut='traite', commentaire=$1, traite_par=$2, date_traitement=now() where signalement_id=$3 and statut='ouvert'",
		commentaire, token.CompteID, signalementID,
	)
	if execErr != nil || tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to resolve the signalement")
		return
	}

	if haveFlaggedMessage {
		if _, archErr := tx.Exec(
			ctx,
			"insert into message_archive (signalement_id, auteur_id, message, date_message) values ($1, $2, $3, $4)",
			signalementID, flaggedAuteurID, flaggedMessage, flaggedDate,
		); archErr != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to archive the flagged message")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to resolve the signalement")
		return
	}

	notifierCompte(ctx, conn, signalePar, "signalement_traite", "Votre signalement a ete traite : "+commentaire, "")

	writeJSON(w, http.StatusOK, map[string]any{"message": "Signalement resolved successfully"})
}
