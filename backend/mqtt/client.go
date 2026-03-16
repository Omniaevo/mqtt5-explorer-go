package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"mqtt5-explorer-go/backend/database"
	"mqtt5-explorer-go/backend/models"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

type Client struct {
	mu           sync.RWMutex
	clients      map[int64]mqtt.Client
	connected    map[int64]bool
	status       map[int64]string
	clientIds    map[int64]string
	settings     map[string]string
	messageChan  chan *models.Message
	disconnectCB func(int64)
	connectCB    func(int64)
}

func NewClient(msgChan chan *models.Message, connectCB, disconnectCB func(int64)) *Client {
	return &Client{
		clients:      make(map[int64]mqtt.Client),
		connected:    make(map[int64]bool),
		status:       make(map[int64]string),
		clientIds:    make(map[int64]string),
		messageChan:  msgChan,
		connectCB:    connectCB,
		disconnectCB: disconnectCB,
	}
}

func (c *Client) UpdateSettings(settings map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.settings = settings
}

func (c *Client) Connect(ctx context.Context, conn *models.Connection) error {
	log.Printf("[MQTT] Starting connect to %s:%d", conn.Host, conn.Port)

	settings := c.getSettings()

	c.mu.Lock()
	if client, exists := c.clients[conn.ID]; exists {
		if c.connected[conn.ID] {
			c.mu.Unlock()
			log.Printf("[MQTT] Already connected")
			return nil
		}
		client.Disconnect(0)
	}

	c.status[conn.ID] = "connecting"
	c.mu.Unlock()

	log.Printf("[MQTT] Creating client options")
	var clientID string
	if dc, ok := settings["defaultClientId"]; ok && dc != "" {
		clientID = dc
	} else {
		clientID = "m5g-" + uuid.New().String()
	}

	keepalive := uint16(120)
	if k, ok := settings["keepalive"]; ok && k != "" {
		var kInt int
		fmt.Sscanf(k, "%d", &kInt)
		keepalive = uint16(kInt)
	}

	connTimeout := uint(20)
	if ct, ok := settings["connectionTimeout"]; ok && ct != "" {
		var ctInt int
		fmt.Sscanf(ct, "%d", &ctInt)
		connTimeout = uint(ctInt)
	}

	brokerURL := c.buildBrokerURL(conn)
	opts := mqtt.NewClientOptions().
		SetClientID(clientID).
		AddBroker(brokerURL).
		SetKeepAlive(time.Duration(keepalive) * time.Second).
		SetConnectTimeout(time.Duration(connTimeout) * time.Second).
		SetOnConnectHandler(func(client mqtt.Client) {
			c.mu.Lock()
			c.connected[conn.ID] = true
			c.status[conn.ID] = "connected"
			c.mu.Unlock()

			if c.connectCB != nil {
				c.connectCB(conn.ID)
			}

			c.subscribeToDefaults(client, conn)
		}).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			c.mu.Lock()
			c.connected[conn.ID] = false
			c.status[conn.ID] = "disconnected"
			c.mu.Unlock()

			if c.disconnectCB != nil {
				c.disconnectCB(conn.ID)
			}
		})

	if conn.Username != "" {
		opts.SetUsername(conn.Username)
	}
	if conn.Password != "" {
		opts.SetPassword(conn.Password)
	}

	if conn.Protocol == "mqtts" || conn.Protocol == "wss" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: !conn.ValidateCert,
		}

		if conn.CAFile != "" {
			caCert, err := os.ReadFile(conn.CAFile)
			if err == nil {
				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				tlsConfig.RootCAs = caCertPool
			}
		}

		if conn.ClientCert != "" && conn.ClientKey != "" {
			cert, err := tls.LoadX509KeyPair(conn.ClientCert, conn.ClientKey)
			if err == nil {
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}

		opts.SetTLSConfig(tlsConfig)
	}

	if conn.MQTTVersion == 5 {
		opts.SetProtocolVersion(5)
	}

	opts.SetAutoReconnect(true)
	maxReconnects := 2
	if mr, ok := settings["maxReconnects"]; ok && mr != "" {
		fmt.Sscanf(mr, "%d", &maxReconnects)
	}
	if maxReconnects > 0 {
		opts.SetMaxReconnectInterval(time.Duration(maxReconnects) * time.Second)
	} else {
		opts.SetMaxReconnectInterval(2 * time.Second)
	}

	client := mqtt.NewClient(opts)

	log.Printf("[MQTT] Calling client.Connect()")
	token := client.Connect()
	log.Printf("[MQTT] Connect() returned, waiting for token")
	waitTimeout := time.Duration(connTimeout) * time.Second
	if !token.WaitTimeout(waitTimeout) {
		log.Printf("[MQTT] Connection timeout")
		c.mu.Lock()
		c.status[conn.ID] = "connection_timeout"
		c.mu.Unlock()
		return fmt.Errorf("connection timeout after %v", waitTimeout)
	}

	if token.Error() != nil {
		log.Printf("[MQTT] Connect error: %v", token.Error())
		c.mu.Lock()
		c.status[conn.ID] = "error"
		c.mu.Unlock()
		return token.Error()
	}

	log.Printf("[MQTT] Connection successful, storing client")
	c.mu.Lock()
	c.clients[conn.ID] = client
	c.connected[conn.ID] = true
	c.status[conn.ID] = "connected"
	c.clientIds[conn.ID] = clientID
	c.mu.Unlock()

	log.Printf("[MQTT] Connect completed successfully")
	return nil
}

