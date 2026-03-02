from __future__ import annotations

import grpc

from sentinel.gen.sentinel.v1 import sentinel_pb2 as pb
from sentinel.gen.sentinel.v1 import sentinel_pb2_grpc
from sentinel.server.converters import (
    alert_to_proto,
    delivery_item_from_proto,
    inventory_item_to_proto,
    pagination_response,
    prediction_to_proto,
    redistribution_to_proto,
)
from sentinel.store.alert_store import AlertStore
from sentinel.store.inventory_store import InventoryStore


class SentinelServiceServicer(sentinel_pb2_grpc.SentinelServiceServicer):
    def __init__(self, alert_store: AlertStore, inventory_store: InventoryStore) -> None:
        self._alerts = alert_store
        self._inventory = inventory_store

    def ListAlerts(self, request, context):
        page = request.pagination.page if request.pagination.page > 0 else 1
        per_page = request.pagination.per_page if request.pagination.per_page > 0 else 25

        alerts, total = self._alerts.list(
            severity=request.severity,
            status=request.status,
            page=page,
            per_page=per_page,
        )
        return pb.ListAlertsResponse(
            alerts=[alert_to_proto(a) for a in alerts],
            pagination=pagination_response(page, per_page, total),
        )

    def GetAlertSummary(self, request, context):
        s = self._alerts.summary()
        return pb.GetAlertSummaryResponse(
            total=s.total,
            critical=s.critical,
            warning=s.warning,
            info=s.info,
            unacknowledged=s.unacknowledged,
        )

    def GetAlert(self, request, context):
        if not request.alert_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("alert_id is required")
            return pb.GetAlertResponse()

        alert = self._alerts.get(request.alert_id)
        if alert is None:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(f"alert {request.alert_id} not found")
            return pb.GetAlertResponse()

        return pb.GetAlertResponse(alert=alert_to_proto(alert))

    def AcknowledgeAlert(self, request, context):
        if not request.alert_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("alert_id is required")
            return pb.AcknowledgeAlertResponse()

        alert = self._alerts.acknowledge(request.alert_id, request.acknowledged_by)
        if alert is None:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(f"alert {request.alert_id} not found")
            return pb.AcknowledgeAlertResponse()

        return pb.AcknowledgeAlertResponse(alert=alert_to_proto(alert))

    def DismissAlert(self, request, context):
        if not request.alert_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("alert_id is required")
            return pb.DismissAlertResponse()

        alert = self._alerts.dismiss(request.alert_id, request.reason)
        if alert is None:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(f"alert {request.alert_id} not found")
            return pb.DismissAlertResponse()

        return pb.DismissAlertResponse(alert=alert_to_proto(alert))

    def GetInventory(self, request, context):
        page = request.pagination.page if request.pagination.page > 0 else 1
        per_page = request.pagination.per_page if request.pagination.per_page > 0 else 25

        items, total = self._inventory.list(
            site_id=request.site_id,
            page=page,
            per_page=per_page,
        )
        return pb.GetInventoryResponse(
            items=[inventory_item_to_proto(i) for i in items],
            pagination=pagination_response(page, per_page, total),
        )

    def GetInventoryItem(self, request, context):
        if not request.item_code:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("item_code is required")
            return pb.GetInventoryItemResponse()

        item = self._inventory.get(request.item_code, request.site_id)
        if item is None:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(f"inventory item {request.item_code} not found")
            return pb.GetInventoryItemResponse()

        return pb.GetInventoryItemResponse(item=inventory_item_to_proto(item))

    def RecordDelivery(self, request, context):
        if not request.site_id:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("site_id is required")
            return pb.RecordDeliveryResponse()

        if not request.items:
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("at least one delivery item is required")
            return pb.RecordDeliveryResponse()

        line_items = [delivery_item_from_proto(di) for di in request.items]
        record = self._inventory.record_delivery(
            site_id=request.site_id,
            items=line_items,
            received_by=request.received_by,
            delivery_date=request.delivery_date,
        )
        return pb.RecordDeliveryResponse(
            delivery_id=record.delivery_id,
            items_recorded=len(line_items),
        )

    def GetPredictions(self, request, context):
        preds = self._inventory.predictions(site_id=request.site_id)
        return pb.GetPredictionsResponse(
            predictions=[prediction_to_proto(p) for p in preds],
        )

    def GetRedistribution(self, request, context):
        sugs = self._inventory.redistribution(site_id=request.site_id)
        return pb.GetRedistributionResponse(
            suggestions=[redistribution_to_proto(s) for s in sugs],
        )
