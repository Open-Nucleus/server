from __future__ import annotations

from sentinel.gen.sentinel.v1 import sentinel_pb2 as pb
from sentinel.gen.common.v1 import metadata_pb2 as common_pb
from sentinel.store.models import (
    Alert,
    DeliveryLineItem,
    InventoryItem,
    RedistributionSuggestion,
    SupplyPrediction,
)


def alert_to_proto(a: Alert) -> pb.Alert:
    return pb.Alert(
        id=a.id,
        type=a.type,
        severity=a.severity,
        status=a.status,
        title=a.title,
        description=a.description,
        patient_id=a.patient_id,
        created_at=a.created_at,
        acknowledged_at=a.acknowledged_at,
        acknowledged_by=a.acknowledged_by,
    )


def inventory_item_to_proto(item: InventoryItem) -> pb.InventoryItem:
    return pb.InventoryItem(
        item_code=item.item_code,
        display=item.display,
        quantity=item.quantity,
        unit=item.unit,
        site_id=item.site_id,
        last_updated=item.last_updated,
        reorder_level=item.reorder_level,
    )


def prediction_to_proto(p: SupplyPrediction) -> pb.SupplyPrediction:
    return pb.SupplyPrediction(
        item_code=p.item_code,
        display=p.display,
        current_quantity=p.current_quantity,
        predicted_days_remaining=p.predicted_days_remaining,
        risk_level=p.risk_level,
        recommended_action=p.recommended_action,
    )


def redistribution_to_proto(r: RedistributionSuggestion) -> pb.RedistributionSuggestion:
    return pb.RedistributionSuggestion(
        item_code=r.item_code,
        from_site=r.from_site,
        to_site=r.to_site,
        suggested_quantity=r.suggested_quantity,
        rationale=r.rationale,
    )


def delivery_item_from_proto(di: pb.DeliveryItem) -> DeliveryLineItem:
    return DeliveryLineItem(
        item_code=di.item_code,
        quantity=di.quantity,
        unit=di.unit,
        batch_number=di.batch_number,
        expiry_date=di.expiry_date,
    )


def pagination_response(page: int, per_page: int, total: int) -> common_pb.PaginationResponse:
    total_pages = (total + per_page - 1) // per_page if per_page > 0 else 0
    return common_pb.PaginationResponse(
        page=page,
        per_page=per_page,
        total=total,
        total_pages=total_pages,
    )
