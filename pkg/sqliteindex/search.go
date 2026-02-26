package sqliteindex

import (
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// SearchPatients performs FTS5 full-text search on the patients table per spec §6.3.
func (idx *sqliteIndex) SearchPatients(query string, opts fhir.PaginationOpts) ([]*fhir.PatientRow, *fhir.Pagination, error) {
	countQuery := `SELECT COUNT(*) FROM patients_fts fts
		JOIN patients p ON p.id = fts.id
		WHERE patients_fts MATCH ? AND p.active = 1`

	var total int
	if err := idx.db.QueryRow(countQuery, query).Scan(&total); err != nil {
		return nil, nil, err
	}

	pg := paginate(opts, total)

	sqlQuery := `SELECT p.id, p.family_name, p.given_names, p.gender, p.birth_date, p.site_id, p.active, p.last_updated, p.git_blob_hash, p.fhir_json
		FROM patients_fts fts
		JOIN patients p ON p.id = fts.id
		WHERE patients_fts MATCH ? AND p.active = 1
		ORDER BY rank
		LIMIT ? OFFSET ?`

	rows, err := idx.db.Query(sqlQuery, query, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.PatientRow
	for rows.Next() {
		p := &fhir.PatientRow{}
		var active int
		if err := rows.Scan(&p.ID, &p.FamilyName, &p.GivenNames, &p.Gender, &p.BirthDate, &p.SiteID, &active, &p.LastUpdated, &p.GitBlobHash, &p.FHIRJson); err != nil {
			return nil, nil, err
		}
		p.Active = active == 1
		results = append(results, p)
	}
	return results, pg, rows.Err()
}
