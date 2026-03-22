import { useState } from 'react';
import { useParams, useNavigate } from '@tanstack/react-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Edit3,
  Trash2,
  Activity,
  FileText,
  Heart,
  Pill,
  AlertTriangle,
  Syringe,
  Scissors,
  Shield,
  GitCommit,
  Plus,
  ClipboardList,
} from 'lucide-react';
import { apiGet, apiDelete } from '@/lib/api-client';
import { EncounterDialog } from '@/features/patients/dialogs/encounter-dialog';
import { ObservationDialog } from '@/features/patients/dialogs/observation-dialog';
import { ConditionDialog } from '@/features/patients/dialogs/condition-dialog';
import { MedicationRequestDialog } from '@/features/patients/dialogs/medication-request-dialog';
import { AllergyDialog } from '@/features/patients/dialogs/allergy-dialog';
import { ImmunizationDialog } from '@/features/patients/dialogs/immunization-dialog';
import { ProcedureDialog } from '@/features/patients/dialogs/procedure-dialog';
// EraseDialog available at '@/features/patients/dialogs/erase-dialog' if needed
import { API } from '@/lib/api-paths';
import { timeAgo, toDisplayDate } from '@/lib/date-utils';
import { capitalize } from '@/lib/string-utils';
import { cn } from '@/lib/utils';
import {
  DataTableCard,
  ConfirmDialog,
  LoadingSkeleton,
  ErrorState,
  PageHeader,
  StatusIndicator,
} from '@/components';
import type { ApiEnvelope, ClinicalListResponse, ConsentSummary } from '@/types';

/* ---------- types ---------- */

interface FhirPatientResource {
  resourceType: string;
  id?: string;
  name?: Array<{ use?: string; family?: string; given?: string[] }>;
  gender?: string;
  birthDate?: string;
  active?: boolean;
  address?: Array<{
    use?: string;
    line?: string[];
    city?: string;
    state?: string;
    postalCode?: string;
    country?: string;
  }>;
  telecom?: Array<{ system?: string; value?: string; use?: string }>;
  meta?: { lastUpdated?: string; versionId?: string };
}

interface HistoryEntry {
  commit: string;
  message: string;
  timestamp: string;
  author?: string;
}

/* ---------- helpers ---------- */

function patientDisplayName(patient: FhirPatientResource): string {
  const name = patient.name?.[0];
  if (!name) return 'Unknown';
  const parts: string[] = [];
  if (name.family) parts.push(name.family);
  if (name.given?.length) parts.push(name.given.join(' '));
  return parts.join(', ') || 'Unknown';
}

function calculateAge(birthDate: string): number {
  const birth = new Date(birthDate);
  const today = new Date();
  let age = today.getFullYear() - birth.getFullYear();
  const monthDiff = today.getMonth() - birth.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    age--;
  }
  return age;
}

function extractCodeDisplay(resource: Record<string, unknown>, field: string): string {
  const cc = resource[field] as
    | { coding?: Array<{ display?: string; code?: string }>; text?: string }
    | undefined;
  if (!cc) return '--';
  if (cc.text) return cc.text;
  if (cc.coding?.[0]?.display) return cc.coding[0].display;
  if (cc.coding?.[0]?.code) return cc.coding[0].code;
  return '--';
}

/* ---------- tab definitions ---------- */

const TABS = [
  { id: 'overview', label: 'Overview', icon: ClipboardList },
  { id: 'encounters', label: 'Encounters', icon: FileText },
  { id: 'vitals', label: 'Vitals', icon: Activity },
  { id: 'conditions', label: 'Conditions', icon: Heart },
  { id: 'medications', label: 'Medications', icon: Pill },
  { id: 'allergies', label: 'Allergies', icon: AlertTriangle },
  { id: 'immunizations', label: 'Immunizations', icon: Syringe },
  { id: 'procedures', label: 'Procedures', icon: Scissors },
  { id: 'consent', label: 'Consent', icon: Shield },
  { id: 'history', label: 'History', icon: GitCommit },
] as const;

type TabId = (typeof TABS)[number]['id'];

/* ---------- component ---------- */

