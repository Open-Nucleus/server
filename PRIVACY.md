# Privacy & Data Protection

Open Nucleus privacy controls for compliance with health data protection regulations in target deployment regions.

## Crypto-Erasure

The primary mechanism for honoring data deletion requests. When a patient's data must be erased:

```
DELETE /api/v1/patients/{id}/erase
```

### What happens

1. **Key destruction** — The patient's encryption key (`.nucleus/keys/{patient_id}.key`) is deleted from Git and committed. Without this key, all of the patient's FHIR data in Git becomes permanently unreadable ciphertext.

2. **Index purge** — All SQLite rows referencing the patient are deleted in a single transaction across 10 tables (patients, encounters, observations, conditions, medication_requests, allergy_intolerances, immunizations, procedures, flags, patient_summaries).

3. **Git history retained** — Encrypted files remain in Git history as undecryptable binary data. This preserves the integrity of the Git hash chain and Merkle anchoring proofs without exposing patient data.

4. **Audit trail** — The erasure event is logged for compliance audit.

### What remains after erasure

| Data | Retained | Readable |
|------|----------|----------|
| Encrypted FHIR files in Git history | Yes | No — key destroyed |
| SQLite index rows | No | — |
| Git commit messages mentioning patient ID | Yes | Yes |
| Wrapped encryption key | No | — |
| Audit log entry of erasure | Yes | Yes |

### Limitations

- Git commit messages may contain the patient UUID. These cannot be removed without rewriting Git history, which would break sync integrity across nodes.
- If another node has not yet received the key destruction commit, it may still hold a valid key. Sync propagates the deletion, but there is a window during which disconnected nodes retain access.

## Compliance Matrix

| Regulation | Requirement | Implementation |
|------------|-------------|----------------|
| **GDPR Art 17** (EU) | Right to erasure | Crypto-erasure endpoint destroys per-patient DEK |
| **SA POPIA** Sec 11(2)(d) | Destruction of personal information | Same crypto-erasure mechanism |
| **Kenya DPA** Sec 40 | Right to deletion | Same crypto-erasure mechanism |
| **Nigeria NDPA** Sec 3.1(10) | Right to erasure | Same crypto-erasure mechanism |
| **GDPR Art 25** | Data protection by design | Per-patient encryption, minimal SQLite index |
| **GDPR Art 32** | Security of processing | AES-256-GCM encryption, TLS, Ed25519 auth |
| **HIPAA** §164.312(a)(2)(iv) | Encryption and decryption | AES-256-GCM at rest, TLS in transit |

## Data Minimization

### SQLite Search Index

The SQLite database stores only extracted search fields needed for query operations:

| Resource | Indexed Fields |
|----------|---------------|
| Patient | name (family, given), birth_date, gender, identifiers, active status |
| Encounter | patient_id, status, class_code, period_start, period_end |
| Observation | patient_id, status, category, code, effective_date, value |
| Condition | patient_id, clinical_status, verification_status, code, onset_date |
| MedicationRequest | patient_id, status, intent, medication_code |
| AllergyIntolerance | patient_id, clinical_status, verification_status, code |
| Immunization | patient_id, status, vaccine_code, occurrence_date |
| Procedure | patient_id, status, code, performed_date |

The full FHIR JSON (with narrative text, notes, extensions, and other rich data) exists only in Git, encrypted.

### Data Flow

```
Clinical Write Request
    |
    v
Validate FHIR JSON (cleartext)
    |
    v
Extract search fields (cleartext) ──> SQLite index (fields only)
    |
    v
Encrypt full FHIR JSON ──> Git commit (ciphertext)
    |
    v
Return resource + git commit info
```

## Consent Model

Open Nucleus does not implement a consent management system in V1. In the target deployment context (emergency medicine, disaster relief, military forward operating bases), the treating clinician's professional judgment serves as the basis for data processing under:

- GDPR Art 9(2)(c): vital interests
- GDPR Art 9(2)(h): healthcare provision
- POPIA Sec 27(1)(d): protecting the data subject's legitimate interests

A formal consent management system with granular purpose tracking is planned for future versions.

## Data Retention

Git retains all historical versions of clinical data indefinitely (as encrypted ciphertext). This is by design:

1. **Clinical audit trail** — Medical records must be retained for legal and clinical continuity purposes.
2. **Sync integrity** — Git history cannot be selectively rewritten without breaking the sync mesh.
3. **Merkle proofs** — Anchored Merkle roots reference specific Git states. Rewriting history would invalidate integrity proofs.

Crypto-erasure provides the right to erasure without rewriting history: the data exists but is mathematically inaccessible.

## Contact

For privacy-related inquiries, contact the data controller for your deployment site.
