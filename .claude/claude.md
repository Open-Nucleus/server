# Open Nucleus

Open Nucleus is an open-source, offline-first electronic health record (EHR) system designed for military forward operating bases, disaster relief zones, and small clinics in sub-Saharan Africa. It assumes zero connectivity as the default and treats network access as a bonus.

## Core Architecture

**Microservices in Go** (Patient, Sync, Auth, Formulary, Anchor services) plus a **Python Sentinel Agent**, fronted by a Go API Gateway on port 8080 (REST/JSON). The Flutter frontend lives in a separate repo (`open-nucleus-app`) and consumes the gateway as a pure REST client.

**Dual-layer data model:** FHIR R4 resources are stored as JSON files in a **Git repository** (source of truth) with a **SQLite database** as a rebuildable query index. Every clinical write commits to Git first, then upserts SQLite. If SQLite is lost, it rebuilds from Git.

**Git-based sync:** Nodes discover each other via Wi-Fi Direct, Bluetooth, or local network and sync using Git fetch/merge/push. A FHIR-aware merge driver classifies conflicts into auto-merge (safe), review (flag for clinician), or block (clinical safety risk). Transport is pluggable and automatic.

**Sentinel Agent:** A "sleeper" AI agent that wakes on sync events, crawls the merged dataset for epidemiological outbreak signals, cross-site medication conflicts, missed referral follow-ups, and supply stockout predictions. V1 is rule-based using WHO IDSR thresholds.

**IOTA Tangle anchoring:** Git Merkle roots are periodically anchored to the IOTA Tangle (feeless), providing cryptographic proof of data integrity for regulatory compliance, humanitarian accountability, and supply chain provenance.

## Key Specs

- All services communicate via gRPC internally
- Auth uses Ed25519 keypairs with offline-verifiable JWTs
- Target hardware: Raspberry Pi 4 or Android tablet
- FHIR R4 compliant for interoperability with global health systems
- Licensed AGPLv3