export default function PatientDetailPage() {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { id } = useParams({ strict: false }) as any;
  const patientId = id as string;
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState<TabId>('overview');
  const [openDialog, setOpenDialog] = useState<string | null>(null);
  const [eraseDialogOpen, setEraseDialogOpen] = useState(false);

  /* ---------- patient query ---------- */
  const {
    data: patientEnvelope,
    isLoading: patientLoading,
    error: patientError,
  } = useQuery<ApiEnvelope<FhirPatientResource>>({
    queryKey: ['patient', patientId],
    queryFn: () => apiGet<FhirPatientResource>(API.patients.get(patientId)),
    enabled: !!patientId,
  });

  const patient = patientEnvelope?.data;
  const patientName = patient ? patientDisplayName(patient) : '';

  /* ---------- clinical queries (per-tab) ---------- */
  const encounterQuery = useQuery<ApiEnvelope<ClinicalListResponse>>({
    queryKey: ['patient', patientId, 'encounters'],
    queryFn: () =>
      apiGet<ClinicalListResponse>(API.patients.encounters.list(patientId)),
    enabled: !!patientId && (activeTab === 'encounters' || activeTab === 'overview'),
  });

  const observationQuery = useQuery<ApiEnvelope<ClinicalListResponse>>({
    queryKey: ['patient', patientId, 'observations'],
    queryFn: () =>
      apiGet<ClinicalListResponse>(API.patients.observations.list(patientId)),
    enabled: !!patientId && (activeTab === 'vitals' || activeTab === 'overview'),
  });

  const conditionQuery = useQuery<ApiEnvelope<ClinicalListResponse>>({
    queryKey: ['patient', patientId, 'conditions'],
    queryFn: () =>
      apiGet<ClinicalListResponse>(API.patients.conditions.list(patientId)),
    enabled: !!patientId && (activeTab === 'conditions' || activeTab === 'overview'),
  });

  const medicationQuery = useQuery<ApiEnvelope<ClinicalListResponse>>({
    queryKey: ['patient', patientId, 'medications'],
    queryFn: () =>
      apiGet<ClinicalListResponse>(
        API.patients.medicationRequests.list(patientId),
      ),
    enabled: !!patientId && (activeTab === 'medications' || activeTab === 'overview'),
  });

  const allergyQuery = useQuery<ApiEnvelope<ClinicalListResponse>>({
    queryKey: ['patient', patientId, 'allergies'],
    queryFn: () =>
      apiGet<ClinicalListResponse>(
        API.patients.allergyIntolerances.list(patientId),
      ),
    enabled: !!patientId && (activeTab === 'allergies' || activeTab === 'overview'),
  });

  const immunizationQuery = useQuery<ApiEnvelope<ClinicalListResponse>>({
    queryKey: ['patient', patientId, 'immunizations'],
    queryFn: () =>
      apiGet<ClinicalListResponse>(
        API.patients.immunizations.list(patientId),
      ),
    enabled: !!patientId && activeTab === 'immunizations',
  });

  const procedureQuery = useQuery<ApiEnvelope<ClinicalListResponse>>({
    queryKey: ['patient', patientId, 'procedures'],
    queryFn: () =>
      apiGet<ClinicalListResponse>(API.patients.procedures.list(patientId)),
    enabled: !!patientId && activeTab === 'procedures',
  });

  const consentQuery = useQuery<ApiEnvelope<{ consents: ConsentSummary[] }>>({
    queryKey: ['patient', patientId, 'consents'],
    queryFn: () =>
      apiGet<{ consents: ConsentSummary[] }>(
        API.patients.consents.list(patientId),
      ),
    enabled: !!patientId && activeTab === 'consent',
  });

  const historyQuery = useQuery<ApiEnvelope<{ entries: HistoryEntry[] }>>({
    queryKey: ['patient', patientId, 'history'],
    queryFn: () =>
      apiGet<{ entries: HistoryEntry[] }>(API.patients.history(patientId)),
    enabled: !!patientId && activeTab === 'history',
  });

  /* ---------- erase mutation ---------- */
  const eraseMutation = useMutation({
    mutationFn: () => apiDelete(API.patients.erase(patientId)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients'] });
      navigate({ to: '/patients' });
    },
  });

  /* ---------- loading / error states ---------- */
  if (patientLoading) {
    return (
      <div className="page-padding">
        <LoadingSkeleton count={8} />
      </div>
    );
  }

  if (patientError || !patient) {
    return (
      <div className="page-padding">
        <ErrorState
          message="Failed to load patient"
          details={patientError ? (patientError as Error).message : 'Patient not found'}
        />
      </div>
    );
  }

  /* ---------- render ---------- */
  return (
    <div className="page-padding">
      <PageHeader
        title={patientName || 'Patient Detail'}
        breadcrumbs={[
          { label: 'Dashboard', path: '/dashboard' },
          { label: 'Patients', path: '/patients' },
          { label: patientName || 'Detail' },
        ]}
        actions={
          <button
            type="button"
            onClick={() =>
              navigate({
                to: '/patients/$id/edit',
                params: { id: patientId },
              })
            }
            className={cn(
              'inline-flex items-center gap-2 px-4 py-2 text-xs font-mono uppercase tracking-wider cursor-pointer',
              'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
              'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
              'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
              'transition-colors duration-150 rounded-[var(--radius-sm)]',
            )}
          >
            <Edit3 size={14} />
            Edit
          </button>
        }
      />

      <div className="flex gap-6 h-full">
        {/* ===== Left panel ===== */}
        <div
          className={cn(
            'w-72 shrink-0 rounded-[var(--radius-lg)]',
            'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
            'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
            'p-5 flex flex-col gap-5 self-start',
          )}
        >
          {/* Name */}
          <div>
            <h2 className="text-xl font-bold text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] leading-tight">
              {patientDisplayName(patient)}
            </h2>
            <div className="flex items-center gap-2 mt-2">
              <span
                className={cn(
                  'inline-flex items-center px-2 py-0.5 rounded-full',
                  'text-[10px] font-mono font-semibold uppercase tracking-wider',
                  'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
                  'text-[var(--color-muted)]',
                )}
              >
                {capitalize(patient.gender ?? 'unknown')}
              </span>
              <StatusIndicator
                status={patient.active !== false ? 'active' : 'inactive'}
                label={patient.active !== false ? 'Active' : 'Inactive'}
                size="sm"
              />
            </div>
          </div>

          {/* Demographics */}
          <div className="space-y-3">
            <DetailRow label="Birth Date" value={patient.birthDate ?? '--'} />
            {patient.birthDate && (
              <DetailRow label="Age" value={`${calculateAge(patient.birthDate)} years`} />
            )}
            {patient.address?.[0] && (
              <DetailRow
                label="Address"
                value={[
                  patient.address[0].line?.join(', '),
                  patient.address[0].city,
                  patient.address[0].state,
                  patient.address[0].postalCode,
                  patient.address[0].country,
                ]
                  .filter(Boolean)
                  .join(', ') || '--'}
              />
            )}
            {patient.telecom?.map((t, i) => (
              <DetailRow
                key={i}
                label={capitalize(t.system ?? 'contact')}
                value={t.value ?? '--'}
              />
            ))}
            {patient.meta?.lastUpdated && (
              <DetailRow
                label="Last Updated"
                value={timeAgo(patient.meta.lastUpdated)}
              />
            )}
          </div>

          {/* Patient ID */}
          <div>
            <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
              Patient ID
            </span>
            <p className="font-mono text-xs text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] mt-0.5 break-all">
              {patientId}
            </p>
          </div>

          {/* Quick actions */}
          <div className="flex flex-col gap-2 mt-auto">
            <button
              type="button"
              onClick={() =>
                navigate({
                  to: '/patients/$id/edit',
                  params: { id: patientId },
                })
              }
              className={cn(
                'flex items-center justify-center gap-2 w-full px-4 py-2 text-xs font-mono uppercase tracking-wider cursor-pointer',
                'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
                'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
                'transition-colors duration-150 rounded-[var(--radius-sm)]',
              )}
            >
              <Edit3 size={14} />
              Edit
            </button>
            <button
              type="button"
              onClick={() => setEraseDialogOpen(true)}
              className={cn(
                'flex items-center justify-center gap-2 w-full px-4 py-2 text-xs font-mono uppercase tracking-wider cursor-pointer',
                'border border-[var(--color-error)] text-[var(--color-error)]',
                'hover:bg-[var(--color-error)] hover:text-white',
                'transition-colors duration-150 rounded-[var(--radius-sm)]',
              )}
            >
              <Trash2 size={14} />
              Erase
            </button>
          </div>
        </div>

        {/* ===== Right panel ===== */}
        <div className="flex-1 flex flex-col gap-4 min-w-0">
          {/* Tab bar */}
          <div className="flex gap-1 overflow-x-auto border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)] pb-px">
            {TABS.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  type="button"
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={cn(
                    'flex items-center gap-1.5 px-3 py-2 text-[11px] font-mono uppercase tracking-wider whitespace-nowrap cursor-pointer',
                    'border-b-2 transition-colors duration-150',
                    isActive
                      ? 'border-[var(--color-ink)] dark:border-[var(--color-sidebar-text)] text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]'
                      : 'border-transparent text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
                  )}
                >
                  <Icon size={14} />
                  {tab.label}
                </button>
              );
            })}
          </div>

          {/* Tab content */}
          <div className="flex-1 overflow-y-auto">
            {activeTab === 'overview' && (
              <OverviewTab
                encounters={encounterQuery.data?.data?.resources?.length ?? 0}
                observations={observationQuery.data?.data?.resources?.length ?? 0}
                conditions={conditionQuery.data?.data?.resources?.length ?? 0}
                medications={medicationQuery.data?.data?.resources?.length ?? 0}
                allergies={allergyQuery.data?.data?.resources?.length ?? 0}
                loading={
                  encounterQuery.isLoading ||
                  observationQuery.isLoading ||
                  conditionQuery.isLoading ||
                  medicationQuery.isLoading ||
                  allergyQuery.isLoading
                }
              />
            )}

            {activeTab === 'encounters' && (
              <ClinicalTable
                title="Encounters"
                queryResult={encounterQuery}
                columns={[
                  {
                    key: 'status',
                    header: 'Status',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs">
                        {capitalize(String(r.status ?? '--'))}
                      </span>
                    ),
                  },
                  {
                    key: 'class',
                    header: 'Class',
                    render: (r: Record<string, unknown>) => {
                      const cls = r.class as { code?: string } | undefined;
                      return (
                        <span className="font-mono text-xs">
                          {capitalize(cls?.code ?? '--')}
                        </span>
                      );
                    },
                  },
                  {
                    key: 'period_start',
                    header: 'Start',
                    render: (r: Record<string, unknown>) => {
                      const period = r.period as { start?: string } | undefined;
                      return (
                        <span className="font-mono text-xs tabular-nums">
                          {period?.start ? toDisplayDate(period.start) : '--'}
                        </span>
                      );
                    },
                  },
                  {
                    key: 'period_end',
                    header: 'End',
                    render: (r: Record<string, unknown>) => {
                      const period = r.period as { end?: string } | undefined;
                      return (
                        <span className="font-mono text-xs tabular-nums">
                          {period?.end ? toDisplayDate(period.end) : '--'}
                        </span>
                      );
                    },
                  },
                ]}
                addLabel="Add Encounter"
                onAdd={() => setOpenDialog('encounter')}
                emptyTitle="No encounters"
                emptySubtitle="Record the first encounter for this patient."
              />
            )}

            {activeTab === 'vitals' && (
              <ClinicalTable
                title="Observations / Vitals"
                queryResult={observationQuery}
                columns={[
                  {
                    key: 'code',
                    header: 'Code',
                    render: (r: Record<string, unknown>) => (
                      <span className="text-sm">{extractCodeDisplay(r, 'code')}</span>
                    ),
                  },
                  {
                    key: 'value',
                    header: 'Value',
                    render: (r: Record<string, unknown>) => {
                      const vq = r.valueQuantity as
                        | { value?: number; unit?: string }
                        | undefined;
                      if (vq) {
                        return (
                          <span className="font-mono text-xs tabular-nums">
                            {vq.value ?? '--'} {vq.unit ?? ''}
                          </span>
                        );
                      }
                      return (
                        <span className="font-mono text-xs">
                          {String(r.valueString ?? '--')}
                        </span>
                      );
                    },
                  },
                  {
                    key: 'effective',
                    header: 'Date',
                    render: (r: Record<string, unknown>) => {
                      const dt =
                        (r.effectiveDateTime as string) ??
                        (r.effectivePeriod as { start?: string })?.start;
                      return (
                        <span className="font-mono text-xs tabular-nums">
                          {dt ? toDisplayDate(dt) : '--'}
                        </span>
                      );
                    },
                  },
                ]}
                addLabel="Add Vital"
                onAdd={() => setOpenDialog('observation')}
                emptyTitle="No observations"
                emptySubtitle="Record the first vital sign or observation."
              />
            )}

            {activeTab === 'conditions' && (
              <ClinicalTable
                title="Conditions"
                queryResult={conditionQuery}
                columns={[
                  {
                    key: 'code',
                    header: 'Condition',
                    render: (r: Record<string, unknown>) => (
                      <span className="text-sm">{extractCodeDisplay(r, 'code')}</span>
                    ),
                  },
                  {
                    key: 'clinicalStatus',
                    header: 'Clinical Status',
                    render: (r: Record<string, unknown>) => {
                      const cs = r.clinicalStatus as
                        | { coding?: Array<{ code?: string }> }
                        | undefined;
                      return (
                        <span className="font-mono text-xs">
                          {capitalize(cs?.coding?.[0]?.code ?? '--')}
                        </span>
                      );
                    },
                  },
                  {
                    key: 'recordedDate',
                    header: 'Recorded',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs tabular-nums">
                        {r.recordedDate
                          ? toDisplayDate(String(r.recordedDate))
                          : '--'}
                      </span>
                    ),
                  },
                ]}
                addLabel="Add Condition"
                onAdd={() => setOpenDialog('condition')}
                emptyTitle="No conditions"
                emptySubtitle="No conditions have been recorded."
              />
            )}

            {activeTab === 'medications' && (
              <ClinicalTable
                title="Medication Requests"
                queryResult={medicationQuery}
                columns={[
                  {
                    key: 'medication',
                    header: 'Medication',
                    render: (r: Record<string, unknown>) => (
                      <span className="text-sm">
                        {extractCodeDisplay(r, 'medicationCodeableConcept')}
                      </span>
                    ),
                  },
                  {
                    key: 'status',
                    header: 'Status',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs">
                        {capitalize(String(r.status ?? '--'))}
                      </span>
                    ),
                  },
                  {
                    key: 'authoredOn',
                    header: 'Authored On',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs tabular-nums">
                        {r.authoredOn
                          ? toDisplayDate(String(r.authoredOn))
                          : '--'}
                      </span>
                    ),
                  },
                ]}
                addLabel="Add Medication"
                onAdd={() => setOpenDialog('medication')}
                emptyTitle="No medication requests"
                emptySubtitle="No medications have been prescribed."
              />
            )}

            {activeTab === 'allergies' && (
              <ClinicalTable
                title="Allergy Intolerances"
                queryResult={allergyQuery}
                columns={[
                  {
                    key: 'substance',
                    header: 'Substance',
                    render: (r: Record<string, unknown>) => (
                      <span className="text-sm">{extractCodeDisplay(r, 'code')}</span>
                    ),
                  },
                  {
                    key: 'criticality',
                    header: 'Criticality',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs">
                        {capitalize(String(r.criticality ?? '--'))}
                      </span>
                    ),
                  },
                  {
                    key: 'type',
                    header: 'Type',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs">
                        {capitalize(String(r.type ?? '--'))}
                      </span>
                    ),
                  },
                ]}
                addLabel="Add Allergy"
                onAdd={() => setOpenDialog('allergy')}
                emptyTitle="No allergies recorded"
                emptySubtitle="No known allergies or intolerances."
              />
            )}

            {activeTab === 'immunizations' && (
              <ClinicalTable
                title="Immunizations"
                queryResult={immunizationQuery}
                columns={[
                  {
                    key: 'vaccine',
                    header: 'Vaccine',
                    render: (r: Record<string, unknown>) => (
                      <span className="text-sm">
                        {extractCodeDisplay(r, 'vaccineCode')}
                      </span>
                    ),
                  },
                  {
                    key: 'status',
                    header: 'Status',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs">
                        {capitalize(String(r.status ?? '--'))}
                      </span>
                    ),
                  },
                  {
                    key: 'date',
                    header: 'Date',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs tabular-nums">
                        {r.occurrenceDateTime
                          ? toDisplayDate(String(r.occurrenceDateTime))
                          : '--'}
                      </span>
                    ),
                  },
                ]}
                addLabel="Add Immunization"
                onAdd={() => setOpenDialog('immunization')}
                emptyTitle="No immunizations"
                emptySubtitle="No immunizations recorded."
              />
            )}

            {activeTab === 'procedures' && (
              <ClinicalTable
                title="Procedures"
                queryResult={procedureQuery}
                columns={[
                  {
                    key: 'code',
                    header: 'Procedure',
                    render: (r: Record<string, unknown>) => (
                      <span className="text-sm">{extractCodeDisplay(r, 'code')}</span>
                    ),
                  },
                  {
                    key: 'status',
                    header: 'Status',
                    render: (r: Record<string, unknown>) => (
                      <span className="font-mono text-xs">
                        {capitalize(String(r.status ?? '--'))}
                      </span>
                    ),
                  },
                  {
                    key: 'date',
                    header: 'Date',
                    render: (r: Record<string, unknown>) => {
                      const dt =
                        (r.performedDateTime as string) ??
                        (r.performedPeriod as { start?: string })?.start;
                      return (
                        <span className="font-mono text-xs tabular-nums">
                          {dt ? toDisplayDate(dt) : '--'}
                        </span>
                      );
                    },
                  },
                ]}
                addLabel="Add Procedure"
                onAdd={() => setOpenDialog('procedure')}
                emptyTitle="No procedures"
                emptySubtitle="No procedures recorded."
              />
            )}

            {activeTab === 'consent' && (
              <ConsentTab
                consents={consentQuery.data?.data?.consents ?? []}
                loading={consentQuery.isLoading}
                error={
                  consentQuery.error
                    ? (consentQuery.error as Error).message
                    : undefined
                }
              />
            )}

            {activeTab === 'history' && (
              <HistoryTab
                entries={historyQuery.data?.data?.entries ?? []}
                loading={historyQuery.isLoading}
                error={
                  historyQuery.error
                    ? (historyQuery.error as Error).message
                    : undefined
                }
              />
            )}
          </div>
        </div>

        {/* Erase confirmation dialog */}
        <ConfirmDialog
          open={eraseDialogOpen}
          onOpenChange={setEraseDialogOpen}
          title="Erase Patient"
          description="This will permanently crypto-erase all data for this patient. The encryption key will be destroyed and all index entries purged. This action cannot be undone."
          confirmLabel="Erase Permanently"
          variant="destructive"
          onConfirm={() => eraseMutation.mutate()}
        />

        {/* Clinical form dialogs */}
        <EncounterDialog
          open={openDialog === 'encounter'}
          onClose={() => setOpenDialog(null)}
          patientId={patientId}
        />
        <ObservationDialog
          open={openDialog === 'observation'}
          onClose={() => setOpenDialog(null)}
          patientId={patientId}
        />
        <ConditionDialog
          open={openDialog === 'condition'}
          onClose={() => setOpenDialog(null)}
          patientId={patientId}
        />
        <MedicationRequestDialog
          open={openDialog === 'medication'}
          onClose={() => setOpenDialog(null)}
          patientId={patientId}
        />
        <AllergyDialog
          open={openDialog === 'allergy'}
          onClose={() => setOpenDialog(null)}
          patientId={patientId}
        />
        <ImmunizationDialog
          open={openDialog === 'immunization'}
          onClose={() => setOpenDialog(null)}
          patientId={patientId}
        />
        <ProcedureDialog
          open={openDialog === 'procedure'}
          onClose={() => setOpenDialog(null)}
          patientId={patientId}
        />
      </div>
    </div>
  );
}

