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


## Workflow Rules

1. **Plan Mode Default**
   - Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
   - If something goes sideways, STOP and re-plan immediately
   - Use plan mode for verification steps, not just building

2. **Subagent Strategy**
   - Use multi-agents liberally to keep main context window clean
   - Offload research, exploration, and parallel analysis to subagents
   - One task per subagent for focused execution

3. **Self-Improvement Loop**
   - After ANY correction from the user: update `tasks/lessons.md` with the pattern
   - Review lessons at session start

4. **Verification Before Done**
   - Never mark a task complete without proving it works
   - Run tests, check logs, demonstrate correctness

5. **Demand Elegance (Balanced)**
   - For non-trivial changes: pause and ask "is there a more elegant way?"
   - Skip this for simple, obvious fixes

6. **Autonomous Bug Fixing**
   - When given a bug report: just fix it. No hand-holding.
   - Go fix failing CI tests without being told how

7. **Housekeeping**
   - Keep your name out of git commits
   - Make a git commit after every major feature
   - if building frontend, i prefer black and white design in typewriter format

## Task Management

1. Plan First: Write plan to `tasks/todo.md` with checkable items
2. Verify Plan: Check in before starting implementation
3. Track Progress: Mark items complete as you go
4. Explain Changes: High-level summary at each step
5. Document Results: Add review section to `tasks/todo.md`
6. Capture Lessons: Update `tasks/lessons.md` after corrections

## Core Principles

- **Simplicity First:** Make every change as simple as possible. Impact minimal code.
- **No Laziness:** Find root causes. No temporary fixes. Senior developer standards.