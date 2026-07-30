package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zippyra/backend/shared/db"
)

var (
	ErrDeviceNotFound          = errors.New("device not found")
	ErrGateIDAlreadyExists     = errors.New("gate_id already exists for this store")
	ErrDeviceAlreadyDecommissioned = errors.New("device already decommissioned")
	ErrAlertNotFound           = errors.New("alert not found")
)

type Repository interface {
	CreateDevice(ctx context.Context, device *Device) error
	GetDeviceByID(ctx context.Context, id string) (*Device, error)
	GetDeviceByGateID(ctx context.Context, storeID, gateID string) (*Device, error)
	ListDevices(ctx context.Context, storeID, statusFilter, deviceTypeFilter string, page, pageSize int) ([]*Device, int64, error)
	UpdateDeviceStatus(ctx context.Context, id, status string) error
	UpdateDeviceHeartbeat(ctx context.Context, id string, ts time.Time, firmware *string) error
	DecommissionDevice(ctx context.Context, id string) error

	CreateAlert(ctx context.Context, alert *DeviceAlert) error
	GetUnresolvedAlert(ctx context.Context, deviceID, alertType string) (*DeviceAlert, error)
	ResolveDeviceAlerts(ctx context.Context, deviceID, alertType string) error
	ResolveAlertByID(ctx context.Context, id string) error
	ListAlerts(ctx context.Context, storeID string, resolved *bool) ([]*DeviceAlert, error)

	InsertHeartbeatHypertable(ctx context.Context, hb *DeviceHeartbeat) error
	GetHeartbeatsHypertable(ctx context.Context, deviceID string, from, to time.Time) ([]*DeviceHeartbeat, error)
}

type PostgresRepository struct {
	database *db.DB
}

func NewPostgresRepository(database *db.DB) *PostgresRepository {
	return &PostgresRepository{database: database}
}

