from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class Alert:
    id: str
    type: str
    severity: str  # critical, warning, info
    status: str  # active, acknowledged, dismissed
    title: str
    description: str
    patient_id: str = ""
    created_at: str = ""
    acknowledged_at: str = ""
    acknowledged_by: str = ""
    # Extended fields for FHIR output
    ai_generated: bool = False
    ai_model: str = ""
    ai_confidence: float = 0.0
    ai_reasoning: str = ""
    reflection_iterations: int = 0
    rule_validated: bool = False
    evidence: dict = field(default_factory=dict)
    category: str = ""


@dataclass
class AlertSummary:
    total: int = 0
    critical: int = 0
    warning: int = 0
    info: int = 0
    unacknowledged: int = 0


@dataclass
class InventoryItem:
    item_code: str
    display: str
    quantity: int
    unit: str
    site_id: str
    last_updated: str = ""
    reorder_level: int = 0


@dataclass
class DeliveryRecord:
    delivery_id: str
    site_id: str
    items: list[DeliveryLineItem] = field(default_factory=list)
    received_by: str = ""
    delivery_date: str = ""


@dataclass
class DeliveryLineItem:
    item_code: str
    quantity: int
    unit: str
    batch_number: str = ""
    expiry_date: str = ""


@dataclass
class SupplyPrediction:
    item_code: str
    display: str
    current_quantity: int
    predicted_days_remaining: int
    risk_level: str  # critical, high, moderate, low
    recommended_action: str


@dataclass
class RedistributionSuggestion:
    item_code: str
    from_site: str
    to_site: str
    suggested_quantity: int
    rationale: str
