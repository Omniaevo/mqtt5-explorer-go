# MQTT5 Explorer Go - Specification Document

## 1. Project Overview

**Project Name:** MQTT5 Explorer Go  
**Project Type:** Cross-platform Desktop Application  
**Core Feature Summary:** A professional MQTT client for developers to connect to MQTT brokers, monitor messages, and send messages with full MQTT 5.0 support.  
**Target Users:** Developers, IoT engineers, and system administrators working with MQTT protocol.

---

## 2. Technology Stack

- **Framework:** Wails (Go + WebView)
- **Frontend:** Vue 3 with TypeScript
- **UI Library:** Radix Vue
- **Icons:** MDI Icons (@mdi/font)
- **Database:** SQLite (via modernc.org/sqlite)
- **MQTT Library:** paho.mqtt.golang
- **Charting:** Chart.js with vue-chartjs

---

## 3. UI/UX Specification

### 3.1 Window Management

- **Main Window:** Single window application with sidebar navigation
- **Minimum Size:** 1024x768
- **Default Size:** 1280x800
- **System Tray:** Application minimizes to system tray when configured
- **Window Controls:** Native window frame with standard controls

### 3.2 Layout Structure

```
┌─────────────────────────────────────────────────────────────────┐
│  Title Bar (Native)                                    [─][□][×]│
├───────────┬─────────────────────────────────────────────────────┤
│           │  Header: Connection Status / Search Bar            │
│  Sidebar  ├─────────────────────────────────────────────────────┤
│           │                                                     │
│  - Home   │  Main Content Area                                  │
│  - Tree   │  (Connection List / Message Tree / Message Detail) │
│  - Charts │                                                     │
│  - Settings                                                     │
│           │                                                     │
├───────────┴─────────────────────────────────────────────────────┤
│  Status Bar: Connection Info / Message Count                   │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 Visual Design

#### Color Palette

**Light Mode (Default):**
- Background Primary: `#FFFFFF`
- Background Secondary: `#F5F5F7`
- Background Tertiary: `#E8E8ED`
- Text Primary: `#1D1D1F`
- Text Secondary: `#6E6E73`
- Border: `#D2D2D7`
- Accent (Default Blue): `#007AFF`
- Success: `#34C759`
- Warning: `#FF9500`
- Error: `#FF3B30`

**Dark Mode:**
- Background Primary: `#1C1C1E`
- Background Secondary: `#2C2C2E`
- Background Tertiary: `#3A3A3C`
- Text Primary: `#F5F5F7`
- Text Secondary: `#A1A1A6`
- Border: `#48484A`
- Accent: `#0A84FF`

#### Typography

- **Font Family:** -apple-system, BlinkMacSystemFont, "SF Pro Display", "Segoe UI", Roboto, sans-serif
- **Heading 1:** 24px, 600 weight
- **Heading 2:** 20px, 600 weight
- **Heading 3:** 16px, 600 weight
- **Body:** 14px, 400 weight
- **Caption:** 12px, 400 weight
- **Monospace:** "SF Mono", Monaco, "Cascadia Code", Consolas, monospace

#### Spacing System

- Base unit: 4px
- xs: 4px
- sm: 8px
- md: 16px
- lg: 24px
- xl: 32px
- 2xl: 48px

#### Visual Effects

- Border radius: 8px (cards), 6px (buttons), 4px (inputs)
- Shadow (light): `0 2px 8px rgba(0,0,0,0.08)`
- Shadow (dark): `0 2px 8px rgba(0,0,0,0.32)`
- Transitions: 150ms ease-out

### 3.4 Components

#### Sidebar Navigation
- Width: 220px (collapsible to 60px)
- Icons with labels
- Active state with accent color background
- Hover state with subtle background

#### Connection Card
- Displays: Name, Protocol, Host:Port, Status indicator
- Actions: Edit, Delete, Connect/Disconnect
- States: Disconnected (gray), Connected (green), Error (red)

