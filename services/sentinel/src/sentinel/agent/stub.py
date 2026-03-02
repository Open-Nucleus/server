from __future__ import annotations

import logging
from typing import Any

from sentinel.store.alert_store import AlertStore

logger = logging.getLogger(__name__)


class StubAgent:
    """Placeholder agent until open-sentinel is available.

    Delegates feedback to the alert store and logs all operations.
    """

    def __init__(self, alert_store: AlertStore) -> None:
        self._alerts = alert_store

    async def run(self) -> None:
        logger.info(
            "StubAgent: open-sentinel not configured. "
            "Agent will not perform surveillance. "
            "Alerts are served from seed data only."
        )

    async def process_feedback(
        self,
        alert_id: str,
        outcome: str,
        feedback: str = "",
        reviewer_id: str = "",
    ) -> bool:
        logger.info(
            "StubAgent: feedback received for %s: outcome=%s, reviewer=%s",
            alert_id, outcome, reviewer_id,
        )

        if outcome == "confirmed":
            self._alerts.acknowledge(alert_id, reviewer_id)
        elif outcome == "dismissed":
            self._alerts.dismiss(alert_id, feedback or "clinician dismissed")

        return True

    def status(self) -> dict[str, Any]:
        return {
            "agent": "stub",
            "open_sentinel_configured": False,
            "message": "open-sentinel not configured — running with seed data only",
        }
