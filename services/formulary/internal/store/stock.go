package store

import (
	"database/sql"
	"fmt"
	"time"
)

type StockLevel struct {
	SiteID               string
	MedicationCode       string
	Quantity             int
	Unit                 string
	LastUpdated          string
	EarliestExpiry       string
	DailyConsumptionRate float64
}

type StockStore struct {
	db *sql.DB
}

func NewStockStore(db *sql.DB) *StockStore {
	return &StockStore{db: db}
}

func (s *StockStore) Get(siteID, medicationCode string) (*StockLevel, error) {
	row := s.db.QueryRow(
		`SELECT site_id, medication_code, quantity, unit, last_updated, earliest_expiry, daily_consumption_rate
		 FROM stock_levels WHERE site_id = ? AND medication_code = ?`,
		siteID, medicationCode,
	)
	var sl StockLevel
	err := row.Scan(&sl.SiteID, &sl.MedicationCode, &sl.Quantity, &sl.Unit,
		&sl.LastUpdated, &sl.EarliestExpiry, &sl.DailyConsumptionRate)
	if err == sql.ErrNoRows {
		return &StockLevel{
			SiteID:         siteID,
			MedicationCode: medicationCode,
			Unit:           "units",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get stock level: %w", err)
	}
	return &sl, nil
}

func (s *StockStore) Upsert(sl *StockLevel) error {
	sl.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO stock_levels (site_id, medication_code, quantity, unit, last_updated, earliest_expiry, daily_consumption_rate)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(site_id, medication_code)
		 DO UPDATE SET quantity=excluded.quantity, unit=excluded.unit, last_updated=excluded.last_updated,
		               earliest_expiry=excluded.earliest_expiry, daily_consumption_rate=excluded.daily_consumption_rate`,
		sl.SiteID, sl.MedicationCode, sl.Quantity, sl.Unit,
		sl.LastUpdated, sl.EarliestExpiry, sl.DailyConsumptionRate,
	)
	if err != nil {
		return fmt.Errorf("upsert stock level: %w", err)
	}
	return nil
}

func (s *StockStore) ListBySite(siteID string) ([]*StockLevel, error) {
	rows, err := s.db.Query(
		`SELECT site_id, medication_code, quantity, unit, last_updated, earliest_expiry, daily_consumption_rate
		 FROM stock_levels WHERE site_id = ?`,
		siteID,
	)
	if err != nil {
		return nil, fmt.Errorf("list stock by site: %w", err)
	}
	defer rows.Close()
	var result []*StockLevel
	for rows.Next() {
		var sl StockLevel
		if err := rows.Scan(&sl.SiteID, &sl.MedicationCode, &sl.Quantity, &sl.Unit,
			&sl.LastUpdated, &sl.EarliestExpiry, &sl.DailyConsumptionRate); err != nil {
			return nil, err
		}
		result = append(result, &sl)
	}
	return result, rows.Err()
}

func (s *StockStore) ListByMedication(medicationCode string) ([]*StockLevel, error) {
	rows, err := s.db.Query(
		`SELECT site_id, medication_code, quantity, unit, last_updated, earliest_expiry, daily_consumption_rate
		 FROM stock_levels WHERE medication_code = ?`,
		medicationCode,
	)
	if err != nil {
		return nil, fmt.Errorf("list stock by medication: %w", err)
	}
	defer rows.Close()
	var result []*StockLevel
	for rows.Next() {
		var sl StockLevel
		if err := rows.Scan(&sl.SiteID, &sl.MedicationCode, &sl.Quantity, &sl.Unit,
			&sl.LastUpdated, &sl.EarliestExpiry, &sl.DailyConsumptionRate); err != nil {
			return nil, err
		}
		result = append(result, &sl)
	}
	return result, rows.Err()
}

func (s *StockStore) RecordDelivery(id, siteID, receivedBy, deliveryDate string, itemsRecorded int) error {
	_, err := s.db.Exec(
		`INSERT INTO deliveries (id, site_id, received_by, delivery_date, items_recorded, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, siteID, receivedBy, deliveryDate, itemsRecorded, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("record delivery: %w", err)
	}
	return nil
}