#### Message Tree
- Hierarchical tree view by topic levels
- Expand/collapse nodes
- Message count badge per topic
- Last message timestamp
- Search with regex support

#### Message Detail Panel
- Topic name (copyable)
- Timestamp
- QoS level
- Retain flag
- Payload (with format toggle: Raw/JSON/Text)
- MQTT 5.0 Properties:
  - Content Type
  - User Properties (key-value pairs)
  - Response Topic
  - Correlation Data
  - Message Expiry
  - Topic Alias

#### Send Message Dialog
- Topic input with autocomplete from subscriptions
- QoS selector (0, 1, 2)
- Retain checkbox
- Payload editor with JSON validation
- MQTT 5.0 options (collapsible):
  - Content Type input
  - User Properties (add/remove key-value)
  - Response Topic
  - Correlation Data (hex)
  - Message Expiry Interval

#### Settings Panel
- Grouped settings with section headers
- Toggle switches for boolean options
- Number inputs with validation
- Color picker for accent color

---

## 4. Functional Specification

### 4.1 Home Page - Connection Management

#### Create Connection
- **Required Fields:**
  - Connection Name (string, unique)
  - MQTT Version (dropdown: 3.1, 3.1.1, 5.0)
  - Protocol (dropdown: mqtt, mqtts, ws, wss)
  - Host (string, hostname or IP)
  - Port (number, 1-65535)
  - Username (string, optional)
  - Password (string, optional, masked)

- **Advanced Properties (Expandable):**
  - Validate Certificate (boolean, default: true for mqtts/wss)
  - CA File (file path, optional)
  - Client Certificate (file path, optional)
  - Client Key (file path, optional)
  - Default Topic Subscriptions (list of topic patterns, optional)

#### Connection Actions
- Save connection (persist to SQLite)
- Edit existing connection
- Delete connection (with confirmation dialog)
- Duplicate connection
- Export connection to JSON
- Import connection from JSON
- Connect/Disconnect toggle

#### Search Connections
- Real-time search by connection name
- Search by host
- Filter by protocol
- Filter by connection status

### 4.2 MQTT Connection

#### Connection Behavior
- Auto-generate client ID (format: `m5` + UUID v4)
- Configurable keepalive interval
- Configurable reconnect period
- Configurable max reconnect attempts
- Configurable connection timeout
- Clean session flag (configurable)

#### TLS/SSL Configuration
- CA certificate file
- Client certificate file
- Client key file
- Certificate validation toggle

#### MQTT 5.0 Features
- User Properties support
- Content Type support
- Topic Alias support
- Message Expiry Interval
- Response Topic
- Correlation Data

### 4.3 Message Tree

#### Display
- Hierarchical topic structure (split by `/`)
- Expandable/collapsible nodes
- Message count per topic
- Last message timestamp
- QoS indicators (icons)

#### Search
- Text search in topic names
- Regular expression search
- Case-sensitive toggle
- Real-time filtering

#### Caching
- Configurable cache size per topic (global setting)
- SQLite storage for cached messages
- Automatic cleanup of old messages beyond cache limit
- Cache persistence across app restarts

### 4.4 Message Detail

#### Basic Properties
- Topic (full path)
- Payload (hex, text, JSON formatted)
- QoS (0, 1, 2)
- Retain flag
- Timestamp (received)
- Client ID (sender)

#### MQTT 5.0 Properties (when applicable)
- Content Type
- User Properties (table format)
- Response Topic
- Correlation Data
- Message Expiry Interval
- Topic Alias
- Subscription Identifier

### 4.5 Send Message

- Topic input with subscription autocomplete
- QoS selection (0, 1, 2)
- Retain flag
- Payload input (text area)
- JSON payload validation
- MQTT 5.0 options:
  - Content Type
  - User Properties (key-value editor)
  - Response Topic
  - Correlation Data
  - Message Expiry Interval

### 4.6 Charts

#### Chart Types
- Line chart (time series)
- Value display (latest numeric value)

