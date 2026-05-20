package api

import (
	"encoding/json"
	"fmt"

	"dochandler/internal/model"
)

func FetchDocumentTypes(baseURL string, apiToken string) (map[int]string, error) {
	typeMap := make(map[int]string)
	page := 1

	for {
		body, err := Get(baseURL, apiToken, "/api/document_types/", map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": "100",
		})
		if err != nil {
			return nil, err
		}

		var dtList model.DocumentTypeList
		if err := json.Unmarshal(body, &dtList); err != nil {
			return nil, err
		}

		for _, dt := range dtList.Results {
			typeMap[dt.ID] = dt.Name
		}

		if dtList.Next == nil {
			break
		}
		page++
	}

	return typeMap, nil
}
