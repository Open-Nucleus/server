package model

// Roles as defined in spec section 12.2.
const (
	RoleCHW             = "community_health_worker"
	RoleNurse           = "nurse"
	RolePhysician       = "physician"
	RoleSiteAdmin       = "site_administrator"
	RoleRegionalAdmin   = "regional_administrator"
)

// Permissions used throughout the system.
const (
	PermPatientRead       = "patient:read"
	PermPatientWrite      = "patient:write"
	PermEncounterRead     = "encounter:read"
	PermEncounterWrite    = "encounter:write"
	PermObservationRead   = "observation:read"
	PermObservationWrite  = "observation:write"
	PermConditionRead     = "condition:read"
	PermConditionWrite    = "condition:write"
	PermMedicationRead    = "medication:read"
	PermMedicationWrite   = "medication:write"
	PermAllergyRead       = "allergy:read"
	PermAllergyWrite      = "allergy:write"
	PermConflictRead      = "conflict:read"
	PermConflictResolve   = "conflict:resolve"
	PermAlertRead         = "alert:read"
	PermAlertWrite        = "alert:write"
	PermSyncRead          = "sync:read"
	PermSyncTrigger       = "sync:trigger"
	PermFormularyRead     = "formulary:read"
	PermFormularyWrite    = "formulary:write"
	PermAnchorRead        = "anchor:read"
	PermAnchorTrigger     = "anchor:trigger"
	PermSupplyRead        = "supply:read"
	PermSupplyWrite       = "supply:write"
)

// RolePermissions maps each role to its allowed permissions per spec section 12.2.
var RolePermissions = map[string][]string{
	RoleCHW: {
		PermPatientRead, PermObservationRead, PermObservationWrite,
		PermAlertRead, PermSyncRead,
	},
	RoleNurse: {
		PermPatientRead, PermEncounterRead, PermEncounterWrite,
		PermObservationRead, PermObservationWrite,
		PermMedicationRead, PermConditionRead,
		PermAllergyRead, PermAlertRead, PermSyncRead,
	},
	RolePhysician: {
		PermPatientRead, PermPatientWrite,
		PermEncounterRead, PermEncounterWrite,
		PermObservationRead, PermObservationWrite,
		PermConditionRead, PermConditionWrite,
		PermMedicationRead, PermMedicationWrite,
		PermAllergyRead, PermAllergyWrite,
		PermConflictRead, PermConflictResolve,
		PermAlertRead, PermAlertWrite,
		PermSyncRead, PermFormularyRead,
		PermAnchorRead, PermSupplyRead,
	},
	RoleSiteAdmin: {
		PermPatientRead, PermPatientWrite,
		PermEncounterRead, PermEncounterWrite,
		PermObservationRead, PermObservationWrite,
		PermConditionRead, PermConditionWrite,
		PermMedicationRead, PermMedicationWrite,
		PermAllergyRead, PermAllergyWrite,
		PermConflictRead, PermConflictResolve,
		PermAlertRead, PermAlertWrite,
		PermSyncRead, PermSyncTrigger,
		PermFormularyRead, PermFormularyWrite,
		PermAnchorRead, PermAnchorTrigger,
		PermSupplyRead, PermSupplyWrite,
	},
	RoleRegionalAdmin: {
		PermPatientRead, PermPatientWrite,
		PermEncounterRead, PermEncounterWrite,
		PermObservationRead, PermObservationWrite,
		PermConditionRead, PermConditionWrite,
		PermMedicationRead, PermMedicationWrite,
		PermAllergyRead, PermAllergyWrite,
		PermConflictRead, PermConflictResolve,
		PermAlertRead, PermAlertWrite,
		PermSyncRead, PermSyncTrigger,
		PermFormularyRead, PermFormularyWrite,
		PermAnchorRead, PermAnchorTrigger,
		PermSupplyRead, PermSupplyWrite,
	},
}

// HasPermission checks if a role includes a specific permission.
func HasPermission(role, permission string) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}
