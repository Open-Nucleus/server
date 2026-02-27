package merge

import "encoding/json"

// ConflictLevel indicates how a merge conflict should be handled.
type ConflictLevel int

const (
	AutoMerge ConflictLevel = iota // safe to merge automatically
	Review                         // flag for clinician review
	Block                          // clinical safety risk, must be manually resolved
)

func (cl ConflictLevel) String() string {
	switch cl {
	case AutoMerge:
		return "auto_merge"
	case Review:
		return "review"
	case Block:
		return "block"
	default:
		return "unknown"
	}
}

// ParseConflictLevel converts a string to a ConflictLevel.
func ParseConflictLevel(s string) ConflictLevel {
	switch s {
	case "auto_merge":
		return AutoMerge
	case "review":
		return Review
	case "block":
		return Block
	default:
		return Review
	}
}

// FieldMergeStrategy defines how to resolve a single-field conflict.
type FieldMergeStrategy int

const (
	LatestTimestamp FieldMergeStrategy = iota // use the version with the latest lastUpdated
	KeepBoth                                  // concatenate/union arrays
	PreferLocal                               // always prefer local version
)

// SyncPriority defines the priority tier for sync ordering.
type SyncPriority int

const (
	Tier1Critical  SyncPriority = 1 // alerts, revocations, flags
	Tier2Active    SyncPriority = 2 // active encounters, patients, medications
	Tier3Clinical  SyncPriority = 3 // observations, conditions
	Tier4Resolved  SyncPriority = 4 // closed encounters, resolved conditions
	Tier5History   SyncPriority = 5 // history, audit
)

func (sp SyncPriority) String() string {
	switch sp {
	case Tier1Critical:
		return "critical"
	case Tier2Active:
		return "active"
	case Tier3Clinical:
		return "clinical"
	case Tier4Resolved:
		return "resolved"
	case Tier5History:
		return "history"
	default:
		return "unknown"
	}
}

// ConflictResult holds the output of a merge operation.
type ConflictResult struct {
	Level         ConflictLevel
	Reason        string
	MergedDoc     json.RawMessage // the merged result (if auto-merge)
	ChangedFields []string
	FieldDiffs    []FieldDiff
}

// FieldDiff describes a difference in a single field between local and remote.
type FieldDiff struct {
	Path       string
	LocalValue any
	RemoteValue any
	BaseValue   any
	Strategy   FieldMergeStrategy
}
