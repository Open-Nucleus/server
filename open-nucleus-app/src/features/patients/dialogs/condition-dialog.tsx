import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { CONDITION_STATUSES } from '@/lib/fhir-codes';
import { toFhirDate } from '@/lib/date-utils';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from './clinical-dialog';
import type { WriteResponse } from '@/types';

interface ConditionDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function ConditionDialog({ open, onClose, patientId }: ConditionDialogProps) {
  const queryClient = useQueryClient();

  const [code, setCode] = useState('');
  const [display, setDisplay] = useState('');
  const [clinicalStatus, setClinicalStatus] = useState<string>('active');
  const [recordedDate, setRecordedDate] = useState('');

  const mutation = useMutation({
    mutationFn: () => {
      const condition: Record<string, unknown> = {
        resourceType: 'Condition',
        clinicalStatus: {
          coding: [
            {
              system: 'http://terminology.hl7.org/CodeSystem/condition-clinical',
              code: clinicalStatus,
            },
          ],
        },
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
        ...(recordedDate && {
          recordedDate: toFhirDate(new Date(recordedDate)),
        }),
      };
      return apiPost<WriteResponse>(API.patients.conditions.create(patientId), condition);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'conditions'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setCode('');
    setDisplay('');
    setClinicalStatus('active');
    setRecordedDate('');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="New Condition"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to create condition') : null}
    >
      <div>
        <FieldLabel>Code (ICD / SNOMED)</FieldLabel>
        <input
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="e.g. 38341003"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Display Name</FieldLabel>
        <input
          type="text"
          value={display}
          onChange={(e) => setDisplay(e.target.value)}
          placeholder="e.g. Hypertension"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Clinical Status</FieldLabel>
        <select
          value={clinicalStatus}
          onChange={(e) => setClinicalStatus(e.target.value)}
          className={selectClass}
        >
          {CONDITION_STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      <div>
        <FieldLabel>Recorded Date</FieldLabel>
        <input
          type="date"
          value={recordedDate}
          onChange={(e) => setRecordedDate(e.target.value)}
          className={inputClass}
        />
      </div>
    </ClinicalDialog>
  );
}
