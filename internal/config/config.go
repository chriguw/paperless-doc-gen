package config

import (
	"fmt"
	"os"
)

type Correspondent struct {
	Name string
	Slug string
}

// Correspondents is the list of all correspondents to process
var Correspondents = []Correspondent{
	{Name: "Gemeindeverwaltung", Slug: "gemeindeverwaltung"},
	{Name: "AXA", Slug: "axa"},
	{Name: "Lättigarage", Slug: "lattigarage"},
	{Name: "Steuerverwaltung", Slug: "steuerverwaltung"},
	{Name: "Krankenkasse", Slug: "krankenkasse"},
	// add more here...
}

const (
	BaseURL              = "http://192.168.1.147:8001"                // Deine Paperless-NGX URL
	APIToken             = "47f183c4fb99f65851521bd9035c2c3cc23f6526" // Dein API Token
	CustomFieldIDAmount  = 2
	CustomFieldIDInvoice = 1
	InvoiceOnly          = true
	FontRegular          = "fonts/DejaVuSans.ttf"
	FontBold             = "fonts/DejaVuSans-Bold.ttf"
	ReportsDir           = "reports"
)

func OutputPDF(correspondentName string) string {
	return fmt.Sprintf("%s/Übersicht_%s.pdf", ReportsDir, correspondentName)
}

// EnsureReportsDir creates the reports directory if it does not exist
func EnsureReportsDir() error {
	return os.MkdirAll(ReportsDir, 0755)
}
