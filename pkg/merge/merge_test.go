package merge

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Diff Tests ---

func TestDiffResources_NonOverlapping(t *testing.T) {
	local := json.RawMessage(`{"name":"Alice","gender":"female"}`)
	remote := json.RawMessage(`{"name":"Alice","telecom":["555-1234"]}`)

	diffs, err := DiffResources(local, remote)
	require.NoError(t, err)
	assert.True(t, len(diffs) >= 1) // gender and telecom differ
}

func TestDiffResources_Overlapping(t *testing.T) {
	base := json.RawMessage(`{"name":"Alice","gender":"female"}`)
	local := json.RawMessage(`{"name":"Alice B","gender":"female"}`)
	remote := json.RawMessage(`{"name":"Alice C","gender":"female"}`)

	diffs, err := DiffResourcesWithBase(base, local, remote)
	require.NoError(t, err)
	assert.Len(t, diffs, 1)
	assert.Equal(t, "name", diffs[0].Path)

	overlapping := OverlappingFields(diffs)
	assert.Len(t, overlapping, 1)
}

func TestDiffResources_NonOverlappingWithBase(t *testing.T) {
	base := json.RawMessage(`{"name":"Alice","gender":"female"}`)
	local := json.RawMessage(`{"name":"Alice B","gender":"female"}`)
	remote := json.RawMessage(`{"name":"Alice","gender":"female","telecom":["555"]}`)

	diffs, err := DiffResourcesWithBase(base, local, remote)
	require.NoError(t, err)
	assert.Len(t, diffs, 2) // name changed by local, telecom added by remote

	nonOverlapping := NonOverlappingFields(diffs)
	assert.Len(t, nonOverlapping, 2)
}

// --- Classify Tests ---

func TestClassify_AutoMerge_NonOverlapping(t *testing.T) {
	c := &Classifier{}
	base := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith"}],"gender":"male"}`)
	local := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith"}],"gender":"male","telecom":[{"value":"555"}]}`)
	remote := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith"}],"gender":"male","address":[{"city":"Lagos"}]}`)

	result := c.Classify("Patient", local, remote, base)
	assert.Equal(t, AutoMerge, result.Level)
}

func TestClassify_Block_AllergyCriticality(t *testing.T) {
	c := &Classifier{}
	base := json.RawMessage(`{"resourceType":"AllergyIntolerance","criticality":"low","code":{"text":"Peanuts"}}`)
	local := json.RawMessage(`{"resourceType":"AllergyIntolerance","criticality":"high","code":{"text":"Peanuts"}}`)
	remote := json.RawMessage(`{"resourceType":"AllergyIntolerance","criticality":"unable-to-assess","code":{"text":"Peanuts"}}`)

	result := c.Classify("AllergyIntolerance", local, remote, base)
	assert.Equal(t, Block, result.Level)
	assert.Contains(t, result.Reason, "criticality")
}

func TestClassify_Block_PatientIdentity(t *testing.T) {
	c := &Classifier{}
	base := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith"}],"birthDate":"1990-01-01"}`)
	local := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Smith-Jones"}],"birthDate":"1990-01-01"}`)
	remote := json.RawMessage(`{"resourceType":"Patient","name":[{"family":"Doe"}],"birthDate":"1990-01-01"}`)

	result := c.Classify("Patient", local, remote, base)
	assert.Equal(t, Block, result.Level)
	assert.Contains(t, result.Reason, "patient identity")
}

func TestClassify_Block_DiagnosisConflict(t *testing.T) {
	c := &Classifier{}
	base := json.RawMessage(`{"resourceType":"Condition","code":{"text":"Malaria"}}`)
	local := json.RawMessage(`{"resourceType":"Condition","code":{"text":"Severe Malaria"}}`)
	remote := json.RawMessage(`{"resourceType":"Condition","code":{"text":"Typhoid"}}`)

	result := c.Classify("Condition", local, remote, base)
	assert.Equal(t, Block, result.Level)
	assert.Contains(t, result.Reason, "diagnosis")
}

func TestClassify_Block_MedicationConflict(t *testing.T) {
	c := &Classifier{}
	base := json.RawMessage(`{"resourceType":"MedicationRequest","medicationCodeableConcept":{"coding":[{"code":"ACT"}]}}`)
	local := json.RawMessage(`{"resourceType":"MedicationRequest","medicationCodeableConcept":{"coding":[{"code":"ACT-HIGH"}]}}`)
	remote := json.RawMessage(`{"resourceType":"MedicationRequest","medicationCodeableConcept":{"coding":[{"code":"ACT-LOW"}]}}`)

	result := c.Classify("MedicationRequest", local, remote, base)
	assert.Equal(t, Block, result.Level)
}

func TestClassify_Block_VitalSigns(t *testing.T) {
	c := &Classifier{}
	base := json.RawMessage(`{"resourceType":"Observation","valueQuantity":{"value":120}}`)
	local := json.RawMessage(`{"resourceType":"Observation","valueQuantity":{"value":140}}`)
	remote := json.RawMessage(`{"resourceType":"Observation","valueQuantity":{"value":90}}`)

	result := c.Classify("Observation", local, remote, base)
	assert.Equal(t, Block, result.Level)
	assert.Contains(t, result.Reason, "vital signs")
}

