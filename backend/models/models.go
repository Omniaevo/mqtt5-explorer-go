package models

import (
	"encoding/json"
	"time"
)

type MQTTVersion int

const (
	MQTTVersion_3_1   MQTTVersion = 3
	MQTTVersion_3_1_1 MQTTVersion = 4
	MQTTVersion_5_0   MQTTVersion = 5
)

type Protocol string

const (
	Protocol_MQTT  Protocol = "mqtt"
	Protocol_MQTTS Protocol = "mqtts"
	Protocol_WS    Protocol = "ws"
	Protocol_WSS   Protocol = "wss"
)

type Connection struct {
	ID                   int64       `json:"id"`
	Name                 string      `json:"name"`
	MQTTVersion          MQTTVersion `json:"mqttVersion"`
	Protocol             Protocol    `json:"protocol"`
	Host                 string      `json:"host"`
	Port                 int         `json:"port"`
	Username             string      `json:"username,omitempty"`
	Password             string      `json:"password,omitempty"`
	ValidateCert         bool        `json:"validateCert"`
	CAFile               string      `json:"caFile,omitempty"`
	ClientCert           string      `json:"clientCert,omitempty"`
	ClientKey            string      `json:"clientKey,omitempty"`
	DefaultSubscriptions string      `json:"defaultSubscriptions,omitempty"`
	CreatedAt            time.Time   `json:"createdAt"`
	UpdatedAt            time.Time   `json:"updatedAt"`
}

type Payload []byte

func (p Payload) MarshalJSON() ([]byte, error) {
	ints := make([]int, len(p))
	for i, b := range p {
		ints[i] = int(b)
	}
	return json.Marshal(ints)
}

func (p *Payload) UnmarshalJSON(data []byte) error {
	var ints []int
	if err := json.Unmarshal(data, &ints); err != nil {
		return err
	}
	*p = make(Payload, len(ints))
	for i, n := range ints {
		(*p)[i] = byte(n)
	}
	return nil
}

type Message struct {
	ID              int64             `json:"id"`
	ConnectionID    int64             `json:"connectionId"`
	Topic           string            `json:"topic"`
	Payload         Payload           `json:"payload"`
	QoS             int               `json:"qos"`
	Retain          bool              `json:"retain"`
	Timestamp       time.Time         `json:"timestamp"`
	ContentType     string            `json:"contentType,omitempty"`
	UserProperties  map[string]string `json:"userProperties,omitempty"`
	ResponseTopic   string            `json:"responseTopic,omitempty"`
	CorrelationData []byte            `json:"correlationData,omitempty"`
	MessageExpiry   *uint32           `json:"messageExpiry,omitempty"`
	TopicAlias      *uint16           `json:"topicAlias,omitempty"`
	ClientID        string            `json:"clientId,omitempty"`
}

type Settings struct {
	Theme             string `json:"theme"`
	AccentColor       string `json:"accentColor"`
	CloseToTray       bool   `json:"closeToTray"`
	MaxCachedMessages int    `json:"maxCachedMessages"`
	DefaultClientID   string `json:"defaultClientId"`
	Keepalive         int    `json:"keepalive"`
	ReconnectPeriod   int    `json:"reconnectPeriod"`
	MaxReconnects     int    `json:"maxReconnects"`
	ConnectionTimeout int    `json:"connectionTimeout"`
}

type ConnectionStatus struct {
	ConnectionID int64  `json:"connectionId"`
	Status       string `json:"status"`
	Error        string `json:"error,omitempty"`
}

type TopicNode struct {
	Name         string                `json:"name"`
	FullTopic    string                `json:"fullTopic"`
	Children     map[string]*TopicNode `json:"children,omitempty"`
	MessageCount int                   `json:"messageCount"`
	LastMessage  *Message              `json:"lastMessage,omitempty"`
}

type SendMessageRequest struct {
	ConnectionID    int64             `json:"connectionId"`
	Topic           string            `json:"topic"`
	Payload         string            `json:"payload"`
	QoS             int               `json:"qos"`
	Retain          bool              `json:"retain"`
	ContentType     string            `json:"contentType,omitempty"`
	UserProperties  map[string]string `json:"userProperties,omitempty"`
	ResponseTopic   string            `json:"responseTopic,omitempty"`
	CorrelationData string            `json:"correlationData,omitempty"`
	MessageExpiry   *uint32           `json:"messageExpiry,omitempty"`
}

func DefaultSettings() Settings {
	return Settings{
		Theme:             "light",
		AccentColor:       "#007AFF",
		CloseToTray:       true,
		MaxCachedMessages: 42,
		DefaultClientID:   "",
		Keepalive:         120,
		ReconnectPeriod:   2,
		MaxReconnects:     2,
		ConnectionTimeout: 20,
	}
}
