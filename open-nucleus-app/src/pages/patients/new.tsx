import { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useMutation } from '@tanstack/react-query';
import { Save, X } from 'lucide-react';
import { apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { cn } from '@/lib/utils';
import { ADMINISTRATIVE_GENDERS } from '@/lib/fhir-codes';
import { capitalize } from '@/lib/string-utils';
import { PageHeader } from '@/components';
import type { WriteResponse } from '@/types';

/* ---------- form state ---------- */

interface PatientFormState {
  givenName: string;
  familyName: string;
  gender: string;
  birthDate: string;
  active: boolean;
  addressLine: string;
  city: string;
  state: string;
  postalCode: string;
  country: string;
  phone: string;
  email: string;
}

const INITIAL_STATE: PatientFormState = {
  givenName: '',
  familyName: '',
  gender: 'unknown',
  birthDate: '',
  active: true,
  addressLine: '',
  city: '',
  state: '',
  postalCode: '',
  country: '',
  phone: '',
  email: '',
};

/* ---------- helpers ---------- */

function buildFhirPatient(form: PatientFormState): Record<string, unknown> {
  const telecom: Array<{ system: string; value: string }> = [];
  if (form.phone.trim()) telecom.push({ system: 'phone', value: form.phone.trim() });
  if (form.email.trim()) telecom.push({ system: 'email', value: form.email.trim() });

  const hasAddress =
    form.addressLine.trim() ||
    form.city.trim() ||
    form.state.trim() ||
    form.postalCode.trim() ||
    form.country.trim();

  const resource: Record<string, unknown> = {
    resourceType: 'Patient',
    name: [
      {
        use: 'official',
        family: form.familyName.trim(),
        given: form.givenName.trim() ? [form.givenName.trim()] : [],
      },
    ],
    gender: form.gender,
    birthDate: form.birthDate || undefined,
    active: form.active,
  };

  if (hasAddress) {
    resource.address = [
      {
        use: 'home',
        line: form.addressLine.trim() ? [form.addressLine.trim()] : [],
        city: form.city.trim() || undefined,
        state: form.state.trim() || undefined,
        postalCode: form.postalCode.trim() || undefined,
        country: form.country.trim() || undefined,
      },
    ];
  }

  if (telecom.length > 0) {
    resource.telecom = telecom;
  }

  return resource;
}

/* ---------- component ---------- */

export default function PatientNewPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<PatientFormState>(INITIAL_STATE);

  /* mutation */
  const mutation = useMutation({
    mutationFn: (data: Record<string, unknown>) =>
      apiPost<WriteResponse>(API.patients.create, data),
    onSuccess: (envelope) => {
      const newId = envelope.data?.resource_id;
      if (newId) {
        navigate({ to: '/patients/$id', params: { id: newId } });
      } else {
        navigate({ to: '/patients' });
      }
    },
  });

  /* field updater */
  const set = <K extends keyof PatientFormState>(
    key: K,
    value: PatientFormState[K],
  ) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  /* submit */
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.familyName.trim()) return;
    const resource = buildFhirPatient(form);
    mutation.mutate(resource);
  };

  return (
    <div className="page-padding max-w-2xl">
      <PageHeader
        title="New Patient"
        breadcrumbs={[
          { label: 'Dashboard', path: '/dashboard' },
          { label: 'Patients', path: '/patients' },
          { label: 'New Patient' },
        ]}
        actions={
          <button
            type="button"
            onClick={() => navigate({ to: '/patients' })}
            className={cn(
              'inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
              'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
              'transition-colors duration-150',
            )}
          >
            <X size={14} />
            Cancel
          </button>
        }
      />

      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Name section */}
        <FormSection title="Name">
          <div className="grid grid-cols-2 gap-4">
            <FormInput
              label="Given Name"
              value={form.givenName}
              onChange={(v) => set('givenName', v)}
              placeholder="John"
            />
            <FormInput
              label="Family Name"
              value={form.familyName}
              onChange={(v) => set('familyName', v)}
              placeholder="Doe"
              required
            />
          </div>
        </FormSection>

        {/* Demographics section */}
        <FormSection title="Demographics">
          <div className="grid grid-cols-3 gap-4">
            <FormSelect
              label="Gender"
              value={form.gender}
              onChange={(v) => set('gender', v)}
              options={ADMINISTRATIVE_GENDERS.map((g) => ({
                value: g,
                label: capitalize(g),
              }))}
            />
            <FormInput
              label="Birth Date"
              value={form.birthDate}
              onChange={(v) => set('birthDate', v)}
              type="date"
            />
            <FormToggle
              label="Active"
              checked={form.active}
              onChange={(v) => set('active', v)}
            />
          </div>
        </FormSection>

        {/* Address section */}
        <FormSection title="Address">
          <div className="space-y-4">
            <FormInput
              label="Street Address"
              value={form.addressLine}
              onChange={(v) => set('addressLine', v)}
              placeholder="123 Main St"
            />
            <div className="grid grid-cols-2 gap-4">
              <FormInput
                label="City"
                value={form.city}
                onChange={(v) => set('city', v)}
                placeholder="Nairobi"
              />
              <FormInput
                label="State / Province"
                value={form.state}
                onChange={(v) => set('state', v)}
                placeholder="Nairobi County"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <FormInput
                label="Postal Code"
                value={form.postalCode}
                onChange={(v) => set('postalCode', v)}
                placeholder="00100"
              />
              <FormInput
                label="Country"
                value={form.country}
                onChange={(v) => set('country', v)}
                placeholder="KE"
              />
            </div>
          </div>
        </FormSection>

        {/* Contact section */}
        <FormSection title="Contact">
          <div className="grid grid-cols-2 gap-4">
            <FormInput
              label="Phone"
              value={form.phone}
              onChange={(v) => set('phone', v)}
              placeholder="+254 700 000 000"
              type="tel"
            />
            <FormInput
              label="Email"
              value={form.email}
              onChange={(v) => set('email', v)}
              placeholder="patient@example.com"
              type="email"
            />
          </div>
        </FormSection>

        {/* Error message */}
        {mutation.isError && (
          <div className="px-4 py-3 text-sm font-mono text-[var(--color-error)] border border-[var(--color-error)] rounded-[var(--radius-sm)]">
            {(mutation.error as Error).message}
          </div>
        )}

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <button
            type="submit"
            disabled={mutation.isPending || !form.familyName.trim()}
            className={cn(
              'inline-flex items-center gap-2 px-5 py-2.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
              'bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)]',
              'hover:opacity-90 transition-opacity duration-150 rounded-[var(--radius-sm)]',
              'disabled:opacity-40 disabled:cursor-not-allowed',
            )}
          >
            <Save size={14} />
            {mutation.isPending ? 'Saving...' : 'Save Patient'}
          </button>
          <button
            type="button"
            onClick={() => navigate({ to: '/patients' })}
            className={cn(
              'px-5 py-2.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
              'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
              'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
              'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
              'transition-colors duration-150 rounded-[var(--radius-sm)]',
            )}
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
}

