package model

// --- Correspondent ---

type CorrespondentList struct {
	Count   int             `json:"count"`
	Next    *string         `json:"next"`
	Results []Correspondent `json:"results"`
}

type Correspondent struct {
	ID                int    `json:"id"`
	Slug              string `json:"slug"`
	Name              string `json:"name"`
	Match             string `json:"match"`
	MatchingAlgorithm int    `json:"matching_algorithm"`
	IsInsensitive     bool   `json:"is_insensitive"`
	DocumentCount     int    `json:"document_count"`
	Owner             int    `json:"owner"`
	UserCanChange     bool   `json:"user_can_change"`
}

// --- Document ---

type DocumentList struct {
	Count   int        `json:"count"`
	Next    *string    `json:"next"`
	Results []Document `json:"results"`
}

type Document struct {
	ID           int           `json:"id"`
	Title        string        `json:"title"`
	Created      string        `json:"created"`
	DocumentType *int          `json:"document_type"`
	CustomFields []CustomField `json:"custom_fields"`
}

type CustomField struct {
	Field int     `json:"field"`
	Value *string `json:"value"`
}

// --- DocumentType ---

type DocumentTypeList struct {
	Count   int            `json:"count"`
	Next    *string        `json:"next"`
	Results []DocumentType `json:"results"`
}

type DocumentType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// --- DocEntry (used for grouping/reporting) ---

type DocEntry struct {
	Doc           Document
	Year          string
	Date          string
	Amount        string
	InvoiceNumber string
	DocTypeName   string
}
