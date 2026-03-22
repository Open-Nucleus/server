import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { toFhirDate } from '@/lib/date-utils';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from '../../patients/dialogs/clinical-dialog';
import type { ConsentGrantRequest } from '@/types';

const SCOPES = ['full-access', 'read-only', 'emergency'] as const;

interface ConsentDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function ConsentDialog({ open, onClose, patientId }: ConsentDialogProps) {
  const queryClient = useQueryClient();

  const [providerId, setProviderId] = useState('');
  const [scope, setScope] = useState<string>('read-only');
  const [periodStart, setPeriodStart] = useState('');
  const [periodEnd, setPeriodEnd] = useState('');

  const mutation = useMutation({
    mutationFn: () => {
      const grantRequest: ConsentGrantRequest = {
        patient_id: patientId,
        provider_id: providerId,
        scope,
        period: {
          start: periodStart ? toFhirDate(new Date(periodStart)) : toFhirDate(new Date()),
          ...(periodEnd && { end: toFhirDate(new Date(periodEnd)) }),
        },
      };
      return apiPost<void>(API.patients.consents.grant(patientId), grantRequest);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'consents'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setProviderId('');
    setScope('read-only');
    setPeriodStart('');
    setPeriodEnd('');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="Grant Consent"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to grant consent') : null}
    >
      <div>
        <FieldLabel>Provider ID</FieldLabel>
        <input
          type="text"
          value={providerId}
          onChange={(e) => setProviderId(e.target.value)}
          placeholder="e.g. practitioner-001"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Scope</FieldLabel>
        <select value={scope} onChange={(e) => setScope(e.target.value)} className={selectClass}>
          {SCOPES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      <div className="flex gap-3">
        <div className="flex-1">
          <FieldLabel>Period Start</FieldLabel>
          <input
            type="date"
            value={periodStart}
            onChange={(e) => setPeriodStart(e.target.value)}
            className={inputClass}
          />
        </div>
        <div className="flex-1">
          <FieldLabel>Period End (optional)</FieldLabel>
          <input
            type="date"
            value={periodEnd}
            onChange={(e) => setPeriodEnd(e.target.value)}
            className={inputClass}
          />
        </div>
      </div>
    </ClinicalDialog>
  );
}
