import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { VITAL_SIGNS } from '@/lib/fhir-codes';
import { toFhirDateTime } from '@/lib/date-utils';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from './clinical-dialog';
import type { WriteResponse } from '@/types';

interface ObservationDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function ObservationDialog({ open, onClose, patientId }: ObservationDialogProps) {
  const queryClient = useQueryClient();

  const [selectedCode, setSelectedCode] = useState(VITAL_SIGNS[0].code);
  const [value, setValue] = useState('');
  const [effectiveDateTime, setEffectiveDateTime] = useState('');

  const selectedVital = VITAL_SIGNS.find((v) => v.code === selectedCode) ?? VITAL_SIGNS[0];

  const mutation = useMutation({
    mutationFn: () => {
      const observation: Record<string, unknown> = {
        resourceType: 'Observation',
        status: 'final',
        code: {
          coding: [
            {
              system: 'http://loinc.org',
              code: selectedVital.code,
              display: selectedVital.display,
            },
          ],
          text: selectedVital.display,
        },
        subject: {
          reference: `Patient/${patientId}`,
        },
        valueQuantity: {
          value: parseFloat(value),
          unit: selectedVital.unit,
          system: 'http://unitsofmeasure.org',
          code: selectedVital.unit,
        },
        ...(effectiveDateTime && {
          effectiveDateTime: toFhirDateTime(new Date(effectiveDateTime)),
        }),
      };
      return apiPost<WriteResponse>(API.patients.observations.create(patientId), observation);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'observations'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setSelectedCode(VITAL_SIGNS[0].code);
    setValue('');
    setEffectiveDateTime('');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="New Observation"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to create observation') : null}
    >
      <div>
        <FieldLabel>Vital Sign</FieldLabel>
        <select
          value={selectedCode}
          onChange={(e) => setSelectedCode(e.target.value)}
          className={selectClass}
        >
          {VITAL_SIGNS.map((v) => (
            <option key={v.code} value={v.code}>
              {v.display} ({v.code})
            </option>
          ))}
        </select>
      </div>

      <div className="flex gap-3">
        <div className="flex-1">
          <FieldLabel>Value</FieldLabel>
          <input
            type="number"
            step="any"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder="0.0"
            className={inputClass}
          />
        </div>
        <div className="w-24">
          <FieldLabel>Unit</FieldLabel>
          <input
            type="text"
            value={selectedVital.unit}
            readOnly
            className={`${inputClass} bg-[var(--color-surface-hover)] dark:bg-[var(--color-surface-dark-hover)] cursor-not-allowed`}
          />
        </div>
      </div>

      <div>
        <FieldLabel>Effective Date/Time</FieldLabel>
        <input
          type="datetime-local"
          value={effectiveDateTime}
          onChange={(e) => setEffectiveDateTime(e.target.value)}
          className={inputClass}
        />
      </div>
    </ClinicalDialog>
  );
}
