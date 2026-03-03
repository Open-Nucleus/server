package fhir

import (
	"net/http"
	"strconv"
	"strings"
)

// FHIRSearchParams holds parsed FHIR search parameters.
type FHIRSearchParams struct {
	Patient string            // extracted from ?patient=Patient/xxx or ?subject=Patient/xxx
	Count   int               // _count (default 25, max 100)
	Offset  int               // _offset (default 0)
	Page    int               // computed: (offset / count) + 1
	Filters map[string]string // all other search params
}

// ParseFHIRSearchParams parses FHIR search parameters from an HTTP request.
func ParseFHIRSearchParams(r *http.Request) *FHIRSearchParams {
	q := r.URL.Query()

	count := queryIntDefault(q.Get("_count"), 25)
	if count < 1 {
		count = 25
	}
	if count > 100 {
		count = 100
	}

	offset := queryIntDefault(q.Get("_offset"), 0)
	if offset < 0 {
		offset = 0
	}

	page := (offset / count) + 1

	// Extract patient reference from patient or subject params
	patient := ""
	if p := q.Get("patient"); p != "" {
		patient = stripFHIRReference(p, "Patient/")
	} else if s := q.Get("subject"); s != "" {
		patient = stripFHIRReference(s, "Patient/")
	}

	// Collect remaining filters (skip pagination and patient params)
	filters := map[string]string{}
	skipParams := map[string]bool{
		"_count": true, "_offset": true, "patient": true, "subject": true,
	}
	for key, vals := range q {
		if skipParams[key] || len(vals) == 0 {
			continue
		}
		filters[key] = vals[0]
	}

	return &FHIRSearchParams{
		Patient: patient,
		Count:   count,
		Offset:  offset,
		Page:    page,
		Filters: filters,
	}
}

func stripFHIRReference(ref, prefix string) string {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, prefix) {
		return ref[len(prefix):]
	}
	return ref
}

func queryIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