/* ============================================================
   Sub-components
   ============================================================ */

/* ---------- Detail row (left panel) ---------- */

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
        {label}
      </span>
      <p className="text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] mt-0.5">
        {value}
      </p>
    </div>
  );
}

/* ---------- Overview tab ---------- */

function OverviewTab({
  encounters,
  observations,
  conditions,
  medications,
  allergies,
  loading,
}: {
  encounters: number;
  observations: number;
  conditions: number;
  medications: number;
  allergies: number;
  loading: boolean;
}) {
  const cards = [
    { label: 'Encounters', count: encounters, icon: FileText },
    { label: 'Observations', count: observations, icon: Activity },
    { label: 'Conditions', count: conditions, icon: Heart },
    { label: 'Medications', count: medications, icon: Pill },
    { label: 'Allergies', count: allergies, icon: AlertTriangle },
  ];

  if (loading) {
    return <LoadingSkeleton count={5} />;
  }

  return (
    <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
      {cards.map((card) => {
        const Icon = card.icon;
        return (
          <div
            key={card.label}
            className={cn(
              'rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
              'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
              'p-4 flex flex-col items-center gap-2',
            )}
          >
            <Icon
              size={20}
              className="text-[var(--color-muted)]"
              strokeWidth={1.5}
            />
            <span className="text-2xl font-bold font-mono tabular-nums text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
              {card.count}
            </span>
            <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
              {card.label}
            </span>
          </div>
        );
      })}
    </div>
  );
}

