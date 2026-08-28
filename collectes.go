package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// collectesHandler covers "rechercher/modifier/creer une collecte" and "voir
// la prochaine collecte" from the use case diagram.
func collectesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerCollecte(w, r, token)
	case http.MethodGet:
		rechercherCollecte(w, r)
	case http.MethodPatch:
		modifierCollecte(w, r, token)
	case http.MethodDelete:
		supprimerCollecte(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

// creerCollecte: staff can create any collecte; a commercant can only create
// a pickup request tied to their own commercant_id.
func creerCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) && !isCommercant(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff or a commercant can request a collecte")
		return
	}

	lieu := r.FormValue("lieu")
	dateCollecte := r.FormValue("date_collecte")
	if lieu == "" || dateCollecte == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "lieu and date_collecte are required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	var commercantID *int
	if isCommercant(token) {
		ownID, ownErr := ownCommercantID(ctx, conn, token.CompteID)
		if ownErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "No commercant profil found for this account")
			return
		}
		commercantID = &ownID
	} else if v := r.FormValue("commercant_id"); v != "" {
		id, convErr := strconv.Atoi(v)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "commercant_id must be a valid integer")
			return
		}
		commercantID = &id
	}

	statut := r.FormValue("statut")
	if statut == "" {
		statut = "planifiee"
	}

	var newID int
	err = conn.QueryRow(
		ctx,
		`insert into collecte (commercant_id, lieu, date_collecte, heure_collecte, description, statut, created_by)
		values ($1, $2, $3, nullif($4, ''), nullif($5, ''), $6, $7) returning collecte_id`,
		commercantID, lieu, dateCollecte, r.FormValue("heure_collecte"), r.FormValue("description"), statut, token.CompteID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create collecte")
		return
	}

	if partenaire := r.FormValue("partenaire"); partenaire != "" {
		_, _ = conn.Exec(
			ctx,
			`update collecte set partenaire_id = (
				insert into partenaire (nom, type_partenaire) values ($1, 'association')
				on conflict (nom) do update set nom = excluded.nom returning partenaire_id
			) where collecte_id = $2`,
			partenaire, newID,
		)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Collecte created successfully", "collecte_id": newID})
}

// rechercherCollecte is open to every authenticated account. ?prochaine=1
// covers "voir la prochaine collecte".
func rechercherCollecte(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if idParam := r.URL.Query().Get("id"); idParam != "" {
		collecteID, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}
		writeCollecteDetail(w, r, conn, collecteID)
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	query := `select collecte.collecte_id, collecte.lieu, collecte.date_collecte::text, collecte.heure_collecte::text,
		coalesce(partenaire.nom, ''), collecte.statut, collecte.description
		from collecte left join partenaire on partenaire.partenaire_id = collecte.partenaire_id`
	conditions := make([]string, 0)
	args := []any{}
	argPos := 1

	if r.URL.Query().Get("prochaine") == "1" {
		conditions = append(conditions, "collecte.date_collecte >= current_date and collecte.statut <> 'annulee'")
	}
	if statut := r.URL.Query().Get("statut"); statut != "" {
		conditions = append(conditions, "collecte.statut = $"+strconv.Itoa(argPos))
		args = append(args, statut)
		argPos++
	}
	if q := r.URL.Query().Get("q"); q != "" {
		conditions = append(conditions, "(collecte.lieu ilike $"+strconv.Itoa(argPos)+" or coalesce(partenaire.nom,'') ilike $"+strconv.Itoa(argPos)+")")
		args = append(args, "%"+q+"%")
		argPos++
	}
	for i, c := range conditions {
		if i == 0 {
			query += " where " + c
		} else {
			query += " and " + c
		}
	}
	query += " order by collecte.date_collecte desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query collectes")
		return
	}
	defer rows.Close()

	collectes := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var lieu, dateCollecte, partenaire, statut string
		var heureCollecte, description *string
		if err := rows.Scan(&id, &lieu, &dateCollecte, &heureCollecte, &partenaire, &statut, &description); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan collectes")
			return
		}
		collectes = append(collectes, map[string]any{
			"collecte_id":    id,
			"lieu":           lieu,
			"date_collecte":  dateCollecte,
			"heure_collecte": heureCollecte,
			"partenaire":     partenaire,
			"statut":         statut,
			"description":    description,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(collectes), "collectes": collectes})
}

