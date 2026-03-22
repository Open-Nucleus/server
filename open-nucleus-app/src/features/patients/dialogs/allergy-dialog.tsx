import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import {
  ALLERGY_CRITICALITIES,
  ALLERGY_TYPES,
  ALLERGY_CATEGORIES,
  CONDITION_STATUSES,
} from '@/lib/fhir-codes';
import { ClinicalDialog, FieldLabel, inputClass, selectClass } from './clinical-dialog';
import type { WriteResponse } from '@/types';

interface AllergyDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
}

export function AllergyDialog({ open, onClose, patientId }: AllergyDialogProps) {
  const queryClient = useQueryClient();

  const [code, setCode] = useState('');
  const [display, setDisplay] = useState('');
  const [clinicalStatus, setClinicalStatus] = useState<string>('active');
  const [criticality, setCriticality] = useState<string>('low');
  const [type, setType] = useState<string>('allergy');
  const [category, setCategory] = useState<string>('medication');

  const mutation = useMutation({
    mutationFn: () => {
      const allergy: Record<string, unknown> = {
        resourceType: 'AllergyIntolerance',
        clinicalStatus: {
          coding: [
            {
              system: 'http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical',
              code: clinicalStatus,
            },
          ],
        },
        type,
        category: [category],
        criticality,
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
        patient: {
          reference: `Patient/${patientId}`,
        },
      };
      return apiPost<WriteResponse>(
        API.patients.allergyIntolerances.create(patientId),
        allergy,
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients', patientId, 'allergyIntolerances'] });
      resetAndClose();
    },
  });

  const resetAndClose = () => {
    setCode('');
    setDisplay('');
    setClinicalStatus('active');
    setCriticality('low');
    setType('allergy');
    setCategory('medication');
    mutation.reset();
    onClose();
  };

  return (
    <ClinicalDialog
      open={open}
      onClose={resetAndClose}
      title="New Allergy / Intolerance"
      onSubmit={() => mutation.mutate()}
      submitting={mutation.isPending}
      error={mutation.isError ? (mutation.error instanceof Error ? mutation.error.message : 'Failed to create allergy') : null}
    >
      <div>
        <FieldLabel>Allergen Code</FieldLabel>
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
          placeholder="e.g. Penicillin"
          className={inputClass}
        />
      </div>

      <div className="flex gap-3">
        <div className="flex-1">
          <FieldLabel>Type</FieldLabel>
          <select value={type} onChange={(e) => setType(e.target.value)} className={selectClass}>
            {ALLERGY_TYPES.map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
        </div>
        <div className="flex-1">
          <FieldLabel>Category</FieldLabel>
          <select value={category} onChange={(e) => setCategory(e.target.value)} className={selectClass}>
            {ALLERGY_CATEGORIES.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="flex gap-3">
        <div className="flex-1">
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
        <div className="flex-1">
          <FieldLabel>Criticality</FieldLabel>
          <select
            value={criticality}
            onChange={(e) => setCriticality(e.target.value)}
            className={selectClass}
          >
            {ALLERGY_CRITICALITIES.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
        </div>
      </div>
    </ClinicalDialog>
  );
}