/* ============================================================
   Form sub-components
   ============================================================ */

function FormSection({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <fieldset
      className={cn(
        'rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
        'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
        'p-4',
      )}
    >
      <legend className="px-2 font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
        {title}
      </legend>
      {children}
    </fieldset>
  );
}

function FormInput({
  label,
  value,
  onChange,
  placeholder,
  type = 'text',
  required = false,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  type?: string;
  required?: boolean;
}) {
  return (
    <label className="flex flex-col gap-1">
      <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
        {label}
        {required && <span className="text-[var(--color-error)] ml-0.5">*</span>}
      </span>
      <input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        required={required}
        className={cn(
          'px-3 py-2 text-sm font-mono rounded-[var(--radius-sm)]',
          'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
          'bg-transparent',
          'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
          'placeholder:text-[var(--color-muted)] placeholder:opacity-50',
          'outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
          'transition-colors duration-150',
        )}
      />
    </label>
  );
}

function FormSelect({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: Array<{ value: string; label: string }>;
}) {
  return (
    <label className="flex flex-col gap-1">
      <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
        {label}
      </span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={cn(
          'px-3 py-2 text-sm font-mono rounded-[var(--radius-sm)]',
          'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
          'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
          'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
          'outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
          'transition-colors duration-150',
        )}
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </label>
  );
}

function FormToggle({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <label className="flex flex-col gap-1">
      <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
        {label}
      </span>
      <div className="flex items-center gap-2 py-2">
        <button
          type="button"
          onClick={() => onChange(!checked)}
          className={cn(
            'relative w-10 h-5 rounded-full transition-colors duration-200 cursor-pointer',
            checked
              ? 'bg-[var(--color-ink)] dark:bg-[var(--color-sidebar-text)]'
              : 'bg-[var(--color-border)] dark:bg-[var(--color-border-dark)]',
          )}
        >
          <span
            className={cn(
              'absolute top-0.5 left-0.5 w-4 h-4 rounded-full transition-transform duration-200',
              checked
                ? 'translate-x-5 bg-[var(--color-paper)] dark:bg-[var(--color-paper-dark)]'
                : 'translate-x-0 bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
            )}
          />
        </button>
        <span className="text-xs font-mono text-[var(--color-muted)]">
          {checked ? 'Active' : 'Inactive'}
        </span>
      </div>
    </label>
  );
}
