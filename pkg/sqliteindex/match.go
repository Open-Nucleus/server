package sqliteindex

import (
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// GetMatchCandidates returns broad candidates for patient matching per spec §7.2.
func (idx *sqliteIndex) GetMatchCandidates(familyName, birthYear string) ([]*fhir.PatientRow, error) {
	query := `SELECT id, family_name, given_names, gender, birth_date, site_id, active, last_updated, git_blob_hash, fhir_json
		FROM patients
		WHERE active = 1
		AND (
			family_name LIKE ? || '%'
			OR birth_date LIKE ? || '%'
		)`

	rows, err := idx.db.Query(query, familyName, birthYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*fhir.PatientRow
	for rows.Next() {
		p := &fhir.PatientRow{}
		var active int
		if err := rows.Scan(&p.ID, &p.FamilyName, &p.GivenNames, &p.Gender, &p.BirthDate, &p.SiteID, &active, &p.LastUpdated, &p.GitBlobHash, &p.FHIRJson); err != nil {
			return nil, err
		}
		p.Active = active == 1
		results = append(results, p)
	}
	return results, rows.Err()
}
