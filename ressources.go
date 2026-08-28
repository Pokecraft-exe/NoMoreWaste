package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ressourcesCuisineHandler implements "Accéder aux recettes/cours de
// cuisine": staff publishes content, every authenticated account can browse
// it.
func ressourcesCuisineHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can publish a recette or cours de cuisine")
			return
		}
		creerRessourceCuisine(w, r, token)
	case http.MethodGet:
		listerRessourcesCuisine(w, r)
	case http.MethodPatch:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can modify a recette")
			return
		}
		modifierRessourceCuisine(w, r)
	case http.MethodDelete:
		if !isStaff(token) {
			writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can delete a recette")
			return
		}
		supprimerRessourceCuisine(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func splitCsv(v string) []string {
	if v == "" {
		return []string{}
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func creerRessourceCuisine(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	titre := r.FormValue("titre")
	contenu := r.FormValue("contenu")
	if titre == "" || contenu == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "titre and contenu are required")
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
		"insert into ressource_cuisine (titre, ingredients, outils, contenu, video, created_by) values ($1, $2, $3, $4, nullif($5,''), $6) returning ressource_cuisine_id",
		titre, splitCsv(r.FormValue("ingredients")), splitCsv(r.FormValue("outils")), contenu, r.FormValue("video"), token.CompteID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the ressource")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Ressource created successfully", "ressource_cuisine_id": newID})
}

func listerRessourcesCuisine(w http.ResponseWriter, r *http.Request) {
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
		var titre, contenu string
		var ingredients, outils []string
		var video *string
		err = conn.QueryRow(ctx, "select titre, ingredients, outils, contenu, video from ressource_cuisine where ressource_cuisine_id=$1", id).
			Scan(&titre, &ingredients, &outils, &contenu, &video)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Ressource not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ressource_cuisine_id": id,
			"titre":                titre,
			"ingredients":          ingredients,
			"outils":               outils,
			"contenu":              contenu,
			"video":                video,
		})
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	query := "select ressource_cuisine_id, titre, ingredients, outils, video from ressource_cuisine"
	args := []any{}
	argPos := 1
	if q := r.URL.Query().Get("q"); q != "" {
		query += " where titre ilike $" + strconv.Itoa(argPos) + " or $" + strconv.Itoa(argPos) + " ilike any(ingredients)"
		args = append(args, "%"+q+"%")
		argPos++
	}
	query += " order by ressource_cuisine_id offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query ressources")
		return
	}
	defer rows.Close()

	ressources := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var titre string
		var ingredients, outils []string
		var video *string
		if err := rows.Scan(&id, &titre, &ingredients, &outils, &video); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan ressources")
			return
		}
		ressources = append(ressources, map[string]any{
			"ressource_cuisine_id": id,
			"titre":                titre,
			"ingredients":          ingredients,
			"outils":               outils,
			"video":                video,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(ressources), "recettes": ressources})
}

func modifierRessourceCuisine(w http.ResponseWriter, r *http.Request) {
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

	updates := make([]string, 0)
	args := make([]any, 0)
	argIndex := 1
	addField := func(column string, value any) {
		updates = append(updates, column+"=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	if v := r.FormValue("titre"); v != "" {
		addField("titre", v)
	}
	if v := r.FormValue("ingredients"); v != "" {
		addField("ingredients", splitCsv(v))
	}
	if v := r.FormValue("outils"); v != "" {
		addField("outils", splitCsv(v))
	}
	if v := r.FormValue("contenu"); v != "" {
		addField("contenu", v)
	}
	if v := r.FormValue("video"); v != "" {
		addField("video", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, idParam)
	tag, err := conn.Exec(ctx, "update ressource_cuisine set "+strings.Join(updates, ", ")+" where ressource_cuisine_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the ressource")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Ressource not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Ressource updated successfully"})
}

func supprimerRessourceCuisine(w http.ResponseWriter, r *http.Request) {
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

	tag, err := conn.Exec(ctx, "delete from ressource_cuisine where ressource_cuisine_id=$1", idParam)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to delete the ressource")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Ressource not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
