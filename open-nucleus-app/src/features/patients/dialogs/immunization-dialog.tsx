import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { IMMUNIZATION_STATUSES } from '@/lib/fhir-codes';
import { toFhirDate } from '@/lib/date-utils';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from './clinical-dialog';
import type { WriteResponse } from '@/types';

interface ImmunizationDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function ImmunizationDialog({ open, onClose, patientId }: ImmunizationDialogProps) {
  const queryClient = useQueryClient();

  const [vaccineCode, setVaccineCode] = useState('');
  const [display, setDisplay] = useState('');
  const [status, setStatus] = useState<string>('completed');
  const [occurrenceDate, setOccurrenceDate] = useState('');
  const [lotNumber, setLotNumber] = useState('');

  const mutation = useMutation({
    mutationFn: () => {
      const immunization: Record<string, unknown> = {
        resourceType: 'Immunization',
        status,
        vaccineCode: {
          coding: [
            {
              system: 'http://hl7.org/fhir/sid/cvx',
              code: vaccineCode,
              display,
            },
          ],
          text: display,
        },
        patient: {
          reference: `Patient/${patientId}`,
        },
        ...(occurrenceDate && {
          occurrenceDateTime: toFhirDate(new Date(occurrenceDate)),
        }),
        ...(lotNumber && { lotNumber }),
      };
      return apiPost<WriteResponse>(API.patients.immunizations.create(patientId), immunization);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'immunizations'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setVaccineCode('');
    setDisplay('');
    setStatus('completed');
    setOccurrenceDate('');
    setLotNumber('');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="New Immunization"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to create immunization') : null}
    >
      <div>
        <FieldLabel>Vaccine Code</FieldLabel>
        <input
          type="text"
          value={vaccineCode}
          onChange={(e) => setVaccineCode(e.target.value)}
          placeholder="e.g. CVX code"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Display Name</FieldLabel>
        <input
          type="text"
          value={display}
          onChange={(e) => setDisplay(e.target.value)}
          placeholder="e.g. COVID-19 mRNA Vaccine"
          className={inputClass}
        />
      </div>

      <div>
        <FieldLabel>Status</FieldLabel>
        <select value={status} onChange={(e) => setStatus(e.target.value)} className={selectClass}>
          {IMMUNIZATION_STATUSES.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      <div className="flex gap-3">
        <div className="flex-1">
          <FieldLabel>Occurrence Date</FieldLabel>
          <input
            type="date"
            value={occurrenceDate}
            onChange={(e) => setOccurrenceDate(e.target.value)}
            className={inputClass}
          />
        </div>
        <div className="flex-1">
          <FieldLabel>Lot Number</FieldLabel>
          <input
            type="text"
            value={lotNumber}
            onChange={(e) => setLotNumber(e.target.value)}
            placeholder="e.g. AAJN11K"
            className={inputClass}
          />
        </div>
      </div>
    </ClinicalDialog>
  );
}
