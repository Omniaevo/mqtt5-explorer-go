package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"mqtt5-explorer-go/backend/database"
	"mqtt5-explorer-go/backend/models"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
)

type Client struct {
	mu            sync.RWMutex
	clients       map[int64]*autopaho.ConnectionManager
	connected     map[int64]bool
	status        map[int64]string
	clientIds     map[int64]string
	settings      map[string]string
	messageChan   chan *models.Message
	disconnectCB  func(int64)
	connectCB     func(int64)
	subscriptions map[int64]map[string]byte
	cmToConnID    map[*autopaho.ConnectionManager]int64
}

func NewClient(msgChan chan *models.Message, connectCB, disconnectCB func(int64)) *Client {
	return &Client{
		clients:       make(map[int64]*autopaho.ConnectionManager),
		connected:     make(map[int64]bool),
		status:        make(map[int64]string),
		clientIds:     make(map[int64]string),
		messageChan:   msgChan,
		connectCB:     connectCB,
		disconnectCB:  disconnectCB,
		subscriptions: make(map[int64]map[string]byte),
		cmToConnID:    make(map[*autopaho.ConnectionManager]int64),
	}
}

func (c *Client) UpdateSettings(settings map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.settings = settings
}

func (c *Client) Connect(ctx context.Context, conn *models.Connection) error {
	settings := c.getSettings()

	c.mu.Lock()
	if _, exists := c.clients[conn.ID]; exists {
		if c.connected[conn.ID] {
			c.mu.Unlock()
			return nil
		}
	}
	c.status[conn.ID] = "connecting"
	c.mu.Unlock()

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
	serverURL, err := url.Parse(brokerURL)
	if err != nil {
		return fmt.Errorf("failed to parse broker URL: %w", err)
	}

	// Build TLS config if needed
	var tlsConfig *tls.Config
	if conn.Protocol == "mqtts" || conn.Protocol == "wss" {
		tlsConfig = &tls.Config{
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
	}

	cfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{serverURL},
		TlsCfg:                        tlsConfig,
		ConnectTimeout:                time.Duration(connTimeout) * time.Second,
		KeepAlive:                     keepalive,
		CleanStartOnInitialConnection: true,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connack *paho.Connack) {
			c.mu.Lock()
			c.connected[conn.ID] = true
			c.status[conn.ID] = "connected"
			c.mu.Unlock()

			if c.connectCB != nil {
				c.connectCB(conn.ID)
			}

			c.subscribeToDefaults(conn)
		},
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
		},
	}

	if conn.Username != "" {
		cfg.ConnectPassword = []byte(conn.Password)
		cfg.ConnectUsername = conn.Username
	}

	cm, err := autopaho.NewConnection(ctx, cfg)
	if err != nil {
		c.mu.Lock()
		c.status[conn.ID] = "error"
		c.mu.Unlock()
		return fmt.Errorf("connection error: %w", err)
	}

	c.mu.Lock()
	c.cmToConnID[cm] = conn.ID
	c.mu.Unlock()

	err = cm.AwaitConnection(ctx)
	if err != nil {
		c.mu.Lock()
		c.status[conn.ID] = "error"
		c.mu.Unlock()
		return fmt.Errorf("await connection error: %w", err)
	}

	c.mu.Lock()
	c.clients[conn.ID] = cm
	c.connected[conn.ID] = true
	c.status[conn.ID] = "connected"
	c.clientIds[conn.ID] = clientID
	c.subscriptions[conn.ID] = make(map[string]byte)
	c.mu.Unlock()

	cm.AddOnPublishReceived(func(pr autopaho.PublishReceived) (bool, error) {
		c.handleMessage(conn.ID, pr.Packet)
		return true, nil
	})

	if c.connectCB != nil {
		c.connectCB(conn.ID)
	}

	c.subscribeToDefaults(conn)

	return nil
}

