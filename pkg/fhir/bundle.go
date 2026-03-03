package fhir

import (
	"encoding/json"
	"fmt"
)

// BundleEntry represents a single entry in a FHIR Bundle.
type BundleEntry struct {
	FullURL    string          // absolute or relative URL
	Resource   json.RawMessage // the FHIR resource JSON
	SearchMode string          // "match" or "include" (for searchset bundles)
}

// BundleLink represents a link in a FHIR Bundle (self, next, previous).
type BundleLink struct {
	Relation string // "self", "next", "previous"
	URL      string
}

// NewSearchBundle builds a FHIR R4 Bundle of type "searchset".
func NewSearchBundle(total int, entries []BundleEntry, links []BundleLink) ([]byte, error) {
	fhirLinks := make([]map[string]any, 0, len(links))
	for _, l := range links {
		fhirLinks = append(fhirLinks, map[string]any{
			"relation": l.Relation,
			"url":      l.URL,
		})
	}

	fhirEntries := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		entry := map[string]any{
			"fullUrl":  e.FullURL,
			"resource": json.RawMessage(e.Resource),
		}
		if e.SearchMode != "" {
			entry["search"] = map[string]any{
				"mode": e.SearchMode,
			}
		}
		fhirEntries = append(fhirEntries, entry)
	}

	bundle := map[string]any{
		"resourceType": "Bundle",
		"type":         "searchset",
		"total":        total,
		"link":         fhirLinks,
		"entry":        fhirEntries,
	}
	return json.Marshal(bundle)
}

// PaginationToLinks converts pagination metadata into FHIR Bundle links.
func PaginationToLinks(pg *Pagination, baseURL string) []BundleLink {
	var links []BundleLink

	// self link
	links = append(links, BundleLink{
		Relation: "self",
		URL:      fmt.Sprintf("%s?_count=%d&_offset=%d", baseURL, pg.PerPage, (pg.Page-1)*pg.PerPage),
	})

	// next link
	if pg.Page < pg.TotalPages {
		links = append(links, BundleLink{
			Relation: "next",
			URL:      fmt.Sprintf("%s?_count=%d&_offset=%d", baseURL, pg.PerPage, pg.Page*pg.PerPage),
		})
	}

	// previous link
	if pg.Page > 1 {
		links = append(links, BundleLink{
			Relation: "previous",
			URL:      fmt.Sprintf("%s?_count=%d&_offset=%d", baseURL, pg.PerPage, (pg.Page-2)*pg.PerPage),
		})
	}

	return links
}