/* ---------- Clinical table (reusable per-tab) ---------- */

interface ClinicalTableProps {
  title: string;
  queryResult: {
    data?: ApiEnvelope<ClinicalListResponse>;
    isLoading: boolean;
    error: unknown;
    refetch?: () => void;
  };
  columns: Array<{
    key: string;
    header: string;
    render: (item: Record<string, unknown>) => React.ReactNode;
  }>;
  addLabel: string;
  onAdd?: () => void;
  emptyTitle: string;
  emptySubtitle: string;
}

function ClinicalTable({
  title,
  queryResult,
  columns,
  addLabel,
  onAdd,
  emptyTitle,
  emptySubtitle,
}: ClinicalTableProps) {
  const resources = queryResult.data?.data?.resources ?? [];

  const handleAdd = () => {
    onAdd?.();
  };

  return (
    <DataTableCard<Record<string, unknown>>
      title={title}
      columns={columns}
      data={resources}
      keyExtractor={(r) => String(r.id ?? Math.random())}
      loading={queryResult.isLoading}
      error={
        queryResult.error
          ? (queryResult.error as Error).message
          : undefined
      }
      onRetry={queryResult.refetch}
      emptyTitle={emptyTitle}
      emptySubtitle={emptySubtitle}
      actions={
        <button
          type="button"
          onClick={handleAdd}
          className={cn(
            'inline-flex items-center gap-1.5 px-3 py-1.5 text-[11px] font-mono uppercase tracking-wider cursor-pointer',
            'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
            'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
            'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
            'transition-colors duration-150 rounded-[var(--radius-sm)]',
          )}
        >
          <Plus size={12} />
          {addLabel}
        </button>
      }
    />
  );
}