func (c *Client) buildBrokerURL(conn *models.Connection) string {
	switch conn.Protocol {
	case "mqtt":
		return fmt.Sprintf("tcp://%s:%d", conn.Host, conn.Port)
	case "mqtts":
		return fmt.Sprintf("ssl://%s:%d", conn.Host, conn.Port)
	case "ws":
		return fmt.Sprintf("ws://%s:%d/mqtt", conn.Host, conn.Port)
	case "wss":
		return fmt.Sprintf("wss://%s:%d/mqtt", conn.Host, conn.Port)
	default:
		return fmt.Sprintf("tcp://%s:%d", conn.Host, conn.Port)
	}
}

func (c *Client) subscribeToDefaults(client mqtt.Client, conn *models.Connection) {
	if conn.DefaultSubscriptions == "" {
		return
	}

	var subs []string
	if err := json.Unmarshal([]byte(conn.DefaultSubscriptions), &subs); err != nil {
		subs = strings.Split(conn.DefaultSubscriptions, ",")
	}

	for _, topic := range subs {
		topic = strings.TrimSpace(topic)
		if topic != "" {
			client.Subscribe(topic, 0, c.createMessageHandler(conn.ID))
		}
	}
}

func (c *Client) createMessageHandler(connectionID int64) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		message := &models.Message{
			ConnectionID: connectionID,
			Topic:        msg.Topic(),
			Payload:      msg.Payload(),
			QoS:          int(msg.Qos()),
			Retain:       msg.Retained(),
			Timestamp:    time.Now(),
		}

		if database.DB != nil {
			database.DB.SaveMessage(context.Background(), message)
		}

		select {
		case c.messageChan <- message:
		default:
		}
	}
}

func (c *Client) Disconnect(ctx context.Context, connectionID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if client, exists := c.clients[connectionID]; exists {
		client.Disconnect(0)
		delete(c.clients, connectionID)
		delete(c.clientIds, connectionID)
		c.connected[connectionID] = false
		c.status[connectionID] = "disconnected"

		database.DB.ClearMessages(context.Background(), connectionID)

		if c.disconnectCB != nil {
			c.disconnectCB(connectionID)
		}
	}

	return nil
}

func (c *Client) SendMessage(ctx context.Context, req *models.SendMessageRequest) error {
	c.mu.RLock()
	client, exists := c.clients[req.ConnectionID]
	connected := c.connected[req.ConnectionID]
	c.mu.RUnlock()

	if !exists || !connected {
		return fmt.Errorf("not connected")
	}

	topic := req.Topic
	qos := byte(req.QoS)
	retain := req.Retain
	payload := []byte(req.Payload)

	token := client.Publish(topic, qos, retain, payload)
	if token.WaitTimeout(10 * time.Second) {
		return token.Error()
	}

	return nil
}

func (c *Client) GetStatus(connectionID int64) (bool, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected[connectionID], c.status[connectionID]
}

func (c *Client) GetClientID(connectionID int64) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.clientIds[connectionID]
}

func (c *Client) getSettings() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings
}

