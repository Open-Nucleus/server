from __future__ import annotations

import time
from dataclasses import asdict

from aiohttp import web

from sentinel.store.alert_store import AlertStore
from sentinel.store.inventory_store import InventoryStore


class HealthServer:
    def __init__(
        self,
        alert_store: AlertStore,
        inventory_store: InventoryStore,
        ollama_sidecar=None,
        port: int = 8090,
        skills: list[str] | None = None,
    ) -> None:
        self._alerts = alert_store
        self._inventory = inventory_store
        self._ollama = ollama_sidecar
        self._port = port
        self._skills = skills or []
        self._start_time = time.time()
        self._app: web.Application | None = None

    def create_app(self) -> web.Application:
        app = web.Application()
        app.router.add_get("/health", self.handle_health)
        app.router.add_get("/status", self.handle_status)
        app.router.add_get("/skills", self.handle_skills)
        app.router.add_get("/alerts/recent", self.handle_alerts_recent)
        app.router.add_get("/alerts/pending", self.handle_alerts_pending)
        app.router.add_post("/alerts/{id}/feedback", self.handle_feedback)
        app.router.add_get("/queue", self.handle_queue)
        app.router.add_get("/llm/status", self.handle_llm_status)
        app.router.add_get("/memory/episodes", self.handle_memory_episodes)
        app.router.add_get("/memory/baselines", self.handle_memory_baselines)
        app.router.add_get("/events/recent", self.handle_events_recent)
        self._app = app
        return app

    async def serve(self) -> None:
        app = self.create_app()
        runner = web.AppRunner(app)
        await runner.setup()
        site = web.TCPSite(runner, "0.0.0.0", self._port)
        await site.start()

    async def handle_health(self, request: web.Request) -> web.Response:
        llm_healthy = False
        if self._ollama is not None:
            llm_healthy = self._ollama.status().get("running", False)

        return web.json_response({
            "status": "healthy",
            "service": "sentinel-agent",
            "uptime_seconds": int(time.time() - self._start_time),
            "llm_available": llm_healthy,
        })

    async def handle_status(self, request: web.Request) -> web.Response:
        summary = self._alerts.summary()
        return web.json_response({
            "state": "running",
            "uptime_seconds": int(time.time() - self._start_time),
            "skills_loaded": len(self._skills),
            "skills": self._skills,
            "alert_summary": {
                "total": summary.total,
                "critical": summary.critical,
                "warning": summary.warning,
                "info": summary.info,
                "unacknowledged": summary.unacknowledged,
            },
            "open_sentinel_configured": False,
        })

    async def handle_skills(self, request: web.Request) -> web.Response:
        skill_list = []
        for s in self._skills:
            skill_list.append({
                "name": s,
                "status": "stub",
                "message": "open-sentinel not configured",
            })
        return web.json_response({"skills": skill_list})

    async def handle_alerts_recent(self, request: web.Request) -> web.Response:
        limit = int(request.query.get("limit", "50"))
        alerts, _ = self._alerts.list(page=1, per_page=limit)
        return web.json_response({
            "alerts": [_alert_dict(a) for a in alerts],
        })

    async def handle_alerts_pending(self, request: web.Request) -> web.Response:
        alerts, _ = self._alerts.list(status="active", page=1, per_page=100)
        return web.json_response({
            "alerts": [_alert_dict(a) for a in alerts],
        })

    async def handle_feedback(self, request: web.Request) -> web.Response:
        alert_id = request.match_info["id"]
        alert = self._alerts.get(alert_id)
        if alert is None:
            return web.json_response(
                {"error": f"alert {alert_id} not found"},
                status=404,
            )

        body = await request.json()
        outcome = body.get("outcome", "")
        if outcome not in ("confirmed", "dismissed", "modified"):
            return web.json_response(
                {"error": "outcome must be one of: confirmed, dismissed, modified"},
                status=400,
            )

        reviewer = body.get("reviewer_id", "")

        if outcome == "confirmed":
            self._alerts.acknowledge(alert_id, reviewer)
        elif outcome == "dismissed":
            reason = body.get("feedback", "clinician dismissed")
            self._alerts.dismiss(alert_id, reason)
        elif outcome == "modified":
            self._alerts.acknowledge(alert_id, reviewer)

        updated = self._alerts.get(alert_id)
        return web.json_response({
            "alert_id": alert_id,
            "outcome": outcome,
            "alert": _alert_dict(updated) if updated else None,
        })

    async def handle_queue(self, request: web.Request) -> web.Response:
        return web.json_response({
            "pending": 0,
            "entries": [],
            "message": "emission queue stub — open-sentinel not configured",
        })

    async def handle_llm_status(self, request: web.Request) -> web.Response:
        if self._ollama is not None:
            return web.json_response(self._ollama.status())
        return web.json_response({
            "enabled": False,
            "running": False,
            "model": None,
            "restart_count": 0,
            "last_restart": None,
        })

    async def handle_memory_episodes(self, request: web.Request) -> web.Response:
        return web.json_response({
            "episodes": [],
            "message": "memory store stub — open-sentinel not configured",
        })

    async def handle_memory_baselines(self, request: web.Request) -> web.Response:
        return web.json_response({
            "baselines": [],
            "message": "memory store stub — open-sentinel not configured",
        })

    async def handle_events_recent(self, request: web.Request) -> web.Response:
        return web.json_response({
            "events": [],
            "message": "event bus stub — open-sentinel not configured",
        })


def _alert_dict(a) -> dict:
    return {
        "id": a.id,
        "type": a.type,
        "severity": a.severity,
        "status": a.status,
        "title": a.title,
        "description": a.description,
        "patient_id": a.patient_id,
        "created_at": a.created_at,
        "acknowledged_at": a.acknowledged_at,
        "acknowledged_by": a.acknowledged_by,
    }