func (p *PostgresRepository) CreateDevice(ctx context.Context, device *Device) error {
	if device.ID == "" {
		device.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	device.CreatedAt = now
	device.UpdatedAt = now
	device.ProvisionedAt = &now

	query := `
		INSERT INTO devices (
			id, store_id, chain_id, device_type, gate_id, label, status,
			iot_thing_name, cert_arn, cert_id, cert_expires_at, device_jwt_kid,
			provisioned_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := p.database.ExecContext(ctx, query,
		device.ID, device.StoreID, device.ChainID, device.DeviceType, device.GateID,
		device.Label, device.Status, device.IoTThingName, device.CertARN, device.CertID,
		device.CertExpiresAt, device.DeviceJWTKid, device.ProvisionedAt, device.CreatedAt, device.UpdatedAt,
	)
	return err
}

func (p *PostgresRepository) GetDeviceByID(ctx context.Context, id string) (*Device, error) {
	query := `
		SELECT id, store_id, chain_id, device_type, gate_id, label, status,
		       iot_thing_name, cert_arn, cert_id, cert_expires_at, device_jwt_kid,
		       last_heartbeat_at, firmware_version, provisioned_at, decommissioned_at,
		       created_at, updated_at
		FROM devices WHERE id = $1
	`
	d := &Device{}
	err := p.database.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &d.StoreID, &d.ChainID, &d.DeviceType, &d.GateID, &d.Label, &d.Status,
		&d.IoTThingName, &d.CertARN, &d.CertID, &d.CertExpiresAt, &d.DeviceJWTKid,
		&d.LastHeartbeatAt, &d.FirmwareVersion, &d.ProvisionedAt, &d.DecommissionedAt,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, ErrDeviceNotFound
	}
	return d, nil
}

func (p *PostgresRepository) GetDeviceByGateID(ctx context.Context, storeID, gateID string) (*Device, error) {
	query := `SELECT id, store_id, chain_id, device_type, gate_id, status FROM devices WHERE store_id = $1 AND gate_id = $2 AND status != 'DECOMMISSIONED' LIMIT 1`
	d := &Device{}
	err := p.database.QueryRowContext(ctx, query, storeID, gateID).Scan(&d.ID, &d.StoreID, &d.ChainID, &d.DeviceType, &d.GateID, &d.Status)
	if err != nil {
		return nil, ErrDeviceNotFound
	}
	return d, nil
}

func (p *PostgresRepository) ListDevices(ctx context.Context, storeID, statusFilter, deviceTypeFilter string, page, pageSize int) ([]*Device, int64, error) {
	return []*Device{}, 0, nil
}

func (p *PostgresRepository) UpdateDeviceStatus(ctx context.Context, id, status string) error {
	query := `UPDATE devices SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := p.database.ExecContext(ctx, query, status, time.Now().UTC(), id)
	return err
}

func (p *PostgresRepository) UpdateDeviceHeartbeat(ctx context.Context, id string, ts time.Time, firmware *string) error {
	query := `UPDATE devices SET last_heartbeat_at = $1, firmware_version = COALESCE($2, firmware_version), updated_at = $3 WHERE id = $4`
	_, err := p.database.ExecContext(ctx, query, ts, firmware, time.Now().UTC(), id)
	return err
}

func (p *PostgresRepository) DecommissionDevice(ctx context.Context, id string) error {
	now := time.Now().UTC()
	query := `UPDATE devices SET status = 'DECOMMISSIONED', decommissioned_at = $1, updated_at = $2 WHERE id = $3`
	_, err := p.database.ExecContext(ctx, query, now, now, id)
	return err
}

func (p *PostgresRepository) CreateAlert(ctx context.Context, alert *DeviceAlert) error {
	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	alert.CreatedAt = time.Now().UTC()
	query := `INSERT INTO device_alerts (id, device_id, store_id, alert_type, detail, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.database.ExecContext(ctx, query, alert.ID, alert.DeviceID, alert.StoreID, alert.AlertType, alert.Detail, alert.CreatedAt)
	return err
}

func (p *PostgresRepository) GetUnresolvedAlert(ctx context.Context, deviceID, alertType string) (*DeviceAlert, error) {
	query := `SELECT id, device_id, store_id, alert_type, created_at FROM device_alerts WHERE device_id = $1 AND alert_type = $2 AND resolved_at IS NULL LIMIT 1`
	a := &DeviceAlert{}
	err := p.database.QueryRowContext(ctx, query, deviceID, alertType).Scan(&a.ID, &a.DeviceID, &a.StoreID, &a.AlertType, &a.CreatedAt)
	if err != nil {
		return nil, ErrAlertNotFound
	}
	return a, nil
}

func (p *PostgresRepository) ResolveDeviceAlerts(ctx context.Context, deviceID, alertType string) error {
	query := `UPDATE device_alerts SET resolved_at = $1 WHERE device_id = $2 AND alert_type = $3 AND resolved_at IS NULL`
	_, err := p.database.ExecContext(ctx, query, time.Now().UTC(), deviceID, alertType)
	return err
}

func (p *PostgresRepository) ResolveAlertByID(ctx context.Context, id string) error {
	query := `UPDATE device_alerts SET resolved_at = $1 WHERE id = $2`
	_, err := p.database.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (p *PostgresRepository) ListAlerts(ctx context.Context, storeID string, resolved *bool) ([]*DeviceAlert, error) {
	return []*DeviceAlert{}, nil
}

func (p *PostgresRepository) InsertHeartbeatHypertable(ctx context.Context, hb *DeviceHeartbeat) error {
	return nil
}

func (p *PostgresRepository) GetHeartbeatsHypertable(ctx context.Context, deviceID string, from, to time.Time) ([]*DeviceHeartbeat, error) {
	return []*DeviceHeartbeat{}, nil
}

// MemoryRepository for testing
type MemoryRepository struct {
	mu         sync.RWMutex
	devices    map[string]*Device
	alerts     map[string]*DeviceAlert
	heartbeats []*DeviceHeartbeat
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		devices:    make(map[string]*Device),
		alerts:     make(map[string]*DeviceAlert),
		heartbeats: []*DeviceHeartbeat{},
	}
}

func (m *MemoryRepository) CreateDevice(ctx context.Context, device *Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if device.GateID != nil && *device.GateID != "" {
		for _, d := range m.devices {
			if d.StoreID == device.StoreID && d.GateID != nil && *d.GateID == *device.GateID && d.Status != StatusDecommissioned {
				return ErrGateIDAlreadyExists
			}
		}
	}

	if device.ID == "" {
		device.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	device.CreatedAt = now
	device.UpdatedAt = now
	device.ProvisionedAt = &now

	cp := *device
	m.devices[device.ID] = &cp
	return nil
}

func (m *MemoryRepository) GetDeviceByID(ctx context.Context, id string) (*Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	d, ok := m.devices[id]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	cp := *d
	if cp.LastHeartbeatAt != nil && time.Since(*cp.LastHeartbeatAt) > 180*time.Second {
		cp.IsStale = true
	}
	return &cp, nil
}

func (m *MemoryRepository) GetDeviceByGateID(ctx context.Context, storeID, gateID string) (*Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, d := range m.devices {
		if d.StoreID == storeID && d.GateID != nil && *d.GateID == gateID && d.Status != StatusDecommissioned {
			cp := *d
			return &cp, nil
		}
	}
	return nil, ErrDeviceNotFound
}

func (m *MemoryRepository) ListDevices(ctx context.Context, storeID, statusFilter, deviceTypeFilter string, page, pageSize int) ([]*Device, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Device
	now := time.Now()

	for _, d := range m.devices {
		if storeID != "" && d.StoreID != storeID {
			continue
		}
		if statusFilter != "" && d.Status != statusFilter {
			continue
		}
		if deviceTypeFilter != "" && d.DeviceType != deviceTypeFilter {
			continue
		}

		cp := *d
		if cp.LastHeartbeatAt != nil && now.Sub(*cp.LastHeartbeatAt) > 180*time.Second {
			cp.IsStale = true
		}
		result = append(result, &cp)
	}

	return result, int64(len(result)), nil
}

func (m *MemoryRepository) UpdateDeviceStatus(ctx context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, ok := m.devices[id]
	if !ok {
		return ErrDeviceNotFound
	}
	d.Status = status
	d.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MemoryRepository) UpdateDeviceHeartbeat(ctx context.Context, id string, ts time.Time, firmware *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, ok := m.devices[id]
	if !ok {
		return ErrDeviceNotFound
	}
	d.LastHeartbeatAt = &ts
	if firmware != nil && *firmware != "" {
		d.FirmwareVersion = firmware
	}
	d.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *MemoryRepository) DecommissionDevice(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	d, ok := m.devices[id]
	if !ok {
		return ErrDeviceNotFound
	}
	if d.Status == StatusDecommissioned {
		return ErrDeviceAlreadyDecommissioned
	}
	now := time.Now().UTC()
	d.Status = StatusDecommissioned
	d.DecommissionedAt = &now
	d.UpdatedAt = now
	return nil
}

func (m *MemoryRepository) CreateAlert(ctx context.Context, alert *DeviceAlert) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if alert.ID == "" {
		alert.ID = uuid.New().String()
	}
	alert.CreatedAt = time.Now().UTC()
	cp := *alert
	m.alerts[alert.ID] = &cp
	return nil
}

func (m *MemoryRepository) GetUnresolvedAlert(ctx context.Context, deviceID, alertType string) (*DeviceAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, a := range m.alerts {
		if a.DeviceID == deviceID && a.AlertType == alertType && a.ResolvedAt == nil {
			cp := *a
			return &cp, nil
		}
	}
	return nil, ErrAlertNotFound
}

func (m *MemoryRepository) ResolveDeviceAlerts(ctx context.Context, deviceID, alertType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, a := range m.alerts {
		if a.DeviceID == deviceID && a.AlertType == alertType && a.ResolvedAt == nil {
			a.ResolvedAt = &now
		}
	}
	return nil
}

func (m *MemoryRepository) ResolveAlertByID(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	a, ok := m.alerts[id]
	if !ok {
		return ErrAlertNotFound
	}
	now := time.Now().UTC()
	a.ResolvedAt = &now
	return nil
}

func (m *MemoryRepository) ListAlerts(ctx context.Context, storeID string, resolved *bool) ([]*DeviceAlert, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*DeviceAlert
	for _, a := range m.alerts {
		if storeID != "" && a.StoreID != storeID {
			continue
		}
		if resolved != nil {
			if *resolved && a.ResolvedAt == nil {
				continue
			}
			if !*resolved && a.ResolvedAt != nil {
				continue
			}
		}
		cp := *a
		result = append(result, &cp)
	}
	return result, nil
}

func (m *MemoryRepository) InsertHeartbeatHypertable(ctx context.Context, hb *DeviceHeartbeat) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := *hb
	m.heartbeats = append(m.heartbeats, &cp)
	return nil
}

func (m *MemoryRepository) GetHeartbeatsHypertable(ctx context.Context, deviceID string, from, to time.Time) ([]*DeviceHeartbeat, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*DeviceHeartbeat
	for _, hb := range m.heartbeats {
		if hb.DeviceID == deviceID {
			if (from.IsZero() || !hb.TS.Before(from)) && (to.IsZero() || !hb.TS.After(to)) {
				cp := *hb
				result = append(result, &cp)
			}
		}
	}
	return result, nil
}
