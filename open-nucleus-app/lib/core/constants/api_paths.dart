/// Every REST endpoint path exposed by the Open Nucleus API gateway.
///
/// Static strings are used for fixed paths; static methods with parameters
/// are used for paths that contain resource IDs.  All paths start with `/`.
class ApiPaths {
  ApiPaths._(); // non-instantiable

  // ---------------------------------------------------------------------------
  // Health
  // ---------------------------------------------------------------------------

  static const String health = '/health';

  // ---------------------------------------------------------------------------
  // Auth
  // ---------------------------------------------------------------------------

  static const String authLogin = '/api/v1/auth/login';
  static const String authRefresh = '/api/v1/auth/refresh';
  static const String authLogout = '/api/v1/auth/logout';
  static const String authWhoami = '/api/v1/auth/whoami';

  // ---------------------------------------------------------------------------
  // Patients
  // ---------------------------------------------------------------------------

  static const String patients = '/api/v1/patients';
  static const String patientsSearch = '/api/v1/patients/search';
  static const String patientsMatch = '/api/v1/patients/match';

  static String patient(String id) => '/api/v1/patients/$id';
  static String patientHistory(String id) => '/api/v1/patients/$id/history';
  static String patientTimeline(String id) => '/api/v1/patients/$id/timeline';
  static String patientErase(String id) => '/api/v1/patients/$id/erase';

  // ---------------------------------------------------------------------------
  // Encounters
  // ---------------------------------------------------------------------------

  static String encounters(String patientId) =>
      '/api/v1/patients/$patientId/encounters';

  static String encounter(String patientId, String encounterId) =>
      '/api/v1/patients/$patientId/encounters/$encounterId';

  // ---------------------------------------------------------------------------
  // Observations
  // ---------------------------------------------------------------------------

  static String observations(String patientId) =>
      '/api/v1/patients/$patientId/observations';

  static String observation(String patientId, String observationId) =>
      '/api/v1/patients/$patientId/observations/$observationId';

  // ---------------------------------------------------------------------------
  // Conditions
  // ---------------------------------------------------------------------------

  static String conditions(String patientId) =>
      '/api/v1/patients/$patientId/conditions';

  static String condition(String patientId, String conditionId) =>
      '/api/v1/patients/$patientId/conditions/$conditionId';

  // ---------------------------------------------------------------------------
  // Medication Requests
  // ---------------------------------------------------------------------------

  static String medicationRequests(String patientId) =>
      '/api/v1/patients/$patientId/medication-requests';

  static String medicationRequest(String patientId, String requestId) =>
      '/api/v1/patients/$patientId/medication-requests/$requestId';

  // ---------------------------------------------------------------------------
  // Allergy Intolerances
  // ---------------------------------------------------------------------------

  static String allergyIntolerances(String patientId) =>
      '/api/v1/patients/$patientId/allergy-intolerances';

  static String allergyIntolerance(String patientId, String allergyId) =>
      '/api/v1/patients/$patientId/allergy-intolerances/$allergyId';

  // ---------------------------------------------------------------------------
  // Immunizations
  // ---------------------------------------------------------------------------

  static String immunizations(String patientId) =>
      '/api/v1/patients/$patientId/immunizations';

  static String immunization(String patientId, String immunizationId) =>
      '/api/v1/patients/$patientId/immunizations/$immunizationId';

  // ---------------------------------------------------------------------------
  // Procedures
  // ---------------------------------------------------------------------------

  static String procedures(String patientId) =>
      '/api/v1/patients/$patientId/procedures';

  static String procedure(String patientId, String procedureId) =>
      '/api/v1/patients/$patientId/procedures/$procedureId';

  // ---------------------------------------------------------------------------
  // Consents (patient-scoped)
  // ---------------------------------------------------------------------------

  static String patientConsents(String patientId) =>
      '/api/v1/patients/$patientId/consents';

  // ---------------------------------------------------------------------------
  // Consents (top-level)
  // ---------------------------------------------------------------------------

  static String consent(String consentId) => '/api/v1/consents/$consentId';

  static String consentVc(String consentId) =>
      '/api/v1/consents/$consentId/vc';

  // ---------------------------------------------------------------------------
  // Practitioners
  // ---------------------------------------------------------------------------

  static const String practitioners = '/api/v1/practitioners';

  static String practitioner(String id) => '/api/v1/practitioners/$id';

  // ---------------------------------------------------------------------------
  // Organizations
  // ---------------------------------------------------------------------------

  static const String organizations = '/api/v1/organizations';

  static String organization(String id) => '/api/v1/organizations/$id';

  // ---------------------------------------------------------------------------
  // Locations
  // ---------------------------------------------------------------------------

