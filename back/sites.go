package main

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// sitesHandler is a minimal read-only listing of the association's sites
// (siege, agences, entrepots, ...), seeded in db/sql.sql. Used to populate
// the site_id dropdown when creating a stock storage location - staff only,
// since it's purely an internal back-office concern.
func sitesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can list sites")
		return
	}
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "select site_id, nom, type_site, ville from site order by site_id")
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query sites")
		return
	}
	defer rows.Close()

	sites := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var nom, typeSite, ville string
		if err := rows.Scan(&id, &nom, &typeSite, &ville); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan sites")
			return
		}
		sites = append(sites, map[string]any{"site_id": id, "nom": nom, "type_site": typeSite, "ville": ville})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(sites), "sites": sites})
}
