package fhir

import (
	"encoding/json"
	"testing"
)

func TestNewSearchBundle_Basic(t *testing.T) {
	entries := []BundleEntry{
		{
			FullURL:    "http://example.com/Patient/1",
			Resource:   json.RawMessage(`{"resourceType":"Patient","id":"1"}`),
			SearchMode: "match",
		},
	}
	links := []BundleLink{
		{Relation: "self", URL: "http://example.com/Patient?_count=10&_offset=0"},
	}

	data, err := NewSearchBundle(1, entries, links)
	if err != nil {
		t.Fatal(err)
	}

	var bundle map[string]any
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle["resourceType"] != "Bundle" {
		t.Errorf("resourceType = %v", bundle["resourceType"])
	}
	if bundle["type"] != "searchset" {
		t.Errorf("type = %v", bundle["type"])
	}
	if int(bundle["total"].(float64)) != 1 {
		t.Errorf("total = %v", bundle["total"])
	}

	entryArr := bundle["entry"].([]any)
	if len(entryArr) != 1 {
		t.Fatalf("entries = %d, want 1", len(entryArr))
	}
	e0 := entryArr[0].(map[string]any)
	if e0["fullUrl"] != "http://example.com/Patient/1" {
		t.Errorf("fullUrl = %v", e0["fullUrl"])
	}
	search := e0["search"].(map[string]any)
	if search["mode"] != "match" {
		t.Errorf("search.mode = %v", search["mode"])
	}
}

func TestNewSearchBundle_Empty(t *testing.T) {
	data, err := NewSearchBundle(0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var bundle map[string]any
	json.Unmarshal(data, &bundle)
	if int(bundle["total"].(float64)) != 0 {
		t.Errorf("total = %v", bundle["total"])
	}
}

func TestPaginationToLinks_FirstPage(t *testing.T) {
	pg := &Pagination{Page: 1, PerPage: 10, Total: 25, TotalPages: 3}
	links := PaginationToLinks(pg, "http://example.com/Patient")

	if len(links) != 2 {
		t.Fatalf("expected 2 links (self+next), got %d", len(links))
	}
	if links[0].Relation != "self" {
		t.Errorf("first link relation = %q", links[0].Relation)
	}
	if links[1].Relation != "next" {
		t.Errorf("second link relation = %q", links[1].Relation)
	}
}

func TestPaginationToLinks_MiddlePage(t *testing.T) {
	pg := &Pagination{Page: 2, PerPage: 10, Total: 25, TotalPages: 3}
	links := PaginationToLinks(pg, "http://example.com/Patient")

	if len(links) != 3 {
		t.Fatalf("expected 3 links (self+next+prev), got %d", len(links))
	}
	relations := map[string]bool{}
	for _, l := range links {
		relations[l.Relation] = true
	}
	for _, r := range []string{"self", "next", "previous"} {
		if !relations[r] {
			t.Errorf("missing %q link", r)
		}
	}
}

func TestPaginationToLinks_LastPage(t *testing.T) {
	pg := &Pagination{Page: 3, PerPage: 10, Total: 25, TotalPages: 3}
	links := PaginationToLinks(pg, "http://example.com/Patient")

	if len(links) != 2 {
		t.Fatalf("expected 2 links (self+prev), got %d", len(links))
	}
	if links[1].Relation != "previous" {
		t.Errorf("second link relation = %q, want 'previous'", links[1].Relation)
	}
}
