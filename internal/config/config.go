package config

import (
	"fmt"
	"os"
)

// Correspondent defines a name/slug pair
type Correspondent struct {
	Name string
	Slug string
}

// WebhookRequest defines the incoming POST body
type WebhookRequest struct {
	Correspondents       []Correspondent `json:"correspondents"`
	CustomFieldIDAmount  int             `json:"custom_field_id_amount"`
	CustomFieldIDInvoice int             `json:"custom_field_id_invoice"`
	InvoiceOnly          bool            `json:"invoice_only"`
	ShowTotals           bool            `json:"show_totals"`
	ShowFooterRow        bool            `json:"show_footer_row"`
	AllCorrespondents    bool            `json:"all_correspondents"`
	CombinedPDF          bool            `json:"combined_pdf"`
}

const (
	FontRegular = "fonts/DejaVuSans.ttf"
	FontBold    = "fonts/DejaVuSans-Bold.ttf"
	ReportsDir  = "reports"
	WebhookPort = ":8080"
)

// BaseURL returns the Paperless-NGX base URL from environment
func BaseURL() string {
	url := os.Getenv("PAPERLESS_URL")
	if url == "" {
		panic("PAPERLESS_URL environment variable is not set")
	}
	return url
}

// APIToken returns the Paperless-NGX API token from environment
func APIToken() string {
	token := os.Getenv("PAPERLESS_TOKEN")
	if token == "" {
		panic("PAPERLESS_TOKEN environment variable is not set")
	}
	return token
}

// OutputPDF returns the full path for the PDF report of a correspondent
func OutputPDF(correspondentName string) string {
	return fmt.Sprintf("%s/Paperless_Correspondent_%s.pdf", ReportsDir, correspondentName)
}

// CombinedOutputPDF returns the full path for the combined PDF report
func CombinedOutputPDF() string {
	return fmt.Sprintf("%s/Paperless_AllCorrespondents.pdf", ReportsDir)
}

// EnsureReportsDir creates the reports directory if it does not exist
func EnsureReportsDir() error {
	return os.MkdirAll(ReportsDir, 0755)
}
