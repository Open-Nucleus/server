from __future__ import annotations

import threading
from datetime import datetime, timezone

from sentinel.store.models import Alert, AlertSummary


class AlertStore:
    def __init__(self) -> None:
        self._alerts: dict[str, Alert] = {}
        self._lock = threading.Lock()

    def add(self, alert: Alert) -> None:
        with self._lock:
            self._alerts[alert.id] = alert

    def get(self, alert_id: str) -> Alert | None:
        with self._lock:
            return self._alerts.get(alert_id)

    def list(
        self,
        severity: str = "",
        status: str = "",
        page: int = 1,
        per_page: int = 25,
    ) -> tuple[list[Alert], int]:
        with self._lock:
            filtered = list(self._alerts.values())

        if severity:
            filtered = [a for a in filtered if a.severity == severity]
        if status:
            filtered = [a for a in filtered if a.status == status]

        total = len(filtered)
        start = (page - 1) * per_page
        end = start + per_page
        return filtered[start:end], total

    def summary(self) -> AlertSummary:
        with self._lock:
            alerts = list(self._alerts.values())

        s = AlertSummary(total=len(alerts))
        for a in alerts:
            if a.severity == "critical":
                s.critical += 1
            elif a.severity == "warning":
                s.warning += 1
            elif a.severity == "info":
                s.info += 1
            if a.status != "acknowledged" and a.status != "dismissed":
                s.unacknowledged += 1
        return s

    def acknowledge(self, alert_id: str, acknowledged_by: str) -> Alert | None:
        with self._lock:
            alert = self._alerts.get(alert_id)
            if alert is None:
                return None
            alert.status = "acknowledged"
            alert.acknowledged_at = datetime.now(timezone.utc).isoformat()
            alert.acknowledged_by = acknowledged_by
            return alert

    def dismiss(self, alert_id: str, reason: str) -> Alert | None:
        with self._lock:
            alert = self._alerts.get(alert_id)
            if alert is None:
                return None
            alert.status = "dismissed"
            alert.description = f"{alert.description} [Dismissed: {reason}]"
            return alert
