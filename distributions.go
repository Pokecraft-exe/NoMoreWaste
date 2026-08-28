package main

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

func distributionsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerDistribution(w, r, token)
	case http.MethodGet:
		rechercherDistribution(w, r)
	case http.MethodPatch:
		modifierDistribution(w, r, token)
	case http.MethodDelete:
		annulerDistribution(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerDistribution(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can create a distribution")
		return
	}

	lieu := r.FormValue("lieu")
	dateDistribution := r.FormValue("date_distribution")
	if lieu == "" || dateDistribution == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "lieu and date_distribution are required")
		return
	}

	statut := r.FormValue("statut")
	if statut == "" {
		statut = "planifiee"
	}
	quota := r.FormValue("quota_par_adherent")
	if quota == "" {
		quota = "1"
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
		`insert into distribution (lieu, date_distribution, heure_distribution, statut, quota_par_adherent, created_by)
		values ($1, $2, nullif($3,''), $4, $5, $6) returning distribution_id`,
		lieu, dateDistribution, r.FormValue("heure_distribution"), statut, quota, token.CompteID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the distribution")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Distribution created successfully", "distribution_id": newID})
}

func rechercherDistribution(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	if idParam := r.URL.Query().Get("id"); idParam != "" {
		distributionID, convErr := strconv.Atoi(idParam)
		if convErr != nil {
			writeAPIError(w, http.StatusBadRequest, "invalid_request", "id must be a valid integer")
			return
		}
		writeDistributionDetail(w, r, conn, distributionID)
		return
	}

	from, size, err := parsePagination(r)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	query := "select distribution_id, lieu, date_distribution::text, heure_distribution::text, statut, quota_par_adherent from distribution"
	conditions := make([]string, 0)
	args := []any{}
	argPos := 1

	if r.URL.Query().Get("prochaine") == "1" {
		conditions = append(conditions, "date_distribution >= current_date and statut <> 'annulee'")
	}
	if statut := r.URL.Query().Get("statut"); statut != "" {
		conditions = append(conditions, "statut = $"+strconv.Itoa(argPos))
		args = append(args, statut)
		argPos++
	}
	if q := r.URL.Query().Get("q"); q != "" {
		conditions = append(conditions, "lieu ilike $"+strconv.Itoa(argPos))
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
	query += " order by date_distribution desc offset $" + strconv.Itoa(argPos) + " limit $" + strconv.Itoa(argPos+1)
	args = append(args, from, size)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query distributions")
		return
	}
	defer rows.Close()

	distributions := make([]map[string]any, 0)
	for rows.Next() {
		var id, quota int
		var lieu, dateDistribution, statut string
		var heureDistribution *string
		if err := rows.Scan(&id, &lieu, &dateDistribution, &heureDistribution, &statut, &quota); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan distributions")
			return
		}
		distributions = append(distributions, map[string]any{
			"distribution_id": id, "lieu": lieu, "date_distribution": dateDistribution,
			"heure_distribution": heureDistribution, "statut": statut, "quota_par_adherent": quota,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"from": from, "size": size, "total": len(distributions), "distributions": distributions})
}

func writeDistributionDetail(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, distributionID int) {
	ctx := r.Context()

	var lieu, dateDistribution, statut string
	var heureDistribution *string
	var quota int
	err := conn.QueryRow(
		ctx,
		"select lieu, date_distribution::text, heure_distribution::text, statut, quota_par_adherent from distribution where distribution_id = $1",
		distributionID,
	).Scan(&lieu, &dateDistribution, &heureDistribution, &statut, &quota)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Distribution not found")
		return
	}

	rows, err := conn.Query(
		ctx,
		`select distribution_denree.stock_produit_id, stock_produit.nom, stock_produit.unite, distribution_denree.quantite,
			coalesce((select sum(reservation.quantite) from reservation where reservation.distribution_id = $1 and reservation.stock_produit_id = distribution_denree.stock_produit_id), 0)
		from distribution_denree inner join stock_produit on stock_produit.stock_produit_id = distribution_denree.stock_produit_id
		where distribution_id = $1`,
		distributionID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query denrees")
		return
	}
	denrees := make([]map[string]any, 0)
	for rows.Next() {
		var stockProduitID int
		var nom, unite string
		var quantite, reserve float64
		if err := rows.Scan(&stockProduitID, &nom, &unite, &quantite, &reserve); err != nil {
			rows.Close()
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan denrees")
			return
		}
		denrees = append(denrees, map[string]any{
			"stock_produit_id": stockProduitID, "nom": nom, "unite": unite, "quantite": quantite, "restant": quantite - reserve,
		})
	}
	rows.Close()

	benevoles, err := conn.Query(
		ctx,
		`select compte.compte_id, personne.nom, personne.prenom, distribution_benevole_affecte.role_mission
		from distribution_benevole_affecte
		inner join benevole on benevole.benevole_id = distribution_benevole_affecte.benevole_id
		inner join personne on personne.personne_id = benevole.personne_id
		inner join compte on compte.personne_id = personne.personne_id
		where distribution_benevole_affecte.distribution_id = $1`,
		distributionID,
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
		"distribution_id":    distributionID,
		"lieu":               lieu,
		"date_distribution":  dateDistribution,
		"heure_distribution": heureDistribution,
		"statut":             statut,
		"quota_par_adherent": quota,
		"denrees":            denrees,
		"benevoles_affectes": affectations,
	})
}

