/** FHIR R4 Coding element. */
export interface Coding {
  system?: string;
  code?: string;
  display?: string;
}

/** FHIR R4 CodeableConcept element. */
export interface CodeableConcept {
  coding?: Coding[];
  text?: string;
}

/** FHIR R4 Reference element. */
export interface FhirReference {
  reference?: string;
  display?: string;
  type?: string;
}

/** FHIR R4 Period element. */
export interface FhirPeriod {
  start?: string;
  end?: string;
}

/** FHIR R4 HumanName element. */
export interface HumanName {
  use?: string;
  family?: string;
  given?: string[];
  prefix?: string[];
}

/** FHIR R4 Quantity element. */
export interface Quantity {
  value?: number;
  comparator?: string;
  unit?: string;
  system?: string;
  code?: string;
}

/** FHIR R4 Identifier element. */
export interface FhirIdentifier {
  system?: string;
  value?: string;
  use?: string;
}

/** FHIR R4 ContactPoint element. */
export interface ContactPoint {
  system?: string;
  value?: string;
  use?: string;
}

/** FHIR R4 Address element. */
export interface FhirAddress {
  use?: string;
  line?: string[];
  city?: string;
  state?: string;
  postalCode?: string;
  country?: string;
}
