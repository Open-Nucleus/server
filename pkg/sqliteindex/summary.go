package sqliteindex

import (
	"time"

	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// UpdateSummary recomputes patient_summaries counts from child tables per spec §5.2.
func (idx *sqliteIndex) UpdateSummary(patientID string) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO patient_summaries (patient_id, encounter_count, active_conditions, active_medications, active_allergies, unresolved_alerts, last_encounter_date, last_updated)
		VALUES (
			?,
			(SELECT COUNT(*) FROM encounters WHERE patient_id = ?),
			(SELECT COUNT(*) FROM conditions WHERE patient_id = ? AND clinical_status = 'active'),
			(SELECT COUNT(*) FROM medication_requests WHERE patient_id = ? AND status = 'active'),
			(SELECT COUNT(*) FROM allergy_intolerances WHERE patient_id = ? AND clinical_status = 'active'),
			(SELECT COUNT(*) FROM flags WHERE patient_id = ? AND status = 'active'),
			(SELECT MAX(period_start) FROM encounters WHERE patient_id = ?),
			?
		)`,
		patientID, patientID, patientID, patientID, patientID, patientID, patientID, time.Now().UTC().Format(time.RFC3339))
	return err
}

// GetPatientBundle retrieves all resources for a patient bundle per spec §6.2.
func (idx *sqliteIndex) GetPatientBundle(patientID string) (*BundleResult, error) {
	patient, err := idx.GetPatient(patientID)
	if err != nil {
		return nil, err
	}
	if patient == nil {
		return nil, nil
	}

	bundle := &BundleResult{Patient: patient}

	// Encounters
	rows, err := idx.db.Query("SELECT id, patient_id, status, class_code, type_code, period_start, period_end, site_id, reason_code, last_updated, git_blob_hash FROM encounters WHERE patient_id = ? ORDER BY period_start DESC", patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &fhir.EncounterRow{}
		if err := rows.Scan(&e.ID, &e.PatientID, &e.Status, &e.ClassCode, &e.TypeCode, &e.PeriodStart, &e.PeriodEnd, &e.SiteID, &e.ReasonCode, &e.LastUpdated, &e.GitBlobHash); err != nil {
			return nil, err
		}
		bundle.Encounters = append(bundle.Encounters, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Observations
	obsRows, err := idx.db.Query("SELECT id, patient_id, encounter_id, status, category, code, code_display, effective_datetime, value_quantity_value, value_quantity_unit, value_string, value_codeable_concept, site_id, last_updated, git_blob_hash FROM observations WHERE patient_id = ? ORDER BY effective_datetime DESC", patientID)
	if err != nil {
		return nil, err
	}
	defer obsRows.Close()
	for obsRows.Next() {
		o := &fhir.ObservationRow{}
		if err := obsRows.Scan(&o.ID, &o.PatientID, &o.EncounterID, &o.Status, &o.Category, &o.Code, &o.CodeDisplay, &o.EffectiveDatetime, &o.ValueQuantityValue, &o.ValueQuantityUnit, &o.ValueString, &o.ValueCodeableConcept, &o.SiteID, &o.LastUpdated, &o.GitBlobHash); err != nil {
			return nil, err
		}
		bundle.Observations = append(bundle.Observations, o)
	}
	if err := obsRows.Err(); err != nil {
		return nil, err
	}

	// Conditions (active only)
	condRows, err := idx.db.Query("SELECT id, patient_id, clinical_status, verification_status, code, code_display, onset_datetime, site_id, last_updated, git_blob_hash FROM conditions WHERE patient_id = ? AND clinical_status = 'active'", patientID)
	if err != nil {
		return nil, err
	}
	defer condRows.Close()
	for condRows.Next() {
		c := &fhir.ConditionRow{}
		if err := condRows.Scan(&c.ID, &c.PatientID, &c.ClinicalStatus, &c.VerificationStatus, &c.Code, &c.CodeDisplay, &c.OnsetDatetime, &c.SiteID, &c.LastUpdated, &c.GitBlobHash); err != nil {
			return nil, err
		}
		bundle.Conditions = append(bundle.Conditions, c)
	}
	if err := condRows.Err(); err != nil {
		return nil, err
	}

	// MedicationRequests (active only)
	medRows, err := idx.db.Query("SELECT id, patient_id, status, intent, medication_code, medication_display, authored_on, site_id, last_updated, git_blob_hash FROM medication_requests WHERE patient_id = ? AND status = 'active'", patientID)
	if err != nil {
		return nil, err
	}
	defer medRows.Close()
	for medRows.Next() {
		m := &fhir.MedicationRequestRow{}
		if err := medRows.Scan(&m.ID, &m.PatientID, &m.Status, &m.Intent, &m.MedicationCode, &m.MedicationDisplay, &m.AuthoredOn, &m.SiteID, &m.LastUpdated, &m.GitBlobHash); err != nil {
			return nil, err
		}
		bundle.MedicationRequests = append(bundle.MedicationRequests, m)
	}
	if err := medRows.Err(); err != nil {
		return nil, err
	}

	// AllergyIntolerances (active only)
	allergyRows, err := idx.db.Query("SELECT id, patient_id, clinical_status, verification_status, type, substance_code, substance_display, criticality, site_id, last_updated, git_blob_hash FROM allergy_intolerances WHERE patient_id = ? AND clinical_status = 'active'", patientID)
	if err != nil {
		return nil, err
	}
	defer allergyRows.Close()
	for allergyRows.Next() {
		a := &fhir.AllergyIntoleranceRow{}
		if err := allergyRows.Scan(&a.ID, &a.PatientID, &a.ClinicalStatus, &a.VerificationStatus, &a.Type, &a.SubstanceCode, &a.SubstanceDisplay, &a.Criticality, &a.SiteID, &a.LastUpdated, &a.GitBlobHash); err != nil {
			return nil, err
		}
		bundle.AllergyIntolerances = append(bundle.AllergyIntolerances, a)
	}
	if err := allergyRows.Err(); err != nil {
		return nil, err
	}

	// Flags (active only)
	flagRows, err := idx.db.Query("SELECT id, patient_id, status, category, code, period_start, period_end, generated_by, site_id, last_updated, git_blob_hash FROM flags WHERE patient_id = ? AND status = 'active'", patientID)
	if err != nil {
		return nil, err
	}
	defer flagRows.Close()
	for flagRows.Next() {
		f := &fhir.FlagRow{}
		if err := flagRows.Scan(&f.ID, &f.PatientID, &f.Status, &f.Category, &f.Code, &f.PeriodStart, &f.PeriodEnd, &f.GeneratedBy, &f.SiteID, &f.LastUpdated, &f.GitBlobHash); err != nil {
			return nil, err
		}
		bundle.Flags = append(bundle.Flags, f)
	}
	return bundle, flagRows.Err()
}