  static const String locations = '/api/v1/locations';

  static String location(String id) => '/api/v1/locations/$id';

  // ---------------------------------------------------------------------------
  // Sync
  // ---------------------------------------------------------------------------

  static const String syncStatus = '/api/v1/sync/status';
  static const String syncPeers = '/api/v1/sync/peers';
  static const String syncTrigger = '/api/v1/sync/trigger';
  static const String syncHistory = '/api/v1/sync/history';
  static const String syncBundleExport = '/api/v1/sync/bundle/export';
  static const String syncBundleImport = '/api/v1/sync/bundle/import';

  // ---------------------------------------------------------------------------
  // Conflicts
  // ---------------------------------------------------------------------------

  static const String conflicts = '/api/v1/conflicts';

  static String conflict(String id) => '/api/v1/conflicts/$id';
  static String conflictResolve(String id) => '/api/v1/conflicts/$id/resolve';
  static String conflictDefer(String id) => '/api/v1/conflicts/$id/defer';

  // ---------------------------------------------------------------------------
  // Alerts
  // ---------------------------------------------------------------------------

  static const String alerts = '/api/v1/alerts';
  static const String alertsSummary = '/api/v1/alerts/summary';

  static String alert(String id) => '/api/v1/alerts/$id';
  static String alertAcknowledge(String id) =>
      '/api/v1/alerts/$id/acknowledge';
  static String alertDismiss(String id) => '/api/v1/alerts/$id/dismiss';

  // ---------------------------------------------------------------------------
  // Formulary
  // ---------------------------------------------------------------------------

  static const String formularyMedications = '/api/v1/formulary/medications';

  static String formularyMedicationsByCategory(String category) =>
      '/api/v1/formulary/medications/category/$category';

  static String formularyMedication(String code) =>
      '/api/v1/formulary/medications/$code';

  static const String formularyCheckInteractions =
      '/api/v1/formulary/check-interactions';
  static const String formularyCheckAllergies =
      '/api/v1/formulary/check-allergies';
  static const String formularyDosingValidate =
      '/api/v1/formulary/dosing/validate';
  static const String formularyDosingOptions =
      '/api/v1/formulary/dosing/options';
  static const String formularyDosingSchedule =
      '/api/v1/formulary/dosing/schedule';

  static String formularyStock(String siteId, String medicationCode) =>
      '/api/v1/formulary/stock/$siteId/$medicationCode';

  static String formularyStockPrediction(
          String siteId, String medicationCode) =>
      '/api/v1/formulary/stock/$siteId/$medicationCode/prediction';

  static const String formularyDeliveries = '/api/v1/formulary/deliveries';

  static String formularyRedistribution(String medicationCode) =>
      '/api/v1/formulary/redistribution/$medicationCode';

  static const String formularyInfo = '/api/v1/formulary/info';

  // ---------------------------------------------------------------------------
  // Anchor
  // ---------------------------------------------------------------------------

  static const String anchorStatus = '/api/v1/anchor/status';
  static const String anchorVerify = '/api/v1/anchor/verify';
  static const String anchorHistory = '/api/v1/anchor/history';
  static const String anchorTrigger = '/api/v1/anchor/trigger';
  static const String anchorDidNode = '/api/v1/anchor/did/node';

  static String anchorDidDevice(String deviceId) =>
      '/api/v1/anchor/did/device/$deviceId';

  static const String anchorDidResolve = '/api/v1/anchor/did/resolve';
  static const String anchorCredentialsIssue =
      '/api/v1/anchor/credentials/issue';
  static const String anchorCredentialsVerify =
      '/api/v1/anchor/credentials/verify';
  static const String anchorCredentials = '/api/v1/anchor/credentials';
  static const String anchorBackends = '/api/v1/anchor/backends';

  static String anchorBackend(String name) =>
      '/api/v1/anchor/backends/$name';

  static const String anchorQueue = '/api/v1/anchor/queue';

  // ---------------------------------------------------------------------------
  // Supply
  // ---------------------------------------------------------------------------

  static const String supplyInventory = '/api/v1/supply/inventory';

  static String supplyInventoryItem(String itemCode) =>
      '/api/v1/supply/inventory/$itemCode';

  static const String supplyDeliveries = '/api/v1/supply/deliveries';
  static const String supplyPredictions = '/api/v1/supply/predictions';
  static const String supplyRedistribution = '/api/v1/supply/redistribution';

  // ---------------------------------------------------------------------------
  // SMART on FHIR
  // ---------------------------------------------------------------------------

  static const String smartClients = '/api/v1/smart/clients';

  static String smartClient(String id) => '/api/v1/smart/clients/$id';
}