/* ---------- Consent tab ---------- */

function ConsentTab({
  consents,
  loading,
  error,
}: {
  consents: ConsentSummary[];
  loading: boolean;
  error?: string;
}) {
  return (
    <DataTableCard<ConsentSummary>
      title="Consent Grants"
      columns={[
        {
          key: 'scope',
          header: 'Scope',
          render: (c) => (
            <span className="font-mono text-xs">{c.scope ?? '--'}</span>
          ),
        },
        {
          key: 'status',
          header: 'Status',
          render: (c) => (
            <StatusIndicator
              status={c.status === 'active' ? 'active' : 'inactive'}
              label={capitalize(c.status)}
              size="sm"
            />
          ),
        },
        {
          key: 'grantor',
          header: 'Grantor',
          render: (c) => (
            <span className="font-mono text-xs">{c.grantor ?? '--'}</span>
          ),
        },
        {
          key: 'period',
          header: 'Period',
          render: (c) => (
            <span className="font-mono text-xs tabular-nums">
              {c.period?.start ? toDisplayDate(c.period.start) : '--'}
              {c.period?.end ? ` - ${toDisplayDate(c.period.end)}` : ''}
            </span>
          ),
        },
      ]}
      data={consents}
      keyExtractor={(c) => c.id}
      loading={loading}
      error={error}
      emptyTitle="No consent records"
      emptySubtitle="No consent grants have been issued for this patient."
    />
  );
}

