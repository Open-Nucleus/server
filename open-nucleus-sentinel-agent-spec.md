# Open Nucleus — Sentinel Agent Service Specification

**Version:** 2.0  
**Date:** March 2026  
**Author:** Dr Akanimoh Osutuk — FibrinLab  
**Repo:** github.com/FibrinLab/open-nucleus  
**Service:** `services/sentinel/`  
**Depends on:** open-sentinel ≥ 0.3.1 (github.com/Open-Nucleus/open-sentinel)  
**Status:** Draft

---

## 1. Service Overview

### 1.1 Role

The Sentinel Agent is a thin service wrapper around the `open-sentinel` library. It connects the library to the Open Nucleus event bus, configures the FhirGit data adapter, wires alert outputs to the Patient Service, manages the Ollama sidecar, and exposes a management API for clinician feedback and manual control.

The Sentinel Agent does NOT implement surveillance logic, LLM reasoning, reflection loops, memory, or guardrails — all of that lives in `open-sentinel`. This service is integration and configuration.

### 1.2 What Changed (V1 → V2)

The underlying `open-sentinel` library has been fundamentally revised from rule-first to LLM-first architecture with 21 agentic design patterns. The wrapper must now:

| V1 Wrapper | V2 Wrapper |
|------------|------------|
| Passed LLM as optional | Passes LLM as required (agent runs degraded without it) |
| Configured flat `AgentState` | Configures three-tier `MemoryStore` (working, episodic, semantic, procedural) |
| No feedback mechanism | Exposes HITL feedback API → `process_feedback()` |
| No hardware awareness | Passes hardware profile → `ResourceManager` |
| No priority handling | Priority-aware skill ordering is handled by library, but wrapper configures the profile |
| Simple alert write-back | Extended alert write-back with AI provenance, reflection metadata, review queue |
| Basic Ollama start/stop | Ollama watchdog with crash recovery and LLM health monitoring |
| Flat event forwarding | Richer lifecycle events forwarded to Open Nucleus event bus |

### 1.3 Service Identity

| Property | Value |
|----------|-------|
| Language | Python |
| Port | 8090 (HTTP health + management + feedback API) |
| Dependencies | `open-sentinel` ≥ 0.3.1, `grpcio` |
| Reads from | Open Nucleus Git repo + SQLite index (via open-sentinel `FhirGitAdapter`) |
| Writes to | Patient Service gRPC (DetectedIssue, Flag resources) |
| Triggered by | `SyncCompleted` events from Sync Service (gRPC stream) |
| Sidecar | Ollama (separate process, same machine) |

### 1.4 Design Principles

- **open-sentinel does the work.** This service is configuration, integration, and API exposure.
- **Skills are the intelligence.** Swapping surveillance capabilities is a config change.
- **LLM is a sidecar, not in-process.** Ollama crash → agent continues in degraded mode.
- **The system works without the Sentinel.** All clinical workflows function if it's not running.
- **Feedback flows back.** Clinician decisions calibrate the agent over time.

---

## 2. Architecture Within Open Nucleus

