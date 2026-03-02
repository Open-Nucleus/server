from sentinel.store.alert_store import AlertStore
from sentinel.store.inventory_store import InventoryStore
from sentinel.store.models import (
    Alert,
    InventoryItem,
    RedistributionSuggestion,
    SupplyPrediction,
)

SEED_ALERTS = [
    Alert(
        id="alert-001",
        type="outbreak",
        severity="critical",
        status="active",
        title="Cholera cluster detected — Boma District",
        description="IDSR threshold exceeded: 5 cases of acute watery diarrhoea (ICD A00) in Boma District within 7 days. WHO threshold: 1 case.",
        patient_id="",
        created_at="2026-02-28T10:30:00Z",
        ai_generated=False,
        rule_validated=True,
        category="idsr_cholera",
    ),
    Alert(
        id="alert-002",
        type="outbreak",
        severity="warning",
        status="active",
        title="Measles case rise — Kinshasa Site",
        description="3 confirmed measles cases in 14 days at site-alpha. IDSR doubles previous baseline for this period.",
        patient_id="",
        created_at="2026-02-27T14:15:00Z",
        ai_generated=False,
        rule_validated=True,
        category="idsr_measles",
    ),
    Alert(
        id="alert-003",
        type="stockout",
        severity="critical",
        status="active",
        title="Amoxicillin stockout imminent — site-alpha",
        description="Current stock: 20 units. Predicted days remaining: 3. Daily consumption rate: 6.5 units/day.",
        patient_id="",
        created_at="2026-02-28T08:00:00Z",
        ai_generated=False,
        rule_validated=True,
        category="stockout_prediction",
    ),
    Alert(
        id="alert-004",
        type="medication_interaction",
        severity="warning",
        status="active",
        title="Metformin + ACE inhibitor interaction — Patient P-1042",
        description="Concurrent prescription of Metformin and Enalapril detected. Risk of hypoglycaemia in renal impairment. Review recommended.",
        patient_id="patient-1042",
        created_at="2026-02-26T16:45:00Z",
        ai_generated=False,
        rule_validated=True,
        category="medication_interaction",
    ),
    Alert(
        id="alert-005",
        type="vital_trend",
        severity="info",
        status="active",
        title="Rising blood pressure trend — Patient P-0817",
        description="Systolic BP trending upward over 3 visits: 128, 135, 142 mmHg. Consider antihypertensive review.",
        patient_id="patient-0817",
        created_at="2026-02-25T09:20:00Z",
        ai_generated=False,
        rule_validated=True,
        category="vital_sign_trend",
    ),
]

SEED_INVENTORY_SITE_ALPHA = [
    InventoryItem(item_code="AMX250", display="Amoxicillin 250mg caps", quantity=20, unit="capsules", site_id="site-alpha", last_updated="2026-02-28T08:00:00Z", reorder_level=50),
    InventoryItem(item_code="PCT500", display="Paracetamol 500mg tabs", quantity=500, unit="tablets", site_id="site-alpha", last_updated="2026-02-28T08:00:00Z", reorder_level=100),
    InventoryItem(item_code="ORS001", display="ORS sachets", quantity=200, unit="sachets", site_id="site-alpha", last_updated="2026-02-28T08:00:00Z", reorder_level=50),
    InventoryItem(item_code="MET500", display="Metformin 500mg tabs", quantity=150, unit="tablets", site_id="site-alpha", last_updated="2026-02-28T08:00:00Z", reorder_level=30),
    InventoryItem(item_code="ART020", display="Artemether/Lumefantrine 20/120mg", quantity=80, unit="tablets", site_id="site-alpha", last_updated="2026-02-28T08:00:00Z", reorder_level=40),
]

SEED_INVENTORY_SITE_BRAVO = [
    InventoryItem(item_code="AMX250", display="Amoxicillin 250mg caps", quantity=300, unit="capsules", site_id="site-bravo", last_updated="2026-02-28T08:00:00Z", reorder_level=50),
    InventoryItem(item_code="PCT500", display="Paracetamol 500mg tabs", quantity=120, unit="tablets", site_id="site-bravo", last_updated="2026-02-28T08:00:00Z", reorder_level=100),
    InventoryItem(item_code="ORS001", display="ORS sachets", quantity=50, unit="sachets", site_id="site-bravo", last_updated="2026-02-28T08:00:00Z", reorder_level=50),
    InventoryItem(item_code="CTX480", display="Cotrimoxazole 480mg tabs", quantity=400, unit="tablets", site_id="site-bravo", last_updated="2026-02-28T08:00:00Z", reorder_level=60),
    InventoryItem(item_code="IB400", display="Ibuprofen 400mg tabs", quantity=250, unit="tablets", site_id="site-bravo", last_updated="2026-02-28T08:00:00Z", reorder_level=50),
]

SEED_PREDICTIONS = [
    SupplyPrediction(item_code="AMX250", display="Amoxicillin 250mg caps", current_quantity=20, predicted_days_remaining=3, risk_level="critical", recommended_action="Urgent resupply required. Request emergency transfer from site-bravo (300 units available)."),
    SupplyPrediction(item_code="ORS001", display="ORS sachets", current_quantity=200, predicted_days_remaining=14, risk_level="moderate", recommended_action="Schedule routine resupply within 7 days."),
    SupplyPrediction(item_code="ART020", display="Artemether/Lumefantrine 20/120mg", current_quantity=80, predicted_days_remaining=10, risk_level="high", recommended_action="Malaria season approaching. Increase stock to 200 units."),
]

SEED_REDISTRIBUTIONS = [
    RedistributionSuggestion(item_code="AMX250", from_site="site-bravo", to_site="site-alpha", suggested_quantity=100, rationale="site-alpha has 20 units (3 days supply), site-bravo has 300 units (46 days supply). Transfer balances risk."),
    RedistributionSuggestion(item_code="ORS001", from_site="site-alpha", to_site="site-bravo", suggested_quantity=50, rationale="site-bravo at reorder level (50). site-alpha has 200 units. Cholera alert active — both sites need adequate ORS stock."),
]


def seed_stores() -> tuple[AlertStore, InventoryStore]:
    alert_store = AlertStore()
    for alert in SEED_ALERTS:
        alert_store.add(alert)

    inventory_store = InventoryStore()
    for item in SEED_INVENTORY_SITE_ALPHA + SEED_INVENTORY_SITE_BRAVO:
        inventory_store.add(item)

    for pred in SEED_PREDICTIONS:
        inventory_store.add_prediction("site-alpha", pred)

    for sug in SEED_REDISTRIBUTIONS:
        inventory_store.add_redistribution("site-alpha", sug)

    return alert_store, inventory_store
