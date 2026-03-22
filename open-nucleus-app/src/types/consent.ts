/** Summary of a FHIR Consent resource. */
export interface ConsentSummary {
  id: string;
  status: string;
  grantor?: string;
  scope?: string;
  period?: {
    start: string;
    end?: string;
  };
  created_at: string;
}

/** Request body for granting consent. */
export interface ConsentGrantRequest {
  patient_id: string;
  provider_id: string;
  scope: string;
  period: {
    start: string;
    end?: string;
  };
}
