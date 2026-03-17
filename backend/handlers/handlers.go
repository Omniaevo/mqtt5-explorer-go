package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"mqtt5-explorer-go/backend/database"
	"mqtt5-explorer-go/backend/models"
	"mqtt5-explorer-go/backend/mqtt"
)

type Handlers struct {
	mqttClient *mqtt.Client
}

func (h *Handlers) SearchConnections(ctx context.Context, query string) ([]models.Connection, error) {
	if query == "" {
		return database.DB.GetAllConnections(ctx)
	}
	return database.DB.SearchConnections(ctx, query)
}

func (h *Handlers) ExportConnections(ctx context.Context, ids []int64) (string, error) {
	type ExportConnection struct {
		models.Connection
		PasswordEncoded string `json:"password,omitempty"`
	}

	connections := make([]ExportConnection, 0)

	for _, id := range ids {
		conn, err := database.DB.GetConnection(ctx, id)
		if err != nil {
			continue
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(conn.Password))
		exportConn := ExportConnection{
			Connection:      *conn,
			PasswordEncoded: encoded,
		}
		exportConn.Password = ""
		connections = append(connections, exportConn)
	}

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (h *Handlers) ImportConnections(ctx context.Context, jsonData string) ([]int64, error) {
	type ImportConnection struct {
		models.Connection
		PasswordEncoded string `json:"password,omitempty"`
	}

	var connections []ImportConnection
	if err := json.Unmarshal([]byte(jsonData), &connections); err != nil {
		return nil, err
	}

	ids := make([]int64, 0)
	for _, conn := range connections {
		conn.ID = 0
		if conn.PasswordEncoded != "" {
			decoded, err := base64.StdEncoding.DecodeString(conn.PasswordEncoded)
			if err == nil {
				conn.Password = string(decoded)
			}
		}
		id, err := database.DB.CreateConnection(ctx, &conn.Connection)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func NewHandlers(mqttClient *mqtt.Client) *Handlers {
	return &Handlers{
		mqttClient: mqttClient,
	}
}

func (h *Handlers) GetSettings(ctx context.Context) (map[string]string, error) {
	return database.DB.GetAllSettings(ctx)
}

func (h *Handlers) SetSetting(ctx context.Context, key, value string) error {
	if err := database.DB.SetSetting(ctx, key, value); err != nil {
		return err
	}

	settings, _ := database.DB.GetAllSettings(ctx)
	h.mqttClient.UpdateSettings(settings)

	return nil
}

func (h *Handlers) GetConnections(ctx context.Context) ([]models.Connection, error) {
	return database.DB.GetAllConnections(ctx)
}

func (h *Handlers) GetConnection(ctx context.Context, id int64) (*models.Connection, error) {
	return database.DB.GetConnection(ctx, id)
}

func (h *Handlers) CreateConnection(ctx context.Context, conn *models.Connection) (int64, error) {
	return database.DB.CreateConnection(ctx, conn)
}

func (h *Handlers) UpdateConnection(ctx context.Context, conn *models.Connection) error {
	return database.DB.UpdateConnection(ctx, conn)
}

func (h *Handlers) DeleteConnection(ctx context.Context, id int64) error {
	connected, _ := h.mqttClient.GetStatus(id)
	if connected {
		h.mqttClient.Disconnect(ctx, id)
	}
	return database.DB.DeleteConnection(ctx, id)
}

func (h *Handlers) Connect(ctx context.Context, id int64) error {
	conn, err := database.DB.GetConnection(ctx, id)
	if err != nil {
		return err
	}

	return h.mqttClient.Connect(ctx, conn)
}

func (h *Handlers) Disconnect(ctx context.Context, id int64) error {
	return h.mqttClient.Disconnect(ctx, id)
}

func (h *Handlers) GetConnectionStatus(ctx context.Context, id int64) (bool, string, error) {
	connected, status := h.mqttClient.GetStatus(id)
	return connected, status, nil
}

func (h *Handlers) GetClientID(ctx context.Context, id int64) (string, error) {
	return h.mqttClient.GetClientID(id), nil
}

func (h *Handlers) GetMessages(ctx context.Context, connectionID int64, topic string, limit int) ([]models.Message, error) {
	if topic != "" {
		return database.DB.GetMessagesByTopic(ctx, connectionID, topic, limit)
	}
	return database.DB.GetMessagesByConnection(ctx, connectionID, limit)
}

func (h *Handlers) SearchMessages(ctx context.Context, connectionID int64, pattern string, useRegex bool) ([]models.Message, error) {
	return database.DB.SearchMessages(ctx, connectionID, pattern, useRegex)
}

func (h *Handlers) ClearMessages(ctx context.Context, connectionID int64) error {
	return database.DB.ClearMessages(ctx, connectionID)
}

func (h *Handlers) SendMessage(ctx context.Context, req *models.SendMessageRequest) error {
	return h.mqttClient.SendMessage(ctx, req)
}

func (h *Handlers) Subscribe(ctx context.Context, connectionID int64, topic string, qos int) error {
	return h.mqttClient.Subscribe(connectionID, topic, byte(qos))
}

func (h *Handlers) Unsubscribe(ctx context.Context, connectionID int64, topic string) error {
	return h.mqttClient.Unsubscribe(connectionID, topic)
}

func (h *Handlers) GetTopicTree(ctx context.Context, connectionID int64) (*models.TopicNode, error) {
	messages, err := database.DB.GetMessagesByConnection(ctx, connectionID, 1000)
	if err != nil {
		return nil, err
	}

	return mqtt.BuildTopicTree(messages), nil
}

func (h *Handlers) GetNumericMessages(ctx context.Context, connectionID int64, topic string, limit int) ([]models.Message, error) {
	return database.DB.GetNumericMessages(ctx, connectionID, topic, limit)
}

func (h *Handlers) GetHostInfo(ctx context.Context) (string, string, error) {
	return mqtt.GetHostInfo()
}

func (h *Handlers) CheckPort(ctx context.Context, host string, port int) bool {
	return mqtt.CheckPort(host, port)
}
