package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"mqtt5-explorer-go/backend/database"
	"mqtt5-explorer-go/backend/handlers"
	"mqtt5-explorer-go/backend/models"
	"mqtt5-explorer-go/backend/mqtt"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/appicon.png
var iconData []byte

var (
	h             *handlers.Handlers
	mqttClient    *mqtt.Client
	messageChan   chan *models.Message
	appCtx        context.Context
	closeToTray   bool
	windowVisible bool
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	appCtx = ctx

	userDataDir := getUserDataDir()
	log.Printf("User data directory: %s", userDataDir)

	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		log.Printf("Failed to create user data directory: %v", err)
	}

	if err := database.Init(userDataDir); err != nil {
		log.Printf("Failed to initialize database: %v", err)
	}

	settings, _ := database.DB.GetAllSettings(context.Background())

	closeToTray = settings["closeToTray"] == "true"

	messageChan = make(chan *models.Message, 1000)
	mqttClient = mqtt.NewClient(messageChan, onConnect, onDisconnect)
	mqttClient.UpdateSettings(settings)

	go messageListener(ctx)

	h = handlers.NewHandlers(mqttClient)

	if closeToTray {
		setupTray(ctx)
		windowVisible = false
	} else {
		windowVisible = true
	}

	log.Printf("Application started successfully")
}

func setupTray(ctx context.Context) {
	log.Printf("Setting up tray icon, icon data length: %d", len(iconData))

	systray.Register(onTrayReady, onTrayExit)
}

func onTrayReady() {
	systray.SetIcon(iconData)
	systray.SetTooltip("MQTT5 Explorer")

	showHideItem := systray.AddMenuItem("Show/Hide MQTT5 Explorer", "Show or hide the main window")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem("Quit MQTT5 Explorer", "Quit the application")

	go func() {
		for {
			select {
			case <-showHideItem.ClickedCh:
				if windowVisible {
					runtime.WindowHide(appCtx)
					windowVisible = false
				} else {
					runtime.WindowShow(appCtx)
					windowVisible = true
				}
			case <-quitItem.ClickedCh:
				if appCtx != nil {
					runtime.Quit(appCtx)
				}
				os.Exit(0)
			}
		}
	}()
}

func onTrayExit() {
}

func getUserDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	appData := filepath.Join(homeDir, ".mqtt5-explorer-go")

	if os.Getenv("XDG_DATA_HOME") != "" {
		appData = filepath.Join(os.Getenv("XDG_DATA_HOME"), "mqtt5-explorer-go")
	}

	if os.Getenv("APPDATA") != "" {
		appData = filepath.Join(os.Getenv("APPDATA"), "mqtt5-explorer-go")
	}

	return appData
}

func (a *App) shutdown(_ context.Context) {
	mqttClient.DisconnectAll()
	database.DB.Close()
	log.Printf("Application shutdown")
}

func (a *App) beforeClose(ctx context.Context) bool {
	if closeToTray {
		runtime.WindowHide(ctx)
		windowVisible = false
		return true
	}
	return false
}

func onConnect(connectionID int64) {
	if appCtx != nil {
		runtime.EventsEmit(appCtx, "connection-status", map[string]interface{}{
			"connectionId": connectionID,
			"status":       "connected",
		})
	}
}

func onDisconnect(connectionID int64) {
	if appCtx != nil {
		runtime.EventsEmit(appCtx, "connection-status", map[string]interface{}{
			"connectionId": connectionID,
			"status":       "disconnected",
		})
	}
}

func messageListener(ctx context.Context) {
	for {
		select {
		case msg := <-messageChan:
			data, _ := json.Marshal(msg)
			runtime.EventsEmit(ctx, "mqtt-message", string(data))
			if len(msg.Payload) == 0 {
				runtime.EventsEmit(ctx, "topic-delete", msg.Topic)
			}
		}
	}
}

// Methods without context - Wails will inject context automatically

func (a *App) GetSettings() (map[string]string, error) {
	return h.GetSettings(context.Background())
}