```
┌─────────────────────────────────────────────────────┐
│                 Open Nucleus Node                     │
│                                                       │
│  ┌──────────┐    SyncCompleted     ┌───────────────┐ │
│  │  Sync    │──────event──────────▶│   Sentinel    │ │
│  │ Service  │     (gRPC stream)    │    Agent      │ │
│  └──────────┘                      │               │ │
│                                    │  open-sentinel │ │
│  ┌──────────┐    Read clinical     │  library:     │ │
│  │ Git Repo │◀─────data────────────│               │ │
│  │ + SQLite │     (FhirGitAdapter) │  • LLM Engine │ │
│  └──────────┘                      │  • Skills     │ │
│                                    │  • Memory     │ │
│  ┌──────────┐    Write alerts      │  • Reflection │ │
│  │ Patient  │◀─────(gRPC)──────────│  • Guardrails │ │
│  │ Service  │   DetectedIssue/Flag │  • Priority   │ │
│  └──────────┘                      │  • Resources  │ │
│                                    └──────┬────────┘ │
│                                           │          │
│  ┌──────────┐    LLM inference     ┌──────▼────────┐ │
│  │  Ollama  │◀─────(HTTP)──────────│ OllamaEngine  │ │
│  │(sidecar) │     localhost:11434  └───────────────┘ │
│  │ watchdog │                                        │
│  └──────────┘                                        │
│                                                       │
│  ┌──────────┐    Critical alerts                     │
│  │   SMS    │◀──────────────────────(SmsOutput)      │
│  │ Gateway  │     Africa's Talking                   │
│  └──────────┘                                        │
│                                                       │
│  ┌──────────┐    Clinician reviews  ┌─────────────┐  │
│  │ Flutter  │────feedback──────────▶│ Feedback API │  │
│  │   App    │    (HTTP POST)       │ :8090        │  │
│  └──────────┘                      └─────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## 3. Integration Points

### 3.1 Event Subscription (← Sync Service)

```python
class SyncEventSubscriber:
    """Connects to Sync Service gRPC and feeds events to open-sentinel."""
    
    def __init__(self, sync_grpc_address: str):
        self.channel = grpc.aio.insecure_channel(sync_grpc_address)
        self.stub = SyncServiceStub(self.channel)
    
    async def subscribe(self) -> AsyncIterator[DataEvent]:
        """Convert Sync Service events to open-sentinel DataEvents."""
        stream = self.stub.SubscribeEvents(
            SubscribeEventsRequest(
                event_types=["sync.completed", "sync.failed"]
            )
        )
        
        async for event in stream:
            if event.type == "sync.completed":
                payload = event.sync_completed_payload
                
                for delta in payload.new_resources + payload.modified_resources:
                    yield DataEvent(
                        event_type=f"resource.{delta.operation}",
                        resource_type=delta.resource_type,
                        resource_id=delta.resource_id,
                        site_id=delta.site_origin,
                        timestamp=datetime.utcnow(),
                        metadata={
                            "sync_id": payload.sync_id,
                            "peer_node_id": payload.peer_node_id,
                            "git_path": delta.path,
                        },
                    )
                
                yield DataEvent(
                    event_type="sync.completed",
                    resource_type="",
                    resource_id=payload.sync_id,
                    site_id=None,
                    timestamp=datetime.utcnow(),
                    metadata={
                        "records_received": payload.records_received,
                        "records_sent": payload.records_sent,
                        "conflicts_found": payload.conflicts_found,
                    },
                )
```

### 3.2 FhirGitAdapter Configuration

```python
from open_sentinel.adapters import FhirGitAdapter

