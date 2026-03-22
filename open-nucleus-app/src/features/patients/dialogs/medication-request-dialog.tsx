import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { MEDICATION_STATUSES } from '@/lib/fhir-codes';
import { toFhirDateTime } from '@/lib/date-utils';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from './clinical-dialog';
import type { WriteResponse } from '@/types';

const INTENTS = ['proposal', 'plan', 'order', 'original-order', 'reflex-order', 'filler-order', 'instance-order', 'option'] as const;

interface MedicationRequestDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function MedicationRequestDialog({ open, onClose, patientId }: MedicationRequestDialogProps) {
  const queryClient = useQueryClient();

  const [medicationCode, setMedicationCode] = useState('');
  const [display, setDisplay] = useState('');
  const [status, setStatus] = useState<string>('active');
  const [intent, setIntent] = useState<string>('order');
  const [dosageText, setDosageText] = useState('');
  const [authoredOn, setAuthoredOn] = useState('');

  const mutation = useMutation({
    mutationFn: () => {
      const medRequest: Record<string, unknown> = {
        resourceType: 'MedicationRequest',
        status,
        intent,
        medicationCodeableConcept: {
          coding: [
            {
              system: 'http://www.nlm.nih.gov/research/umls/rxnorm',
              code: medicationCode,
              display,
            },
          ],
          text: display,
        },
        subject: {
          reference: `Patient/${patientId}`,
        },
        ...(dosageText && {
          dosageInstruction: [
            {
              text: dosageText,
            },
          ],
        }),
        ...(authoredOn && {
          authoredOn: toFhirDateTime(new Date(authoredOn)),
        }),
      };
      return apiPost<WriteResponse>(
        API.patients.medicationRequests.create(patientId),
        medRequest,
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'medicationRequests'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setMedicationCode('');
    setDisplay('');
    setStatus('active');
    setIntent('order');
    setDosageText('');
    setAuthoredOn('');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="New Medication Request"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to create medication request') : null}
    >
      <div>
        <FieldLabel>Medication Code</FieldLabel>
        <input
          type="text"
          value={medicationCode}
          onChange={(e) => setMedicationCode(e.target.value)}
          placeholder="e.g. RxNorm code"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Display Name</FieldLabel>
        <input
          type="text"
          value={display}
          onChange={(e) => setDisplay(e.target.value)}
          placeholder="e.g. Amoxicillin 500mg"
          className={inputClass}
        />
      </div>

      <div className="flex gap-3">
        <div className="flex-1">
          <FieldLabel>Status</FieldLabel>
          <select value={status} onChange={(e) => setStatus(e.target.value)} className={selectClass}>
            {MEDICATION_STATUSES.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>
        <div className="flex-1">
          <FieldLabel>Intent</FieldLabel>
          <select value={intent} onChange={(e) => setIntent(e.target.value)} className={selectClass}>
            {INTENTS.map((i) => (
              <option key={i} value={i}>{i}</option>
            ))}
          </select>
        </div>
      </div>

      <div>
        <FieldLabel>Dosage Instructions</FieldLabel>
        <input
          type="text"
          value={dosageText}
          onChange={(e) => setDosageText(e.target.value)}
          placeholder="e.g. Take 1 tablet three times daily"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Authored On</FieldLabel>
        <input
          type="date"
          value={authoredOn}
          onChange={(e) => setAuthoredOn(e.target.value)}
          className={inputClass}
        />
      </div>
    </ClinicalDialog>
  );
}
