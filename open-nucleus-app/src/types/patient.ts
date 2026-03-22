/** Lightweight patient summary returned from list/search endpoints. */
export interface PatientSummary {
  id: string;
  family_name: string;
  given_names: string[];
  gender: string;
  birth_date: string;
  active: boolean;
  site_id?: string;
  last_updated?: string;
  has_alerts?: boolean;
}

/** Raw FHIR Patient resource. */
export type PatientBundle = Record<string, unknown>;

/** Response from creating or updating a resource (commit info). */
export interface WriteResponse {
  resource_id: string;
  commit: string;
}
