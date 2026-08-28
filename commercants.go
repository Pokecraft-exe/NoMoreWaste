package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// commercantsHandler implements "gerer les adhesions des commercants
// (informations generales, identification, ...)". Staff manages every
// commercant; a commercant account can only read its own record.
func commercantsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodGet:
		getCommercant(w, r, token)
	case http.MethodPatch:
		modifierCommercant(w, r, token)
	case http.MethodDelete:
		supprimerCommercant(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func ownCommercantID(ctx context.Context, conn *pgx.Conn, compteID int) (int, error) {
	var commercantID int
	err := conn.QueryRow(
		ctx,
		"select commercant.commercant_id from commercant inner join compte on compte.personne_id = commercant.personne_id where compte.compte_id = $1",
		compteID,
	).Scan(&commercantID)
	return commercantID, err
}

func getCommercant(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		commercantID, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}
		if !isStaff(token) {
			ownID, ownErr := ownCommercantID(ctx, conn, token.CompteID)
			if ownErr != nil || ownID != commercantID {
				writeAPIError(w, http.StatusForbidden, "access_denied", "You can only access your own commercant profil")
				return
			}
		}
		writeOneCommercant(w, ctx, conn, commercantID)
		return
	}

	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Listing commercants is staff-only. Query your own profil with ?id=")
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	query := "select commercant_id, raison_sociale, identifiant_legal, email, telephone, adresse, code_postal, ville, pays, actif from commercant"
	args := []any{}
	argPos := 1
	if q := r.URL.Query().Get("q"); q != "" {
		query += " where raison_sociale ilike $" + strconv.Itoa(argPos)
		args = append(args, "%"+q+"%")
		argPos++
	}
	query += " order by commercant_id offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query commercants")
		return
	}
	defer rows.Close()

	commercants := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var raisonSociale, adresse, codePostal, ville, pays, actif string
		var identifiantLegal, email, telephone *string
		if err := rows.Scan(&id, &raisonSociale, &identifiantLegal, &email, &telephone, &adresse, &codePostal, &ville, &pays, &actif); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan commercants")
			return
		}
		commercants = append(commercants, map[string]any{
			"commercant_id":     id,
			"raison_sociale":    raisonSociale,
			"identifiant_legal": identifiantLegal,
			"email":             email,
			"telephone":         telephone,
			"adresse":           adresse,
			"code_postal":       codePostal,
			"ville":             ville,
			"pays":              pays,
			"actif":             actif == "1",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(commercants), "commercants": commercants})
}

func writeOneCommercant(w http.ResponseWriter, ctx context.Context, conn *pgx.Conn, commercantID int) {
	var raisonSociale, adresse, codePostal, ville, pays, actif string
	var identifiantLegal, email, telephone *string
	err := conn.QueryRow(
		ctx,
		"select raison_sociale, identifiant_legal, email, telephone, adresse, code_postal, ville, pays, actif from commercant where commercant_id=$1",
		commercantID,
	).Scan(&raisonSociale, &identifiantLegal, &email, &telephone, &adresse, &codePostal, &ville, &pays, &actif)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Commercant not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"commercant_id":     commercantID,
		"raison_sociale":    raisonSociale,
		"identifiant_legal": identifiantLegal,
		"email":             email,
		"telephone":         telephone,
		"adresse":           adresse,
		"code_postal":       codePostal,
		"ville":             ville,
		"pays":              pays,
		"actif":             actif == "1",
	})
}

func modifierCommercant(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	commercantID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid commercant id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if !isStaff(token) {
		ownID, ownErr := ownCommercantID(ctx, conn, token.CompteID)
		if ownErr != nil || ownID != commercantID {
			writeAPIError(w, http.StatusForbidden, "access_denied", "You can only modify your own commercant profil")
			return
		}
	}

	updates := make([]string, 0)
	args := make([]any, 0)
	argIndex := 1
	addField := func(column, value string) {
		updates = append(updates, column+"=$"+strconv.Itoa(argIndex))
		args = append(args, value)
		argIndex++
	}

	if v := r.FormValue("email"); v != "" {
		addField("email", v)
	}
	if v := r.FormValue("telephone"); v != "" {
		addField("telephone", v)
	}
	if v := r.FormValue("adresse"); v != "" {
		addField("adresse", v)
	}
	if v := r.FormValue("code_postal"); v != "" {
		if !codePostalRegexp.MatchString(v) {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "code_postal must be exactly 5 digits")
			return
		}
		addField("code_postal", v)
	}
	if v := r.FormValue("ville"); v != "" {
		addField("ville", v)
	}
	if isStaff(token) {
		if v := r.FormValue("actif"); v != "" {
			addField("actif", v)
		}
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, commercantID)
	query := "update commercant set " + strings.Join(updates, ", ") + " where commercant_id=$" + strconv.Itoa(argIndex)
	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update commercant")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Commercant not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Commercant updated successfully"})
}

func supprimerCommercant(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can deactivate a commercant")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	commercantID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid commercant id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	// Never hard-delete: a commercant may have historical collectes.
	tag, err := conn.Exec(ctx, "update commercant set actif='0' where commercant_id=$1", commercantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to deactivate commercant")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Commercant not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Commercant deactivated successfully"})
}

// adhesionCommercantHandler manages the paid subscription of a commercant
// (creation + automatic renewal reminder system share the same rappel_*
// pattern as adherents, via /api/v1/adherents/rappel).
func adhesionCommercantHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can manage a commercant's adhesion")
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerAdhesionCommercant(w, r, token)
	case http.MethodGet:
		listerAdhesionsCommercant(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerAdhesionCommercant(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	commercantIDParam := r.FormValue("commercant_id")
	forfaitLibelle := r.FormValue("forfait")
	if commercantIDParam == "" || forfaitLibelle == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "commercant_id and forfait are required")
		return
	}
	commercantID, err := strconv.Atoi(commercantIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "commercant_id must be a valid integer")
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
		`insert into adhesion_commercant (commercant_id, forfait_id, date_debut, date_fin, created_by)
		select $1, forfait_id, current_date, (current_date + interval '1 year')::date, $3
		from forfait where libelle = $2
		returning adhesion_commercant_id`,
		commercantID, forfaitLibelle, token.CompteID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Unknown commercant_id or forfait")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Adhesion created successfully", "adhesion_commercant_id": newID})
}

func listerAdhesionsCommercant(w http.ResponseWriter, r *http.Request) {
	commercantIDParam := r.URL.Query().Get("commercant_id")
	if commercantIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'commercant_id' parameter")
		return
	}
	commercantID, err := strconv.Atoi(commercantIDParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid commercant_id")
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
		"select adhesion_commercant_id, forfait_id, date_debut, date_fin, statut from adhesion_commercant where commercant_id=$1 order by date_debut desc",
		commercantID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query adhesions")
		return
	}
	defer rows.Close()

	adhesions := make([]map[string]any, 0)
	for rows.Next() {
		var id, forfaitID int
		var dateDebut, statut string
		var dateFin *string
		if err := rows.Scan(&id, &forfaitID, &dateDebut, &dateFin, &statut); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan adhesions")
			return
		}
		adhesions = append(adhesions, map[string]any{
			"adhesion_commercant_id": id,
			"forfait_id":             forfaitID,
			"date_debut":             dateDebut,
			"date_fin":               dateFin,
			"statut":                 statut,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"commercant_id": commercantID, "total": len(adhesions), "adhesions": adhesions})
}
