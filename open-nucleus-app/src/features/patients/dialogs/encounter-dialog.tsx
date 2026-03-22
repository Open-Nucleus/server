import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { ENCOUNTER_STATUSES } from '@/lib/fhir-codes';
import { toFhirDateTime } from '@/lib/date-utils';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from './clinical-dialog';
import type { WriteResponse } from '@/types';

interface EncounterDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function EncounterDialog({ open, onClose, patientId }: EncounterDialogProps) {
  const queryClient = useQueryClient();

  const [status, setStatus] = useState<string>('in-progress');
  const [classCode, setClassCode] = useState('');
  const [periodStart, setPeriodStart] = useState('');
  const [periodEnd, setPeriodEnd] = useState('');

  const mutation = useMutation({
    mutationFn: () => {
      const encounter: Record<string, unknown> = {
        resourceType: 'Encounter',
        status,
        class: {
          system: 'http://terminology.hl7.org/CodeSystem/v3-ActCode',
          code: classCode || 'AMB',
          display: classCode || 'ambulatory',
        },
        subject: {
          reference: `Patient/${patientId}`,
        },
        period: {
          ...(periodStart && { start: toFhirDateTime(new Date(periodStart)) }),
          ...(periodEnd && { end: toFhirDateTime(new Date(periodEnd)) }),
        },
      };
      return apiPost<WriteResponse>(API.patients.encounters.create(patientId), encounter);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'encounters'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setStatus('in-progress');
    setClassCode('');
    setPeriodStart('');
    setPeriodEnd('');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="New Encounter"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to create encounter') : null}
    >
      <div>
        <FieldLabel>Status</FieldLabel>
        <select value={status} onChange={(e) => setStatus(e.target.value)} className={selectClass}>
          {ENCOUNTER_STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      <div>
        <FieldLabel>Class Code</FieldLabel>
        <input
          type="text"
          value={classCode}
          onChange={(e) => setClassCode(e.target.value)}
          placeholder="e.g. AMB, EMER, IMP"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Period Start</FieldLabel>
        <input
          type="datetime-local"
          value={periodStart}
          onChange={(e) => setPeriodStart(e.target.value)}
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Period End</FieldLabel>
        <input
          type="datetime-local"
          value={periodEnd}
          onChange={(e) => setPeriodEnd(e.target.value)}
          className={inputClass}
        />
      </div>
    </ClinicalDialog>
  );
}
