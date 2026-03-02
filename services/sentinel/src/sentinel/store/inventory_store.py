from __future__ import annotations

import threading
import uuid
from datetime import datetime, timezone

from sentinel.store.models import (
    DeliveryRecord,
    DeliveryLineItem,
    InventoryItem,
    RedistributionSuggestion,
    SupplyPrediction,
)


class InventoryStore:
    def __init__(self) -> None:
        # keyed by (site_id, item_code)
        self._items: dict[tuple[str, str], InventoryItem] = {}
        self._deliveries: list[DeliveryRecord] = []
        self._predictions: dict[str, list[SupplyPrediction]] = {}  # site_id -> list
        self._redistributions: dict[str, list[RedistributionSuggestion]] = {}  # site_id -> list
        self._lock = threading.Lock()

    def add(self, item: InventoryItem) -> None:
        with self._lock:
            self._items[(item.site_id, item.item_code)] = item

    def add_prediction(self, site_id: str, pred: SupplyPrediction) -> None:
        with self._lock:
            self._predictions.setdefault(site_id, []).append(pred)

    def add_redistribution(self, site_id: str, sug: RedistributionSuggestion) -> None:
        with self._lock:
            self._redistributions.setdefault(site_id, []).append(sug)

    def list(
        self,
        site_id: str = "",
        page: int = 1,
        per_page: int = 25,
    ) -> tuple[list[InventoryItem], int]:
        with self._lock:
            items = list(self._items.values())

        if site_id:
            items = [i for i in items if i.site_id == site_id]

        total = len(items)
        start = (page - 1) * per_page
        end = start + per_page
        return items[start:end], total

    def get(self, item_code: str, site_id: str = "") -> InventoryItem | None:
        with self._lock:
            if site_id:
                return self._items.get((site_id, item_code))
            # If no site_id, return first match
            for key, item in self._items.items():
                if key[1] == item_code:
                    return item
            return None

    def record_delivery(
        self,
        site_id: str,
        items: list[DeliveryLineItem],
        received_by: str,
        delivery_date: str,
    ) -> DeliveryRecord:
        delivery_id = f"del-{uuid.uuid4().hex[:8]}"
        record = DeliveryRecord(
            delivery_id=delivery_id,
            site_id=site_id,
            items=items,
            received_by=received_by,
            delivery_date=delivery_date,
        )

        with self._lock:
            self._deliveries.append(record)
            now = datetime.now(timezone.utc).isoformat()
            for di in items:
                key = (site_id, di.item_code)
                if key in self._items:
                    self._items[key].quantity += di.quantity
                    self._items[key].last_updated = now

        return record

    def predictions(self, site_id: str = "") -> list[SupplyPrediction]:
        with self._lock:
            if site_id:
                return list(self._predictions.get(site_id, []))
            result = []
            for preds in self._predictions.values():
                result.extend(preds)
            return result

    def redistribution(self, site_id: str = "") -> list[RedistributionSuggestion]:
        with self._lock:
            if site_id:
                return list(self._redistributions.get(site_id, []))
            result = []
            for sugs in self._redistributions.values():
                result.extend(sugs)
            return result