func writeCollecteDetail(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, collecteID int) {
	ctx := r.Context()

	var lieu, dateCollecte, statut, stockMisAJour string
	var heureCollecte, description, partenaire *string
	var commercantID *int
	err := conn.QueryRow(
		ctx,
		`select collecte.lieu, collecte.date_collecte::text, collecte.heure_collecte::text, collecte.statut, collecte.description,
			collecte.stock_mis_a_jour, collecte.commercant_id, partenaire.nom
		from collecte left join partenaire on partenaire.partenaire_id = collecte.partenaire_id
		where collecte.collecte_id = $1`,
		collecteID,
	).Scan(&lieu, &dateCollecte, &heureCollecte, &statut, &description, &stockMisAJour, &commercantID, &partenaire)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Collecte not found")
		return
	}

	rows, err := conn.Query(
		ctx,
		`select collecte_denree.collecte_denree_id, collecte_denree.stock_produit_id, stock_produit.nom, collecte_denree.quantite,
			collecte_denree.non_perissable, collecte_denree.dlc::text, collecte_denree.propose_par, collecte_denree.propose_par_type,
			collecte_denree.confirmee, collecte_denree.date_ajout::text
		from collecte_denree inner join stock_produit on stock_produit.stock_produit_id = collecte_denree.stock_produit_id
		where collecte_denree.collecte_id = $1 order by collecte_denree.date_ajout`,
		collecteID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query denrees")
		return
	}
	denrees := make([]map[string]any, 0)
	for rows.Next() {
		var id, stockProduitID, proposePar int
		var nom, proposeParType string
		var quantite float64
		var nonPerissable string
		var dlc *string
		var confirmee string
		var dateAjout string
		if err := rows.Scan(&id, &stockProduitID, &nom, &quantite, &nonPerissable, &dlc, &proposePar, &proposeParType, &confirmee, &dateAjout); err != nil {
			rows.Close()
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan denrees")
			return
		}
		denrees = append(denrees, map[string]any{
			"collecte_denree_id": id, "stock_produit_id": stockProduitID, "nom": nom, "quantite": quantite,
			"non_perissable": nonPerissable == "1", "dlc": dlc, "propose_par": proposePar, "propose_par_type": proposeParType,
			"confirmee": confirmee == "1", "date_ajout": dateAjout,
		})
	}
	rows.Close()

	benevoles, err := conn.Query(
		ctx,
		`select compte.compte_id, personne.nom, personne.prenom, collecte_benevole_affecte.role_mission
		from collecte_benevole_affecte
		inner join benevole on benevole.benevole_id = collecte_benevole_affecte.benevole_id
		inner join personne on personne.personne_id = benevole.personne_id
		inner join compte on compte.personne_id = personne.personne_id
		where collecte_benevole_affecte.collecte_id = $1`,
		collecteID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query affectations")
		return
	}
	affectations := make([]map[string]any, 0)
	for benevoles.Next() {
		var benevoleCompteID int
		var nom, prenom, roleMission string
		if err := benevoles.Scan(&benevoleCompteID, &nom, &prenom, &roleMission); err != nil {
			benevoles.Close()
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan affectations")
			return
		}
		affectations = append(affectations, map[string]any{"benevole_id": benevoleCompteID, "nom": nom, "prenom": prenom, "role_mission": roleMission})
	}
	benevoles.Close()

	writeJSON(w, http.StatusOK, map[string]any{
		"collecte_id":        collecteID,
		"lieu":               lieu,
		"date_collecte":      dateCollecte,
		"heure_collecte":     heureCollecte,
		"partenaire":         partenaire,
		"commercant_id":      commercantID,
		"statut":             statut,
		"description":        description,
		"stock_mis_a_jour":   stockMisAJour == "1",
		"denrees":            denrees,
		"benevoles_affectes": affectations,
	})
}

func modifierCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can modify a collecte")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	collecteID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid collecte id")
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

	if v := r.FormValue("lieu"); v != "" {
		addField("lieu", v)
	}
	if v := r.FormValue("date_collecte"); v != "" {
		addField("date_collecte", v)
	}
	if v := r.FormValue("heure_collecte"); v != "" {
		addField("heure_collecte", v)
	}
	if v := r.FormValue("statut"); v != "" {
		addField("statut", v)
	}
	if v := r.FormValue("description"); v != "" {
		addField("description", v)
	}

	if v := r.FormValue("partenaire"); v != "" {
		_, _ = conn.Exec(
			ctx,
			`update collecte set partenaire_id = (
				insert into partenaire (nom, type_partenaire) values ($1, 'association')
				on conflict (nom) do update set nom = excluded.nom returning partenaire_id
			) where collecte_id = $2`,
			v, collecteID,
		)
	}

	if len(updates) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"message": "Collecte updated successfully"})
		return
	}

	args = append(args, collecteID)
	tag, err := conn.Exec(ctx, "update collecte set "+strings.Join(updates, ", ")+" where collecte_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update collecte")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Collecte not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Collecte updated successfully"})
}

func supprimerCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can cancel a collecte")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	collecteID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid collecte id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	// A cancelled collecte is kept for history, not hard-deleted.
	tag, err := conn.Exec(ctx, "update collecte set statut='annulee' where collecte_id=$1", collecteID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to cancel collecte")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Collecte not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Collecte cancelled successfully"})
}

func canManageCollecteDenrees(ctx context.Context, conn *pgx.Conn, token *IntrospectionPayload, collecteID int) bool {
	if isStaff(token) {
		return true
	}
	if !isBenevole(token) {
		return false
	}
	var count int
	_ = conn.QueryRow(
		ctx,
		`select count(*) from collecte_benevole_affecte
		inner join benevole on benevole.benevole_id = collecte_benevole_affecte.benevole_id
		inner join compte on compte.personne_id = benevole.personne_id
		where collecte_benevole_affecte.collecte_id=$1 and compte.compte_id=$2`,
		collecteID, token.CompteID,
	).Scan(&count)
	return count > 0
}

func canProposeCollecteDenrees(ctx context.Context, conn *pgx.Conn, token *IntrospectionPayload, collecteID int) bool {
	if !isCommercant(token) {
		return false
	}
	ownID, err := ownCommercantID(ctx, conn, token.CompteID)
	if err != nil {
		return false
	}
	var commercantID *int
	if err := conn.QueryRow(ctx, "select commercant_id from collecte where collecte_id=$1", collecteID).Scan(&commercantID); err != nil {
		return false
	}
	return commercantID != nil && *commercantID == ownID
}

func ensureStockProduitID(ctx context.Context, conn *pgx.Conn, nom string) (int, error) {
	var id int
	err := conn.QueryRow(ctx, "select stock_produit_id from stock_produit where lower(nom)=lower($1)", nom).Scan(&id)
	if err == nil {
		return id, nil
	}
	return id, conn.QueryRow(ctx, "insert into stock_produit (nom) values ($1) returning stock_produit_id", nom).Scan(&id)
}

func collecteDenreesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		ajouterDenreeCollecte(w, r, token)
	case http.MethodPatch:
		toggleDenreeConfirmee(w, r, token)
	case http.MethodDelete:
		supprimerDenreeCollecte(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func ajouterDenreeCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	collecteID, err := strconv.Atoi(r.FormValue("collecte_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "collecte_id must be a valid integer")
		return
	}
	stockProduitID, err := strconv.Atoi(r.FormValue("stock_produit_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "stock_produit_id must be a valid integer")
		return
	}
	quantite := r.FormValue("quantite")
	if quantite == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "quantite is required")
		return
	}
	nonPerissableBool := r.FormValue("non_perissable") == "1" || strings.EqualFold(r.FormValue("non_perissable"), "true")
	dlc := r.FormValue("dlc")
	if !nonPerissableBool && dlc == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "dlc is required unless non_perissable is set")
		return
	}
	nonPerissable := "0"
	if nonPerissableBool {
		nonPerissable = "1"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	canManage := canManageCollecteDenrees(ctx, conn, token, collecteID)
	canPropose := canProposeCollecteDenrees(ctx, conn, token, collecteID)
	if !canManage && !canPropose {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You are not allowed to add a denree to this collecte")
		return
	}

	proposeParType := "commercant"
	if canManage {
		proposeParType = "staff"
	}

	var newID int
	err = conn.QueryRow(
		ctx,
		`insert into collecte_denree (collecte_id, stock_produit_id, quantite, non_perissable, dlc, propose_par, propose_par_type)
		values ($1, $2, $3, $4, nullif($5,'')::date, $6, $7) returning collecte_denree_id`,
		collecteID, stockProduitID, quantite, nonPerissable, dlc, token.CompteID, proposeParType,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Failed to add the denree")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Denree added successfully", "collecte_denree_id": newID})
}

func toggleDenreeConfirmee(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	denreeID, err := strconv.Atoi(r.URL.Query().Get("id"))
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

	var collecteID int
	if err := conn.QueryRow(ctx, "select collecte_id from collecte_denree where collecte_denree_id=$1", denreeID).Scan(&collecteID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Denree not found")
		return
	}
	if !canManageCollecteDenrees(ctx, conn, token, collecteID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You are not allowed to confirm this denree")
		return
	}

	confirmee := "0"
	if r.FormValue("confirmee") == "1" || strings.EqualFold(r.FormValue("confirmee"), "true") {
		confirmee = "1"
	}
	if _, err := conn.Exec(ctx, "update collecte_denree set confirmee=$1 where collecte_denree_id=$2", confirmee, denreeID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the denree")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Denree updated successfully"})
}

func supprimerDenreeCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	denreeID, err := strconv.Atoi(r.URL.Query().Get("id"))
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

	var collecteID int
	if err := conn.QueryRow(ctx, "select collecte_id from collecte_denree where collecte_denree_id=$1", denreeID).Scan(&collecteID); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Denree not found")
		return
	}
	if !canManageCollecteDenrees(ctx, conn, token, collecteID) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "You are not allowed to remove this denree")
		return
	}

	if _, err := conn.Exec(ctx, "delete from collecte_denree where collecte_denree_id=$1", denreeID); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to remove the denree")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func confirmerCollecteHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can confirm a collecte")
		return
	}

	collecteID, err := strconv.Atoi(r.FormValue("collecte_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "collecte_id must be a valid integer")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	tag, err := conn.Exec(ctx, "update collecte set stock_mis_a_jour = '1' where collecte_id = $1 and stock_mis_a_jour = '0'", collecteID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to confirm the collecte")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Collecte not found, or already confirmed")
		return
	}

	_, err = conn.Exec(
		ctx,
		`insert into stock_produit (nom, quantite_disponible)
		select stock_produit.nom, collecte_denree.quantite from collecte_denree
		inner join stock_produit on stock_produit.stock_produit_id = collecte_denree.stock_produit_id
		where collecte_denree.collecte_id = $1 and collecte_denree.confirmee = '1'
		on conflict (nom) do update set quantite_disponible = stock_produit.quantite_disponible + excluded.quantite_disponible`,
		collecteID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update stock")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Collecte confirmed, stock updated"})
}

func collecteBenevolesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can affect benevoles to a collecte")
		return
	}
	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	collecteID, err := strconv.Atoi(r.FormValue("collecte_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "collecte_id must be a valid integer")
		return
	}
	benevoleCompteID, err := strconv.Atoi(r.FormValue("benevole_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "benevole_id must be a valid integer")
		return
	}
	roleMission := r.FormValue("role_mission")
	if roleMission == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "role_mission is required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	benevoleID, err := ensureBenevoleID(ctx, conn, benevoleCompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "benevole_id does not refer to a valid compte")
		return
	}

	_, err = conn.Exec(
		ctx,
		`insert into collecte_benevole_affecte (collecte_id, benevole_id, role_mission) values ($1, $2, $3)
		on conflict (collecte_id, benevole_id) do update set role_mission = excluded.role_mission`,
		collecteID, benevoleID, roleMission,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to affect the benevole")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Benevole affected successfully"})
}

