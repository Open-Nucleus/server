import grpc
import pytest

from sentinel.gen.sentinel.v1 import sentinel_pb2 as pb
from sentinel.gen.common.v1 import metadata_pb2 as common_pb


@pytest.mark.asyncio
class TestGrpcServicer:
    async def test_get_alert_summary(self, grpc_stub):
        resp = await grpc_stub.GetAlertSummary(pb.GetAlertSummaryRequest())
        assert resp.total == 5
        assert resp.critical == 2
        assert resp.warning == 2
        assert resp.info == 1
        assert resp.unacknowledged >= 3  # may be less if other tests mutated store

    async def test_list_alerts(self, grpc_stub):
        resp = await grpc_stub.ListAlerts(pb.ListAlertsRequest(
            pagination=common_pb.PaginationRequest(page=1, per_page=10),
        ))
        assert len(resp.alerts) == 5
        assert resp.pagination.total == 5

    async def test_list_alerts_filter_severity(self, grpc_stub):
        resp = await grpc_stub.ListAlerts(pb.ListAlertsRequest(
            severity="critical",
            pagination=common_pb.PaginationRequest(page=1, per_page=10),
        ))
        assert len(resp.alerts) == 2
        assert all(a.severity == "critical" for a in resp.alerts)

    async def test_get_alert(self, grpc_stub):
        resp = await grpc_stub.GetAlert(pb.GetAlertRequest(alert_id="alert-001"))
        assert resp.alert.id == "alert-001"
        assert resp.alert.severity == "critical"
        assert resp.alert.title == "Cholera cluster detected — Boma District"

    async def test_get_alert_not_found(self, grpc_stub):
        with pytest.raises(grpc.aio.AioRpcError) as exc_info:
            await grpc_stub.GetAlert(pb.GetAlertRequest(alert_id="nonexistent"))
        assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND

    async def test_get_alert_empty_id(self, grpc_stub):
        with pytest.raises(grpc.aio.AioRpcError) as exc_info:
            await grpc_stub.GetAlert(pb.GetAlertRequest(alert_id=""))
        assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT

    async def test_acknowledge_alert(self, grpc_stub):
        resp = await grpc_stub.AcknowledgeAlert(pb.AcknowledgeAlertRequest(
            alert_id="alert-001",
            acknowledged_by="dr-smith",
        ))
        assert resp.alert.status == "acknowledged"
        assert resp.alert.acknowledged_by == "dr-smith"

    async def test_acknowledge_alert_not_found(self, grpc_stub):
        with pytest.raises(grpc.aio.AioRpcError) as exc_info:
            await grpc_stub.AcknowledgeAlert(pb.AcknowledgeAlertRequest(
                alert_id="nonexistent",
            ))
        assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND

    async def test_dismiss_alert(self, grpc_stub):
        resp = await grpc_stub.DismissAlert(pb.DismissAlertRequest(
            alert_id="alert-002",
            reason="false positive",
        ))
        assert resp.alert.status == "dismissed"

    async def test_get_inventory(self, grpc_stub):
        resp = await grpc_stub.GetInventory(pb.GetInventoryRequest(
            pagination=common_pb.PaginationRequest(page=1, per_page=25),
        ))
        assert len(resp.items) == 10
        assert resp.pagination.total == 10

    async def test_get_inventory_by_site(self, grpc_stub):
        resp = await grpc_stub.GetInventory(pb.GetInventoryRequest(
            site_id="site-alpha",
            pagination=common_pb.PaginationRequest(page=1, per_page=25),
        ))
        assert len(resp.items) == 5
        assert all(i.site_id == "site-alpha" for i in resp.items)

    async def test_get_inventory_item(self, grpc_stub):
        resp = await grpc_stub.GetInventoryItem(pb.GetInventoryItemRequest(
            item_code="AMX250",
            site_id="site-alpha",
        ))
        assert resp.item.item_code == "AMX250"
        assert resp.item.display == "Amoxicillin 250mg caps"

    async def test_get_inventory_item_not_found(self, grpc_stub):
        with pytest.raises(grpc.aio.AioRpcError) as exc_info:
            await grpc_stub.GetInventoryItem(pb.GetInventoryItemRequest(
                item_code="NONEXISTENT",
            ))
        assert exc_info.value.code() == grpc.StatusCode.NOT_FOUND

    async def test_record_delivery(self, grpc_stub):
        resp = await grpc_stub.RecordDelivery(pb.RecordDeliveryRequest(
            site_id="site-alpha",
            items=[pb.DeliveryItem(
                item_code="AMX250",
                quantity=100,
                unit="capsules",
                batch_number="B001",
                expiry_date="2027-12-01",
            )],
            received_by="nurse-jones",
            delivery_date="2026-03-01",
        ))
        assert resp.delivery_id.startswith("del-")
        assert resp.items_recorded == 1

    async def test_record_delivery_no_site(self, grpc_stub):
        with pytest.raises(grpc.aio.AioRpcError) as exc_info:
            await grpc_stub.RecordDelivery(pb.RecordDeliveryRequest(
                items=[pb.DeliveryItem(item_code="X", quantity=1, unit="u")],
            ))
        assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT

    async def test_get_predictions(self, grpc_stub):
        resp = await grpc_stub.GetPredictions(pb.GetPredictionsRequest(
            site_id="site-alpha",
        ))
        assert len(resp.predictions) == 3
        assert resp.predictions[0].item_code == "AMX250"
        assert resp.predictions[0].risk_level == "critical"

    async def test_get_redistribution(self, grpc_stub):
        resp = await grpc_stub.GetRedistribution(pb.GetRedistributionRequest(
            site_id="site-alpha",
        ))
        assert len(resp.suggestions) == 2
        assert resp.suggestions[0].from_site == "site-bravo"
