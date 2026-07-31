package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

type NotificationRepository interface {
	UpsertDeviceToken(ctx context.Context, token *DeviceToken) error
	DeactivateDeviceToken(ctx context.Context, userID, deviceID string) error
	GetActiveDeviceTokens(ctx context.Context, userID string) ([]*DeviceToken, error)
	MarkDeviceTokenInactive(ctx context.Context, id string) error

	GetPreference(ctx context.Context, userID string, notifType NotificationType) (*NotificationPreference, error)
	UpsertPreference(ctx context.Context, pref *NotificationPreference) error
	ListPreferences(ctx context.Context, userID string) ([]*NotificationPreference, error)

	CreateNotificationLog(ctx context.Context, nLog *NotificationLog) (bool, error)
	UpdateNotificationLogChannel(ctx context.Context, id, channelSent string) error
	ListUserInbox(ctx context.Context, userID string, page, limit int) ([]*NotificationLog, error)
	MarkNotificationRead(ctx context.Context, userID, id string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)

	GetWhatsAppTemplateConfig(ctx context.Context, templateKey, language string) (*WhatsAppTemplateConfig, error)
	ListWhatsAppTemplateConfigs(ctx context.Context) ([]*WhatsAppTemplateConfig, error)
	UpsertWhatsAppTemplateConfig(ctx context.Context, config *WhatsAppTemplateConfig) error

	ListOpsAlertChannelsForType(ctx context.Context, alertType string) ([]*OpsAlertChannel, error)
	ListAllOpsAlertChannels(ctx context.Context) ([]*OpsAlertChannel, error)
	CreateOpsAlertChannel(ctx context.Context, ch *OpsAlertChannel) error
	UpdateOpsAlertChannel(ctx context.Context, ch *OpsAlertChannel) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) UpsertDeviceToken(ctx context.Context, token *DeviceToken) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	query := `
		INSERT INTO device_tokens (id, user_id, user_type, fcm_token, platform, device_id, is_active, last_used_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
		ON CONFLICT (user_id, device_id)
		DO UPDATE SET fcm_token = EXCLUDED.fcm_token, platform = EXCLUDED.platform, is_active = true, last_used_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.UserType, token.FCMToken, token.Platform, token.DeviceID)
	return err
}

func (r *PostgresRepository) DeactivateDeviceToken(ctx context.Context, userID, deviceID string) error {
	query := `UPDATE device_tokens SET is_active = false WHERE user_id = $1 AND device_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, deviceID)
	return err
}

