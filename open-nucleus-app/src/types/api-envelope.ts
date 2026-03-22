/** Standard response wrapper from the Go backend. */
export interface ApiEnvelope<T> {
  status: 'success' | 'error';
  data?: T;
  error?: ErrorBody;
  pagination?: Pagination;
  warnings?: Warning[];
  git?: GitInfo;
  meta?: Meta;
}

export interface ErrorBody {
  code: string;
  message: string;
  details?: unknown;
}

export interface Pagination {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface Warning {
  severity: string;
  type: string;
  description: string;
  interacting_medication?: string;
  source?: string;
}

export interface GitInfo {
  commit: string;
  message: string;
}

export interface Meta {
  request_id: string;
  duration_ms: number;
  node_id: string;
}