adapter = FhirGitAdapter(
    repo_path="/var/lib/open-nucleus/data",
    sqlite_path="/var/lib/open-nucleus/index.db",
    event_source=sync_subscriber,
)
```

The `FhirGitAdapter` translates open-sentinel's generic query interface into SQLite queries against the Open Nucleus index:

| open-sentinel Query | SQLite Translation |
|--------------------|-------------------|
| `query("Condition", {"code": "A00"})` | `SELECT * FROM conditions WHERE code LIKE 'A00%'` |
| `aggregate("Condition", ["site_id", "week"], "count", {...})` | `SELECT site_id, strftime('%Y-W%W', date) as week, COUNT(*) ... GROUP BY 1, 2` |
| `count("Encounter", {"status": "finished"})` | `SELECT COUNT(*) FROM encounters WHERE status = 'finished'` |

### 3.3 Alert Write-Back (→ Patient Service)

Expanded for V3.1's richer alert structure with AI provenance and reflection metadata:

```python
class NucleusFhirFlagOutput(AlertOutput):
    """Write alerts to Open Nucleus Patient Service as FHIR DetectedIssue."""
    
    def __init__(self, patient_service_address: str):
        self.channel = grpc.aio.insecure_channel(patient_service_address)
        self.stub = PatientServiceStub(self.channel)
    
    def name(self): return "nucleus-fhir-flag"
    
    def accepts(self, alert):
        return alert.patient_id is not None or alert.site_id is not None
    
    async def emit(self, alert: Alert) -> bool:
        detected_issue = self._alert_to_fhir(alert)
        response = await self.stub.CreateDetectedIssue(
            CreateDetectedIssueRequest(
                resource_json=json.dumps(detected_issue).encode(),
                patient_id=alert.patient_id,
            )
        )
        return response.success
    
    def _alert_to_fhir(self, alert: Alert) -> dict:
        resource = {
            "resourceType": "DetectedIssue",
            "status": "preliminary",
            "severity": self._map_severity(alert.severity),
            "code": {
                "coding": [{
                    "system": "https://open-nucleus.dev/sentinel/alert-category",
                    "code": alert.category,
                    "display": alert.title,
                }]
            },
            "detail": alert.description,
            "identified": alert.created_at.isoformat() + "Z",
        }
        
        if alert.patient_id:
            resource["patient"] = {"reference": f"Patient/{alert.patient_id}"}
        
        # AI provenance — always present on every alert
        tags = []
        
        if alert.ai_generated:
            tags.append({
                "system": "https://open-nucleus.dev/sentinel/ai-provenance",
                "code": "ai-generated",
                "display": (f"AI-generated by {alert.ai_model}, "
                           f"confidence: {alert.ai_confidence:.2f}, "
                           f"reflections: {alert.reflection_iterations}"),
            })
        else:
            tags.append({
                "system": "https://open-nucleus.dev/sentinel/ai-provenance",
                "code": "rule-only",
                "display": "Rule-based detection (LLM unavailable)",
            })
        
        if alert.rule_validated:
            tags.append({
                "system": "https://open-nucleus.dev/sentinel/ai-provenance",
                "code": "rule-validated",
                "display": "LLM finding confirmed by deterministic rules",
            })
        
        resource["meta"] = {"tag": tags}
        
        # AI reasoning as extension
        extensions = []
        if alert.ai_reasoning:
            extensions.append({
                "url": "https://open-nucleus.dev/sentinel/ai-reasoning",
                "valueString": alert.ai_reasoning,
            })
        
        if alert.evidence:
            extensions.append({
                "url": "https://open-nucleus.dev/sentinel/evidence",
                "valueString": json.dumps(alert.evidence),
            })
        
        if extensions:
            resource["extension"] = extensions
        
        return resource
    
    def _map_severity(self, sentinel_severity: str) -> str:
        return {
            "critical": "high",
            "high": "high",
            "moderate": "moderate",
            "low": "low",
        }.get(sentinel_severity, "moderate")
```

### 3.4 Ollama Sidecar with Watchdog

The LLM runs as a separate process. V2 adds crash recovery and health monitoring:

```python
class OllamaSidecar:
    """Manage the Ollama process with watchdog for crash recovery."""
    
    def __init__(self, config: OllamaConfig):
        self.config = config
        self.process = None
        self._restart_count = 0
        self._max_restarts = 5
        self._last_restart = None
    
    async def start(self):
        if not self.config.enabled:
            return
        if await self._is_running():
            return
        
        self.process = await asyncio.create_subprocess_exec(
            "ollama", "serve",
            env={
                **os.environ,
                "OLLAMA_HOST": f"127.0.0.1:{self.config.port}",
                "OLLAMA_NUM_PARALLEL": "1",
                "OLLAMA_MAX_LOADED_MODELS": "1",
            }
        )
        await self._wait_ready(timeout=60)
        await self._ensure_model(self.config.model)
    
    async def stop(self):
        if self.process:
            self.process.terminate()
            await self.process.wait()
    
    async def health(self) -> bool:
        return await self._is_running()
    
    async def watchdog_loop(self):
        """Run as background task. Detects Ollama crashes and restarts."""
        while True:
            await asyncio.sleep(30)
            
            if not self.config.enabled:
                continue
            
            if not await self._is_running():
                if self._restart_count >= self._max_restarts:
                    # Too many restarts — give up, agent runs degraded
                    logging.error(
                        "Ollama exceeded max restarts (%d). "
                        "Agent running in degraded mode.",
                        self._max_restarts,
                    )
                    continue
                
                logging.warning("Ollama not running. Attempting restart...")
                self._restart_count += 1
                self._last_restart = datetime.utcnow()
                
                try:
                    await self.start()
                    logging.info("Ollama restarted successfully.")
                except Exception as e:
                    logging.error("Ollama restart failed: %s", e)
    
    def status(self) -> dict:
        return {
            "enabled": self.config.enabled,
            "running": self.process is not None and self.process.returncode is None,
            "model": self.config.model,
            "restart_count": self._restart_count,
            "last_restart": self._last_restart.isoformat() if self._last_restart else None,
        }
```

---

## 4. Service Startup

```python
# services/sentinel/main.py

