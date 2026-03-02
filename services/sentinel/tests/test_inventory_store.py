from sentinel.store.inventory_store import InventoryStore
from sentinel.store.models import DeliveryLineItem, InventoryItem


def _make_item(code: str, site: str = "site-alpha", qty: int = 100) -> InventoryItem:
    return InventoryItem(
        item_code=code,
        display=f"Test {code}",
        quantity=qty,
        unit="units",
        site_id=site,
        last_updated="2026-01-01T00:00:00Z",
        reorder_level=10,
    )


class TestInventoryStore:
    def test_add_and_get(self):
        store = InventoryStore()
        store.add(_make_item("X001"))
        item = store.get("X001", "site-alpha")
        assert item is not None
        assert item.quantity == 100

    def test_get_not_found(self):
        store = InventoryStore()
        assert store.get("nonexistent") is None

    def test_get_without_site(self):
        store = InventoryStore()
        store.add(_make_item("X001"))
        item = store.get("X001")
        assert item is not None

    def test_list_all(self, inventory_store):
        items, total = inventory_store.list()
        assert total == 10  # 5 per site
        assert len(items) == 10

    def test_list_by_site(self, inventory_store):
        items, total = inventory_store.list(site_id="site-alpha")
        assert total == 5
        assert all(i.site_id == "site-alpha" for i in items)

    def test_list_pagination(self, inventory_store):
        items, total = inventory_store.list(page=1, per_page=3)
        assert total == 10
        assert len(items) == 3

    def test_record_delivery(self, inventory_store):
        # Get initial quantity
        initial = inventory_store.get("AMX250", "site-alpha")
        assert initial is not None
        initial_qty = initial.quantity

        line_items = [
            DeliveryLineItem(
                item_code="AMX250",
                quantity=50,
                unit="capsules",
                batch_number="BATCH001",
                expiry_date="2027-06-01",
            ),
        ]
        record = inventory_store.record_delivery(
            site_id="site-alpha",
            items=line_items,
            received_by="nurse-jones",
            delivery_date="2026-03-01",
        )
        assert record.delivery_id.startswith("del-")
        assert len(record.items) == 1

        # Verify quantity increased
        updated = inventory_store.get("AMX250", "site-alpha")
        assert updated.quantity == initial_qty + 50

    def test_predictions(self, inventory_store):
        preds = inventory_store.predictions("site-alpha")
        assert len(preds) == 3
        assert preds[0].item_code == "AMX250"
        assert preds[0].risk_level == "critical"

    def test_predictions_empty_site(self, inventory_store):
        preds = inventory_store.predictions("site-nonexistent")
        assert preds == []

    def test_redistribution(self, inventory_store):
        sugs = inventory_store.redistribution("site-alpha")
        assert len(sugs) == 2
        assert sugs[0].from_site == "site-bravo"
        assert sugs[0].to_site == "site-alpha"

    def test_redistribution_empty_site(self, inventory_store):
        sugs = inventory_store.redistribution("site-nonexistent")
        assert sugs == []
