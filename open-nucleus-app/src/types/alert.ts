/** A Sentinel alert (outbreak signal, medication conflict, etc.). */
export interface AlertDetail {
  id: string;
  severity: string;
  type: string;
  message: string;
  patient_id?: string;
  source?: string;
  created_at: string;
  acknowledged: boolean;
}

/** Summary counts for the alerts dashboard. */
export interface AlertSummary {
  total: number;
  critical: number;
  warning: number;
  info: number;
  unacknowledged: number;
}
