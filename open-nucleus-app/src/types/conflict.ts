/** A merge conflict requiring clinician review. */
export interface ConflictDetail {
  id: string;
  type: string;
  severity: string;
  resources: Record<string, unknown>[];
  suggestions: string[];
}