// participationCollecteHandler implements "participer a la prochaine
// collecte": a benevole can join or leave a collecte.
func participationCollecteHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		participerCollecte(w, r, token)
	case http.MethodDelete:
		quitterCollecte(w, r, token)
	case http.MethodGet:
		listerParticipationsCollecte(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

// ensureBenevoleID resolves a compte_id to its benevole_id, creating the
// benevole row on first use so the API can keep speaking compte_id
// everywhere even though benevole is its own table in this schema.
func ensureBenevoleID(ctx context.Context, conn *pgx.Conn, compteID int) (int, error) {
	var personneID int
	if err := conn.QueryRow(ctx, "select personne_id from compte where compte_id=$1", compteID).Scan(&personneID); err != nil {
		return 0, err
	}
	var benevoleID int
	err := conn.QueryRow(ctx, "select benevole_id from benevole where personne_id=$1", personneID).Scan(&benevoleID)
	if err == nil {
		return benevoleID, nil
	}
	err = conn.QueryRow(ctx, "insert into benevole (personne_id, statut) values ($1, 'actif') returning benevole_id", personneID).Scan(&benevoleID)
	return benevoleID, err
}

func ownBenevoleID(ctx context.Context, conn *pgx.Conn, compteID int) (int, error) {
	return ensureBenevoleID(ctx, conn, compteID)
}

func participerCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isBenevole(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only benevoles can join a collecte")
		return
	}
	collecteIDParam := r.FormValue("collecte_id")
	if collecteIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required field: collecte_id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	benevoleID, err := ensureBenevoleID(ctx, conn, token.CompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No benevole profil found for this account")
		return
	}

	_, err = conn.Exec(
		ctx,
		"insert into participation_collecte (collecte_id, benevole_id, role_mission, commentaire) values ($1, $2, nullif($3,''), nullif($4,'')) on conflict (collecte_id, benevole_id) do nothing",
		collecteIDParam, benevoleID, r.FormValue("role_mission"), r.FormValue("commentaire"),
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to join the collecte")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func quitterCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	collecteIDParam := r.URL.Query().Get("collecte_id")
	if collecteIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing required field: collecte_id")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	benevoleID, err := ensureBenevoleID(ctx, conn, token.CompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No benevole profil found for this account")
		return
	}

	_, err = conn.Exec(ctx, "delete from participation_collecte where collecte_id=$1 and benevole_id=$2", collecteIDParam, benevoleID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to leave the collecte")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func listerParticipationsCollecte(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can list a collecte's participants")
		return
	}
	collecteID, err := strconv.Atoi(r.URL.Query().Get("collecte_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "collecte_id must be a valid integer")
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
		`select compte.compte_id, personne.nom, personne.prenom, participation_collecte.role_mission, participation_collecte.commentaire
		from participation_collecte
		inner join benevole on benevole.benevole_id = participation_collecte.benevole_id
		inner join personne on personne.personne_id = benevole.personne_id
		inner join compte on compte.personne_id = personne.personne_id
		where participation_collecte.collecte_id = $1`,
		collecteID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query participants")
		return
	}
	defer rows.Close()

	benevoles := make([]map[string]any, 0)
	for rows.Next() {
		var benevoleCompteID int
		var nom, prenom string
		var roleMission, commentaire *string
		if err := rows.Scan(&benevoleCompteID, &nom, &prenom, &roleMission, &commentaire); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan participants")
			return
		}
		benevoles = append(benevoles, map[string]any{
			"benevole_id": benevoleCompteID, "nom": nom, "prenom": prenom, "role_mission": roleMission, "commentaire": commentaire,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"collecte_id": collecteID, "total": len(benevoles), "benevoles": benevoles})
}