func (c *Client) Subscribe(connectionID int64, topic string, qos byte) error {
	c.mu.RLock()
	client, exists := c.clients[connectionID]
	connected := c.connected[connectionID]
	c.mu.RUnlock()

	if !exists || !connected {
		return fmt.Errorf("not connected")
	}

	token := client.Subscribe(topic, qos, c.createMessageHandler(connectionID))
	if token.WaitTimeout(10 * time.Second) {
		return token.Error()
	}

	return fmt.Errorf("subscribe timeout")
}

func (c *Client) Unsubscribe(connectionID int64, topic string) error {
	c.mu.RLock()
	client, exists := c.clients[connectionID]
	connected := c.connected[connectionID]
	c.mu.RUnlock()

	if !exists || !connected {
		return fmt.Errorf("not connected")
	}

	token := client.Unsubscribe(topic)
	if token.WaitTimeout(10 * time.Second) {
		return token.Error()
	}

	return nil
}

func (c *Client) SearchTopics(connectionID int64, pattern string, useRegex bool) ([]string, error) {
	c.mu.RLock()
	_, connected := c.clients[connectionID]
	c.mu.RUnlock()

	if !connected {
		return nil, fmt.Errorf("not connected")
	}

	messages, err := database.DB.SearchMessages(context.Background(), connectionID, pattern, useRegex)
	if err != nil {
		return nil, err
	}

	topics := make([]string, 0)
	seen := make(map[string]bool)
	for _, msg := range messages {
		if !seen[msg.Topic] {
			seen[msg.Topic] = true
			topics = append(topics, msg.Topic)
		}
	}

	return topics, nil
}

func MatchTopic(topic, pattern string) bool {
	regex := strings.ReplaceAll(pattern, "/", "\\/")
	regex = strings.ReplaceAll(regex, "#", ".*")
	regex = strings.ReplaceAll(regex, "+", "[^/]+")
	regex = "^" + regex + "$"

	matched, _ := regexp.MatchString(regex, topic)
	return matched
}

func BuildTopicTree(messages []models.Message) *models.TopicNode {
	root := &models.TopicNode{
		Name:     "root",
		Children: make(map[string]*models.TopicNode),
	}

	topicStats := make(map[string]int)
	topicLastMsg := make(map[string]*models.Message)

	for _, msg := range messages {
		topicStats[msg.Topic]++
		topicLastMsg[msg.Topic] = &msg
	}

	for topic := range topicStats {
		parts := strings.Split(topic, "/")
		current := root
		var pathParts []string

		for i, part := range parts {
			if current.Children == nil {
				current.Children = make(map[string]*models.TopicNode)
			}

			pathParts = append(pathParts, part)
			currentPath := strings.Join(pathParts, "/")

			if _, exists := current.Children[part]; !exists {
				current.Children[part] = &models.TopicNode{
					Name:         part,
					FullTopic:    currentPath,
					Children:     make(map[string]*models.TopicNode),
					MessageCount: 0,
				}
			}

			current = current.Children[part]

			if i == len(parts)-1 {
				current.FullTopic = topic
				current.MessageCount = topicStats[topic]
				current.LastMessage = topicLastMsg[topic]
			}
		}
	}

	propagateCounts(root)

	return root
}

func propagateCounts(node *models.TopicNode) int {
	if node == nil {
		return 0
	}

	if node.Children == nil || len(node.Children) == 0 {
		return node.MessageCount
	}

	total := node.MessageCount
	for _, child := range node.Children {
		childCount := propagateCounts(child)
		if child.FullTopic != node.FullTopic {
			total += childCount
		}
	}
	node.MessageCount = total

	return total
}

func GetHostInfo() (string, string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", "", err
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return hostname, "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			return hostname, ipNet.IP.String(), nil
		}
	}

	return hostname, "", nil
}

func CheckPort(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (c *Client) DisconnectAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id := range c.clients {
		if client, exists := c.clients[id]; exists {
			client.Disconnect(0)
		}
		c.connected[id] = false
		c.status[id] = "disconnected"
	}

	database.DB.ClearAllMessages(context.Background())

	c.clients = make(map[int64]mqtt.Client)
	c.clientIds = make(map[int64]string)
	c.connected = make(map[int64]bool)
	c.status = make(map[int64]string)
}
