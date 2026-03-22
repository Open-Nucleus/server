/** An item in the supply chain inventory. */
export interface InventoryItem {
  code: string;
  display: string;
  current: number;
  reorder_level: number;
  unit_cost?: number;
}

/** Predicted stockout or reorder event. */
export interface SupplyPrediction {
  item_code: string;
  predicted_date: string;
  quantity: number;
}
