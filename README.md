# 📄 dochandler

A Go webhook service that fetches documents from a [Paperless-NGX](https://docs.paperless-ngx.com/) instance,
groups them by year and generates nicely formatted PDF reports per correspondent.

## ✨ Features

- 🌐 Runs as a **webhook HTTP listener** — triggered by a simple POST request
- 📋 Fetches documents for configured correspondents via Paperless-NGX REST API
- 🗂️ Groups documents by year with subtotals
- 💰 Parses and sums CHF amounts from custom fields
- 📄 Generates PDF reports including:
  - Clickable document ID links back to Paperless-NGX
  - Year sections with totals
  - Grand total row
  - Missing amount highlighting in red
  - Alternating row colors
  - Optional combined PDF with all correspondents + grand summary page
- 🖥️ Returns a JSON response with processing results
- 🔧 Fully configurable via POST body parameters

---

## 📁 Project Structure

```
dochandler/
├── cmd/
│   └── dochandler/
│       └── main.go                  # Webhook listener & entry point
├── internal/
│   ├── config/
│   │   └── config.go                # Configuration, env vars & request structs
│   ├── api/
│   │   ├── client.go                # HTTP helper for Paperless-NGX API
│   │   ├── correspondents.go        # Correspondent API calls
│   │   ├── documents.go             # Document API calls
│   │   └── doctypes.go              # Document type API calls
│   ├── model/
│   │   └── model.go                 # Data structures
│   └── report/
│       └── pdf.go                   # PDF generation
├── fonts/
│   ├── DejaVuSans.ttf               # Regular font (not committed to git)
│   └── DejaVuSans-Bold.ttf          # Bold font (not committed to git)
├── reports/                         # Generated PDFs (not committed to git)
├── .gitignore
├── go.mod
└── README.md
```

---

## 🔧 Prerequisites

- [Go](https://golang.org/dl/) 1.21 or higher
- A running [Paperless-NGX](https://docs.paperless-ngx.com/) instance
- DejaVu fonts (see [Installation](#-installation))

---

## 🚀 Installation

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/dochandler.git
cd dochandler
```

### 2. Download the fonts

```bash
mkdir -p fonts

# Option A: Download from SourceForge
curl -L -o dejavu.zip "https://sourceforge.net/projects/dejavu/files/dejavu/2.37/dejavu-fonts-ttf-2.37.zip/download"
unzip dejavu.zip "dejavu-fonts-ttf-2.37/ttf/DejaVuSans.ttf" "dejavu-fonts-ttf-2.37/ttf/DejaVuSans-Bold.ttf"
mv dejavu-fonts-ttf-2.37/ttf/DejaVuSans.ttf fonts/
mv dejavu-fonts-ttf-2.37/ttf/DejaVuSans-Bold.ttf fonts/
rm -rf dejavu-fonts-ttf-2.37 dejavu.zip

# Option B: Copy from system fonts (macOS)
cp "/System/Library/Fonts/Supplemental/Arial.ttf" fonts/DejaVuSans.ttf
cp "/System/Library/Fonts/Supplemental/Arial Bold.ttf" fonts/DejaVuSans-Bold.ttf

# Option C: Copy from system fonts (Linux)
cp /usr/share/fonts/truetype/dejavu/DejaVuSans.ttf fonts/
cp /usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf fonts/
```

### 3. Install dependencies

```bash
go mod tidy
```

---

## ⚙️ Configuration

### Environment Variables

The following environment variables **must** be set before starting the service:

| Variable | Description | Example |
|---|---|---|
| `PAPERLESS_URL` | Base URL of your Paperless-NGX instance | `http://192.168.1.147:8001` |
| `PAPERLESS_TOKEN` | API token for authentication | `abc123xyz...` |

#### Set environment variables

```bash
# Linux / macOS
export PAPERLESS_URL="http://192.168.1.147:8001"
export PAPERLESS_TOKEN="your-api-token-here"

# Windows (PowerShell)
$env:PAPERLESS_URL="http://192.168.1.147:8001"
$env:PAPERLESS_TOKEN="your-api-token-here"
```

#### Using a .env file (optional)

Create a `.env` file in the project root (already in `.gitignore`):

```env
PAPERLESS_URL=http://192.168.1.147:8001
PAPERLESS_TOKEN=your-api-token-here
```

Then load it before running:

```bash
export $(cat .env | xargs) && go run cmd/dochandler/main.go
```

### Getting your Paperless-NGX API Token

1. Log in to your Paperless-NGX instance
2. Navigate to **Settings → API Token**
3. Copy the token and set it as `PAPERLESS_TOKEN`

### Finding Custom Field IDs

```bash
curl -H "Authorization: Token YOUR-API-TOKEN" \
     http://YOUR-PAPERLESS-URL:PORT/api/custom_fields/
```

---

## ▶️ Running the Service

Always run from the **project root** so that the `fonts/` path resolves correctly:

```bash
go run cmd/dochandler/main.go
```

Or build and run:

```bash
go build -o bin/dochandler cmd/dochandler/main.go
./bin/dochandler
```

The service starts and listens on port **8080**:

```
🚀 dochandler webhook listening on :8080
📡 POST http://localhost:8080/webhook
```

---

## 📡 Webhook API

### Endpoint

```
POST http://localhost:8080/webhook
Content-Type: application/json
```

### Request Body Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `correspondents` | `array` | No* | List of correspondents to process. Only used when `all_correspondents` is `false` |
| `correspondents[].name` | `string` | No* | Display name of the correspondent (e.g. `"Gemeindeverwaltung"`) |
| `correspondents[].slug` | `string` | No* | URL slug of the correspondent in Paperless-NGX (e.g. `"gemeindeverwaltung"`) |
| `custom_field_id_amount` | `int` | Yes | ID of the custom field containing the invoice amount (e.g. `2`) |
| `custom_field_id_invoice` | `int` | Yes | ID of the custom field containing the invoice number (e.g. `1`) |
| `invoice_only` | `bool` | Yes | If `true`, only documents with type `Rechnung` are included in the report |
| `show_totals` | `bool` | Yes | If `true`, year subtotal rows and grand total row are shown in the PDF |
| `show_footer_row` | `bool` | Yes | If `true`, a summary footer is shown at the bottom of each correspondent section with total documents, missing amounts and total CHF |
| `all_correspondents` | `bool` | Yes | If `true`, all correspondents are fetched directly from Paperless-NGX and the `correspondents` array is ignored |
| `combined_pdf` | `bool` | Yes | If `true`, all correspondents are combined into a single PDF with each correspondent starting on a new page and a grand summary page at the end. If `false`, one PDF per correspondent is generated |

> *Required only when `all_correspondents` is `false`

---

### Example Requests

#### Generate a combined PDF for all correspondents in Paperless-NGX

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "custom_field_id_amount": 2,
    "custom_field_id_invoice": 1,
    "invoice_only": false,
    "show_totals": true,
    "show_footer_row": true,
    "all_correspondents": true,
    "combined_pdf": true
  }'
```

#### Generate individual PDFs for specific correspondents only

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "correspondents": [
      {"name": "Gemeindeverwaltung", "slug": "gemeindeverwaltung"},
      {"name": "Versicherung", "slug": "versicherung"},
      {"name": "Garage", "slug": "garage"},
      {"name": "Baumarkt", "slug": "baumarkt"}
    ],
    "custom_field_id_amount": 2,
    "custom_field_id_invoice": 1,
    "invoice_only": false,
    "show_totals": true,
    "show_footer_row": true,
    "all_correspondents": false,
    "combined_pdf": false
  }'
```

#### Invoices only, no totals, combined PDF

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "custom_field_id_amount": 2,
    "custom_field_id_invoice": 1,
    "invoice_only": true,
    "show_totals": false,
    "show_footer_row": true,
    "all_correspondents": true,
    "combined_pdf": true
  }'
```

---

### Response

```json
{
  "status": "ok",
  "pdf": "reports/Übersicht_AllCorrespondents.pdf",
  "results": [
    {
      "name": "Gemeindeverwaltung",
      "docs": 12,
      "missing": 1,
      "amount": 20.30,
      "pdf": ""
    },
    {
      "name": "Versicherung",
      "docs": 8,
      "missing": 0,
      "amount": 150.00,
      "pdf": ""
    },
    {
      "name": "Garage",
      "docs": 4,
      "missing": 2,
      "amount": 178.50,
      "pdf": ""
    }
  ],
  "errors": []
}
```

#### Response Fields

| Field | Type | Description |
|---|---|---|
| `status` | `string` | `"ok"` if the request was processed successfully |
| `pdf` | `string` | Path to the combined PDF. Only set when `combined_pdf` is `true` |
| `results` | `array` | One entry per processed correspondent |
| `results[].name` | `string` | Correspondent name |
| `results[].docs` | `int` | Total number of documents processed |
| `results[].missing` | `int` | Number of documents with missing amount |
| `results[].amount` | `float` | Total CHF amount across all documents |
| `results[].pdf` | `string` | Path to individual PDF. Only set when `combined_pdf` is `false` |
| `results[].error` | `string` | Error message if processing failed for this correspondent |
| `errors` | `array` | List of all errors across all correspondents |

---

## 📄 PDF Report Layout

### Per Correspondent Section

| Element | Description |
|---|---|
| **Title** | `Übersicht <CorrespondentName>` |
| **Subtitle** | Filter mode, document count, missing count |
| **Year header** | Light blue bar with the year |
| **Column headers** | ID, Date, Invoice No., Type, Amount, Title |
| **ID column** | Clickable link — opens document directly in Paperless-NGX |
| **Alternating rows** | White and light blue striping |
| **Missing amount rows** | Highlighted in light red |
| **Year total row** | Shown when `show_totals` is `true` |
| **Grand total row** | Shown when `show_totals` is `true` |
| **Footer summary** | Shown when `show_footer_row` is `true` — total docs, missing, total CHF |

### Combined PDF Summary Page

A final summary page is appended at the end when `combined_pdf` is `true`:

| Column | Description |
|---|---|
| **Korrespondent** | Correspondent name |
| **Dokumente** | Total document count |
| **Fehlend** | Missing amount count (shown in red if > 0) |
| **Betrag** | Total CHF amount |
| **Total row** | Grand totals across all correspondents |

### Output Files

```
reports/
├── Übersicht_AllCorrespondents.pdf    # when combined_pdf = true
├── Übersicht_Gemeindeverwaltung.pdf   # when combined_pdf = false
├── Übersicht_Versicherung.pdf         # when combined_pdf = false
└── Übersicht_Garage.pdf               # when combined_pdf = false
```

---

## 🛠️ Built With

- [Go](https://golang.org/) — Programming language
- [gofpdf](https://github.com/jung-kurt/gofpdf) — PDF generation
- [Paperless-NGX REST API](https://docs.paperless-ngx.com/api/) — Document source
- [DejaVu Fonts](https://dejavu-fonts.github.io/) — UTF-8 font with full umlaut support

---

## 📝 License

MIT License — feel free to use and adapt for your own Paperless-NGX setup.
