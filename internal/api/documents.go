package api

import (
	"encoding/json"
	"fmt"

	"dochandler/internal/model"
)

func FetchAllDocuments(baseURL string, apiToken string, correspondentID int) ([]model.Document, error) {
	var allDocs []model.Document
	page := 1

	for {
		body, err := Get(baseURL, apiToken, "/api/documents/", map[string]string{
			"correspondent__id": fmt.Sprintf("%d", correspondentID),
			"page":              fmt.Sprintf("%d", page),
			"page_size":         "100",
		})
		if err != nil {
			return nil, err
		}

		var docList model.DocumentList
		if err := json.Unmarshal(body, &docList); err != nil {
			return nil, err
		}

		allDocs = append(allDocs, docList.Results...)

		if docList.Next == nil {
			break
		}
		page++
	}

	return allDocs, nil
}
