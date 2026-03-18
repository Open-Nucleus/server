/// RBAC permissions and role definitions mirroring the Go backend
/// (`internal/model/rbac.go`).
///
/// The [rolePermissions] map is the single source of truth on the Flutter side
/// and **must** stay in sync with the backend whenever roles or permissions
/// change.
class Permissions {
  Permissions._();

  // ---------------------------------------------------------------------------
  // Permission constants
  // ---------------------------------------------------------------------------

  static const String patientRead = 'patient:read';
  static const String patientWrite = 'patient:write';
  static const String encounterRead = 'encounter:read';
  static const String encounterWrite = 'encounter:write';
  static const String observationRead = 'observation:read';
  static const String observationWrite = 'observation:write';
  static const String conditionRead = 'condition:read';
  static const String conditionWrite = 'condition:write';
  static const String medicationRead = 'medication:read';
  static const String medicationWrite = 'medication:write';
  static const String allergyRead = 'allergy:read';
  static const String allergyWrite = 'allergy:write';
  static const String conflictRead = 'conflict:read';
  static const String conflictResolve = 'conflict:resolve';
  static const String alertRead = 'alert:read';
  static const String alertWrite = 'alert:write';
  static const String syncRead = 'sync:read';
  static const String syncTrigger = 'sync:trigger';
  static const String formularyRead = 'formulary:read';
  static const String formularyWrite = 'formulary:write';
  static const String anchorRead = 'anchor:read';
  static const String anchorTrigger = 'anchor:trigger';
  static const String supplyRead = 'supply:read';
  static const String supplyWrite = 'supply:write';
  static const String smartLaunch = 'smart:launch';
  static const String smartRegister = 'smart:register';
  static const String deviceManage = 'device:manage';
  static const String consentRead = 'consent:read';
  static const String consentWrite = 'consent:write';

  // ---------------------------------------------------------------------------
  // Role constants
  // ---------------------------------------------------------------------------

  static const String roleCHW = 'community_health_worker';
  static const String roleNurse = 'nurse';
  static const String rolePhysician = 'physician';
  static const String roleSiteAdmin = 'site_administrator';
  static const String roleRegionalAdmin = 'regional_administrator';

  static const List<String> allRoles = [
    roleCHW,
    roleNurse,
    rolePhysician,
    roleSiteAdmin,
    roleRegionalAdmin,
  ];

  // ---------------------------------------------------------------------------
  // Role -> permissions mapping (mirrors internal/model/rbac.go exactly)
  // ---------------------------------------------------------------------------

  static const Map<String, List<String>> rolePermissions = {
    roleCHW: [
      patientRead,
      observationRead,
      observationWrite,
      alertRead,
      syncRead,
    ],
    roleNurse: [
      patientRead,
      encounterRead,
      encounterWrite,
      observationRead,
      observationWrite,
      medicationRead,
      conditionRead,
      allergyRead,
      alertRead,
      syncRead,
    ],
    rolePhysician: [
      patientRead,
      patientWrite,
      encounterRead,
      encounterWrite,
      observationRead,
      observationWrite,
      conditionRead,
      conditionWrite,
      medicationRead,
      medicationWrite,
      allergyRead,
      allergyWrite,
      conflictRead,
      conflictResolve,
      alertRead,
      alertWrite,
      syncRead,
      formularyRead,
      anchorRead,
      supplyRead,
      smartLaunch,
      consentRead,
      consentWrite,
    ],
    roleSiteAdmin: [
      patientRead,
      patientWrite,
      encounterRead,
      encounterWrite,
      observationRead,
      observationWrite,
      conditionRead,
      conditionWrite,
      medicationRead,
      medicationWrite,
      allergyRead,
      allergyWrite,
      conflictRead,
      conflictResolve,
      alertRead,
      alertWrite,
      syncRead,
      syncTrigger,
      formularyRead,
      formularyWrite,
      anchorRead,
      anchorTrigger,
      supplyRead,
      supplyWrite,
      deviceManage,
      smartLaunch,
      smartRegister,
      consentRead,
      consentWrite,
    ],
    roleRegionalAdmin: [
      patientRead,
      patientWrite,
      encounterRead,
      encounterWrite,
      observationRead,
      observationWrite,
      conditionRead,
      conditionWrite,
      medicationRead,
      medicationWrite,
      allergyRead,
      allergyWrite,
      conflictRead,
      conflictResolve,
      alertRead,
      alertWrite,
      syncRead,
      syncTrigger,
      formularyRead,
      formularyWrite,
      anchorRead,
      anchorTrigger,
      supplyRead,
      supplyWrite,
      deviceManage,
      smartLaunch,
      smartRegister,
      consentRead,
      consentWrite,
    ],
  };

  /// Returns `true` if [role] includes the given [permission].
  static bool hasPermission(String role, String permission) {
    final perms = rolePermissions[role];
    if (perms == null) return false;
    return perms.contains(permission);
  }

  /// Returns the display-friendly name for a role.
  static String roleDisplayName(String role) {
    switch (role) {
      case roleCHW:
        return 'Community Health Worker';
      case roleNurse:
        return 'Nurse';
      case rolePhysician:
        return 'Physician';
      case roleSiteAdmin:
        return 'Site Administrator';
      case roleRegionalAdmin:
        return 'Regional Administrator';
      default:
        return role;
    }
  }
}
