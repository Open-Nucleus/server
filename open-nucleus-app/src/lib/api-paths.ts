/**
 * All REST API endpoint paths for the Open Nucleus Go backend.
 * Mirrors the chi router defined in internal/router/router.go.
 */
export const API = {
  // ---------- public ----------
  health: '/health',
  fhirMetadata: '/fhir/metadata',
  smartConfiguration: '/.well-known/smart-configuration',

  // ---------- auth ----------
  auth: {
    login: '/api/v1/auth/login',
    refresh: '/api/v1/auth/refresh',
    logout: '/api/v1/auth/logout',
    whoami: '/api/v1/auth/whoami',
  },

  // ---------- SMART OAuth2 ----------
  smart: {
    authorize: '/auth/smart/authorize',
    token: '/auth/smart/token',
    revoke: '/auth/smart/revoke',
    introspect: '/auth/smart/introspect',
    register: '/auth/smart/register',
    launch: '/auth/smart/launch',
  },

  // ---------- patients ----------
  patients: {
    list: '/api/v1/patients',
    create: '/api/v1/patients',
    search: '/api/v1/patients/search',
    match: '/api/v1/patients/match',
    get: (id: string) => `/api/v1/patients/${id}` as const,
    update: (id: string) => `/api/v1/patients/${id}` as const,
    delete: (id: string) => `/api/v1/patients/${id}` as const,
    history: (id: string) => `/api/v1/patients/${id}/history` as const,
    timeline: (id: string) => `/api/v1/patients/${id}/timeline` as const,
    erase: (id: string) => `/api/v1/patients/${id}/erase` as const,

    // encounters
    encounters: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/encounters` as const,
      create: (patientId: string) =>
        `/api/v1/patients/${patientId}/encounters` as const,
      get: (patientId: string, encounterId: string) =>
        `/api/v1/patients/${patientId}/encounters/${encounterId}` as const,
      update: (patientId: string, encounterId: string) =>
        `/api/v1/patients/${patientId}/encounters/${encounterId}` as const,
    },

    // observations
    observations: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/observations` as const,
      create: (patientId: string) =>
        `/api/v1/patients/${patientId}/observations` as const,
      get: (patientId: string, observationId: string) =>
        `/api/v1/patients/${patientId}/observations/${observationId}` as const,
    },

    // conditions
    conditions: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/conditions` as const,
      create: (patientId: string) =>
        `/api/v1/patients/${patientId}/conditions` as const,
      update: (patientId: string, conditionId: string) =>
        `/api/v1/patients/${patientId}/conditions/${conditionId}` as const,
    },

    // medication requests
    medicationRequests: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/medication-requests` as const,
      create: (patientId: string) =>
        `/api/v1/patients/${patientId}/medication-requests` as const,
      update: (patientId: string, medicationRequestId: string) =>
        `/api/v1/patients/${patientId}/medication-requests/${medicationRequestId}` as const,
    },

    // allergy intolerances
    allergyIntolerances: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/allergy-intolerances` as const,
      create: (patientId: string) =>
        `/api/v1/patients/${patientId}/allergy-intolerances` as const,
      update: (patientId: string, allergyId: string) =>
        `/api/v1/patients/${patientId}/allergy-intolerances/${allergyId}` as const,
    },

    // immunizations
    immunizations: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/immunizations` as const,
      create: (patientId: string) =>
        `/api/v1/patients/${patientId}/immunizations` as const,
      get: (patientId: string, immunizationId: string) =>
        `/api/v1/patients/${patientId}/immunizations/${immunizationId}` as const,
    },

    // procedures
    procedures: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/procedures` as const,
      create: (patientId: string) =>
        `/api/v1/patients/${patientId}/procedures` as const,
      get: (patientId: string, procedureId: string) =>
        `/api/v1/patients/${patientId}/procedures/${procedureId}` as const,
    },

    // consents (patient-scoped)
    consents: {
      list: (patientId: string) =>
        `/api/v1/patients/${patientId}/consents` as const,
      grant: (patientId: string) =>
        `/api/v1/patients/${patientId}/consents` as const,
    },
  },

  // ---------- consent management (top-level) ----------
  consents: {
    revoke: (consentId: string) =>
      `/api/v1/consents/${consentId}` as const,
    issueVC: (consentId: string) =>
      `/api/v1/consents/${consentId}/vc` as const,
  },

  // ---------- practitioners ----------
  practitioners: {
    list: '/api/v1/practitioners',
    create: '/api/v1/practitioners',
    get: (id: string) => `/api/v1/practitioners/${id}` as const,
    update: (id: string) => `/api/v1/practitioners/${id}` as const,
  },

  // ---------- organizations ----------
  organizations: {
    list: '/api/v1/organizations',
    create: '/api/v1/organizations',
    get: (id: string) => `/api/v1/organizations/${id}` as const,
  },

  // ---------- locations ----------
  locations: {
    list: '/api/v1/locations',
    create: '/api/v1/locations',
    get: (id: string) => `/api/v1/locations/${id}` as const,
  },

  // ---------- sync ----------
  sync: {
    status: '/api/v1/sync/status',
    peers: '/api/v1/sync/peers',
    trigger: '/api/v1/sync/trigger',
    history: '/api/v1/sync/history',
    exportBundle: '/api/v1/sync/bundle/export',
    importBundle: '/api/v1/sync/bundle/import',
  },

  // ---------- conflicts ----------
  conflicts: {
    list: '/api/v1/conflicts',
    get: (id: string) => `/api/v1/conflicts/${id}` as const,
    resolve: (id: string) => `/api/v1/conflicts/${id}/resolve` as const,
    defer: (id: string) => `/api/v1/conflicts/${id}/defer` as const,
  },

  // ---------- alerts ----------
  alerts: {
    list: '/api/v1/alerts',
    summary: '/api/v1/alerts/summary',
    get: (id: string) => `/api/v1/alerts/${id}` as const,
    acknowledge: (id: string) => `/api/v1/alerts/${id}/acknowledge` as const,
    dismiss: (id: string) => `/api/v1/alerts/${id}/dismiss` as const,
  },

  // ---------- formulary ----------
  formulary: {
    medications: '/api/v1/formulary/medications',
    medicationsByCategory: (category: string) =>
      `/api/v1/formulary/medications/category/${category}` as const,
    medication: (code: string) =>
      `/api/v1/formulary/medications/${code}` as const,
    checkInteractions: '/api/v1/formulary/check-interactions',
    checkAllergies: '/api/v1/formulary/check-allergies',
    dosingValidate: '/api/v1/formulary/dosing/validate',
    dosingOptions: '/api/v1/formulary/dosing/options',
    dosingSchedule: '/api/v1/formulary/dosing/schedule',
    stockLevel: (siteId: string, medicationCode: string) =>
      `/api/v1/formulary/stock/${siteId}/${medicationCode}` as const,
    updateStock: (siteId: string, medicationCode: string) =>
      `/api/v1/formulary/stock/${siteId}/${medicationCode}` as const,
    stockPrediction: (siteId: string, medicationCode: string) =>
      `/api/v1/formulary/stock/${siteId}/${medicationCode}/prediction` as const,
    deliveries: '/api/v1/formulary/deliveries',
    redistribution: (medicationCode: string) =>
      `/api/v1/formulary/redistribution/${medicationCode}` as const,
    info: '/api/v1/formulary/info',
  },

  // ---------- anchor ----------
  anchor: {
    status: '/api/v1/anchor/status',
    verify: '/api/v1/anchor/verify',
    history: '/api/v1/anchor/history',
    trigger: '/api/v1/anchor/trigger',
    didNode: '/api/v1/anchor/did/node',
    didDevice: (deviceId: string) =>
      `/api/v1/anchor/did/device/${deviceId}` as const,
    didResolve: '/api/v1/anchor/did/resolve',
    credentialsIssue: '/api/v1/anchor/credentials/issue',
    credentialsVerify: '/api/v1/anchor/credentials/verify',
    credentialsList: '/api/v1/anchor/credentials',
    backends: '/api/v1/anchor/backends',
    backend: (name: string) => `/api/v1/anchor/backends/${name}` as const,
    queue: '/api/v1/anchor/queue',
  },

  // ---------- supply ----------
  supply: {
    inventory: '/api/v1/supply/inventory',
    inventoryItem: (itemCode: string) =>
      `/api/v1/supply/inventory/${itemCode}` as const,
    deliveries: '/api/v1/supply/deliveries',
    predictions: '/api/v1/supply/predictions',
    redistribution: '/api/v1/supply/redistribution',
  },

  // ---------- SMART client management ----------
  smartClients: {
    list: '/api/v1/smart/clients',
    get: (id: string) => `/api/v1/smart/clients/${id}` as const,
    update: (id: string) => `/api/v1/smart/clients/${id}` as const,
    delete: (id: string) => `/api/v1/smart/clients/${id}` as const,
  },

  // ---------- FHIR R4 REST ----------
  fhir: {
    search: (resourceType: string) => `/fhir/${resourceType}` as const,
    read: (resourceType: string, id: string) =>
      `/fhir/${resourceType}/${id}` as const,
    create: (resourceType: string) => `/fhir/${resourceType}` as const,
    update: (resourceType: string, id: string) =>
      `/fhir/${resourceType}/${id}` as const,
    delete: (resourceType: string, id: string) =>
      `/fhir/${resourceType}/${id}` as const,
  },

  // ---------- WebSocket (stub) ----------
  ws: '/api/v1/ws',
} as const;
