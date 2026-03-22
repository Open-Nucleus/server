/** A medication in the formulary. */
export interface MedicationDetail {
  code: string;
  display: string;
  form?: string;
  route?: string;
  category?: string;
  available: boolean;
  who_essential?: boolean;
  therapeutic_class?: string;
  common_frequencies?: string[];
  strength?: string;
  unit?: string;
}