func TestClassify_Review_OverlappingNonClinical(t *testing.T) {
	c := &Classifier{}
	base := json.RawMessage(`{"resourceType":"Patient","text":{"status":"generated","div":"old"},"meta":{"versionId":"1"}}`)
	local := json.RawMessage(`{"resourceType":"Patient","text":{"status":"generated","div":"new-local"},"meta":{"versionId":"2"}}`)
	remote := json.RawMessage(`{"resourceType":"Patient","text":{"status":"generated","div":"new-remote"},"meta":{"versionId":"3"}}`)

	result := c.Classify("Patient", local, remote, base)
	// text and meta overlapping but non-clinical → Review
	assert.Equal(t, Review, result.Level)
}

func TestClassify_NilFormulary(t *testing.T) {
	c := &Classifier{Formulary: nil}
	base := json.RawMessage(`{"resourceType":"MedicationRequest","note":[{"text":"take daily"}]}`)
	local := json.RawMessage(`{"resourceType":"MedicationRequest","note":[{"text":"take daily with food"}]}`)
	remote := json.RawMessage(`{"resourceType":"MedicationRequest","note":[{"text":"take daily before meals"}]}`)

	// Should not panic with nil formulary
	result := c.Classify("MedicationRequest", local, remote, base)
	assert.NotEqual(t, AutoMerge, result.Level) // overlapping changes in note
}

// --- Merge Strategy Tests ---

func TestGetFieldStrategy(t *testing.T) {
	assert.Equal(t, LatestTimestamp, GetFieldStrategy("Patient", "name"))
	assert.Equal(t, KeepBoth, GetFieldStrategy("Patient", "telecom"))
	assert.Equal(t, KeepBoth, GetFieldStrategy("Observation", "note"))
	assert.Equal(t, LatestTimestamp, GetFieldStrategy("Patient", "unknown_field"))
	assert.Equal(t, LatestTimestamp, GetFieldStrategy("UnknownType", "field"))
}

func TestMergeFields_NonOverlapping(t *testing.T) {
	driver := NewDriver(nil)

	base := json.RawMessage(`{"resourceType":"Patient","meta":{"lastUpdated":"2024-01-01T00:00:00Z"},"name":[{"family":"Smith"}]}`)
	local := json.RawMessage(`{"resourceType":"Patient","meta":{"lastUpdated":"2024-01-02T00:00:00Z"},"name":[{"family":"Smith"}],"telecom":[{"value":"555"}]}`)
	remote := json.RawMessage(`{"resourceType":"Patient","meta":{"lastUpdated":"2024-01-01T12:00:00Z"},"name":[{"family":"Smith"}],"address":[{"city":"Lagos"}]}`)

	merged, err := driver.MergeFields("Patient", base, local, remote)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(merged, &result)
	require.NoError(t, err)

	assert.NotNil(t, result["telecom"])
	assert.NotNil(t, result["address"])
}

func TestMergeFields_KeepBothStrategy(t *testing.T) {
	driver := NewDriver(nil)

	base := json.RawMessage(`{"resourceType":"Patient","meta":{"lastUpdated":"2024-01-01T00:00:00Z"},"telecom":[{"value":"555"}]}`)
	local := json.RawMessage(`{"resourceType":"Patient","meta":{"lastUpdated":"2024-01-02T00:00:00Z"},"telecom":[{"value":"555"},{"value":"666"}]}`)
	remote := json.RawMessage(`{"resourceType":"Patient","meta":{"lastUpdated":"2024-01-02T00:00:00Z"},"telecom":[{"value":"555"},{"value":"777"}]}`)

	merged, err := driver.MergeFields("Patient", base, local, remote)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(merged, &result)
	require.NoError(t, err)

	telecoms := result["telecom"].([]any)
	assert.Len(t, telecoms, 3) // 555, 666, 777
}

// --- Priority Tests ---

func TestPriority_Tier1_Flag(t *testing.T) {
	data := json.RawMessage(`{"resourceType":"Flag","status":"active"}`)
	assert.Equal(t, Tier1Critical, ClassifyResource("patients/p1/flags/f1.json", data))
}

func TestPriority_Tier2_ActivePatient(t *testing.T) {
	data := json.RawMessage(`{"resourceType":"Patient","active":true}`)
	assert.Equal(t, Tier2Active, ClassifyResource("patients/p1/Patient.json", data))
}

func TestPriority_Tier2_ActiveEncounter(t *testing.T) {
	data := json.RawMessage(`{"resourceType":"Encounter","status":"in-progress"}`)
	assert.Equal(t, Tier2Active, ClassifyResource("patients/p1/encounters/e1.json", data))
}

func TestPriority_Tier3_Observation(t *testing.T) {
	data := json.RawMessage(`{"resourceType":"Observation","status":"final"}`)
	assert.Equal(t, Tier3Clinical, ClassifyResource("patients/p1/observations/o1.json", data))
}

func TestPriority_Tier4_FinishedEncounter(t *testing.T) {
	data := json.RawMessage(`{"resourceType":"Encounter","status":"finished"}`)
	assert.Equal(t, Tier4Resolved, ClassifyResource("patients/p1/encounters/e1.json", data))
}

func TestPriority_Tier4_ResolvedCondition(t *testing.T) {
	data := json.RawMessage(`{"resourceType":"Condition","clinicalStatus":{"coding":[{"code":"resolved"}]}}`)
	assert.Equal(t, Tier4Resolved, ClassifyResource("patients/p1/conditions/c1.json", data))
}
