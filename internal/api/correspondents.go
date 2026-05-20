package api

import (
	"encoding/json"
	"fmt"

	"dochandler/internal/model"
)

func FindCorrespondentID(slug string) (int, error) {
	body, err := Get("/api/correspondents/", map[string]string{
		"slug":      slug,
		"page_size": "100",
	})
	if err != nil {
		return 0, err
	}

	var list model.CorrespondentList
	if err := json.Unmarshal(body, &list); err != nil {
		return 0, err
	}

	for _, c := range list.Results {
		if c.Slug == slug {
			return c.ID, nil
		}
	}

	// Debug output
	fmt.Println("Gefundene Korrespondenten:")
	for _, c := range list.Results {
		fmt.Printf("  ID: %d  Slug: %q  Name: %q\n", c.ID, c.Slug, c.Name)
	}

	return 0, fmt.Errorf("Korrespondent mit Slug '%s' nicht gefunden", slug)
}

func FetchAllCorrespondents() ([]model.Correspondent, error) {
	var all []model.Correspondent
	page := 1

	for {
		body, err := Get("/api/correspondents/", map[string]string{
			"page":      fmt.Sprintf("%d", page),
			"page_size": "100",
		})
		if err != nil {
			return nil, err
		}

		var list model.CorrespondentList
		if err := json.Unmarshal(body, &list); err != nil {
			return nil, err
		}

		all = append(all, list.Results...)

		if list.Next == nil {
			break
		}
		page++
	}

	return all, nil
}