#### Chart Configuration
- Select topic to visualize
- Time range selector
- Auto-refresh toggle

#### Data Handling
- Parse numeric payloads
- Handle JSON payloads with numeric fields
- Timestamp on x-axis
- Value on y-axis

### 4.7 System Tray

#### Tray Icon
- Application icon in system tray
- Tooltip showing connection status
- Context menu:
  - Show/Hide Window
  - Connection status
  - Quit Application

#### Close to Tray
- Configurable in settings (default: true)
- When enabled, close button minimizes to tray
- When disabled, close button quits application

### 4.8 Settings

#### Appearance
- **Theme:** Light / Dark mode toggle (default: Light)
- **Accent Color:** Color picker (default: #007AFF blue)

#### Behavior
- **Close to Tray:** Boolean toggle (default: true)
- **Max Cached Messages:** Number input (default: 7)

#### MQTT Defaults
- **Client ID Prefix:** String (default: "m5")
- **Keepalive:** Number in seconds (default: 120)
- **Reconnect Period:** Number in seconds (default: 2)
- **Max Reconnects:** Number (default: 2, 0 = unlimited)
- **Connection Timeout:** Number in seconds (default: 20)

### 4.9 Data Persistence

#### SQLite Schema

```sql
-- Settings table
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Connections table
CREATE TABLE connections (
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
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Messages cache table
CREATE TABLE messages (
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
    correlation_data TEXT,
    message_expiry INTEGER,
    topic_alias INTEGER,
    FOREIGN KEY (connection_id) REFERENCES connections(id) ON DELETE CASCADE
);

CREATE INDEX idx_messages_connection_topic ON messages(connection_id, topic);
CREATE INDEX idx_messages_timestamp ON messages(timestamp);
```

#### Export/Import
- Export single connection to JSON
- Export all connections to JSON
- Import connections from JSON
- Validate JSON structure on import

---

## 5. Acceptance Criteria

### 5.1 Core Functionality
- [ ] Application launches without errors on Linux, macOS, Windows
- [ ] Can create, edit, delete MQTT connections
- [ ] Can connect to MQTT brokers (mqtt, mqtts, ws, wss)
- [ ] Can receive and display incoming messages in tree view
- [ ] Can search messages by topic with regex support
- [ ] Can send messages with full MQTT 5.0 properties
- [ ] Messages are cached in SQLite and persist across restarts

### 5.2 UI/UX
- [ ] MacOS-like appearance with proper typography and spacing
- [ ] Radix Vue components used throughout
- [ ] MDI icons displayed correctly
- [ ] Light/Dark mode toggles correctly
- [ ] Accent color changes applied throughout UI
- [ ] Sidebar navigation works correctly

### 5.3 System Integration
- [ ] System tray icon appears
- [ ] Close to tray works when enabled
- [ ] Settings persist correctly in SQLite

### 5.4 Data Management
- [ ] Connections saved to SQLite
- [ ] Settings saved to SQLite
- [ ] Messages cached in SQLite
- [ ] Export to JSON works
- [ ] Import from JSON works

### 5.5 Charts
- [ ] Can select topic for charting
- [ ] Numeric messages displayed in line chart
- [ ] Chart updates with new messages

---

## 6. File Structure

```
mqtt5-explorer-go/
├── frontend/
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   │   ├── common/
│   │   │   ├── connections/
│   │   │   ├── messages/
│   │   │   ├── charts/
│   │   │   └── settings/
│   │   ├── composables/
│   │   ├── layouts/
│   │   ├── pages/
│   │   ├── stores/
│   │   ├── styles/
│   │   ├── types/
│   │   ├── App.vue
│   │   └── main.ts
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── database/
│   │   └── database.go
│   ├── mqtt/
│   │   └── client.go
│   ├── handlers/
│   │   ├── connections.go
│   │   ├── messages.go
│   │   └── settings.go
│   └── models/
│       └── models.go
├── wails.json
├── README.md
└── SPEC.md
```