func modifierDistribution(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can modify a distribution")
		return
	}

	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}
	distributionID, err := strconv.Atoi(idParam)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Invalid distribution id")
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
	if v := r.FormValue("date_distribution"); v != "" {
		addField("date_distribution", v)
	}
	if v := r.FormValue("heure_distribution"); v != "" {
		addField("heure_distribution", v)
	}
	if v := r.FormValue("statut"); v != "" {
		addField("statut", v)
	}
	if v := r.FormValue("quota_par_adherent"); v != "" {
		addField("quota_par_adherent", v)
	}

	if len(updates) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No fields to update")
		return
	}

	args = append(args, distributionID)
	tag, err := conn.Exec(ctx, "update distribution set "+strings.Join(updates, ", ")+" where distribution_id=$"+strconv.Itoa(argIndex), args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to update the distribution")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Distribution not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Distribution updated successfully"})
}

func annulerDistribution(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can cancel a distribution")
		return
	}

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

	tag, err := conn.Exec(ctx, "update distribution set statut='annulee' where distribution_id=$1", idParam)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to cancel the distribution")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Distribution not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Distribution cancelled successfully"})
}

func distributionDenreesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can allocate denrees to a distribution")
		return
	}
	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	distributionID, err := strconv.Atoi(r.FormValue("distribution_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "distribution_id must be a valid integer")
		return
	}
	stockProduitID, err := strconv.Atoi(r.FormValue("stock_produit_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "stock_produit_id must be a valid integer")
		return
	}
	quantite := r.FormValue("quantite")
	if quantite == "" {
		quantite = "0"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(
		ctx,
		`insert into distribution_denree (distribution_id, stock_produit_id, quantite) values ($1, $2, $3)
		on conflict (distribution_id, stock_produit_id) do update set quantite = excluded.quantite`,
		distributionID, stockProduitID, quantite,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to allocate the denree")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Denree allocated successfully"})
}

func distributionBenevolesHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can affect benevoles to a distribution")
		return
	}
	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
		return
	}

	distributionID, err := strconv.Atoi(r.FormValue("distribution_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "distribution_id must be a valid integer")
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
		`insert into distribution_benevole_affecte (distribution_id, benevole_id, role_mission) values ($1, $2, $3)
		on conflict (distribution_id, benevole_id) do update set role_mission = excluded.role_mission`,
		distributionID, benevoleID, roleMission,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to affect the benevole")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "Benevole affected successfully"})
}

func participationDistributionHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		participerDistribution(w, r, token)
	case http.MethodGet:
		listerParticipationsDistribution(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func participerDistribution(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isBenevole(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only benevoles can join a distribution")
		return
	}
	distributionID := r.FormValue("distribution_id")
	if distributionID == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "distribution_id is required")
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

	_, err = conn.Exec(ctx, "insert into distribution_benevole_participant (distribution_id, benevole_id) values ($1, $2) on conflict do nothing", distributionID, benevoleID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to join the distribution")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func listerParticipationsDistribution(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can list a distribution's participants")
		return
	}
	distributionID, err := strconv.Atoi(r.URL.Query().Get("distribution_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "distribution_id must be a valid integer")
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
		`select compte.compte_id, personne.nom, personne.prenom
		from distribution_benevole_participant
		inner join benevole on benevole.benevole_id = distribution_benevole_participant.benevole_id
		inner join personne on personne.personne_id = benevole.personne_id
		inner join compte on compte.personne_id = personne.personne_id
		where distribution_benevole_participant.distribution_id = $1`,
		distributionID,
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
		if err := rows.Scan(&benevoleCompteID, &nom, &prenom); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan participants")
			return
		}
		benevoles = append(benevoles, map[string]any{"benevole_id": benevoleCompteID, "nom": nom, "prenom": prenom})
	}

	writeJSON(w, http.StatusOK, map[string]any{"distribution_id": distributionID, "total": len(benevoles), "benevoles": benevoles})
}

