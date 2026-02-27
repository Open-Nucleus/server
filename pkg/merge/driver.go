package merge

import (
	"encoding/json"
	"fmt"
	"time"
)

// Driver performs FHIR-aware three-way merges.
type Driver struct {
	Classifier *Classifier
}

// NewDriver creates a merge driver with an optional formulary checker.
func NewDriver(formulary FormularyChecker) *Driver {
	return &Driver{
		Classifier: &Classifier{Formulary: formulary},
	}
}

// MergeFile performs a three-way merge on a FHIR resource.
func (d *Driver) MergeFile(resourceType, resourceID, patientID string, base, local, remote json.RawMessage) ConflictResult {
	result := d.Classifier.Classify(resourceType, local, remote, base)

	if result.Level == AutoMerge {
		merged, err := d.MergeFields(resourceType, base, local, remote)
		if err != nil {
			result.Level = Review
			result.Reason = fmt.Sprintf("auto-merge failed: %v", err)
		} else {
			result.MergedDoc = merged
		}
	}

	return result
}

// MergeFields applies field-level merge strategies to produce a merged document.
func (d *Driver) MergeFields(resourceType string, base, local, remote json.RawMessage) (json.RawMessage, error) {
	var baseMap map[string]any
	if base != nil {
		if err := json.Unmarshal(base, &baseMap); err != nil {
			return nil, fmt.Errorf("unmarshal base: %w", err)
		}
	} else {
		baseMap = make(map[string]any)
	}

	var localMap, remoteMap map[string]any
	if err := json.Unmarshal(local, &localMap); err != nil {
		return nil, fmt.Errorf("unmarshal local: %w", err)
	}
	if err := json.Unmarshal(remote, &remoteMap); err != nil {
		return nil, fmt.Errorf("unmarshal remote: %w", err)
	}

	merged := make(map[string]any)

	// Start with local as the base for the merge
	for k, v := range localMap {
		merged[k] = v
	}

	// Apply remote changes where appropriate
	allKeys := make(map[string]bool)
	for k := range localMap {
		allKeys[k] = true
	}
	for k := range remoteMap {
		allKeys[k] = true
	}

	for key := range allKeys {
		lv := localMap[key]
		rv := remoteMap[key]
		bv := baseMap[key]

		localChanged := !jsonEqual(lv, bv)
		remoteChanged := !jsonEqual(rv, bv)

		switch {
		case !localChanged && !remoteChanged:
			// No changes, keep local (already in merged)
		case localChanged && !remoteChanged:
			// Only local changed, keep local (already in merged)
		case !localChanged && remoteChanged:
			// Only remote changed, take remote
			if rv != nil {
				merged[key] = rv
			} else {
				delete(merged, key)
			}
		case localChanged && remoteChanged:
			// Both changed — apply strategy
			strategy := GetFieldStrategy(resourceType, key)
			mergedValue := applyStrategy(strategy, lv, rv, localMap, remoteMap)
			if mergedValue != nil {
				merged[key] = mergedValue
			}
		}
	}

	return json.Marshal(merged)
}

// applyStrategy applies a field merge strategy.
func applyStrategy(strategy FieldMergeStrategy, localVal, remoteVal any, localDoc, remoteDoc map[string]any) any {
	switch strategy {
	case KeepBoth:
		return mergeArrays(localVal, remoteVal)
	case PreferLocal:
		return localVal
	case LatestTimestamp:
		return pickLatest(localVal, remoteVal, localDoc, remoteDoc)
	default:
		return localVal
	}
}

// mergeArrays unions two array values, deduplicating by JSON equality.
func mergeArrays(a, b any) any {
	aArr, aOk := toSlice(a)
	bArr, bOk := toSlice(b)

	if !aOk && !bOk {
		return a
	}
	if !aOk {
		return b
	}
	if !bOk {
		return a
	}

	result := make([]any, len(aArr))
	copy(result, aArr)

	for _, bItem := range bArr {
		found := false
		for _, aItem := range aArr {
			if jsonEqual(aItem, bItem) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, bItem)
		}
	}
	return result
}

func toSlice(v any) ([]any, bool) {
	if v == nil {
		return nil, false
	}
	if s, ok := v.([]any); ok {
		return s, true
	}
	return nil, false
}

// pickLatest picks the value from the version with the more recent lastUpdated.
func pickLatest(localVal, remoteVal any, localDoc, remoteDoc map[string]any) any {
	localTime := extractLastUpdated(localDoc)
	remoteTime := extractLastUpdated(remoteDoc)

	if remoteTime.After(localTime) {
		return remoteVal
	}
	return localVal
}

func extractLastUpdated(doc map[string]any) time.Time {
	meta, ok := doc["meta"].(map[string]any)
	if !ok {
		return time.Time{}
	}
	lu, ok := meta["lastUpdated"].(string)
	if !ok {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, lu)
	if err != nil {
		return time.Time{}
	}
	return t
}