import asyncio
from open_sentinel import SentinelAgent, AgentConfig
from open_sentinel.adapters import FhirGitAdapter
from open_sentinel.llm import OllamaEngine
from open_sentinel.memory import MemoryStore
from open_sentinel.resources import ResourceManager
from open_sentinel.outputs import FileOutput, WebhookOutput, ConsoleOutput
from open_sentinel.skills import (
    IdsrCholeraSkill, IdsrMeaslesSkill, IdsrMeningitisSkill,
    IdsrYellowFeverSkill, IdsrEbolaSkill,
    MalariaTrendSkill, MedicationMissedDoseSkill,
    MedicationInteractionRetroSkill,
    StockoutPredictionSkill, StockoutCriticalSkill,
    ImmunisationGapSkill, TbTreatmentSkill,
    MaternalRiskSkill, MissedReferralSkill,
    VitalSignTrendSkill, SyndromicSurveillanceSkill,
)


async def main():
    config = load_config("/etc/open-nucleus/sentinel.yaml")
    
    # 1. Start Ollama sidecar (if configured)
    ollama_sidecar = OllamaSidecar(config.ollama)
    await ollama_sidecar.start()
    
    # 2. Connect to Sync Service event stream
    sync_subscriber = SyncEventSubscriber(config.sync_grpc_address)
    
    # 3. Configure data adapter
    adapter = FhirGitAdapter(
        repo_path=config.repo_path,
        sqlite_path=config.sqlite_path,
        event_source=sync_subscriber,
    )
    
    # 4. Configure LLM engine
    llm = None
    if config.ollama.enabled and await ollama_sidecar.health():
        llm = OllamaEngine(
            base_url=f"http://127.0.0.1:{config.ollama.port}",
            model=config.ollama.model,
            timeout=config.ollama.timeout,
            allow_patient_data=config.ollama.allow_patient_data,
        )
    
    # 5. Configure alert outputs
    outputs = [
        NucleusFhirFlagOutput(config.patient_service_address),
        FileOutput(path=config.alert_log_path),
    ]
    if config.sms.enabled:
        outputs.append(SmsOutput(config.sms))
    if config.webhook.enabled:
        outputs.append(WebhookOutput(config.webhook))
    
    # 6. Load skills
    skills = load_enabled_skills(config.skills)
    if config.custom_skills_dir and Path(config.custom_skills_dir).exists():
        skills.extend(load_skills_from_directory(config.custom_skills_dir))
    
    # 7. Create agent with V3.1 configuration
    agent = SentinelAgent(
        data_adapter=adapter,
        llm=llm,
        skills=skills,
        outputs=outputs,
        config=AgentConfig(
            state_db_path=config.state_db_path,
            hardware=config.hardware_profile,     # NEW: pi4_8gb, hub_16gb, etc.
            skill_config=config.skill_overrides,
            max_critical_per_hour=config.max_critical_per_hour,
        ),
    )
    
    # 8. Start background tasks
    asyncio.create_task(ollama_sidecar.watchdog_loop())
    
    # 9. Start health/management/feedback HTTP server
    health_server = HealthServer(agent, ollama_sidecar, port=config.http_port)
    asyncio.create_task(health_server.serve())
    
    # 10. Run the sleeper loop (blocks forever)
    await agent.run()


if __name__ == "__main__":
    asyncio.run(main())