func (c *Client) handleMessage(connectionID int64, publish *paho.Publish) {
	message := &models.Message{
		ConnectionID: connectionID,
		Topic:        publish.Topic,
		Payload:      publish.Payload,
		QoS:          int(publish.QoS),
		Retain:       publish.Retain,
		Timestamp:    time.Now(),
	}

	// Extract MQTT 5 properties
	if publish.Properties != nil {
		if publish.Properties.ResponseTopic != "" {
			message.ResponseTopic = publish.Properties.ResponseTopic
		}
		if publish.Properties.CorrelationData != nil {
			message.CorrelationData = publish.Properties.CorrelationData
		}
		if publish.Properties.MessageExpiry != nil {
			message.MessageExpiry = publish.Properties.MessageExpiry
		}
		if publish.Properties.TopicAlias != nil {
			message.TopicAlias = publish.Properties.TopicAlias
		}
		if publish.Properties.ContentType != "" {
			message.ContentType = publish.Properties.ContentType
		}
		if publish.Properties.User != nil {
			userProps := make(map[string]string)
			for _, prop := range publish.Properties.User {
				userProps[prop.Key] = prop.Value
			}
			message.UserProperties = userProps
		}
	}

	if database.DB != nil {
		database.DB.SaveMessage(context.Background(), message)
	}

	select {
	case c.messageChan <- message:
	default:
	}
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

func (c *Client) subscribeToDefaults(conn *models.Connection) {
	if conn == nil || conn.DefaultSubscriptions == "" {
		return
	}

	var subs []string
	if err := json.Unmarshal([]byte(conn.DefaultSubscriptions), &subs); err != nil {
		subs = strings.Split(conn.DefaultSubscriptions, ",")
	}

	for _, topic := range subs {
		topic = strings.TrimSpace(topic)
		if topic != "" {
			c.Subscribe(conn.ID, topic, 0)
		}
	}
}

func (c *Client) Disconnect(ctx context.Context, connectionID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cm, exists := c.clients[connectionID]; exists {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cm.Disconnect(ctx)
		delete(c.clients, connectionID)
		delete(c.clientIds, connectionID)
		delete(c.subscriptions, connectionID)
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
	cm, exists := c.clients[req.ConnectionID]
	connected := c.connected[req.ConnectionID]
	c.mu.RUnlock()

	if !exists || !connected {
		return fmt.Errorf("not connected")
	}

	publish := &paho.Publish{
		Topic:   req.Topic,
		QoS:     byte(req.QoS),
		Retain:  req.Retain,
		Payload: []byte(req.Payload),
	}

	// Add MQTT 5 properties if set
	if req.ContentType != "" || req.ResponseTopic != "" || req.CorrelationData != "" ||
		req.MessageExpiry != nil || len(req.UserProperties) > 0 {

		props := &paho.PublishProperties{}

		if req.ContentType != "" {
			props.ContentType = req.ContentType
		}
		if req.ResponseTopic != "" {
			props.ResponseTopic = req.ResponseTopic
		}
		if req.CorrelationData != "" {
			props.CorrelationData = []byte(req.CorrelationData)
		}
		if req.MessageExpiry != nil {
			props.MessageExpiry = req.MessageExpiry
		}
		if len(req.UserProperties) > 0 {
			for k, v := range req.UserProperties {
				props.User = append(props.User, paho.UserProperty{Key: k, Value: v})
			}
		}

		publish.Properties = props
	}

	_, err := cm.Publish(ctx, publish)
	if err != nil {
		return fmt.Errorf("publish error: %w", err)
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
	cm, exists := c.clients[connectionID]
	connected := c.connected[connectionID]
	c.mu.RUnlock()

	if !exists || !connected {
		return fmt.Errorf("not connected")
	}

	sub := &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: topic,
				QoS:   qos,
			},
		},
	}

	_, err := cm.Subscribe(context.Background(), sub)
	if err != nil {
		return fmt.Errorf("subscribe error: %w", err)
	}

	c.mu.Lock()
	if c.subscriptions[connectionID] != nil {
		c.subscriptions[connectionID][topic] = qos
	}
	c.mu.Unlock()

	return nil
}

func (c *Client) Unsubscribe(connectionID int64, topic string) error {
	c.mu.RLock()
	cm, exists := c.clients[connectionID]
	connected := c.connected[connectionID]
	c.mu.RUnlock()

	if !exists || !connected {
		return fmt.Errorf("not connected")
	}

	unsub := &paho.Unsubscribe{
		Topics: []string{topic},
	}

	_, err := cm.Unsubscribe(context.Background(), unsub)
	if err != nil {
		return fmt.Errorf("unsubscribe error: %w", err)
	}

	// Remove subscription
	c.mu.Lock()
	if c.subscriptions[connectionID] != nil {
		delete(c.subscriptions[connectionID], topic)
	}
	c.mu.Unlock()

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for id := range c.clients {
		if cm, exists := c.clients[id]; exists {
			cm.Disconnect(ctx)
		}
		c.connected[id] = false
		c.status[id] = "disconnected"
	}

	database.DB.ClearAllMessages(context.Background())

	c.clients = make(map[int64]*autopaho.ConnectionManager)
	c.clientIds = make(map[int64]string)
	c.connected = make(map[int64]bool)
	c.status = make(map[int64]string)
	c.subscriptions = make(map[int64]map[string]byte)
}
