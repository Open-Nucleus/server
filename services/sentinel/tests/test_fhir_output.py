from sentinel.fhir_output import alert_to_fhir, EmissionQueue
from sentinel.store.models import Alert


def _make_alert(**kwargs) -> Alert:
    defaults = dict(
        id="test-001",
        type="outbreak",
        severity="critical",
        status="active",
        title="Test Alert",
        description="Test description",
        created_at="2026-02-28T10:30:00Z",
        category="idsr_cholera",
    )
    defaults.update(kwargs)
    return Alert(**defaults)


class TestAlertToFhir:
    def test_basic_structure(self):
        fhir = alert_to_fhir(_make_alert())
        assert fhir["resourceType"] == "DetectedIssue"
        assert fhir["status"] == "preliminary"
        assert fhir["severity"] == "high"  # critical maps to high
        assert fhir["detail"] == "Test description"
        assert fhir["identified"] == "2026-02-28T10:30:00Z"

    def test_severity_mapping(self):
        for sentinel_sev, fhir_sev in [
            ("critical", "high"),
            ("high", "high"),
            ("moderate", "moderate"),
            ("low", "low"),
            ("unknown", "moderate"),
        ]:
            fhir = alert_to_fhir(_make_alert(severity=sentinel_sev))
            assert fhir["severity"] == fhir_sev

    def test_patient_reference(self):
        fhir = alert_to_fhir(_make_alert(patient_id="patient-123"))
        assert fhir["patient"] == {"reference": "Patient/patient-123"}

    def test_no_patient_reference(self):
        fhir = alert_to_fhir(_make_alert(patient_id=""))
        assert "patient" not in fhir

    def test_rule_only_provenance(self):
        fhir = alert_to_fhir(_make_alert(ai_generated=False))
        tags = fhir["meta"]["tag"]
        assert any(t["code"] == "rule-only" for t in tags)

    def test_ai_generated_provenance(self):
        fhir = alert_to_fhir(_make_alert(
            ai_generated=True,
            ai_model="gemma2:2b",
            ai_confidence=0.85,
            reflection_iterations=2,
        ))
        tags = fhir["meta"]["tag"]
        ai_tag = next(t for t in tags if t["code"] == "ai-generated")
        assert "gemma2:2b" in ai_tag["display"]
        assert "0.85" in ai_tag["display"]
        assert "2" in ai_tag["display"]

    def test_rule_validated_tag(self):
        fhir = alert_to_fhir(_make_alert(rule_validated=True))
        tags = fhir["meta"]["tag"]
        assert any(t["code"] == "rule-validated" for t in tags)

    def test_ai_reasoning_extension(self):
        fhir = alert_to_fhir(_make_alert(
            ai_reasoning="Cluster analysis showed spatial clustering"
        ))
        exts = fhir.get("extension", [])
        assert any(
            e["url"] == "https://open-nucleus.dev/sentinel/ai-reasoning"
            for e in exts
        )

    def test_evidence_extension(self):
        fhir = alert_to_fhir(_make_alert(evidence={"cases": 5, "threshold": 1}))
        exts = fhir.get("extension", [])
        assert any(
            e["url"] == "https://open-nucleus.dev/sentinel/evidence"
            for e in exts
        )

    def test_no_extensions_when_empty(self):
        fhir = alert_to_fhir(_make_alert())
        assert "extension" not in fhir


class TestEmissionQueue:
    def test_enqueue(self):
        q = EmissionQueue()
        q.enqueue(_make_alert())
        assert q.pending == 1

    def test_entries(self):
        q = EmissionQueue()
        q.enqueue(_make_alert(id="a1"))
        q.enqueue(_make_alert(id="a2"))
        entries = q.entries()
        assert len(entries) == 2
        assert entries[0]["alert_id"] == "a1"
