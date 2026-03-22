/** A vital sign definition with its LOINC code, display name, and default unit. */
export interface VitalSignDef {
  code: string;
  display: string;
  unit: string;
}

/** Common vital sign LOINC codes used in observation forms. */
export const VITAL_SIGNS: VitalSignDef[] = [
  { code: '8310-5', display: 'Body Temperature', unit: 'degC' },
  { code: '8867-4', display: 'Heart Rate', unit: '/min' },
  { code: '85354-9', display: 'Blood Pressure Panel', unit: 'mmHg' },
  { code: '8480-6', display: 'Systolic Blood Pressure', unit: 'mmHg' },
  { code: '8462-4', display: 'Diastolic Blood Pressure', unit: 'mmHg' },
  { code: '29463-7', display: 'Body Weight', unit: 'kg' },
  { code: '8302-2', display: 'Body Height', unit: 'cm' },
  { code: '2708-6', display: 'Oxygen Saturation (SpO2)', unit: '%' },
  { code: '9279-1', display: 'Respiratory Rate', unit: '/min' },
];

/** FHIR Encounter status codes. */
export const ENCOUNTER_STATUSES = [
  'planned',
  'arrived',
  'triaged',
  'in-progress',
  'onleave',
  'finished',
  'cancelled',
] as const;

/** FHIR Condition clinical status codes. */
export const CONDITION_STATUSES = [
  'active',
  'recurrence',
  'relapse',
  'inactive',
  'remission',
  'resolved',
] as const;

/** FHIR AllergyIntolerance criticality codes. */
export const ALLERGY_CRITICALITIES = [
  'low',
  'high',
  'unable-to-assess',
] as const;

/** FHIR MedicationRequest status codes. */
export const MEDICATION_STATUSES = [
  'active',
  'on-hold',
  'cancelled',
  'completed',
  'entered-in-error',
  'stopped',
  'draft',
  'unknown',
] as const;

/** FHIR Immunization status codes. */
export const IMMUNIZATION_STATUSES = [
  'completed',
  'entered-in-error',
  'not-done',
] as const;

/** FHIR Procedure status codes. */
export const PROCEDURE_STATUSES = [
  'preparation',
  'in-progress',
  'not-done',
  'on-hold',
  'stopped',
  'completed',
  'entered-in-error',
  'unknown',
] as const;

/** FHIR Observation status codes. */
export const OBSERVATION_STATUSES = [
  'registered',
  'preliminary',
  'final',
  'amended',
  'corrected',
  'cancelled',
  'entered-in-error',
  'unknown',
] as const;

/** FHIR administrative gender codes. */
export const ADMINISTRATIVE_GENDERS = [
  'male',
  'female',
  'other',
  'unknown',
] as const;

/** FHIR AllergyIntolerance category codes. */
export const ALLERGY_CATEGORIES = [
  'food',
  'medication',
  'environment',
  'biologic',
] as const;

/** FHIR AllergyIntolerance type codes. */
export const ALLERGY_TYPES = [
  'allergy',
  'intolerance',
] as const;

/** FHIR Consent status codes. */
export const CONSENT_STATUSES = [
  'draft',
  'proposed',
  'active',
  'rejected',
  'inactive',
  'entered-in-error',
] as const;
