from sentinel.store.alert_store import AlertStore
from sentinel.store.models import Alert


def _make_alert(id: str, severity: str = "warning", status: str = "active") -> Alert:
    return Alert(
        id=id,
        type="test",
        severity=severity,
        status=status,
        title=f"Test alert {id}",
        description=f"Description for {id}",
        created_at="2026-01-01T00:00:00Z",
    )


class TestAlertStore:
    def test_add_and_get(self):
        store = AlertStore()
        alert = _make_alert("a1")
        store.add(alert)
        assert store.get("a1") is not None
        assert store.get("a1").title == "Test alert a1"

    def test_get_not_found(self):
        store = AlertStore()
        assert store.get("nonexistent") is None

    def test_list_all(self, alert_store):
        alerts, total = alert_store.list()
        assert total == 5
        assert len(alerts) == 5

    def test_list_filter_severity(self, alert_store):
        alerts, total = alert_store.list(severity="critical")
        assert total == 2
        assert all(a.severity == "critical" for a in alerts)

    def test_list_filter_status(self, alert_store):
        alerts, total = alert_store.list(status="active")
        assert total == 5
        assert all(a.status == "active" for a in alerts)

    def test_list_pagination(self):
        store = AlertStore()
        for i in range(10):
            store.add(_make_alert(f"a{i}"))
        page1, total = store.list(page=1, per_page=3)
        assert total == 10
        assert len(page1) == 3
        page2, _ = store.list(page=2, per_page=3)
        assert len(page2) == 3
        page4, _ = store.list(page=4, per_page=3)
        assert len(page4) == 1

    def test_summary(self, alert_store):
        s = alert_store.summary()
        assert s.total == 5
        assert s.critical == 2
        assert s.warning == 2
        assert s.info == 1
        assert s.unacknowledged == 5

    def test_acknowledge(self, alert_store):
        alert = alert_store.acknowledge("alert-001", "dr-smith")
        assert alert is not None
        assert alert.status == "acknowledged"
        assert alert.acknowledged_by == "dr-smith"
        assert alert.acknowledged_at != ""

        # Summary should reflect the change
        s = alert_store.summary()
        assert s.unacknowledged == 4

    def test_acknowledge_not_found(self, alert_store):
        assert alert_store.acknowledge("nonexistent", "dr-smith") is None

    def test_dismiss(self, alert_store):
        alert = alert_store.dismiss("alert-002", "false positive")
        assert alert is not None
        assert alert.status == "dismissed"
        assert "false positive" in alert.description

    def test_dismiss_not_found(self, alert_store):
        assert alert_store.dismiss("nonexistent", "reason") is None
