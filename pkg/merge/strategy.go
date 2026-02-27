package merge

// fieldStrategies maps resource type + field path to a merge strategy.
var fieldStrategies = map[string]map[string]FieldMergeStrategy{
	"Patient": {
		"name":      LatestTimestamp,
		"telecom":   KeepBoth,
		"address":   KeepBoth,
		"contact":   KeepBoth,
		"extension": KeepBoth,
		"meta":      LatestTimestamp,
		"text":      LatestTimestamp,
	},
	"Encounter": {
		"participant": KeepBoth,
		"location":    KeepBoth,
		"meta":        LatestTimestamp,
		"text":        LatestTimestamp,
	},
	"Observation": {
		"performer": KeepBoth,
		"note":      KeepBoth,
		"meta":      LatestTimestamp,
		"text":      LatestTimestamp,
	},
	"Condition": {
		"note":     KeepBoth,
		"evidence": KeepBoth,
		"meta":     LatestTimestamp,
		"text":     LatestTimestamp,
	},
	"MedicationRequest": {
		"note":              KeepBoth,
		"reasonReference":   KeepBoth,
		"supportingInfo":    KeepBoth,
		"meta":              LatestTimestamp,
		"text":              LatestTimestamp,
	},
	"AllergyIntolerance": {
		"note": KeepBoth,
		"meta": LatestTimestamp,
		"text": LatestTimestamp,
	},
}

// GetFieldStrategy returns the merge strategy for a specific field.
// Defaults to LatestTimestamp if no specific strategy is defined.
func GetFieldStrategy(resourceType, fieldPath string) FieldMergeStrategy {
	if strategies, ok := fieldStrategies[resourceType]; ok {
		if strategy, ok := strategies[fieldPath]; ok {
			return strategy
		}
	}
	return LatestTimestamp
}
