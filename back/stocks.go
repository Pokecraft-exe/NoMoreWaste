package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// stocksHandler manages the storage locations ("gerer les stocks"). Internal
// to the association - staff only.
func stocksHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can manage stocks")
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerStock(w, r)
	case http.MethodGet:
		listerStocks(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerStock(w http.ResponseWriter, r *http.Request) {
	siteIDParam := r.FormValue("site_id")
	nom := r.FormValue("nom")
	typeStock := r.FormValue("type_stock")
	if siteIDParam == "" || nom == "" || typeStock == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "site_id, nom and type_stock are required")
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
		"insert into stock (site_id, nom, type_stock, capacite_max) values ($1, $2, $3, nullif($4, '')::numeric) returning stock_id",
		siteIDParam, nom, typeStock, r.FormValue("capacite_max"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid site_id/type_stock or a stock with this name already exists on this site")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Stock created successfully", "stock_id": newID})
}

func listerStocks(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := "select stock_id, site_id, nom, type_stock, capacite_max, actif from stock"
	args := []any{}
	if siteID := r.URL.Query().Get("site_id"); siteID != "" {
		query += " where site_id = $1"
		args = append(args, siteID)
	}
	query += " order by stock_id"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query stocks")
		return
	}
	defer rows.Close()

	stocks := make([]map[string]any, 0)
	for rows.Next() {
		var id, siteID int
		var nom, typeStock, actif string
		var capaciteMax *float64
		if err := rows.Scan(&id, &siteID, &nom, &typeStock, &capaciteMax, &actif); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan stocks")
			return
		}
		stocks = append(stocks, map[string]any{
			"stock_id":     id,
			"site_id":      siteID,
			"nom":          nom,
			"type_stock":   typeStock,
			"capacite_max": capaciteMax,
			"actif":        actif == "1",
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(stocks), "stocks": stocks})
}

// produitsHandler references every product collected by its barcode, stores
// it and makes it retrievable ("chaque produit rapporte au siege devra etre
// reference (code barre), stocke et retrouvable tres rapidement").
func produitsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can manage produits")
		return
	}

	switch r.Method {
	case http.MethodPut:
		referencerProduit(w, r)
	case http.MethodGet:
		rechercherProduit(w, r)
	case http.MethodPatch:
		modifierProduit(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func referencerProduit(w http.ResponseWriter, r *http.Request) {
	barcode := r.FormValue("barcode")
	nom := r.FormValue("nom")
	if barcode == "" || nom == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "barcode and nom are required")
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
		`insert into produit (collecte_id, categorie_produit_id, stock_id, barcode, nom, description, quantite, poids_kg, date_peremption)
		values (nullif($1,'')::int, nullif($2,'')::int, nullif($3,'')::int, $4, $5, nullif($6,''), coalesce(nullif($7,'')::int, 1), nullif($8,'')::numeric, nullif($9,'')::date)
		returning produit_id`,
		r.FormValue("collecte_id"), r.FormValue("categorie_produit_id"), r.FormValue("stock_id"), barcode, nom,
		r.FormValue("description"), r.FormValue("quantite"), r.FormValue("poids_kg"), r.FormValue("date_peremption"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to reference produit (barcode must be unique)")
		return
	}

	if r.FormValue("stock_id") != "" {
		_, _ = conn.Exec(
			ctx,
			"insert into mouvement_stock (produit_id, stock_id, type_mouvement, quantite, commentaire) values ($1, nullif($2,'')::int, 'entree', coalesce(nullif($3,'')::int,1), 'Reception initiale')",
			newID, r.FormValue("stock_id"), r.FormValue("quantite"),
		)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Produit referenced successfully", "produit_id": newID, "barcode": barcode})
}

func rechercherProduit(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if barcode := r.URL.Query().Get("barcode"); barcode != "" {
		var produitID, quantite int
		var nom, etat, dateReception string
		var stock, site *string
		err = conn.QueryRow(
			ctx,
			"select produit_id, nom, quantite, etat, date_reception, stock, site from v_stock_barcode where barcode = $1",
			barcode,
		).Scan(&produitID, &nom, &quantite, &etat, &dateReception, &stock, &site)
		if err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Produit not found for this barcode")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"produit_id":     produitID,
			"barcode":        barcode,
			"nom":            nom,
			"quantite":       quantite,
			"etat":           etat,
			"date_reception": dateReception,
			"stock":          stock,
			"site":           site,
		})
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	query := "select produit_id, barcode, nom, quantite, etat, date_reception, stock, site from v_stock_barcode"
	args := []any{}
	argPos := 1
	if q := r.URL.Query().Get("q"); q != "" {
		query += " where nom ilike $" + strconv.Itoa(argPos) + " or barcode ilike $" + strconv.Itoa(argPos)
		args = append(args, "%"+q+"%")
		argPos++
	}
	query += " order by date_reception desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query produits")
		return
	}
	defer rows.Close()

	produits := make([]map[string]any, 0)
	for rows.Next() {
		var produitID, quantite int
		var barcode, nom, etat, dateReception string
		var stock, site *string
		if err := rows.Scan(&produitID, &barcode, &nom, &quantite, &etat, &dateReception, &stock, &site); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan produits")
			return
		}
		produits = append(produits, map[string]any{
			"produit_id":     produitID,
			"barcode":        barcode,
			"nom":            nom,
			"quantite":       quantite,
			"etat":           etat,
			"date_reception": dateReception,
			"stock":          stock,
			"site":           site,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(produits), "produits": produits})
}

