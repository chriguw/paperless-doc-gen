package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"dochandler/internal/api"
	"dochandler/internal/config"
	"dochandler/internal/model"
	"dochandler/internal/report"
)

func getYear(created string) string {
	if len(created) >= 4 {
		return created[:4]
	}
	return "????"
}

func isRechnung(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "Rechnung")
}

func getCustomFieldValue(doc model.Document, fieldID int) string {
	for _, cf := range doc.CustomFields {
		if cf.Field == fieldID {
			if cf.Value == nil || *cf.Value == "" {
				return ""
			}
			return *cf.Value
		}
	}
	return ""
}

func buildYearMap(req config.WebhookRequest, docs []model.Document, docTypes map[int]string) ([]string, map[string][]model.DocEntry, float64, int, int) {
	yearMap := make(map[string][]model.DocEntry)

	for _, doc := range docs {
		docTypeName := "-"
		if doc.DocumentType != nil {
			if name, ok := docTypes[*doc.DocumentType]; ok {
				docTypeName = name
			}
		}

		if req.InvoiceOnly && !isRechnung(docTypeName) {
			continue
		}

		year := getYear(doc.Created)
		entry := model.DocEntry{
			Doc:           doc,
			Year:          year,
			Date:          doc.Created[:10],
			Amount:        getCustomFieldValue(doc, req.CustomFieldIDAmount),
			InvoiceNumber: getCustomFieldValue(doc, req.CustomFieldIDInvoice),
			DocTypeName:   docTypeName,
		}
		yearMap[year] = append(yearMap[year], entry)
	}

	years := make([]string, 0, len(yearMap))
	for y := range yearMap {
		years = append(years, y)
	}
	sort.Strings(years)

	totalAll := 0.0
	missingAll := 0
	totalDocs := 0

	for _, year := range years {
		entries := yearMap[year]
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Date < entries[j].Date
		})
		yearMap[year] = entries

		for _, e := range entries {
			if e.Amount != "" {
				if val, ok := report.ParseAmount(e.Amount); ok {
					totalAll += val
				}
			} else {
				missingAll++
			}
		}
		totalDocs += len(entries)
	}

	return years, yearMap, totalAll, totalDocs, missingAll
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req config.WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	baseURL := config.BaseURL()
	apiToken := config.APIToken()

	log.Printf("📥 Webhook received — AllCorrespondents: %v  CombinedPDF: %v  InvoiceOnly: %v\n",
		req.AllCorrespondents, req.CombinedPDF, req.InvoiceOnly)

	// Ensure reports directory exists
	if err := config.EnsureReportsDir(); err != nil {
		http.Error(w, fmt.Sprintf("reports dir error: %v", err), http.StatusInternalServerError)
		return
	}

	// Load document types
	docTypes, err := api.FetchDocumentTypes(baseURL, apiToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching document types: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("📂 %d document types loaded\n", len(docTypes))

	// Build correspondent list
	var correspondents []config.Correspondent
	if req.AllCorrespondents {
		all, err := api.FetchAllCorrespondents(baseURL, apiToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("error fetching correspondents: %v", err), http.StatusInternalServerError)
			return
		}
		for _, c := range all {
			correspondents = append(correspondents, config.Correspondent{
				Name: c.Name,
				Slug: c.Slug,
			})
		}
		log.Printf("📋 %d correspondents fetched from Paperless\n", len(correspondents))
	} else {
		correspondents = req.Correspondents
		log.Printf("📋 %d correspondents from request\n", len(correspondents))
	}

	// Create combined PDF writer if needed
	var combinedWriter *report.CombinedPDFWriter
	if req.CombinedPDF {
		combinedWriter, err = report.NewCombinedPDFWriter(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating combined PDF: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Process each correspondent
	type result struct {
		Name    string  `json:"name"`
		Docs    int     `json:"docs"`
		Missing int     `json:"missing"`
		Amount  float64 `json:"amount"`
		PDF     string  `json:"pdf,omitempty"`
		Error   string  `json:"error,omitempty"`
	}
	var results []result
	var errors []string

	for _, correspondent := range correspondents {
		log.Printf("🔄 Processing: %s\n", correspondent.Name)

		corrID, err := api.FindCorrespondentID(baseURL, apiToken, correspondent.Slug)
		if err != nil {
			msg := fmt.Sprintf("%s: %v", correspondent.Name, err)
			errors = append(errors, msg)
			results = append(results, result{Name: correspondent.Name, Error: msg})
			continue
		}

		docs, err := api.FetchAllDocuments(baseURL, apiToken, corrID)
		if err != nil {
			msg := fmt.Sprintf("%s: %v", correspondent.Name, err)
			errors = append(errors, msg)
			results = append(results, result{Name: correspondent.Name, Error: msg})
			continue
		}

		years, yearMap, totalAll, totalDocs, missingAll := buildYearMap(req, docs, docTypes)

		if req.CombinedPDF {
			combinedWriter.AddCorrespondent(correspondent.Name, years, yearMap, totalAll, totalDocs, missingAll)
			results = append(results, result{
				Name:    correspondent.Name,
				Docs:    totalDocs,
				Missing: missingAll,
				Amount:  totalAll,
			})
		} else {
			outputPDF := config.OutputPDF(correspondent.Name)
			if err := report.GeneratePDF(req, correspondent.Name, outputPDF, years, yearMap, totalAll, totalDocs, missingAll); err != nil {
				msg := fmt.Sprintf("%s: %v", correspondent.Name, err)
				errors = append(errors, msg)
				results = append(results, result{Name: correspondent.Name, Error: msg})
				continue
			}
			results = append(results, result{
				Name:    correspondent.Name,
				Docs:    totalDocs,
				Missing: missingAll,
				Amount:  totalAll,
				PDF:     outputPDF,
			})
			log.Printf("✅ PDF saved: %s\n", outputPDF)
		}
	}

	// Save combined PDF
	if req.CombinedPDF {
		if err := combinedWriter.Save(); err != nil {
			http.Error(w, fmt.Sprintf("error saving combined PDF: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Combined PDF saved: %s\n", config.CombinedOutputPDF())
	}

	// Return JSON response
	type response struct {
		Status  string   `json:"status"`
		PDF     string   `json:"pdf,omitempty"`
		Results []result `json:"results"`
		Errors  []string `json:"errors,omitempty"`
	}

	resp := response{
		Status:  "ok",
		Results: results,
		Errors:  errors,
	}
	if req.CombinedPDF {
		resp.PDF = config.CombinedOutputPDF()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// Validate env vars on startup
	_ = config.BaseURL()
	_ = config.APIToken()

	http.HandleFunc("/webhook", handleWebhook)

	log.Printf("🚀 dochandler webhook listening on %s\n", config.WebhookPort)
	log.Printf("📡 POST http://localhost%s/webhook\n", config.WebhookPort)
	log.Fatal(http.ListenAndServe(config.WebhookPort, nil))
}
