package main

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// stockProduitsHandler is the lightweight product-type catalog (nom + unite
// + quantite_disponible) used for collecte/distribution denree quantities -
// distinct from stock/produit which track individual barcoded items.
func stockProduitsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerStockProduit(w, r)
	case http.MethodGet:
		rechercherStockProduit(w, r)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerStockProduit(w http.ResponseWriter, r *http.Request) {
	nom := r.FormValue("nom")
	if nom == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "nom is required")
		return
	}
	unite := r.FormValue("unite")
	if unite == "" {
		unite = "unites"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var existingID int
	err = conn.QueryRow(ctx, "select stock_produit_id from stock_produit where lower(nom) = lower(trim($1))", nom).Scan(&existingID)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"message": "Produit already exists", "stock_produit_id": existingID})
		return
	}

	var newID int
	err = conn.QueryRow(ctx, "insert into stock_produit (nom, unite) values ($1, $2) returning stock_produit_id", nom, unite).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the produit")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Produit created successfully", "stock_produit_id": newID})
}

func rechercherStockProduit(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := "select stock_produit_id, nom, unite, quantite_disponible from stock_produit"
	args := []any{}
	if q := r.URL.Query().Get("q"); q != "" {
		query += " where nom ilike $1"
		args = append(args, "%"+q+"%")
	}
	query += " order by nom limit 20"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query produits")
		return
	}
	defer rows.Close()

	produits := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var nom, unite string
		var quantiteDisponible float64
		if err := rows.Scan(&id, &nom, &unite, &quantiteDisponible); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan produits")
			return
		}
		produits = append(produits, map[string]any{
			"stock_produit_id": id, "nom": nom, "unite": unite, "quantite_disponible": quantiteDisponible,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(produits), "produits": produits})
}
