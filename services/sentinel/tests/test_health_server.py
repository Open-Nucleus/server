import pytest
from aiohttp import web
from aiohttp.test_utils import AioHTTPTestCase, TestClient, TestServer

from sentinel.http.health_server import HealthServer
from sentinel.store.seed import seed_stores


@pytest.fixture
def stores():
    return seed_stores()


@pytest.fixture
async def http_client(stores, aiohttp_client):
    alert_store, inventory_store = stores
    server = HealthServer(
        alert_store=alert_store,
        inventory_store=inventory_store,
        port=0,
        skills=["idsr_cholera", "stockout_prediction"],
    )
    app = server.create_app()
    client = await aiohttp_client(app)
    return client


@pytest.mark.asyncio
class TestHealthServer:
    async def test_health(self, http_client):
        resp = await http_client.get("/health")
        assert resp.status == 200
        data = await resp.json()
        assert data["status"] == "healthy"
        assert data["service"] == "sentinel-agent"
        assert "uptime_seconds" in data

    async def test_status(self, http_client):
        resp = await http_client.get("/status")
        assert resp.status == 200
        data = await resp.json()
        assert data["state"] == "running"
        assert data["skills_loaded"] == 2
        assert data["alert_summary"]["total"] == 5

    async def test_skills(self, http_client):
        resp = await http_client.get("/skills")
        assert resp.status == 200
        data = await resp.json()
        assert len(data["skills"]) == 2
        assert data["skills"][0]["status"] == "stub"

    async def test_alerts_recent(self, http_client):
        resp = await http_client.get("/alerts/recent")
        assert resp.status == 200
        data = await resp.json()
        assert len(data["alerts"]) == 5

    async def test_alerts_pending(self, http_client):
        """Run before feedback mutations."""
        resp = await http_client.get("/alerts/pending")
        assert resp.status == 200
        data = await resp.json()
        assert len(data["alerts"]) >= 3  # at least 3 active

    async def test_feedback_confirm(self, http_client):
        resp = await http_client.post(
            "/alerts/alert-001/feedback",
            json={"outcome": "confirmed", "reviewer_id": "dr-smith"},
        )
        assert resp.status == 200
        data = await resp.json()
        assert data["outcome"] == "confirmed"
        assert data["alert"]["status"] == "acknowledged"

    async def test_feedback_not_found(self, http_client):
        resp = await http_client.post(
            "/alerts/nonexistent/feedback",
            json={"outcome": "confirmed"},
        )
        assert resp.status == 404

    async def test_feedback_invalid_outcome(self, http_client):
        resp = await http_client.post(
            "/alerts/alert-001/feedback",
            json={"outcome": "invalid"},
        )
        assert resp.status == 400

    async def test_queue(self, http_client):
        resp = await http_client.get("/queue")
        assert resp.status == 200
        data = await resp.json()
        assert data["pending"] == 0

    async def test_llm_status(self, http_client):
        resp = await http_client.get("/llm/status")
        assert resp.status == 200
        data = await resp.json()
        assert data["enabled"] is False

    async def test_memory_episodes(self, http_client):
        resp = await http_client.get("/memory/episodes")
        assert resp.status == 200
        data = await resp.json()
        assert data["episodes"] == []

    async def test_memory_baselines(self, http_client):
        resp = await http_client.get("/memory/baselines")
        assert resp.status == 200

    async def test_events_recent(self, http_client):
        resp = await http_client.get("/events/recent")
        assert resp.status == 200