```

---

## 5. Health, Management, and Feedback API

### 5.1 Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Service health check (includes LLM status) |
| `/status` | GET | Agent status: loaded skills, hardware profile, memory stats, last run |
| `/skills` | GET | List loaded skills with priority, category, degraded status |
| `/skills/{name}/run` | POST | Manually trigger a specific skill |
| `/sweep` | POST | Run all scheduled skills immediately |
| `/alerts/recent` | GET | Recent alerts from memory store (default: last 50) |
| `/alerts/pending` | GET | Alerts awaiting clinician review (`outcome = pending`) |
| `/alerts/{id}/feedback` | POST | Submit clinician feedback (confirm/dismiss/modify) |
| `/queue` | GET | Pending emission queue |
| `/llm/status` | GET | Ollama health, model info, restart count |
| `/memory/episodes` | GET | Recent episodic memory entries |
| `/memory/baselines` | GET | Learned baselines by skill and site |
| `/events/recent` | GET | Recent lifecycle events (last 100) |

### 5.2 Feedback Endpoint

The critical new integration point. Clinician feedback flows back into open-sentinel's learning/calibration loop:

```python
class HealthServer:
    
    async def handle_feedback(self, request):
        """POST /alerts/{id}/feedback
        
        Body:
        {
            "outcome": "confirmed" | "dismissed" | "modified",
            "feedback": "Optional clinician note",
            "modified_severity": "Optional new severity if modified",
            "reviewer_id": "clinician-001"
        }
        """
        alert_id = request.match_info["id"]
        body = await request.json()
        
        outcome = body["outcome"]
        feedback = body.get("feedback")
        
        # Delegate to open-sentinel's feedback processing
        # This triggers:
        #   1. Alert outcome update in memory
        #   2. Episode outcome update
        #   3. Skill confidence threshold calibration
        #   4. skill.calibrated event emission
        await self.agent.process_feedback(
            alert_id=alert_id,
            outcome=outcome,
            feedback=feedback,
        )
        
        return web.json_response({
            "status": "ok",
            "alert_id": alert_id,
            "outcome": outcome,
        })
```

### 5.3 Status Response

```json
{
    "agent": {
        "status": "running",
        "hardware_profile": "pi4_8gb",
        "uptime_seconds": 86400,
        "last_wake": "2026-03-02T06:00:00Z",
        "next_scheduled_wake": "2026-03-02T12:00:00Z"
    },
    "skills": {
        "loaded": 16,
        "degraded": 1,
        "last_sweep": "2026-03-02T06:00:12Z",
        "alerts_today": 3
    },
    "llm": {
        "available": true,
        "model": "phi3:mini",
        "inference_count_today": 47,
        "avg_inference_ms": 22400,
        "ollama_restarts": 0
    },
    "memory": {
        "episodes_stored": 1240,
        "baselines_tracked": 48,
        "alerts_pending_review": 2,
        "emission_queue_depth": 0
    }
}
```

---

## 6. Configuration

```yaml
# /etc/open-nucleus/sentinel.yaml

sentinel:
  http_port: 8090
  state_db_path: /var/lib/open-nucleus/sentinel-state.db
  alert_log_path: /var/log/open-nucleus/sentinel-alerts.jsonl
  custom_skills_dir: /etc/open-nucleus/sentinel-skills/
  
  # Hardware profile — determines reflection limits, concurrent skills, model
  # Options: pi4_4gb, pi4_8gb, uconsole_8gb, hub_16gb, hub_32gb
  hardware_profile: pi4_8gb
  
  # Safety
  max_critical_per_hour: 10
  
  # Open Nucleus integration
  repo_path: /var/lib/open-nucleus/data
  sqlite_path: /var/lib/open-nucleus/index.db
  sync_grpc_address: localhost:50052
  patient_service_address: localhost:50051
  
  # Skills (with optional config overrides)
  skills:
    - name: idsr-cholera
      enabled: true
    - name: idsr-measles
      enabled: true
    - name: idsr-meningitis
      enabled: true
      config:
        belt_threshold: 10
    - name: idsr-yellow-fever
      enabled: true
    - name: idsr-ebola
      enabled: true
    - name: malaria-trend
      enabled: true
      config:
        season_start_month: 4
        alert_threshold_multiplier: 1.5
    - name: medication-missed-dose
      enabled: true
    - name: medication-interaction-retro
      enabled: true
    - name: stockout-prediction
      enabled: true
      config:
        warning_days: 14
        critical_days: 7
    - name: stockout-critical
      enabled: true
    - name: immunisation-gap
      enabled: true
    - name: tb-treatment-completion
      enabled: true
    - name: maternal-risk-scoring
      enabled: true
    - name: missed-referral
      enabled: true
    - name: vital-sign-trend
      enabled: true
    - name: syndromic-surveillance
      enabled: true
  
  # Alert routing
  sms:
    enabled: true
    provider: africas-talking
    api_key_env: AT_API_KEY
    recipients:
      - "+234XXXXXXXXXX"
    filter:
      min_severity: critical
  
  webhook:
    enabled: false
    url: ""
  
  # Ollama sidecar
  ollama:
    enabled: true
    port: 11434
    model: phi3:mini
    allow_patient_data: true    # Safe: runs locally
    timeout: 60
    max_restarts: 5