func (r *PostgresRepository) GetActiveDeviceTokens(ctx context.Context, userID string) ([]*DeviceToken, error) {
	query := `SELECT id, user_id, user_type, fcm_token, platform, device_id, is_active, last_used_at, created_at
	          FROM device_tokens WHERE user_id = $1 AND is_active = true`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*DeviceToken
	for rows.Next() {
		t := &DeviceToken{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.UserType, &t.FCMToken, &t.Platform, &t.DeviceID, &t.IsActive, &t.LastUsedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

func (r *PostgresRepository) MarkDeviceTokenInactive(ctx context.Context, id string) error {
	query := `UPDATE device_tokens SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) GetPreference(ctx context.Context, userID string, notifType NotificationType) (*NotificationPreference, error) {
	query := `SELECT user_id, user_type, notification_type, channel, updated_at
	          FROM notification_preferences WHERE user_id = $1 AND notification_type = $2`
	row := r.db.QueryRowContext(ctx, query, userID, notifType)
	pref := &NotificationPreference{}
	if err := row.Scan(&pref.UserID, &pref.UserType, &pref.NotificationType, &pref.Channel, &pref.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return pref, nil
}

func (r *PostgresRepository) UpsertPreference(ctx context.Context, pref *NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (user_id, user_type, notification_type, channel, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id, notification_type)
		DO UPDATE SET channel = EXCLUDED.channel, updated_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, pref.UserID, pref.UserType, pref.NotificationType, pref.Channel)
	return err
}

func (r *PostgresRepository) ListPreferences(ctx context.Context, userID string) ([]*NotificationPreference, error) {
	query := `SELECT user_id, user_type, notification_type, channel, updated_at
	          FROM notification_preferences WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []*NotificationPreference
	for rows.Next() {
		p := &NotificationPreference{}
		if err := rows.Scan(&p.UserID, &p.UserType, &p.NotificationType, &p.Channel, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func (r *PostgresRepository) CreateNotificationLog(ctx context.Context, nLog *NotificationLog) (bool, error) {
	if nLog.ID == "" {
		nLog.ID = uuid.New().String()
	}
	query := `
		INSERT INTO notification_log (id, user_id, user_type, notification_type, channel_sent, title, body, deep_link, source_event_type, source_event_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (source_event_type, source_event_id, user_id, notification_type)
		DO NOTHING;
	`
	res, err := r.db.ExecContext(ctx, query, nLog.ID, nLog.UserID, nLog.UserType, nLog.NotificationType, nLog.ChannelSent, nLog.Title, nLog.Body, nLog.DeepLink, nLog.SourceEventType, nLog.SourceEventID)
	if err != nil {
		return false, err
	}
	rowsAffected, _ := res.RowsAffected()
	return rowsAffected > 0, nil
}

func (r *PostgresRepository) UpdateNotificationLogChannel(ctx context.Context, id, channelSent string) error {
	query := `UPDATE notification_log SET channel_sent = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, channelSent, id)
	return err
}

func (r *PostgresRepository) ListUserInbox(ctx context.Context, userID string, page, limit int) ([]*NotificationLog, error) {
	offset := (page - 1) * limit
	query := `SELECT id, user_id, user_type, notification_type, channel_sent, title, body, deep_link, source_event_type, source_event_id, read_at, created_at
	          FROM notification_log WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*NotificationLog
	for rows.Next() {
		l := &NotificationLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.UserType, &l.NotificationType, &l.ChannelSent, &l.Title, &l.Body, &l.DeepLink, &l.SourceEventType, &l.SourceEventID, &l.ReadAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *PostgresRepository) MarkNotificationRead(ctx context.Context, userID, id string) error {
	query := `UPDATE notification_log SET read_at = NOW() WHERE id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, userID)
	return err
}

func (r *PostgresRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notification_log WHERE user_id = $1 AND read_at IS NULL`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *PostgresRepository) GetWhatsAppTemplateConfig(ctx context.Context, templateKey, language string) (*WhatsAppTemplateConfig, error) {
	if language == "" {
		language = "en"
	}
	query := `SELECT template_key, whatsapp_template_name, is_approved, language, updated_at
	          FROM whatsapp_template_config WHERE template_key = $1 AND language = $2`
	row := r.db.QueryRowContext(ctx, query, templateKey, language)
	cfg := &WhatsAppTemplateConfig{}
	if err := row.Scan(&cfg.TemplateKey, &cfg.WhatsAppTemplateName, &cfg.IsApproved, &cfg.Language, &cfg.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return cfg, nil
}

func (r *PostgresRepository) ListWhatsAppTemplateConfigs(ctx context.Context) ([]*WhatsAppTemplateConfig, error) {
	query := `SELECT template_key, whatsapp_template_name, is_approved, language, updated_at FROM whatsapp_template_config`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*WhatsAppTemplateConfig
	for rows.Next() {
		c := &WhatsAppTemplateConfig{}
		if err := rows.Scan(&c.TemplateKey, &c.WhatsAppTemplateName, &c.IsApproved, &c.Language, &c.UpdatedAt); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

func (r *PostgresRepository) UpsertWhatsAppTemplateConfig(ctx context.Context, config *WhatsAppTemplateConfig) error {
	query := `
		INSERT INTO whatsapp_template_config (template_key, whatsapp_template_name, is_approved, language, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (template_key)
		DO UPDATE SET whatsapp_template_name = EXCLUDED.whatsapp_template_name, is_approved = EXCLUDED.is_approved, language = EXCLUDED.language, updated_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, config.TemplateKey, config.WhatsAppTemplateName, config.IsApproved, config.Language)
	return err
}

func (r *PostgresRepository) ListOpsAlertChannelsForType(ctx context.Context, alertType string) ([]*OpsAlertChannel, error) {
	query := `SELECT id, channel_type, target, alert_types, is_active, created_at FROM ops_alert_channels WHERE is_active = true`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*OpsAlertChannel
	for rows.Next() {
		ch := &OpsAlertChannel{}
		var alertTypesJSON []byte
		if err := rows.Scan(&ch.ID, &ch.ChannelType, &ch.Target, &alertTypesJSON, &ch.IsActive, &ch.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(alertTypesJSON, &ch.AlertTypes)

		for _, at := range ch.AlertTypes {
			if at == alertType || at == "*" {
				channels = append(channels, ch)
				break
			}
		}
	}
	return channels, nil
}

func (r *PostgresRepository) ListAllOpsAlertChannels(ctx context.Context) ([]*OpsAlertChannel, error) {
	query := `SELECT id, channel_type, target, alert_types, is_active, created_at FROM ops_alert_channels`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*OpsAlertChannel
	for rows.Next() {
		ch := &OpsAlertChannel{}
		var alertTypesJSON []byte
		if err := rows.Scan(&ch.ID, &ch.ChannelType, &ch.Target, &alertTypesJSON, &ch.IsActive, &ch.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(alertTypesJSON, &ch.AlertTypes)
		channels = append(channels, ch)
	}
	return channels, nil
}

func (r *PostgresRepository) CreateOpsAlertChannel(ctx context.Context, ch *OpsAlertChannel) error {
	if ch.ID == "" {
		ch.ID = uuid.New().String()
	}
	alertTypesJSON, _ := json.Marshal(ch.AlertTypes)
	query := `INSERT INTO ops_alert_channels (id, channel_type, target, alert_types, is_active, created_at)
	          VALUES ($1, $2, $3, $4, $5, NOW())`
	_, err := r.db.ExecContext(ctx, query, ch.ID, ch.ChannelType, ch.Target, alertTypesJSON, ch.IsActive)
	return err
}

func (r *PostgresRepository) UpdateOpsAlertChannel(ctx context.Context, ch *OpsAlertChannel) error {
	alertTypesJSON, _ := json.Marshal(ch.AlertTypes)
	query := `UPDATE ops_alert_channels SET channel_type = $1, target = $2, alert_types = $3, is_active = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query, ch.ChannelType, ch.Target, alertTypesJSON, ch.IsActive, ch.ID)
	return err
}

// -------------------------------------------------------------------
// In-Memory Repository Fallback (for unit tests / mock mode)
// -------------------------------------------------------------------
type MemoryRepository struct {
	tokens       map[string]*DeviceToken             // token_id -> DeviceToken
	userTokens   map[string]map[string]*DeviceToken  // user_id -> device_id -> DeviceToken
	preferences  map[string]*NotificationPreference // user_id:notif_type -> NotificationPreference
	logs         map[string]*NotificationLog        // key -> NotificationLog
	userLogs     map[string][]*NotificationLog      // user_id -> []NotificationLog
	waTemplates  map[string]*WhatsAppTemplateConfig // template_key:lang -> Config
	opsChannels  map[string]*OpsAlertChannel        // id -> OpsAlertChannel
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tokens:      make(map[string]*DeviceToken),
		userTokens:  make(map[string]map[string]*DeviceToken),
		preferences: make(map[string]*NotificationPreference),
		logs:        make(map[string]*NotificationLog),
		userLogs:    make(map[string][]*NotificationLog),
		waTemplates: make(map[string]*WhatsAppTemplateConfig),
		opsChannels: make(map[string]*OpsAlertChannel),
	}
}

func (r *MemoryRepository) UpsertDeviceToken(ctx context.Context, token *DeviceToken) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	if _, ok := r.userTokens[token.UserID]; !ok {
		r.userTokens[token.UserID] = make(map[string]*DeviceToken)
	}
	now := time.Now()
	token.IsActive = true
	token.LastUsedAt = &now
	token.CreatedAt = now

	r.tokens[token.ID] = token
	r.userTokens[token.UserID][token.DeviceID] = token
	return nil
}

func (r *MemoryRepository) DeactivateDeviceToken(ctx context.Context, userID, deviceID string) error {
	if dtMap, ok := r.userTokens[userID]; ok {
		if tok, ok := dtMap[deviceID]; ok {
			tok.IsActive = false
		}
	}
	return nil
}

func (r *MemoryRepository) GetActiveDeviceTokens(ctx context.Context, userID string) ([]*DeviceToken, error) {
	var active []*DeviceToken
	if dtMap, ok := r.userTokens[userID]; ok {
		for _, tok := range dtMap {
			if tok.IsActive {
				active = append(active, tok)
			}
		}
	}
	return active, nil
}

func (r *MemoryRepository) MarkDeviceTokenInactive(ctx context.Context, id string) error {
	if tok, ok := r.tokens[id]; ok {
		tok.IsActive = false
	}
	return nil
}

func (r *MemoryRepository) GetPreference(ctx context.Context, userID string, notifType NotificationType) (*NotificationPreference, error) {
	key := fmt.Sprintf("%s:%s", userID, notifType)
	if pref, ok := r.preferences[key]; ok {
		return pref, nil
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) UpsertPreference(ctx context.Context, pref *NotificationPreference) error {
	key := fmt.Sprintf("%s:%s", pref.UserID, pref.NotificationType)
	pref.UpdatedAt = time.Now()
	r.preferences[key] = pref
	return nil
}

func (r *MemoryRepository) ListPreferences(ctx context.Context, userID string) ([]*NotificationPreference, error) {
	var list []*NotificationPreference
	for _, p := range r.preferences {
		if p.UserID == userID {
			list = append(list, p)
		}
	}
	return list, nil
}

func (r *MemoryRepository) CreateNotificationLog(ctx context.Context, nLog *NotificationLog) (bool, error) {
	key := fmt.Sprintf("%s:%s:%s:%s", nLog.SourceEventType, nLog.SourceEventID, nLog.UserID, nLog.NotificationType)
	if _, exists := r.logs[key]; exists {
		return false, nil // 0 rows affected, idempotency duplicate
	}
	if nLog.ID == "" {
		nLog.ID = uuid.New().String()
	}
	nLog.CreatedAt = time.Now()
	r.logs[key] = nLog
	r.userLogs[nLog.UserID] = append(r.userLogs[nLog.UserID], nLog)
	return true, nil
}

func (r *MemoryRepository) UpdateNotificationLogChannel(ctx context.Context, id, channelSent string) error {
	for _, l := range r.logs {
		if l.ID == id {
			l.ChannelSent = channelSent
			break
		}
	}
	return nil
}

func (r *MemoryRepository) ListUserInbox(ctx context.Context, userID string, page, limit int) ([]*NotificationLog, error) {
	logs := r.userLogs[userID]
	start := (page - 1) * limit
	if start >= len(logs) {
		return []*NotificationLog{}, nil
	}
	end := start + limit
	if end > len(logs) {
		end = len(logs)
	}
	return logs[start:end], nil
}

func (r *MemoryRepository) MarkNotificationRead(ctx context.Context, userID, id string) error {
	now := time.Now()
	for _, l := range r.userLogs[userID] {
		if l.ID == id {
			l.ReadAt = &now
			break
		}
	}
	return nil
}

func (r *MemoryRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	count := 0
	for _, l := range r.userLogs[userID] {
		if l.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

func (r *MemoryRepository) GetWhatsAppTemplateConfig(ctx context.Context, templateKey, language string) (*WhatsAppTemplateConfig, error) {
	if language == "" {
		language = "en"
	}
	key := fmt.Sprintf("%s:%s", templateKey, language)
	if cfg, ok := r.waTemplates[key]; ok {
		return cfg, nil
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) ListWhatsAppTemplateConfigs(ctx context.Context) ([]*WhatsAppTemplateConfig, error) {
	var list []*WhatsAppTemplateConfig
	for _, c := range r.waTemplates {
		list = append(list, c)
	}
	return list, nil
}

func (r *MemoryRepository) UpsertWhatsAppTemplateConfig(ctx context.Context, config *WhatsAppTemplateConfig) error {
	key := fmt.Sprintf("%s:%s", config.TemplateKey, config.Language)
	config.UpdatedAt = time.Now()
	r.waTemplates[key] = config
	return nil
}

func (r *MemoryRepository) ListOpsAlertChannelsForType(ctx context.Context, alertType string) ([]*OpsAlertChannel, error) {
	var list []*OpsAlertChannel
	for _, ch := range r.opsChannels {
		if !ch.IsActive {
			continue
		}
		for _, at := range ch.AlertTypes {
			if at == alertType || at == "*" {
				list = append(list, ch)
				break
			}
		}
	}
	return list, nil
}

func (r *MemoryRepository) ListAllOpsAlertChannels(ctx context.Context) ([]*OpsAlertChannel, error) {
	var list []*OpsAlertChannel
	for _, ch := range r.opsChannels {
		list = append(list, ch)
	}
	return list, nil
}

func (r *MemoryRepository) CreateOpsAlertChannel(ctx context.Context, ch *OpsAlertChannel) error {
	if ch.ID == "" {
		ch.ID = uuid.New().String()
	}
	ch.CreatedAt = time.Now()
	r.opsChannels[ch.ID] = ch
	return nil
}

func (r *MemoryRepository) UpdateOpsAlertChannel(ctx context.Context, ch *OpsAlertChannel) error {
	r.opsChannels[ch.ID] = ch
	return nil
}

var _ = pq.Array
