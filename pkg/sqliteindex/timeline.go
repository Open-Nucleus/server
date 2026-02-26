package sqliteindex

import (
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// GetTimeline returns a chronological view of clinical events per spec §6.5.
func (idx *sqliteIndex) GetTimeline(patientID string, opts fhir.PaginationOpts) ([]fhir.TimelineEvent, *fhir.Pagination, error) {
	countQuery := `SELECT COUNT(*) FROM (
		SELECT id FROM encounters WHERE patient_id = ?
		UNION ALL
		SELECT id FROM observations WHERE patient_id = ?
		UNION ALL
		SELECT id FROM conditions WHERE patient_id = ?
		UNION ALL
		SELECT id FROM flags WHERE patient_id = ?
	)`
	var total int
	if err := idx.db.QueryRow(countQuery, patientID, patientID, patientID, patientID).Scan(&total); err != nil {
		return nil, nil, err
	}

	pg := paginate(opts, total)

	query := `SELECT type, id, date, fhir_json FROM (
		SELECT 'encounter' as type, id, period_start as date, fhir_json FROM encounters WHERE patient_id = ?
		UNION ALL
		SELECT 'observation' as type, id, effective_datetime as date, fhir_json FROM observations WHERE patient_id = ?
		UNION ALL
		SELECT 'condition' as type, id, COALESCE(onset_datetime, last_updated) as date, fhir_json FROM conditions WHERE patient_id = ?
		UNION ALL
		SELECT 'flag' as type, id, COALESCE(period_start, last_updated) as date, fhir_json FROM flags WHERE patient_id = ?
	) ORDER BY date DESC LIMIT ? OFFSET ?`

	rows, err := idx.db.Query(query, patientID, patientID, patientID, patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var events []fhir.TimelineEvent
	for rows.Next() {
		var e fhir.TimelineEvent
		if err := rows.Scan(&e.EventType, &e.ResourceID, &e.Date, &e.FHIRJson); err != nil {
			return nil, nil, err
		}
		events = append(events, e)
	}
	return events, pg, rows.Err()
}
