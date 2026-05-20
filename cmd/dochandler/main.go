package main

import (
	"dochandler/internal/api"
	"dochandler/internal/config"
	"dochandler/internal/model"
	"dochandler/internal/report"
	"fmt"
	"log"
	"sort"
	"strings"
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

func processCorrespondent(correspondent config.Correspondent, docTypes map[int]string) error {
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Korrespondent: %-45s ║\n", correspondent.Name)
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	corrID, err := api.FindCorrespondentID(correspondent.Slug)
	if err != nil {
		return fmt.Errorf("Korrespondent '%s' nicht gefunden: %w", correspondent.Name, err)
	}
	fmt.Printf("Korrespondent gefunden: ID %d\n\n", corrID)

	fmt.Println("Lade Dokumente...")
	docs, err := api.FetchAllDocuments(corrID)
	if err != nil {
		return fmt.Errorf("Fehler beim Laden der Dokumente: %w", err)
	}
	fmt.Printf("%d Dokumente gefunden.\n\n", len(docs))

	if config.InvoiceOnly {
		fmt.Println("⚠️  InvoiceOnly = true → nur Dokumente vom Typ 'Rechnung' werden angezeigt.\n")
	}

	// --- Dokumente nach Jahr gruppieren ---
	yearMap := make(map[string][]model.DocEntry)
	for _, doc := range docs {
		docTypeName := "-"
		if doc.DocumentType != nil {
			if name, ok := docTypes[*doc.DocumentType]; ok {
				docTypeName = name
			}
		}

		if config.InvoiceOnly && !isRechnung(docTypeName) {
			continue
		}

		year := getYear(doc.Created)
		entry := model.DocEntry{
			Doc:           doc,
			Year:          year,
			Date:          doc.Created[:10],
			Amount:        getCustomFieldValue(doc, config.CustomFieldIDAmount),
			InvoiceNumber: getCustomFieldValue(doc, config.CustomFieldIDInvoice),
			DocTypeName:   docTypeName,
		}
		yearMap[year] = append(yearMap[year], entry)
	}

	// --- Jahre sortieren ---
	years := make([]string, 0, len(yearMap))
	for y := range yearMap {
		years = append(years, y)
	}
	sort.Strings(years)

	// --- Totals berechnen & Einträge sortieren ---
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

	// --- Terminal Ausgabe ---
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%-6s  %-12s  %-16s  %-14s  %-16s  %s\n", "ID", "Datum", "Rechnungs-Nr.", "Betrag", "Dokumenttyp", "Titel")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for _, year := range years {
		fmt.Printf("\n── %s ──\n", year)
		for _, e := range yearMap[year] {
			inv := e.InvoiceNumber
			if inv == "" {
				inv = "-"
			}
			amt := "-"
			if e.Amount != "" {
				if val, ok := report.ParseAmount(e.Amount); ok {
					amt = fmt.Sprintf("CHF %8.2f", val)
				}
			}
			fmt.Printf("%-6d  %-12s  %-16s  %-14s  %-16s  %s\n", e.Doc.ID, e.Date, inv, amt, e.DocTypeName, e.Doc.Title)
		}
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Gesamttotal: CHF %.2f  |  %d Dokument(e)  |  %d ohne Betrag\n", totalAll, totalDocs, missingAll)

	// --- PDF generieren ---
	outputPDF := config.OutputPDF(correspondent.Name)
	fmt.Printf("\nGeneriere PDF '%s'...\n", outputPDF)
	if err := report.GeneratePDF(correspondent.Name, outputPDF, years, yearMap, totalAll, totalDocs, missingAll); err != nil {
		return fmt.Errorf("Fehler beim PDF generieren: %w", err)
	}
	fmt.Printf("✅ PDF erfolgreich gespeichert: %s\n", outputPDF)

	return nil
}

func main() {
	// Ensure reports directory exists
	if err := config.EnsureReportsDir(); err != nil {
		log.Fatalf("Fehler beim Erstellen des Reports-Ordners: %v", err)
	}
	fmt.Printf("📁 Reports werden gespeichert in: %s/\n\n", config.ReportsDir)

	// Load document types once for all correspondents
	fmt.Println("Lade Dokumenttypen...")
	docTypes, err := api.FetchDocumentTypes()
	if err != nil {
		log.Fatalf("Fehler beim Laden der Dokumenttypen: %v", err)
	}
	fmt.Printf("%d Dokumenttypen geladen.\n\n", len(docTypes))

	// Process each correspondent
	errors := []string{}
	for _, correspondent := range config.Correspondents {
		if err := processCorrespondent(correspondent, docTypes); err != nil {
			fmt.Printf("⚠️  Fehler bei '%s': %v\n", correspondent.Name, err)
			errors = append(errors, fmt.Sprintf("%s: %v", correspondent.Name, err))
		}
	}

	// Summary
	fmt.Printf("\n╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Zusammenfassung                                             ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  %d Korrespondent(en) verarbeitet                             \n", len(config.Correspondents))
	fmt.Printf("║  %d Fehler                                                    \n", len(errors))
	for _, e := range errors {
		fmt.Printf("║  ⚠️  %s\n", e)
	}
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")
}
