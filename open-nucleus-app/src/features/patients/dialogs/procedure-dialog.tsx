import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { PROCEDURE_STATUSES } from '@/lib/fhir-codes';
import { toFhirDateTime } from '@/lib/date-utils';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from './clinical-dialog';
import type { WriteResponse } from '@/types';

interface ProcedureDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function ProcedureDialog({ open, onClose, patientId }: ProcedureDialogProps) {
  const queryClient = useQueryClient();

  const [code, setCode] = useState('');
  const [display, setDisplay] = useState('');
  const [status, setStatus] = useState<string>('completed');
  const [performedDateTime, setPerformedDateTime] = useState('');

  const mutation = useMutation({
    mutationFn: () => {
      const procedure: Record<string, unknown> = {
        resourceType: 'Procedure',
        status,
        code: {
          coding: [
            {
              system: 'http://snomed.info/sct',
              code,
              display,
            },
          ],
          text: display,
        },
        subject: {
          reference: `Patient/${patientId}`,
        },
        ...(performedDateTime && {
          performedDateTime: toFhirDateTime(new Date(performedDateTime)),
        }),
      };
      return apiPost<WriteResponse>(API.patients.procedures.create(patientId), procedure);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'procedures'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setCode('');
    setDisplay('');
    setStatus('completed');
    setPerformedDateTime('');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="New Procedure"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to create procedure') : null}
    >
      <div>
        <FieldLabel>Procedure Code</FieldLabel>
        <input
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="e.g. SNOMED code"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Display Name</FieldLabel>
        <input
          type="text"
          value={display}
          onChange={(e) => setDisplay(e.target.value)}
          placeholder="e.g. Appendectomy"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Status</FieldLabel>
        <select value={status} onChange={(e) => setStatus(e.target.value)} className={selectClass}>
          {PROCEDURE_STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      <div>
        <FieldLabel>Performed Date/Time</FieldLabel>
        <input
          type="datetime-local"
          value={performedDateTime}
          onChange={(e) => setPerformedDateTime(e.target.value)}
          className={inputClass}
        />
      </div>
    </ClinicalDialog>
  );
}
