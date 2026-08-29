package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// documentsHandler implements "Chaque livraison donnera lieu à l'émission
// d'un récapitulatif au format PDF": staff generates a one-page summary PDF
// for a collecte (or a tournee / planning_service), stored on disk and
// tracked in document_genere.
func documentsHandler(w http.ResponseWriter, r *http.Request) {
	token := tryAuth(w, r)
	if !token.Active {
		return
	}
	if !isStaff(token) {
		writeAPIError(w, http.StatusForbidden, "access_denied", "Only staff can generate documents")
		return
	}

	switch r.Method {
	case http.MethodPut:
		genererRecapitulatif(w, r, token)
	case http.MethodGet:
		if r.URL.Query().Get("id") != "" {
			telechargerDocument(w, r)
		} else {
			listerDocuments(w, r)
		}
	default:
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "The URI does not support the requested method.")
	}
}

// telechargerDocument streams a previously generated PDF back to the caller
// (Content-Disposition: attachment) so the back-office can offer an actual
// "Telecharger le recapitulatif PDF" button/link.
func telechargerDocument(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	documentID, err := strconv.Atoi(idParam)
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

	var nomFichier, cheminFichier string
	if err := conn.QueryRow(ctx, "select nom_fichier, chemin_fichier from document_genere where document_genere_id=$1", documentID).Scan(&nomFichier, &cheminFichier); err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "Document not found")
		return
	}

	// chemin_fichier is stored relative to DATA_DIR (see genererRecapitulatif);
	// filepath.Base guards against it ever containing a path traversal.
	data, err := os.ReadFile(filepath.Join(DATA_DIR, filepath.Base(cheminFichier)))
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "not_found", "The PDF file is no longer available")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+nomFichier+"\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func genererRecapitulatif(w http.ResponseWriter, r *http.Request, token *IntrospectionPayload) {
	collecteIDParam := r.FormValue("collecte_id")
	tourneeIDParam := r.FormValue("tournee_id")
	planningServiceIDParam := r.FormValue("planning_service_id")
	if collecteIDParam == "" && tourneeIDParam == "" && planningServiceIDParam == "" {
		writeAPIError(w, http.StatusBadRequest, "invalid_request", "One of collecte_id, tournee_id or planning_service_id is required")
		return
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	lines := []string{"NO MORE WASTE - Recapitulatif", "Genere le " + time.Now().Format("2006-01-02 15:04")}

	if collecteIDParam != "" {
		var lieu, dateCollecte, statut string
		if err := conn.QueryRow(ctx, "select lieu, date_collecte, statut from collecte where collecte_id=$1", collecteIDParam).Scan(&lieu, &dateCollecte, &statut); err != nil {
			writeAPIError(w, http.StatusNotFound, "not_found", "Collecte not found")
			return
		}
		lines = append(lines, "Collecte #"+collecteIDParam, "Lieu: "+lieu, "Date: "+dateCollecte, "Statut: "+statut)

		rows, err := conn.Query(ctx, "select barcode, nom, quantite, poids_kg from produit where collecte_id=$1 order by produit_id", collecteIDParam)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var barcode, nom string
				var quantite int
				var poids *float64
				if rows.Scan(&barcode, &nom, &quantite, &poids) == nil {
					line := fmt.Sprintf("- %s (%s) x%d", nom, barcode, quantite)
					if poids != nil {
						line += fmt.Sprintf(" %.2fkg", *poids)
					}
					lines = append(lines, line)
				}
			}
		}
	}

	pdfBytes := generateSimplePDF(lines)

	if err := os.MkdirAll(DATA_DIR, 0o755); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to prepare the documents directory")
		return
	}

	fileName := fmt.Sprintf("recapitulatif_%d.pdf", time.Now().UnixNano())
	if err := os.WriteFile(filepath.Join(DATA_DIR, fileName), pdfBytes, 0o644); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to write the PDF file")
		return
	}

	var newID int
	err = conn.QueryRow(
		ctx,
		`insert into document_genere (collecte_id, tournee_id, planning_service_id, type_document, nom_fichier, chemin_fichier, genere_par)
		values (nullif($1,'')::int, nullif($2,'')::int, nullif($3,'')::int, 'pdf', $4, $5, $6)
		returning document_genere_id`,
		collecteIDParam, tourneeIDParam, planningServiceIDParam, fileName, fileName, token.CompteID,
	).Scan(&newID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to record the generated document")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":            "Document generated successfully",
		"document_genere_id": newID,
		"nom_fichier":        fileName,
	})
}

func listerDocuments(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, DATABASE_PUBLIC_URL)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to connect to the database")
		return
	}
	defer conn.Close(ctx)

	query := "select document_genere_id, collecte_id, tournee_id, planning_service_id, type_document, nom_fichier, date_generation from document_genere"
	args := []any{}
	if collecteID := r.URL.Query().Get("collecte_id"); collecteID != "" {
		query += " where collecte_id = $1"
		args = append(args, collecteID)
	}
	query += " order by date_generation desc"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to query documents")
		return
	}
	defer rows.Close()

	documents := make([]map[string]any, 0)
	for rows.Next() {
		var id int
		var collecteID, tourneeID, planningServiceID *int
		var typeDocument, nomFichier, dateGeneration string
		if err := rows.Scan(&id, &collecteID, &tourneeID, &planningServiceID, &typeDocument, &nomFichier, &dateGeneration); err != nil {
			writeAPIError(w, http.StatusInternalServerError, "internal_server_error", "Failed to scan documents")
			return
		}
		documents = append(documents, map[string]any{
			"document_genere_id":  id,
			"collecte_id":         collecteID,
			"tournee_id":          tourneeID,
			"planning_service_id": planningServiceID,
			"type_document":       typeDocument,
			"nom_fichier":         nomFichier,
			"date_generation":     dateGeneration,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"total": len(documents), "documents": documents})
}

// generateSimplePDF writes a minimal, dependency-free single-page PDF
// (Helvetica, one line of text per lines[i]) with a correct xref table.
func generateSimplePDF(lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT /F1 12 Tf 72 740 Td 16 TL\n")
	for _, line := range lines {
		content.WriteString("(" + pdfEscape(line) + ") Tj T*\n")
	}
	content.WriteString("ET")

	objects := []string{
		"<</Type/Catalog/Pages 2 0 R>>",
		"<</Type/Pages/Kids[3 0 R]/Count 1>>",
		"<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Resources<</Font<</F1 4 0 R>>>>/Contents 5 0 R>>",
		"<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>",
		"<</Length " + strconv.Itoa(content.Len()) + ">>\nstream\n" + content.String() + "\nendstream",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		buf.WriteString(strconv.Itoa(i+1) + " 0 obj\n" + obj + "\nendobj\n")
	}

	xrefStart := buf.Len()
	buf.WriteString("xref\n0 " + strconv.Itoa(len(objects)+1) + "\n0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString("trailer\n<</Size " + strconv.Itoa(len(objects)+1) + "/Root 1 0 R>>\nstartxref\n" + strconv.Itoa(xrefStart) + "\n%%EOF")

	return buf.Bytes()
}

func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}
