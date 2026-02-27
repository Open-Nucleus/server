package auth

// Permission constants for RBAC.
const (
	// Patient permissions
	PermPatientRead   = "patient:read"
	PermPatientWrite  = "patient:write"
	PermPatientDelete = "patient:delete"
	PermPatientMatch  = "patient:match"

	// Encounter permissions
	PermEncounterRead  = "encounter:read"
	PermEncounterWrite = "encounter:write"

	// Observation permissions
	PermObservationRead  = "observation:read"
	PermObservationWrite = "observation:write"

	// Condition permissions
	PermConditionRead  = "condition:read"
	PermConditionWrite = "condition:write"

	// Medication permissions
	PermMedicationRead  = "medication:read"
	PermMedicationWrite = "medication:write"

	// Allergy permissions
	PermAllergyRead  = "allergy:read"
	PermAllergyWrite = "allergy:write"

	// Formulary permissions
	PermFormularyRead  = "formulary:read"
	PermFormularyWrite = "formulary:write"

	// Sync permissions
	PermSyncTrigger = "sync:trigger"
	PermSyncStatus  = "sync:status"
	PermSyncBundle  = "sync:bundle"

	// Conflict permissions
	PermConflictRead    = "conflict:read"
	PermConflictResolve = "conflict:resolve"

	// Sentinel / Alert permissions
	PermAlertRead        = "alert:read"
	PermAlertAcknowledge = "alert:acknowledge"

	// Anchor permissions
	PermAnchorRead    = "anchor:read"
	PermAnchorTrigger = "anchor:trigger"

	// Supply permissions
	PermSupplyRead  = "supply:read"
	PermSupplyWrite = "supply:write"

	// Admin permissions
	PermDeviceManage = "device:manage"
	PermRoleAssign   = "role:assign"

	// Flag permissions
	PermFlagRead  = "flag:read"
	PermFlagWrite = "flag:write"
)

// Role constants.
const (
	RoleCHW           = "chw"
	RoleNurse         = "nurse"
	RolePhysician     = "physician"
	RoleSiteAdmin     = "site-admin"
	RoleRegionalAdmin = "regional-admin"
)

// RoleDefinition describes a role with its display name and permissions.
type RoleDefinition struct {
	Code        string
	Display     string
	Permissions []string
	SiteScope   string // "local" or "regional"
}

var roleDefinitions = map[string]RoleDefinition{
	RoleCHW: {
		Code:    RoleCHW,
		Display: "Community Health Worker",
		Permissions: []string{
			PermPatientRead, PermPatientWrite, PermPatientMatch,
			PermEncounterRead, PermEncounterWrite,
			PermObservationRead, PermObservationWrite,
			PermConditionRead,
			PermAllergyRead,
			PermMedicationRead,
			PermFormularyRead,
			PermSyncStatus,
			PermAlertRead,
			PermFlagRead,
		},
		SiteScope: "local",
	},
	RoleNurse: {
		Code:    RoleNurse,
		Display: "Nurse",
		Permissions: []string{
			PermPatientRead, PermPatientWrite, PermPatientMatch,
			PermEncounterRead, PermEncounterWrite,
			PermObservationRead, PermObservationWrite,
			PermConditionRead, PermConditionWrite,
			PermAllergyRead, PermAllergyWrite,
			PermMedicationRead, PermMedicationWrite,
			PermFormularyRead,
			PermSyncStatus,
			PermConflictRead,
			PermAlertRead, PermAlertAcknowledge,
			PermSupplyRead,
			PermFlagRead, PermFlagWrite,
		},
		SiteScope: "local",
	},
	RolePhysician: {
		Code:    RolePhysician,
		Display: "Physician",
		Permissions: []string{
			PermPatientRead, PermPatientWrite, PermPatientDelete, PermPatientMatch,
			PermEncounterRead, PermEncounterWrite,
			PermObservationRead, PermObservationWrite,
			PermConditionRead, PermConditionWrite,
			PermAllergyRead, PermAllergyWrite,
			PermMedicationRead, PermMedicationWrite,
			PermFormularyRead, PermFormularyWrite,
			PermSyncStatus, PermSyncTrigger,
			PermConflictRead, PermConflictResolve,
			PermAlertRead, PermAlertAcknowledge,
			PermAnchorRead,
			PermSupplyRead,
			PermFlagRead, PermFlagWrite,
		},
		SiteScope: "local",
	},
	RoleSiteAdmin: {
		Code:    RoleSiteAdmin,
		Display: "Site Administrator",
		Permissions: []string{
			PermPatientRead, PermPatientWrite, PermPatientDelete, PermPatientMatch,
			PermEncounterRead, PermEncounterWrite,
			PermObservationRead, PermObservationWrite,
			PermConditionRead, PermConditionWrite,
			PermAllergyRead, PermAllergyWrite,
			PermMedicationRead, PermMedicationWrite,
			PermFormularyRead, PermFormularyWrite,
			PermSyncStatus, PermSyncTrigger, PermSyncBundle,
			PermConflictRead, PermConflictResolve,
			PermAlertRead, PermAlertAcknowledge,
			PermAnchorRead, PermAnchorTrigger,
			PermSupplyRead, PermSupplyWrite,
			PermDeviceManage, PermRoleAssign,
			PermFlagRead, PermFlagWrite,
		},
		SiteScope: "local",
	},
	RoleRegionalAdmin: {
		Code:    RoleRegionalAdmin,
		Display: "Regional Administrator",
		Permissions: []string{
			PermPatientRead, PermPatientWrite, PermPatientDelete, PermPatientMatch,
			PermEncounterRead, PermEncounterWrite,
			PermObservationRead, PermObservationWrite,
			PermConditionRead, PermConditionWrite,
			PermAllergyRead, PermAllergyWrite,
			PermMedicationRead, PermMedicationWrite,
			PermFormularyRead, PermFormularyWrite,
			PermSyncStatus, PermSyncTrigger, PermSyncBundle,
			PermConflictRead, PermConflictResolve,
			PermAlertRead, PermAlertAcknowledge,
			PermAnchorRead, PermAnchorTrigger,
			PermSupplyRead, PermSupplyWrite,
			PermDeviceManage, PermRoleAssign,
			PermFlagRead, PermFlagWrite,
		},
		SiteScope: "regional",
	},
}

// HasPermission checks if a role has a specific permission.
func HasPermission(role, permission string) bool {
	def, ok := roleDefinitions[role]
	if !ok {
		return false
	}
	for _, p := range def.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetRole returns the role definition for a given role code.
func GetRole(code string) (RoleDefinition, bool) {
	def, ok := roleDefinitions[code]
	return def, ok
}

// AllRoles returns all role definitions.
func AllRoles() []RoleDefinition {
	roles := make([]RoleDefinition, 0, len(roleDefinitions))
	for _, def := range roleDefinitions {
		roles = append(roles, def)
	}
	return roles
}

// ValidRole returns true if the role code is a valid predefined role.
func ValidRole(code string) bool {
	_, ok := roleDefinitions[code]
	return ok
}