```

---

## 7. Event Forwarding

The Sentinel Agent forwards open-sentinel lifecycle events to the Open Nucleus event bus for monitoring and dashboards:

```python
class NucleusEventForwarder:
    """Forward open-sentinel events to Open Nucleus event bus."""
    
    def __init__(self, agent: SentinelAgent, event_bus_address: str):
        self.agent = agent
        self.event_bus = NucleusEventBus(event_bus_address)
    
    def register(self):
        """Register hooks on open-sentinel's event bus."""
        events_to_forward = [
            "agent.started", "agent.wake", "agent.sleep",
            "skill.started", "skill.completed", "skill.degraded", "skill.error",
            "skill.reflecting", "skill.calibrated", "skill.deferred",
            "llm.inference.started", "llm.inference.completed", "llm.crashed",
            "alert.emitted", "alert.gated", "alert.reviewed",
        ]
        
        for event_name in events_to_forward:
            self.agent.events.subscribe(event_name, self._forward)
    
    async def _forward(self, event_name, payload):
        await self.event_bus.publish(
            topic=f"sentinel.{event_name}",
            payload=payload,
            source="sentinel-agent",
        )
```

---

## 8. Deployment

### 8.1 Systemd Service

```ini
[Unit]
Description=Open Nucleus Sentinel Agent
After=open-nucleus-patient.service open-nucleus-sync.service
Wants=open-nucleus-sync.service

[Service]
Type=simple
User=open-nucleus
ExecStart=/usr/bin/python3 -m services.sentinel.main
WorkingDirectory=/opt/open-nucleus
Environment=SENTINEL_CONFIG=/etc/open-nucleus/sentinel.yaml
Restart=always
RestartSec=10
MemoryMax=100M

[Install]
WantedBy=multi-user.target
```

### 8.2 Ollama Sidecar Service

```ini
[Unit]
Description=Ollama LLM Sidecar for Sentinel
After=network.target

[Service]
Type=simple
User=open-nucleus
ExecStart=/usr/bin/ollama serve
Environment=OLLAMA_HOST=127.0.0.1:11434
Environment=OLLAMA_NUM_PARALLEL=1
Environment=OLLAMA_MAX_LOADED_MODELS=1
Restart=always
RestartSec=30
MemoryMax=6G

[Install]
WantedBy=multi-user.target
```

### 8.3 Hardware Configurations

| Profile | RAM | LLM | Concurrent Skills | Max Reflections | Model | Notes |
|---------|-----|-----|-------------------|-----------------|-------|-------|
| `pi4_4gb` | 4GB | No | 4 | 0 | None | Rules only. All skills run, LLM-only skills skip. |
| `pi4_8gb` | 8GB | Yes | 2 | 2 | phi3:mini (3.8B) | Primary field deployment target. |
| `uconsole_8gb` | 8GB | Yes | 2 | 2 | phi3:mini (3.8B) | Portable deployment. |
| `hub_16gb` | 16GB | Yes | 4 | 3 | llama3.2:3b | Regional hub. |
| `hub_32gb` | 32GB | Yes | 8 | 3 | llama3.2:8b | District/national hub. |

### 8.4 Without LLM

```yaml
ollama:
  enabled: false
