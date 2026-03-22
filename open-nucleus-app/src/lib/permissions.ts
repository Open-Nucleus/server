/**
 * RBAC permission constants mirroring the Go backend (internal/model/rbac.go).
 */
export const Permission = {
  PatientRead: 'patient:read',
  PatientWrite: 'patient:write',
  EncounterRead: 'encounter:read',
  EncounterWrite: 'encounter:write',
  ObservationRead: 'observation:read',
  ObservationWrite: 'observation:write',
  ConditionRead: 'condition:read',
  ConditionWrite: 'condition:write',
  MedicationRead: 'medication:read',
  MedicationWrite: 'medication:write',
  AllergyRead: 'allergy:read',
  AllergyWrite: 'allergy:write',
  ConflictRead: 'conflict:read',
  ConflictResolve: 'conflict:resolve',
  AlertRead: 'alert:read',
  AlertWrite: 'alert:write',
  SyncRead: 'sync:read',
  SyncTrigger: 'sync:trigger',
  FormularyRead: 'formulary:read',
  FormularyWrite: 'formulary:write',
  AnchorRead: 'anchor:read',
  AnchorTrigger: 'anchor:trigger',
  SupplyRead: 'supply:read',
  SupplyWrite: 'supply:write',
  SmartLaunch: 'smart:launch',
  SmartRegister: 'smart:register',
  DeviceManage: 'device:manage',
  ConsentRead: 'consent:read',
  ConsentWrite: 'consent:write',
} as const;

export type PermissionValue = (typeof Permission)[keyof typeof Permission];

/**
 * Role constants mirroring the Go backend.
 */
export const Role = {
  CommunityHealthWorker: 'community_health_worker',
  Nurse: 'nurse',
  Physician: 'physician',
  SiteAdministrator: 'site_administrator',
  RegionalAdministrator: 'regional_administrator',
} as const;

export type RoleValue = (typeof Role)[keyof typeof Role];

/**
 * Maps each role to its allowed permissions (mirrors Go RolePermissions).
 */
export const RolePermissions: Record<RoleValue, PermissionValue[]> = {
  [Role.CommunityHealthWorker]: [
    Permission.PatientRead,
    Permission.ObservationRead,
    Permission.ObservationWrite,
    Permission.AlertRead,
    Permission.SyncRead,
  ],
  [Role.Nurse]: [
    Permission.PatientRead,
    Permission.EncounterRead,
    Permission.EncounterWrite,
    Permission.ObservationRead,
    Permission.ObservationWrite,
    Permission.MedicationRead,
    Permission.ConditionRead,
    Permission.AllergyRead,
    Permission.AlertRead,
    Permission.SyncRead,
  ],
  [Role.Physician]: [
    Permission.PatientRead,
    Permission.PatientWrite,
    Permission.EncounterRead,
    Permission.EncounterWrite,
    Permission.ObservationRead,
    Permission.ObservationWrite,
    Permission.ConditionRead,
    Permission.ConditionWrite,
    Permission.MedicationRead,
    Permission.MedicationWrite,
    Permission.AllergyRead,
    Permission.AllergyWrite,
    Permission.ConflictRead,
    Permission.ConflictResolve,
    Permission.AlertRead,
    Permission.AlertWrite,
    Permission.SyncRead,
    Permission.FormularyRead,
    Permission.AnchorRead,
    Permission.SupplyRead,
    Permission.SmartLaunch,
    Permission.ConsentRead,
    Permission.ConsentWrite,
  ],
  [Role.SiteAdministrator]: [
    Permission.PatientRead,
    Permission.PatientWrite,
    Permission.EncounterRead,
    Permission.EncounterWrite,
    Permission.ObservationRead,
    Permission.ObservationWrite,
    Permission.ConditionRead,
    Permission.ConditionWrite,
    Permission.MedicationRead,
    Permission.MedicationWrite,
    Permission.AllergyRead,
    Permission.AllergyWrite,
    Permission.ConflictRead,
    Permission.ConflictResolve,
    Permission.AlertRead,
    Permission.AlertWrite,
    Permission.SyncRead,
    Permission.SyncTrigger,
    Permission.FormularyRead,
    Permission.FormularyWrite,
    Permission.AnchorRead,
    Permission.AnchorTrigger,
    Permission.SupplyRead,
    Permission.SupplyWrite,
    Permission.DeviceManage,
    Permission.SmartLaunch,
    Permission.SmartRegister,
    Permission.ConsentRead,
    Permission.ConsentWrite,
  ],
  [Role.RegionalAdministrator]: [
    Permission.PatientRead,
    Permission.PatientWrite,
    Permission.EncounterRead,
    Permission.EncounterWrite,
    Permission.ObservationRead,
    Permission.ObservationWrite,
    Permission.ConditionRead,
    Permission.ConditionWrite,
    Permission.MedicationRead,
    Permission.MedicationWrite,
    Permission.AllergyRead,
    Permission.AllergyWrite,
    Permission.ConflictRead,
    Permission.ConflictResolve,
    Permission.AlertRead,
    Permission.AlertWrite,
    Permission.SyncRead,
    Permission.SyncTrigger,
    Permission.FormularyRead,
    Permission.FormularyWrite,
    Permission.AnchorRead,
    Permission.AnchorTrigger,
    Permission.SupplyRead,
    Permission.SupplyWrite,
    Permission.DeviceManage,
    Permission.SmartLaunch,
    Permission.SmartRegister,
    Permission.ConsentRead,
    Permission.ConsentWrite,
  ],
};

/** Check if a role includes a specific permission. */
export function hasPermission(role: string, permission: string): boolean {
  const perms = RolePermissions[role as RoleValue];
  if (!perms) return false;
  return perms.includes(permission as PermissionValue);
}
