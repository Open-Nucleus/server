package store

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MedicationRecord represents a medication in the drug database.
type MedicationRecord struct {
	Code              string   `json:"code"`
	Display           string   `json:"display"`
	Form              string   `json:"form"`
	Route             string   `json:"route"`
	Category          string   `json:"category"`
	WHOEssential      bool     `json:"who_essential"`
	TherapeuticClass  string   `json:"therapeutic_class"`
	CommonFrequencies []string `json:"common_frequencies"`
	Strength          string   `json:"strength"`
	Unit              string   `json:"unit"`
}

// DrugDB is an in-memory drug database loaded from FHIR Medication JSONs.
type DrugDB struct {
	byCode map[string]*MedicationRecord // ATC code → record
	all    []*MedicationRecord          // all records for iteration
}

// NewDrugDB creates an empty DrugDB.
func NewDrugDB() *DrugDB {
	return &DrugDB{
		byCode: make(map[string]*MedicationRecord),
	}
}

// LoadFromJSON loads medications from a JSON byte slice (array of MedicationRecord).
func (db *DrugDB) LoadFromJSON(data []byte) error {
	var records []MedicationRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("unmarshal medications: %w", err)
	}
	for i := range records {
		r := &records[i]
		db.byCode[r.Code] = r
		db.all = append(db.all, r)
	}
	return nil
}

// LoadSingleJSON loads a single medication from a FHIR-like JSON object.
func (db *DrugDB) LoadSingleJSON(data []byte) error {
	var r MedicationRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("unmarshal medication: %w", err)
	}
	if r.Code == "" {
		return nil // skip empty
	}
	db.byCode[r.Code] = &r
	db.all = append(db.all, &r)
	return nil
}

// Get returns a medication by ATC code.
func (db *DrugDB) Get(code string) (*MedicationRecord, bool) {
	r, ok := db.byCode[strings.ToUpper(code)]
	return r, ok
}

// Search performs case-insensitive substring search on display name and code.
// Optionally filters by category.
func (db *DrugDB) Search(query, category string, page, perPage int) ([]*MedicationRecord, int) {
	queryLower := strings.ToLower(query)
	categoryLower := strings.ToLower(category)

	var matches []*MedicationRecord
	for _, r := range db.all {
		if categoryLower != "" && strings.ToLower(r.Category) != categoryLower {
			continue
		}
		if queryLower != "" {
			displayLower := strings.ToLower(r.Display)
			codeLower := strings.ToLower(r.Code)
			if !strings.Contains(displayLower, queryLower) && !strings.Contains(codeLower, queryLower) {
				continue
			}
		}
		matches = append(matches, r)
	}

	total := len(matches)
	start := (page - 1) * perPage
	if start >= total {
		return nil, total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return matches[start:end], total
}

// ListByCategory returns all medications in a given category.
func (db *DrugDB) ListByCategory(category string, page, perPage int) ([]*MedicationRecord, int) {
	return db.Search("", category, page, perPage)
}

// All returns all loaded medications.
func (db *DrugDB) All() []*MedicationRecord {
	return db.all
}

// Count returns the number of loaded medications.
func (db *DrugDB) Count() int {
	return len(db.all)
}

// Categories returns all unique categories.
func (db *DrugDB) Categories() []string {
	seen := make(map[string]bool)
	var cats []string
	for _, r := range db.all {
		if r.Category != "" && !seen[r.Category] {
			seen[r.Category] = true
			cats = append(cats, r.Category)
		}
	}
	return cats
}
