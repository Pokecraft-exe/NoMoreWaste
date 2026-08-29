package main

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func notifierCompte(ctx context.Context, conn *pgx.Conn, compteID int, typeNotification, message, lien string) {
	_, _ = conn.Exec(
		ctx,
		"insert into notification (compte_id, type_notification, message, lien) values ($1, $2, $3, nullif($4,''))",
		compteID, typeNotification, message, lien,
	)
}

func notificationsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodGet:
		listerNotifications(w, r, token)
	case http.MethodPatch:
		marquerNotificationsLues(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func listerNotifications(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(
		ctx,
		`select notification_id, type_notification, message, lien, date_notification::text, lu
		from notification where compte_id = $1 order by date_notification desc limit 50`,
		token.CompteID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query notifications")
		return
	}
	defer rows.Close()

	notifications := make([]map[string]any, 0)
	nonLues := 0
	for rows.Next() {
		var id int
		var typeNotification, message, dateNotification, lu string
		var lien *string
		if err := rows.Scan(&id, &typeNotification, &message, &lien, &dateNotification, &lu); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan notifications")
			return
		}
		if lu == "0" {
			nonLues++
		}
		notifications = append(notifications, map[string]any{
			"notification_id": id, "type_notification": typeNotification, "message": message,
			"lien": lien, "date_notification": dateNotification, "lu": lu == "1",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(notifications), "non_lues": nonLues, "notifications": notifications})
}

func marquerNotificationsLues(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "update notification set lu='1' where compte_id=$1", token.CompteID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Notifications marked as read"})
}
