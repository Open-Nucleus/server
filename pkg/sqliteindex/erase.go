package sqliteindex

import "fmt"

// DeletePatientData removes all SQLite rows for a patient across all tables.
// This is used by crypto-erasure: after the encryption key is destroyed,
// the search index entries are also removed.
func (idx *sqliteIndex) DeletePatientData(patientID string) error {
	tables := []struct {
		name   string
		column string
	}{
		{"patient_summaries", "patient_id"},
		{"consents", "patient_id"},
		{"flags", "patient_id"},
		{"allergy_intolerances", "patient_id"},
		{"medication_requests", "patient_id"},
		{"conditions", "patient_id"},
		{"observations", "patient_id"},
		{"encounters", "patient_id"},
		{"immunizations", "patient_id"},
		{"procedures", "patient_id"},
		{"patients", "id"},
	}

	tx, err := idx.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	for _, t := range tables {
		if _, err := tx.Exec(
			fmt.Sprintf("DELETE FROM %s WHERE %s = ?", t.name, t.column),
			patientID,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete from %s: %w", t.name, err)
		}
	}

	return tx.Commit()
}
