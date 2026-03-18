package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"mqtt5-explorer-go/backend/models"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

var DB *Database

func Init(userDataDir string) error {
	dbPath := filepath.Join(userDataDir, "mqtt5-explorer.db")

	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.Exec("PRAGMA journal_mode=WAL")

	DB = &Database{db: db}

	if err := DB.createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := DB.initDefaultSettings(); err != nil {
		return fmt.Errorf("failed to initialize default settings: %w", err)
	}

	return nil
}

func (d *Database) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS connections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		mqtt_version INTEGER NOT NULL,
		protocol TEXT NOT NULL,
		host TEXT NOT NULL,
		port INTEGER NOT NULL,
		username TEXT,
		password TEXT,
		validate_cert INTEGER DEFAULT 1,
		ca_file TEXT,
		client_cert TEXT,
		client_key TEXT,
		default_subscriptions TEXT,
		favourite INTEGER DEFAULT 0,
		last_connected TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		connection_id INTEGER NOT NULL,
		topic TEXT NOT NULL,
		payload BLOB NOT NULL,
		qos INTEGER DEFAULT 0,
		retain INTEGER DEFAULT 0,
		timestamp TEXT NOT NULL,
		content_type TEXT,
		user_properties TEXT,
		response_topic TEXT,
		correlation_data BLOB,
		message_expiry INTEGER,
		topic_alias INTEGER,
		client_id TEXT,
		FOREIGN KEY (connection_id) REFERENCES connections(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_messages_connection_topic ON messages(connection_id, topic);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
	`

	_, err := d.db.Exec(schema)
	return err
}

func (d *Database) initDefaultSettings() error {
	defaults := models.DefaultSettings()
	ctx := context.Background()

	settings, err := d.GetAllSettings(ctx)
	if err != nil || len(settings) == 0 {
		settings = map[string]string{
			"theme":             defaults.Theme,
			"accentColor":       defaults.AccentColor,
			"closeToTray":       fmt.Sprintf("%t", defaults.CloseToTray),
			"maxCachedMessages": fmt.Sprintf("%d", defaults.MaxCachedMessages),
			"defaultClientId":   defaults.DefaultClientID,
			"keepalive":         fmt.Sprintf("%d", defaults.Keepalive),
			"reconnectPeriod":   fmt.Sprintf("%d", defaults.ReconnectPeriod),
			"maxReconnects":     fmt.Sprintf("%d", defaults.MaxReconnects),
			"connectionTimeout": fmt.Sprintf("%d", defaults.ConnectionTimeout),
		}

		for key, value := range settings {
			if err := d.SetSetting(ctx, key, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Database) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := d.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (d *Database) SetSetting(ctx context.Context, key, value string) error {
	_, err := d.db.ExecContext(ctx, "INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}

func (d *Database) GetAllSettings(ctx context.Context) (map[string]string, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}

	return settings, nil
}

func (d *Database) CreateConnection(ctx context.Context, conn *models.Connection) (int64, error) {
	conn.CreatedAt = time.Now()
	conn.UpdatedAt = time.Now()

	result, err := d.db.ExecContext(ctx, `
		INSERT INTO connections (name, mqtt_version, protocol, host, port, username, password, 
			validate_cert, ca_file, client_cert, client_key, default_subscriptions, favourite, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		conn.Name, conn.MQTTVersion, conn.Protocol, conn.Host, conn.Port, conn.Username, conn.Password,
		conn.ValidateCert, conn.CAFile, conn.ClientCert, conn.ClientKey, conn.DefaultSubscriptions, conn.Favourite,
		conn.CreatedAt.Format(time.RFC3339), conn.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (d *Database) UpdateConnection(ctx context.Context, conn *models.Connection) error {
	conn.UpdatedAt = time.Now()

	_, err := d.db.ExecContext(ctx, `
		UPDATE connections SET name = ?, mqtt_version = ?, protocol = ?, host = ?, port = ?,
			username = ?, password = ?, validate_cert = ?, ca_file = ?, client_cert = ?, client_key = ?,
			default_subscriptions = ?, favourite = ?, updated_at = ?
		WHERE id = ?`,
		conn.Name, conn.MQTTVersion, conn.Protocol, conn.Host, conn.Port, conn.Username, conn.Password,
		conn.ValidateCert, conn.CAFile, conn.ClientCert, conn.ClientKey, conn.DefaultSubscriptions, conn.Favourite,
		conn.UpdatedAt.Format(time.RFC3339), conn.ID)
	return err
}

func (d *Database) DeleteConnection(ctx context.Context, id int64) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM messages WHERE connection_id = ?", id)
	if err != nil {
		return err
	}
	_, err = d.db.ExecContext(ctx, "DELETE FROM connections WHERE id = ?", id)
	return err
}

func (d *Database) UpdateLastConnected(ctx context.Context, id int64) error {
	_, err := d.db.ExecContext(ctx, "UPDATE connections SET last_connected = ? WHERE id = ?", time.Now().Format(time.RFC3339), id)
	return err
}

func (d *Database) GetConnection(ctx context.Context, id int64) (*models.Connection, error) {
	var conn models.Connection
	var createdAt, updatedAt, lastConnected string

	err := d.db.QueryRowContext(ctx, `
		SELECT id, name, mqtt_version, protocol, host, port, username, password, validate_cert,
			ca_file, client_cert, client_key, default_subscriptions, favourite, last_connected, created_at, updated_at
		FROM connections WHERE id = ?`, id).Scan(
		&conn.ID, &conn.Name, &conn.MQTTVersion, &conn.Protocol, &conn.Host, &conn.Port,
		&conn.Username, &conn.Password, &conn.ValidateCert, &conn.CAFile, &conn.ClientCert,
		&conn.ClientKey, &conn.DefaultSubscriptions, &conn.Favourite, &lastConnected, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	if lastConnected != "" {
		conn.LastConnected, _ = time.Parse(time.RFC3339, lastConnected)
	}
	conn.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	conn.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &conn, nil
}

func (d *Database) GetAllConnections(ctx context.Context) ([]models.Connection, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT id, name, mqtt_version, protocol, host, port, username, password, validate_cert,
			ca_file, client_cert, client_key, default_subscriptions, favourite, last_connected, created_at, updated_at
		FROM connections ORDER BY favourite DESC, last_connected DESC NULLS LAST, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []models.Connection
	for rows.Next() {
		var conn models.Connection
		var createdAt, updatedAt, lastConnected string

		if err := rows.Scan(
			&conn.ID, &conn.Name, &conn.MQTTVersion, &conn.Protocol, &conn.Host, &conn.Port,
			&conn.Username, &conn.Password, &conn.ValidateCert, &conn.CAFile, &conn.ClientCert,
			&conn.ClientKey, &conn.DefaultSubscriptions, &conn.Favourite, &lastConnected, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		if lastConnected != "" {
			conn.LastConnected, _ = time.Parse(time.RFC3339, lastConnected)
		}
		conn.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		conn.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		connections = append(connections, conn)
	}

	return connections, nil
}

func (d *Database) SearchConnections(ctx context.Context, query string) ([]models.Connection, error) {
	searchPattern := "%" + query + "%"
	rows, err := d.db.QueryContext(ctx, `
		SELECT id, name, mqtt_version, protocol, host, port, username, password, validate_cert,
			ca_file, client_cert, client_key, default_subscriptions, favourite, last_connected, created_at, updated_at
		FROM connections WHERE name LIKE ? OR host LIKE ? ORDER BY favourite DESC, last_connected DESC NULLS LAST, name`, searchPattern, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []models.Connection
	for rows.Next() {
		var conn models.Connection
		var createdAt, updatedAt, lastConnected string

		if err := rows.Scan(
			&conn.ID, &conn.Name, &conn.MQTTVersion, &conn.Protocol, &conn.Host, &conn.Port,
			&conn.Username, &conn.Password, &conn.ValidateCert, &conn.CAFile, &conn.ClientCert,
			&conn.ClientKey, &conn.DefaultSubscriptions, &conn.Favourite, &lastConnected, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		if lastConnected != "" {
			conn.LastConnected, _ = time.Parse(time.RFC3339, lastConnected)
		}
		conn.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		conn.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		connections = append(connections, conn)
	}

	return connections, nil
}

func (d *Database) SaveMessage(ctx context.Context, msg *models.Message) (int64, error) {
	if len(msg.Payload) == 0 {
		_, err := d.db.ExecContext(ctx, `DELETE FROM messages WHERE connection_id = ? AND topic = ?`,
			msg.ConnectionID, msg.Topic)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}

	userPropsJSON, _ := json.Marshal(msg.UserProperties)

	result, err := d.db.ExecContext(ctx, `
		INSERT INTO messages (connection_id, topic, payload, qos, retain, timestamp, content_type,
			user_properties, response_topic, correlation_data, message_expiry, topic_alias, client_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ConnectionID, msg.Topic, msg.Payload, msg.QoS, msg.Retain,
		msg.Timestamp.Format(time.RFC3339), msg.ContentType, userPropsJSON,
		msg.ResponseTopic, msg.CorrelationData, msg.MessageExpiry, msg.TopicAlias, msg.ClientID)
	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()

	maxCached, _ := d.GetSetting(ctx, "maxCachedMessages")
	maxMessages := 7
	if maxCached != "" {
		fmt.Sscanf(maxCached, "%d", &maxMessages)
	}

	_, err = d.db.ExecContext(ctx, `
		DELETE FROM messages WHERE connection_id = ? AND topic = ? AND id NOT IN (
			SELECT id FROM messages WHERE connection_id = ? AND topic = ? 
			ORDER BY timestamp DESC LIMIT ?
		)`, msg.ConnectionID, msg.Topic, msg.ConnectionID, msg.Topic, maxMessages)

	return id, err
}

func (d *Database) GetMessagesByTopic(ctx context.Context, connectionID int64, topic string, limit int) ([]models.Message, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT id, connection_id, topic, payload, qos, retain, timestamp, content_type,
			user_properties, response_topic, correlation_data, message_expiry, topic_alias, client_id
		FROM messages WHERE connection_id = ? AND topic = ? ORDER BY timestamp DESC LIMIT ?`,
		connectionID, topic, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return d.scanMessages(rows)
}

func (d *Database) GetMessagesByConnection(ctx context.Context, connectionID int64, limit int) ([]models.Message, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT id, connection_id, topic, payload, qos, retain, timestamp, content_type,
			user_properties, response_topic, correlation_data, message_expiry, topic_alias, client_id
		FROM messages WHERE connection_id = ? ORDER BY timestamp DESC LIMIT ?`,
		connectionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return d.scanMessages(rows)
}

func (d *Database) SearchMessages(ctx context.Context, connectionID int64, pattern string, useRegex bool) ([]models.Message, error) {
	var rows *sql.Rows
	var err error

	if useRegex {
		rows, err = d.db.QueryContext(ctx, `
			SELECT id, connection_id, topic, payload, qos, retain, timestamp, content_type,
				user_properties, response_topic, correlation_data, message_expiry, topic_alias, client_id
			FROM messages WHERE connection_id = ? AND topic GLOB ? ORDER BY timestamp DESC LIMIT 100`,
			connectionID, pattern)
	} else {
		searchPattern := "%" + pattern + "%"
		rows, err = d.db.QueryContext(ctx, `
			SELECT id, connection_id, topic, payload, qos, retain, timestamp, content_type,
				user_properties, response_topic, correlation_data, message_expiry, topic_alias, client_id
			FROM messages WHERE connection_id = ? AND topic LIKE ? ORDER BY timestamp DESC LIMIT 100`,
			connectionID, searchPattern)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return d.scanMessages(rows)
}

func (d *Database) scanMessages(rows *sql.Rows) ([]models.Message, error) {
	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var timestamp, userPropsJSON, corrData []byte

		if err := rows.Scan(
			&msg.ID, &msg.ConnectionID, &msg.Topic, &msg.Payload, &msg.QoS, &msg.Retain,
			&timestamp, &msg.ContentType, &userPropsJSON, &msg.ResponseTopic, &corrData,
			&msg.MessageExpiry, &msg.TopicAlias, &msg.ClientID); err != nil {
			return nil, err
		}

		msg.Timestamp, _ = time.Parse(time.RFC3339, string(timestamp))
		json.Unmarshal(userPropsJSON, &msg.UserProperties)
		msg.CorrelationData = corrData

		messages = append(messages, msg)
	}

	return messages, nil
}

func (d *Database) ClearMessages(ctx context.Context, connectionID int64) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM messages WHERE connection_id = ?", connectionID)
	return err
}

func (d *Database) ClearAllMessages(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM messages")
	return err
}

func (d *Database) GetTopicStats(ctx context.Context, connectionID int64) (map[string]int, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT topic, COUNT(*) as count FROM messages 
		WHERE connection_id = ? GROUP BY topic ORDER BY topic`,
		connectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var topic string
		var count int
		if err := rows.Scan(&topic, &count); err != nil {
			return nil, err
		}
		stats[topic] = count
	}

	return stats, nil
}

func (d *Database) GetNumericMessages(ctx context.Context, connectionID int64, topic string, limit int) ([]models.Message, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT id, connection_id, topic, payload, qos, retain, timestamp, content_type,
			user_properties, response_topic, correlation_data, message_expiry, topic_alias, client_id
		FROM messages WHERE connection_id = ? AND topic = ? ORDER BY timestamp ASC LIMIT ?`,
		connectionID, topic, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return d.scanMessages(rows)
}

func (d *Database) Close() error {
	return DB.db.Close()
}