func modifierProduit(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	produitID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid produit id")
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

	if v := r.FormValue("etat"); v != "" {
		addField("etat", v)
	}
	if v := r.FormValue("stock_id"); v != "" {
		addField("stock_id", v)
	}
	if v := r.FormValue("quantite"); v != "" {
		addField("quantite", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, produitID)
	tag, err := conn.Exec(ctx, "update produit set "+strings.Join(updates, ", ")+" where produit_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update produit")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Produit not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Produit updated successfully"})
}

// mouvementsStockHandler records stock movements (entree, sortie, transfert,
// ajustement) for traceability.
func mouvementsStockHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can record stock movements")
		return
	}

	switch r.Method {
	case http.MethodPut:
		enregistrerMouvementStock(w, r)
	case http.MethodGet:
		listerMouvementsStock(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func enregistrerMouvementStock(w http.ResponseWriter, r *http.Request) {
	produitIDParam := r.FormValue("produit_id")
	typeMouvement := r.FormValue("type_mouvement")
	if produitIDParam == "" || typeMouvement == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "produit_id and type_mouvement are required")
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
		`insert into mouvement_stock (produit_id, stock_id, type_mouvement, quantite, commentaire)
		values ($1, nullif($2,'')::int, $3, coalesce(nullif($4,'')::int,1), nullif($5,''))
		returning mouvement_stock_id`,
		produitIDParam, r.FormValue("stock_id"), typeMouvement, r.FormValue("quantite"), r.FormValue("commentaire"),
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid produit_id/type_mouvement")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Movement recorded successfully", "mouvement_stock_id": newID})
}

func listerMouvementsStock(w http.ResponseWriter, r *http.Request) {
	produitIDParam := r.URL.Query().Get("produit_id")
	if produitIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'produit_id' parameter")
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
		"select mouvement_stock_id, stock_id, type_mouvement, quantite, commentaire, date_mouvement from mouvement_stock where produit_id=$1 order by date_mouvement desc",
		produitIDParam,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query movements")
		return
	}
	defer rows.Close()

	mouvements := make([]map[string]any, 0)
	for rows.Next() {
		var id, quantite int
		var stockID *int
		var typeMouvement, dateMouvement string
		var commentaire *string
		if err := rows.Scan(&id, &stockID, &typeMouvement, &quantite, &commentaire, &dateMouvement); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan movements")
			return
		}
		mouvements = append(mouvements, map[string]any{
			"mouvement_stock_id": id,
			"stock_id":           stockID,
			"type_mouvement":     typeMouvement,
			"quantite":           quantite,
			"commentaire":        commentaire,
			"date_mouvement":     dateMouvement,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(mouvements), "mouvements": mouvements})
}