func ensureAdherentID(ctx context.Context, conn *pgx.Conn, compteID int) (int, error) {
	var personneID int
	if err := conn.QueryRow(ctx, "select personne_id from compte where compte_id=$1", compteID).Scan(&personneID); err != nil {
		return 0, err
	}
	var adherentID int
	err := conn.QueryRow(ctx, "select adherent_id from adherent where personne_id=$1", personneID).Scan(&adherentID)
	if err == nil {
		return adherentID, nil
	}
	err = conn.QueryRow(ctx, "insert into adherent (personne_id) values ($1) returning adherent_id", personneID).Scan(&adherentID)
	return adherentID, err
}

func reservationHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}

	switch r.Method {
	case http.MethodPut:
		creerReservation(w, r, token)
	case http.MethodDelete:
		annulerReservation(w, r, token)
	case http.MethodGet:
		listerReservations(w, r, token)
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

func creerReservation(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isAdherent(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only adherents can reserve denrees")
		return
	}
	distributionID, err := strconv.Atoi(r.FormValue("distribution_id"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "distribution_id must be a valid integer")
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

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	adherentID, err := ensureAdherentID(ctx, conn, token.CompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No adherent profil found for this account")
		return
	}

	var quota int
	if err := conn.QueryRow(ctx, "select quota_par_adherent from distribution where distribution_id=$1", distributionID).Scan(&quota); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Distribution not found")
		return
	}

	var dejaReserve float64
	_ = conn.QueryRow(ctx, "select coalesce(sum(quantite),0) from reservation where distribution_id=$1 and adherent_id=$2", distributionID, adherentID).Scan(&dejaReserve)

	quantiteFloat, _ := strconv.ParseFloat(quantite, 64)
	if dejaReserve+quantiteFloat > float64(quota) {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "Quota depasse pour cette distribution")
		return
	}

	var newID int
	err = conn.QueryRow(
		ctx,
		"insert into reservation (distribution_id, adherent_id, stock_produit_id, quantite) values ($1, $2, $3, $4) returning reservation_id",
		distributionID, adherentID, stockProduitID, quantite,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to create the reservation")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"message": "Reservation created successfully", "reservation_id": newID})
}

func annulerReservation(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	reservationID, err := strconv.Atoi(r.URL.Query().Get("id"))
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

	adherentID, err := ensureAdherentID(ctx, conn, token.CompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No adherent profil found for this account")
		return
	}

	tag, err := conn.Exec(ctx, "delete from reservation where reservation_id=$1 and adherent_id=$2", reservationID, adherentID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to cancel the reservation")
		return
	}
	if tag.RowsAffected() == 0 {
		writeAPIError(w, http.StatusNotFound, "not_found", "Reservation not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func listerReservations(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	if !isAdherent(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only adherents can list their reservations")
		return
	}
	distributionID := r.URL.Query().Get("distribution_id")
	if distributionID == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "distribution_id is required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	adherentID, err := ensureAdherentID(ctx, conn, token.CompteID)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "No adherent profil found for this account")
		return
	}

	rows, err := conn.Query(
		ctx,
		`select reservation.reservation_id, reservation.stock_produit_id, stock_produit.nom, reservation.quantite
		from reservation inner join stock_produit on stock_produit.stock_produit_id = reservation.stock_produit_id
		where reservation.distribution_id = $1 and reservation.adherent_id = $2`,
		distributionID, adherentID,
	)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query reservations")
		return
	}
	defer rows.Close()

	reservations := make([]map[string]any, 0)
	for rows.Next() {
		var id, stockProduitID int
		var nom string
		var quantite float64
		if err := rows.Scan(&id, &stockProduitID, &nom, &quantite); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan reservations")
			return
		}
		reservations = append(reservations, map[string]any{"reservation_id": id, "stock_produit_id": stockProduitID, "nom": nom, "quantite": quantite})
	}

	writeJSON(w, http.StatusOK, map[string]any{"distribution_id": distributionID, "total": len(reservations), "reservations": reservations})
}