/* ---------- History tab ---------- */

function HistoryTab({
  entries,
  loading,
  error,
}: {
  entries: HistoryEntry[];
  loading: boolean;
  error?: string;
}) {
  return (
    <DataTableCard<HistoryEntry>
      title="Git History"
      columns={[
        {
          key: 'commit',
          header: 'Commit',
          render: (e) => (
            <span className="font-mono text-xs">
              {e.commit.substring(0, 8)}
            </span>
          ),
        },
        {
          key: 'message',
          header: 'Message',
          render: (e) => <span className="text-sm">{e.message}</span>,
        },
        {
          key: 'timestamp',
          header: 'When',
          render: (e) => (
            <span className="font-mono text-xs text-[var(--color-muted)]">
              {e.timestamp ? timeAgo(e.timestamp) : '--'}
            </span>
          ),
        },
        {
          key: 'author',
          header: 'Author',
          render: (e) => (
            <span className="font-mono text-xs">{e.author ?? '--'}</span>
          ),
        },
      ]}
      data={entries}
      keyExtractor={(e) => e.commit}
      loading={loading}
      error={error}
      emptyIcon={<GitCommit size={36} strokeWidth={1.5} />}
      emptyTitle="No history"
      emptySubtitle="No commit history available for this patient."
    />
  );
}
