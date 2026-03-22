/** Paginated list of clinical resources (encounters, observations, etc.). */
export interface ClinicalListResponse {
  resources: Record<string, unknown>[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}