hardware_profile: pi4_4gb
```

All 15 rule-based skills run on < 60MB. The syndromic surveillance skill (LLM-only) gracefully skips.

---

## 9. Error Handling

| Error | Behaviour |
|-------|-----------|
| **Sync Service disconnected** | Retry connection every 30s. Scheduled skills still run on their cron. |
| **Patient Service unavailable** | Alerts queued in `emission_queue`. Retried with exponential backoff (30s → 120s, max 10 attempts). |
| **Ollama crash/OOM** | Watchdog detects within 30s. Auto-restart (up to 5 times). Agent switches to degraded mode immediately. `llm.crashed` event emitted. |
| **Ollama timeout** | Individual inference killed after config timeout (default 60s). Skill falls back to rules for that run. Other skills continue. |
| **Skill throws exception** | Caught and logged. `skill.error` event emitted. Other skills continue. Skill retried on next trigger. |
| **SQLite locked** | Retry with backoff (100ms → 400ms, max 5 attempts). WAL mode reduces contention. |
| **Disk full** | Alert to console. Reduce log retention. Memory store stops writing non-critical episodic entries. |
| **Max critical alerts exceeded** | Guardrail pipeline gates further critical alerts from that skill for the hour. `alert.gated` event with reason `rate_limited`. |

---

## 10. Testing

### 10.1 Integration Tests

| Test | Description |
|------|-------------|
| Sync event → alert | Mock SyncCompleted with new Condition → IDSR skill runs → DetectedIssue written to Patient Service |
| Reflection loop | Mock LLM returns hallucinated finding → `critique_findings()` catches it → LLM refines → correct alert emitted |
| Degraded mode | Disable Ollama → rule skills run, LLM skills use fallback, alerts tagged `ai_generated: false` |
| Feedback calibration | Generate alert → submit "dismissed" feedback → verify confidence threshold increased |
| SMS routing | Generate critical alert → verify SMS output receives it |
| Episodic memory | Run skill → generate alert → run skill again → verify episodes injected into LLM context |
| Custom skill loading | Drop skill folder in custom dir → restart → verify loaded and runs |
| Ollama watchdog | Kill Ollama process → verify watchdog restarts it within 60s |
| Priority ordering | Trigger 3 skills simultaneously → verify CRITICAL runs first, gets LLM slot first |
| Emission queue | Disconnect Patient Service → generate alert → verify queued → reconnect → verify delivered |

### 10.2 End-to-End Scenarios

| Scenario | Setup | Expected |
|----------|-------|----------|
| Cholera outbreak | Sync 3 cholera cases from remote site to non-endemic node | Critical alert with LLM reasoning, SMS to admin, DetectedIssue written, episode stored |
| Hallucination caught | LLM claims 5 cases at a site with 0 actual cases | Reflection loop catches it, no false alert emitted |
| Clinician confirms | Cholera alert reviewed, clinician confirms | Episode updated, threshold unchanged, `alert.reviewed` event |
| Clinician dismisses | False positive alert dismissed with feedback | Confidence threshold raised by 0.05, episode updated with feedback, future alerts require higher confidence |
| LLM goes down | Ollama OOMs mid-sweep | Agent switches to degraded mode for remaining skills, watchdog restarts Ollama, next sweep uses LLM again |
| Stockout | Amoxicillin stock below 7-day threshold | Critical stockout alert with LLM-generated redistribution recommendation |

---

## 11. Performance Targets

All targets on Raspberry Pi 4 (8GB, phi3:mini).

| Operation | Target | Notes |
|-----------|--------|-------|
| Service startup | < 5s | Including skill loading, adapter connection, Ollama health check |
| Event processing (rules only) | < 2s | SyncCompleted → rule alerts emitted |
| Event processing (with LLM) | < 90s | SyncCompleted → LLM analysis + 2 reflections → alerts emitted |
| Full sweep (rules only) | < 10s | All 16 skills, rule fallback |
| Full sweep (with LLM) | < 5 min | All 16 skills sequentially through priority queue |
| Feedback processing | < 100ms | Clinician feedback → calibration complete |
| Alert write-back | < 200ms | gRPC to Patient Service |
| Memory (no LLM) | < 60MB RSS | Agent + all skills + memory store |
| Memory (Ollama idle) | < 200MB RSS | Ollama loads model on demand |
| Memory (Ollama inferring) | < 4GB | phi3:mini during inference |

---

## 12. File Structure

```
services/sentinel/
├── main.py                    # Service entry point
├── config.py                  # Config loader + validation
├── sync_subscriber.py         # SyncEventSubscriber (← Sync Service gRPC)
├── fhir_output.py             # NucleusFhirFlagOutput (→ Patient Service gRPC)
├── ollama_sidecar.py          # OllamaSidecar with watchdog
├── event_forwarder.py         # NucleusEventForwarder
├── health_server.py           # HTTP API: health, management, feedback
├── proto/                     # gRPC proto definitions
│   ├── sync_events.proto
│   └── patient_service.proto
└── tests/
    ├── test_integration.py
    ├── test_feedback.py
    ├── test_fhir_output.py
    └── test_scenarios.py
```

---

*Open Nucleus • Sentinel Agent Service Specification V2 • FibrinLab*  
*github.com/FibrinLab/open-nucleus*