func (a *App) SetSetting(key, value string) error {
	return h.SetSetting(context.Background(), key, value)
}

func (a *App) GetConnections() ([]models.Connection, error) {
	return h.GetConnections(context.Background())
}

func (a *App) GetConnection(id int64) (*models.Connection, error) {
	return h.GetConnection(context.Background(), id)
}

func (a *App) CreateConnection(conn *models.Connection) (int64, error) {
	return h.CreateConnection(context.Background(), conn)
}

func (a *App) UpdateConnection(conn *models.Connection) error {
	return h.UpdateConnection(context.Background(), conn)
}

func (a *App) DeleteConnection(id int64) error {
	return h.DeleteConnection(context.Background(), id)
}

func (a *App) SearchConnections(query string) ([]models.Connection, error) {
	return h.SearchConnections(context.Background(), query)
}

func (a *App) ExportConnections(ids []int64) (string, error) {
	return h.ExportConnections(context.Background(), ids)
}

func (a *App) ExportConnectionsToFile(ids []int64) (string, error) {
	if a.ctx == nil {
		return "", nil
	}

	jsonData, err := h.ExportConnections(context.Background(), ids)
	if err != nil {
		return "", err
	}

	now := time.Now()
	filename := now.Format("2006-01-02")
	defaultFilename := "m5g-export-" + filename + ".json"

	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Connections",
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON Files",
				Pattern:     "*.json",
			},
		},
	})

	if err != nil {
		return "", err
	}

	if filePath == "" {
		return "", nil
	}

	err = os.WriteFile(filePath, []byte(jsonData), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (a *App) ImportConnections(jsonData string) ([]int64, error) {
	return h.ImportConnections(context.Background(), jsonData)
}

func (a *App) ImportFromOldVersion(jsonData string) (int, error) {
	return h.ImportFromOldVersion(context.Background(), jsonData)
}

func (a *App) Connect(id int64) error {
	return h.Connect(context.Background(), id)
}

func (a *App) Disconnect(id int64) error {
	return h.Disconnect(context.Background(), id)
}

func (a *App) GetConnectionStatus(id int64) (bool, string, error) {
	return h.GetConnectionStatus(context.Background(), id)
}

func (a *App) GetClientID(id int64) (string, error) {
	return h.GetClientID(context.Background(), id)
}

func (a *App) GetMessages(connectionID int64, topic string, limit int) ([]models.Message, error) {
	return h.GetMessages(context.Background(), connectionID, topic, limit)
}

func (a *App) SearchMessages(connectionID int64, pattern string, useRegex bool) ([]models.Message, error) {
	return h.SearchMessages(context.Background(), connectionID, pattern, useRegex)
}

func (a *App) ClearMessages(connectionID int64) error {
	return h.ClearMessages(context.Background(), connectionID)
}

func (a *App) SendMessage(req *models.SendMessageRequest) error {
	return h.SendMessage(context.Background(), req)
}

func (a *App) DeleteTopicSubtree(connectionID int64, topic string) error {
	return h.DeleteTopicSubtree(context.Background(), connectionID, topic)
}

func (a *App) Subscribe(connectionID int64, topic string, qos int) error {
	return h.Subscribe(context.Background(), connectionID, topic, qos)
}

func (a *App) Unsubscribe(connectionID int64, topic string) error {
	return h.Unsubscribe(context.Background(), connectionID, topic)
}

func (a *App) GetTopicTree(connectionID int64) (*models.TopicNode, error) {
	return h.GetTopicTree(context.Background(), connectionID)
}

func (a *App) GetNumericMessages(connectionID int64, topic string, limit int) ([]models.Message, error) {
	return h.GetNumericMessages(context.Background(), connectionID, topic, limit)
}

func (a *App) GetHostInfo() (string, string, error) {
	return h.GetHostInfo(context.Background())
}

func (a *App) CheckPort(host string, port int) bool {
	return h.CheckPort(context.Background(), host, port)
}

func (a *App) GetVersion() string {
	var config struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}

	json.Unmarshal(WailsJSON, &config)

	return config.Info.ProductVersion
}
