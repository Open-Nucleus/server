package merge

import "encoding/json"

// DiffResources computes field-level differences between local and remote FHIR resources.
func DiffResources(local, remote json.RawMessage) ([]FieldDiff, error) {
	var localMap, remoteMap map[string]any
	if err := json.Unmarshal(local, &localMap); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(remote, &remoteMap); err != nil {
		return nil, err
	}

	var diffs []FieldDiff
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

		if !jsonEqual(lv, rv) {
			diffs = append(diffs, FieldDiff{
				Path:        key,
				LocalValue:  lv,
				RemoteValue: rv,
			})
		}
	}
	return diffs, nil
}

// DiffResourcesWithBase computes field-level differences with a common ancestor.
func DiffResourcesWithBase(base, local, remote json.RawMessage) ([]FieldDiff, error) {
	var baseMap, localMap, remoteMap map[string]any
	if base != nil {
		if err := json.Unmarshal(base, &baseMap); err != nil {
			return nil, err
		}
	} else {
		baseMap = make(map[string]any)
	}
	if err := json.Unmarshal(local, &localMap); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(remote, &remoteMap); err != nil {
		return nil, err
	}

	var diffs []FieldDiff
	allKeys := make(map[string]bool)
	for k := range localMap {
		allKeys[k] = true
	}
	for k := range remoteMap {
		allKeys[k] = true
	}
	for k := range baseMap {
		allKeys[k] = true
	}

	for key := range allKeys {
		lv := localMap[key]
		rv := remoteMap[key]
		bv := baseMap[key]

		localChanged := !jsonEqual(lv, bv)
		remoteChanged := !jsonEqual(rv, bv)

		if localChanged || remoteChanged {
			if !jsonEqual(lv, rv) {
				diffs = append(diffs, FieldDiff{
					Path:        key,
					LocalValue:  lv,
					RemoteValue: rv,
					BaseValue:   bv,
				})
			}
		}
	}
	return diffs, nil
}

// OverlappingFields returns diffs where both local and remote changed from base.
func OverlappingFields(diffs []FieldDiff) []FieldDiff {
	var result []FieldDiff
	for _, d := range diffs {
		localChanged := !jsonEqual(d.LocalValue, d.BaseValue)
		remoteChanged := !jsonEqual(d.RemoteValue, d.BaseValue)
		if localChanged && remoteChanged {
			result = append(result, d)
		}
	}
	return result
}

// NonOverlappingFields returns diffs where only one side changed from base.
func NonOverlappingFields(diffs []FieldDiff) []FieldDiff {
	var result []FieldDiff
	for _, d := range diffs {
		localChanged := !jsonEqual(d.LocalValue, d.BaseValue)
		remoteChanged := !jsonEqual(d.RemoteValue, d.BaseValue)
		if localChanged != remoteChanged {
			result = append(result, d)
		}
	}
	return result
}

// jsonEqual compares two values by their JSON representation.
func jsonEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aj, err1 := json.Marshal(a)
	bj, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aj) == string(bj)
}
