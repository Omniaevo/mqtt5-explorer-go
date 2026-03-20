<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import moment from 'moment'
import { useAppStore } from '../stores/app'
import { useRouter } from 'vue-router'
import type { Connection } from '../types'
import {
  SplitterGroup,
  SplitterPanel,
  SplitterResizeHandle,
  DialogRoot,
  DialogTrigger,
  DialogContent,
  DialogTitle,
  DialogClose,
  DialogDescription,
  DialogOverlay
} from 'radix-vue'

const store = useAppStore()
const router = useRouter()

const searchQuery = ref('')
const showDialog = ref(false)
const editingConnection = ref<Connection | null>(null)
const showDeleteDialog = ref(false)
const connectionToDelete = ref<Connection | null>(null)
const showPassword = ref(false)
const showInfoPassword = ref(false)
const selectedConnection = ref<Connection | null>(null)
const sidebarWidth = ref(380)

const formData = ref({
  name: '',
  mqttVersion: 4,
  protocol: 'mqtt' as 'mqtt' | 'mqtts' | 'ws' | 'wss',
  host: 'localhost',
  port: 1883,
  username: '',
  password: '',
  validateCert: true,
  caFile: '',
  clientCert: '',
  clientKey: '',
  defaultSubscriptions: '#,$SYS/#',
  favourite: false
})

let isCloning = false

const caFileInputRef = ref<HTMLInputElement | null>(null)
const clientCertInputRef = ref<HTMLInputElement | null>(null)
const clientKeyInputRef = ref<HTMLInputElement | null>(null)

function startResize(e: MouseEvent) {
  document.body.style.userSelect = 'none'
  const startX = e.clientX
  const startWidth = sidebarWidth.value

  function doResize(e: MouseEvent) {
    const delta = e.clientX - startX
    const newWidth = Math.min(Math.max(startWidth + delta, 200), 600)
    sidebarWidth.value = newWidth
  }

  function stopResize() {
    document.body.style.userSelect = ''
    document.removeEventListener('mousemove', doResize)
    document.removeEventListener('mouseup', stopResize)
  }

  document.addEventListener('mousemove', doResize, { passive: true })
  document.addEventListener('mouseup', stopResize)
}

function selectFile(field: 'caFile' | 'clientCert' | 'clientKey') {
  if (field === 'caFile') {
    caFileInputRef.value?.click()
  } else if (field === 'clientCert') {
    clientCertInputRef.value?.click()
  } else if (field === 'clientKey') {
    clientKeyInputRef.value?.click()
  }
}

function handleFileChange(event: Event, field: 'caFile' | 'clientCert' | 'clientKey') {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    formData.value[field] = (file as any).path || file.name
  }
}

const filteredConnections = computed(() => {
  const connections = store.connections || []
  if (!searchQuery.value) return connections
  const query = searchQuery.value.toLowerCase()
  return connections.filter(conn =>
    conn.name.toLowerCase().includes(query) ||
    conn.host.toLowerCase().includes(query)
  )
})

const favouriteConnections = computed(() => {
  return (filteredConnections.value || []).filter(conn => conn.favourite)
})

const regularConnections = computed(() => {
  return (filteredConnections.value || []).filter(conn => !conn.favourite)
})

const protocolOptions = [
  { value: 'mqtt', label: 'MQTT' },
  { value: 'mqtts', label: 'MQTTS' },
  { value: 'ws', label: 'WebSocket' },
  { value: 'wss', label: 'WSS' }
]

const mqttVersionOptions = [
  { value: 3, label: '3.1' },
  { value: 4, label: '3.1.1' },
  { value: 5, label: '5.0' }
]

watch(searchQuery, (val) => {
  store.searchConnections(val)
})

watch(showDialog, (isOpen) => {
  if (isOpen) {
    if (!editingConnection.value && !isCloning) {
      openNewConnection()
    }
    isCloning = false
  } else {
    editingConnection.value = null
    resetForm()
  }
})

function resetForm() {
  formData.value = {
    name: '',
    mqttVersion: 4,
    protocol: 'mqtt',
    host: 'localhost',
    port: 1883,
    username: '',
    password: '',
    validateCert: true,
    caFile: '',
    clientCert: '',
    clientKey: '',
    defaultSubscriptions: '#,$SYS/#',
    favourite: false
  }
}

function openNewConnection() {
  editingConnection.value = null
  resetForm()
}

function openEditConnection(conn: Connection) {
  editingConnection.value = conn
  formData.value = {
    name: conn.name,
    mqttVersion: conn.mqttVersion,
    protocol: conn.protocol,
    host: conn.host,
    port: conn.port,
    username: conn.username || '',
    password: conn.password || '',
    validateCert: conn.validateCert,
    caFile: conn.caFile || '',
    clientCert: conn.clientCert || '',
    clientKey: conn.clientKey || '',
    defaultSubscriptions: conn.defaultSubscriptions || '',
    favourite: conn.favourite
  }
  showDialog.value = true
}

function openCloneConnection(conn: Connection) {
  isCloning = true
  editingConnection.value = null
  formData.value = {
    name: '',
    mqttVersion: conn.mqttVersion,
    protocol: conn.protocol,
    host: conn.host,
    port: conn.port,
    username: conn.username || '',
    password: conn.password || '',
    validateCert: conn.validateCert,
    caFile: conn.caFile || '',
    clientCert: conn.clientCert || '',
    clientKey: conn.clientKey || '',
    defaultSubscriptions: conn.defaultSubscriptions || '',
    favourite: false
  }
  showDialog.value = true
}

async function saveConnection() {
  try {
    if (editingConnection.value) {
      await store.updateConnection({
        ...editingConnection.value,
        ...formData.value
      })
    } else {
      await store.createConnection(formData.value)
    }
    showDialog.value = false
  } catch (error) {
    console.error('Failed to save connection:', error)
  }
}

function confirmDeleteConnection(conn: Connection) {
  connectionToDelete.value = conn
  showDeleteDialog.value = true
}

async function deleteConnection() {
  if (connectionToDelete.value) {
    const connId = connectionToDelete.value.id
    await store.deleteConnection(connId)
    showDeleteDialog.value = false
    connectionToDelete.value = null
    if (selectedConnection.value?.id === connId) {
      selectedConnection.value = null
    }
  }
}

function selectConnection(conn: Connection) {
  selectedConnection.value = conn
}

async function connectTo(conn: Connection) {
  try {
    await store.connect(conn.id)
    router.push('/messages')
  } catch (error) {
    console.error('Failed to connect:', error)
  }
}

async function disconnectFrom(conn: Connection) {
  await store.disconnect(conn.id)
}

async function toggleFavorite(conn: Connection) {
  const updated = { ...conn, favourite: !conn.favourite }
  await store.updateConnection(updated)
  if (selectedConnection.value?.id === conn.id) {
    selectedConnection.value = updated
  }
}

function getStatusClass(conn: Connection) {
  const status = store.connectionStatuses[conn.id]?.status
  if (status === 'connected') return 'badge-success'
  if (status === 'connecting') return 'badge-warning'
  if (status === 'error') return 'badge-error'
  return 'badge-default'
}

function getStatusText(conn: Connection) {
  const status = store.connectionStatuses[conn.id]?.status
  if (status === 'connected') return 'Connected'
  if (status === 'connecting') return 'Connecting'
  if (status === 'error') return 'Error'
  return 'Disconnected'
}

function formatLastConnected(timestamp: string): string {
  return moment(timestamp).format('YYYY/MM/DD HH:mm')
}

onMounted(() => {
  store.loadConnections()
})
</script>

<template>
  <div class="home-page">
    <aside class="connections-sidebar" :style="{ width: sidebarWidth + 'px', minWidth: sidebarWidth + 'px' }">
      <div class="sidebar-header">
        <h2>Connections</h2>
        <span class="connection-count" :title="(store.connections?.length || 0) + ' connections'">
          <span class="mdi mdi-ethernet"></span>{{ store.connections?.length || 0 }}
        </span>
      </div>

      <div class="sidebar-toolbar">
        <div class="toolbar-row">
          <div class="search-box">
            <span class="mdi mdi-magnify"></span>
            <input
              v-model="searchQuery"
              type="text"
              class="input"
              placeholder="Search connections..."
            />
          </div>
          <DialogRoot v-model:open="showDialog">
            <DialogOverlay class="dialog-overlay" />
            <DialogTrigger as-child>
              <button class="btn btn-primary">
                <span class="mdi mdi-plus"></span>
                New
              </button>
            </DialogTrigger>
          <DialogContent class="dialog-content">
            <DialogTitle class="dialog-title">
              {{ editingConnection ? 'Edit Connection' : 'New Connection' }}
            </DialogTitle>
            <DialogDescription class="dialog-description">
              Configure your MQTT connection settings
            </DialogDescription>

            <form @submit.prevent="saveConnection" class="connection-form">
              <div class="form-group">
                <label class="label">Connection Name</label>
                <input v-model="formData.name" type="text" class="input w-full" required />
              </div>

              <div class="form-row">
                <div class="form-group">
                  <label class="label">MQTT Version</label>
                  <select v-model="formData.mqttVersion" class="input w-full">
                    <option v-for="opt in mqttVersionOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>
                <div class="form-group">
                  <label class="label">Protocol</label>
                  <select v-model="formData.protocol" class="input w-full">
                    <option v-for="opt in protocolOptions" :key="opt.value" :value="opt.value">
                      {{ opt.label }}
                    </option>
                  </select>
                </div>
              </div>

              <div class="form-row">
                <div class="form-group flex-2">
                  <label class="label">Host</label>
                  <input v-model="formData.host" type="text" class="input w-full" required />
                </div>
                <div class="form-group flex-1">
                  <label class="label">Port</label>
                  <input v-model.number="formData.port" type="number" class="input w-full" required />
                </div>
              </div>

              <div class="form-row">
                <div class="form-group flex-2">
                  <label class="label">Username</label>
                  <input v-model="formData.username" type="text" class="input w-full" />
                </div>
                <div class="form-group flex-2">
                  <label class="label">Password</label>
                  <div class="password-input">
                    <input v-model="formData.password" :type="showPassword ? 'text' : 'password'" class="input w-full" />
                    <button type="button" class="password-toggle" @click="showPassword = !showPassword">
                      <span class="mdi" :class="showPassword ? 'mdi-eye-off' : 'mdi-eye'"></span>
                    </button>
                  </div>
                </div>
              </div>

              <details class="advanced-section">
                <summary class="advanced-toggle"><span class="mdi mdi-chevron-right chevron"></span> Advanced Settings</summary>

                <div class="form-group">
                  <div class="switch-row">
                    <label class="label">Validate Certificate</label>
                    <label class="switch">
                      <input v-model="formData.validateCert" type="checkbox" />
                      <span class="slider"></span>
                    </label>
                  </div>
                </div>

                <div class="form-group">
                  <label class="label">CA File</label>
                  <input ref="caFileInputRef" type="file" @change="handleFileChange($event, 'caFile')" hidden />
                  <input v-model="formData.caFile" type="text" class="input w-full file-input" placeholder="Path to CA certificate" @click="selectFile('caFile')" readonly />
                </div>

                <div class="form-row">
                  <div class="form-group flex-2">
                    <label class="label">Client Certificate</label>
                    <input ref="clientCertInputRef" type="file" @change="handleFileChange($event, 'clientCert')" hidden />
                    <input v-model="formData.clientCert" type="text" class="input w-full file-input" placeholder="Path to client cert" @click="selectFile('clientCert')" readonly />
                  </div>
                  <div class="form-group flex-2">
                    <label class="label">Client Key</label>
                    <input ref="clientKeyInputRef" type="file" @change="handleFileChange($event, 'clientKey')" hidden />
                    <input v-model="formData.clientKey" type="text" class="input w-full file-input" placeholder="Path to client key" @click="selectFile('clientKey')" readonly />
                  </div>
                </div>

                <div class="form-group">
                  <label class="label">Default Subscriptions</label>
                  <input v-model="formData.defaultSubscriptions" type="text" class="input w-full" placeholder="Comma-separated topics" />
                </div>
              </details>

              <div class="form-actions">
                <DialogClose as-child>
                  <button type="button" class="btn btn-secondary">Cancel</button>
                </DialogClose>
                <button type="submit" class="btn btn-primary">
                  {{ editingConnection ? 'Save' : 'Create' }}
                </button>
              </div>
            </form>

            <DialogClose class="dialog-close">
              <span class="mdi mdi-close"></span>
            </DialogClose>
          </DialogContent>
        </DialogRoot>
        </div>
      </div>

      <div class="connections-list">
        <template v-if="favouriteConnections.length">
          <div class="list-separator">
            <span class="separator-text">Favourites</span>
          </div>
          <div
            v-for="conn in favouriteConnections"
            :key="conn.id"
            class="connection-item"
            :class="{ active: selectedConnection?.id === conn.id }"
            @click="selectConnection(conn)"
          >
            <div class="connection-item-main">
              <span class="mdi mdi-star favourite-star" :class="getStatusClass(conn)" :title="getStatusText(conn)"></span>
              <div class="connection-item-info">
                <span class="connection-name">{{ conn.name }}</span>
                <span class="connection-address mono text-muted text-sm">{{ conn.host }}:{{ conn.port }}</span>
              </div>
            </div>
            <button
              v-if="selectedConnection?.id !== conn.id && store.connectionStatuses[conn.id]?.status !== 'connected' && store.connectionStatuses[conn.id]?.status !== 'connecting' && !store.hasActiveConnection"
              class="connect-btn-small"
              @click.stop="connectTo(conn)"
              title="Connect"
            >
              <span class="mdi mdi-play"></span>
            </button>
          </div>
        </template>

        <template v-if="regularConnections.length">
          <div v-if="favouriteConnections.length" class="list-separator">
            <span class="separator-text">Others</span>
          </div>
          <div
            v-for="conn in regularConnections"
            :key="conn.id"
            class="connection-item"
            :class="{ active: selectedConnection?.id === conn.id }"
            @click="selectConnection(conn)"
          >
            <div class="connection-item-main">
              <span class="status-dot" :class="getStatusClass(conn)" :title="getStatusText(conn)"></span>
              <div class="connection-item-info">
                <span class="connection-name">{{ conn.name }}</span>
                <span class="connection-address mono text-muted text-sm">{{ conn.host }}:{{ conn.port }}</span>
              </div>
            </div>
            <button
              v-if="selectedConnection?.id !== conn.id && store.connectionStatuses[conn.id]?.status !== 'connected' && store.connectionStatuses[conn.id]?.status !== 'connecting' && !store.hasActiveConnection"
              class="connect-btn-small"
              @click.stop="connectTo(conn)"
              title="Connect"
            >
              <span class="mdi mdi-play"></span>
            </button>
          </div>
        </template>

        <div v-if="!filteredConnections?.length" class="empty-state">
          <span class="mdi mdi-connection icon"></span>
          <span class="text-muted text-sm">No connections</span>
        </div>
      </div>
    </aside>

    <div class="resize-handle" @mousedown="startResize"></div>

    <main class="connections-main">
      <div v-if="selectedConnection" class="connection-details">
        <div class="details-header">
          <div class="details-header-left">
            <div class="title-row">
              <h2>{{ selectedConnection.name }}</h2>
              <span class="badge" :class="getStatusClass(selectedConnection)">{{ getStatusText(selectedConnection) }}</span>
            </div>
            <span class="connection-caption text-muted text-sm">
              {{ selectedConnection.lastConnected ? formatLastConnected(selectedConnection.lastConnected) : 'Never connected' }}
            </span>
          </div>
          <button class="btn btn-ghost btn-icon" :title="selectedConnection.favourite ? 'Remove from favorites' : 'Add to favorites'" @click="toggleFavorite(selectedConnection)">
            <span class="mdi" :class="selectedConnection.favourite ? 'mdi-star-off' : 'mdi-star-outline'"></span>
          </button>
        </div>

        <div class="details-content">
          <div class="info-grid">
            <div class="info-row">
              <span class="info-label">Protocol</span>
              <span class="info-value">{{ selectedConnection.protocol.toUpperCase() }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">MQTT Version</span>
              <span class="info-value">{{ selectedConnection.mqttVersion === 3 ? '3.1' : selectedConnection.mqttVersion === 4 ? '3.1.1' : '5.0' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Host</span>
              <span class="info-value mono">{{ selectedConnection.host }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Port</span>
              <span class="info-value">{{ selectedConnection.port }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Username</span>
              <span class="info-value">{{ selectedConnection.username || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Password</span>
              <div class="info-password">
                <span class="info-value mono">{{ selectedConnection.password ? (showInfoPassword ? selectedConnection.password : '••••••••') : '-' }}</span>
                <button v-if="selectedConnection.password" type="button" class="password-toggle" @click="showInfoPassword = !showInfoPassword">
                  <span class="mdi" :class="showInfoPassword ? 'mdi-eye-off' : 'mdi-eye'"></span>
                </button>
              </div>
            </div>
            <div class="info-row">
              <span class="info-label">Validate Certificate</span>
              <span class="info-value">{{ selectedConnection.validateCert ? 'Yes' : 'No' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">CA File</span>
              <span class="info-value mono">{{ selectedConnection.caFile || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Client Certificate</span>
              <span class="info-value mono">{{ selectedConnection.clientCert || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Client Key</span>
              <span class="info-value mono">{{ selectedConnection.clientKey || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Default Subscriptions</span>
              <span class="info-value mono">{{ selectedConnection.defaultSubscriptions || '-' }}</span>
            </div>
          </div>

          <div class="details-actions">
            <button class="btn btn-flat btn-flat-secondary" @click="openEditConnection(selectedConnection)">
              <span class="mdi mdi-pencil"></span>
              Edit
            </button>
            <button class="btn btn-flat btn-flat-secondary" @click="openCloneConnection(selectedConnection)">
              <span class="mdi mdi-content-copy"></span>
              Clone
            </button>
            <button class="btn btn-flat btn-flat-danger" @click="confirmDeleteConnection(selectedConnection)">
              <span class="mdi mdi-delete"></span>
              Delete
            </button>
            <div class="details-actions-right">
              <button
                v-if="store.connectionStatuses[selectedConnection.id]?.status === 'connected'"
                class="btn btn-secondary"
                @click="disconnectFrom(selectedConnection)"
              >
                <span class="mdi mdi-link-off"></span>
                Disconnect
              </button>
              <button
                v-else
                class="btn btn-primary"
                :disabled="store.hasActiveConnection"
                @click="connectTo(selectedConnection)"
              >
                <span class="mdi mdi-link"></span>
                Connect
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="empty-details">
        <span class="mdi mdi-ethernet icon"></span>
        <h3>Select a connection</h3>
        <p class="text-muted">Choose a connection from the list to view details</p>
      </div>
    </main>

    <DialogRoot v-model:open="showDeleteDialog">
      <DialogOverlay class="dialog-overlay" />
      <DialogContent class="dialog-content">
        <DialogTitle class="dialog-title">
          Delete Connection
        </DialogTitle>
        <DialogDescription class="dialog-description">
          Are you sure you want to delete "{{ connectionToDelete?.name }}"? This action cannot be undone.
        </DialogDescription>

        <div class="form-actions">
          <DialogClose as-child>
            <button type="button" class="btn btn-secondary">Cancel</button>
          </DialogClose>
          <button type="button" class="btn btn-danger" @click="deleteConnection">
            Delete
          </button>
        </div>

        <DialogClose class="dialog-close" @click="showDeleteDialog = false">
          <span class="mdi mdi-close"></span>
        </DialogClose>
      </DialogContent>
    </DialogRoot>
  </div>
</template>

<style scoped>
.home-page {
  height: 100%;
  display: flex;
  overflow: hidden;
}

.connections-sidebar {
  width: 280px;
  min-width: 200px;
  max-width: 600px;
  background: var(--color-secondary);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
}

.resize-handle {
  width: 4px;
  cursor: col-resize;
  background: transparent;
  transition: background 0.15s ease;
}

.resize-handle:hover {
  background: var(--color-primary);
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
}

.sidebar-header h2 {
  font-size: 14px;
  font-weight: 600;
}

.connection-count {
  font-size: 11px;
  padding: 2px 8px;
  background: var(--color-muted);
  border-radius: 10px;
  color: var(--color-muted-foreground);
  display: flex;
  align-items: center;
  gap: 6px;
}

.connection-count .mdi {
  font-size: 12px;
}

.sidebar-toolbar {
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border);
}

.toolbar-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.search-box {
  flex: 1;
  position: relative;
}

.search-box .mdi {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-muted-foreground);
}

.search-box .input {
  padding-left: 36px;
  width: 100%;
}

.connections-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.list-separator {
  display: flex;
  align-items: center;
  padding: 8px 12px 4px;
  margin-top: 8px;
  border-bottom: 1px solid var(--color-border);
}

.separator-text {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-muted-foreground);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.connection-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s ease;
  margin-bottom: 4px;
}

.connection-item:hover {
  background: var(--color-muted);
}

.connection-item.active {
  background: var(--color-primary);
  color: white;
}

.connection-item.active .connection-address,
.connection-item.active .text-muted {
  color: rgba(255, 255, 255, 0.7) !important;
}

.connection-item-main {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.connection-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.connection-item-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.connection-name {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.connection-address {
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot.badge-success {
  background: #22c55e;
}

.status-dot.badge-warning {
  background: #f59e0b;
}

.status-dot.badge-error {
  background: #ef4444;
}

.status-dot.badge-default {
  background: var(--color-muted-foreground);
}

.favourite-star {
  font-size: 16px !important;
  flex-shrink: 0;
  width: 0;
  margin-right: 16px;
}

.favourite-star.badge-success {
  color: #22c55e !important;
}

.favourite-star.badge-warning {
  color: #f59e0b !important;
}

.favourite-star.badge-error {
  color: #ef4444 !important;
}

.favourite-star.badge-default {
  color: var(--color-muted-foreground) !important;
}

.connect-btn-small {
  width: 24px;
  height: 24px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: var(--color-primary);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.15s ease, background 0.15s ease;
  flex-shrink: 0;
}

.connect-btn-small:hover:not(:disabled) {
  transform: scale(1.1);
  background: var(--color-accent);
}

.connect-btn-small:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.connect-btn-small .mdi {
  font-size: 14px;
}

.connections-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--color-background);
}

.connection-details {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.details-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.details-header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.details-header h2 {
  font-size: 20px;
  font-weight: 600;
}

.connection-caption {
  font-size: 12px;
}

.details-content {
  flex: 1;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 24px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid var(--color-border);
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--color-muted-foreground);
  font-size: 13px;
}

.info-value {
  color: var(--color-foreground);
  font-size: 13px;
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.info-value.mono {
  font-family: monospace;
}

.info-password {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-password .password-toggle {
  position: static;
  padding: 2px;
}

.details-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--color-border);
}

.details-actions-right {
  margin-left: auto;
}

.btn-flat {
  background: transparent;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  transition: background 0.15s ease;
}

.btn-flat-secondary {
  color: var(--color-primary);
}

.btn-flat-secondary:hover {
  background: var(--color-muted);
}

.btn-flat-danger {
  color: var(--color-destructive);
}

.btn-flat-danger:hover {
  background: rgba(239, 68, 68, 0.1);
}

.empty-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 48px;
}

.empty-details .icon {
  font-size: 48px;
  color: var(--color-muted-foreground);
  margin-bottom: 16px;
}

.empty-details h3 {
  font-size: 18px;
  margin-bottom: 8px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  text-align: center;
}

.empty-state .icon {
  font-size: 32px;
  color: var(--color-muted-foreground);
  margin-bottom: 8px;
}

.dialog-content {
  background: var(--color-card);
  border-radius: 8px;
  padding: 24px;
  width: 560px;
  max-width: 90vw;
  max-height: 85vh;
  overflow-y: auto;
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  z-index: 1000;
}

.dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
}

.dialog-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
}

.dialog-description {
  color: var(--color-muted-foreground);
  font-size: 14px;
  margin-bottom: 20px;
}

.connection-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  margin-bottom: 12px;
}

.form-row {
  display: flex;
  gap: 12px;
}

.flex-1 {
  flex: 1;
}

.flex-2 {
  flex: 2;
}

.password-input {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input .input {
  padding-right: 36px;
}

.password-toggle {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-muted-foreground);
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.password-toggle:hover {
  color: var(--color-foreground);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.advanced-section {
  margin-top: 16px;
  padding: 16px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
}

.advanced-toggle {
  cursor: pointer;
  color: var(--color-primary);
  font-weight: 500;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  list-style: none;
}

.advanced-toggle::-webkit-details-marker {
  display: none;
}

.advanced-toggle .chevron {
  font-size: 18px;
  transition: transform 0.2s ease;
}

details[open] .advanced-toggle .chevron {
  transform: rotate(90deg);
}

details[open] .advanced-toggle {
  margin-bottom: 16px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}

.dialog-close {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 4px;
  color: var(--color-foreground);
}

.dialog-close:hover {
  background: var(--color-muted);
}

.dialog-close .mdi {
  font-size: 18px;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.switch-row .label {
  margin-bottom: 0;
}

.switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--color-muted);
  transition: 0.2s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.2s;
  border-radius: 50%;
}

.switch input:checked + .slider {
  background-color: var(--color-primary);
}

.switch input:checked + .slider:before {
  transform: translateX(20px);
}

.file-input {
  cursor: pointer;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  flex-shrink: 0;
}

.badge-success {
  background: #22c55e;
  color: white;
}

.badge-warning {
  background: #f59e0b;
  color: white;
}

.badge-error {
  background: #ef4444;
  color: white;
}

.badge-default {
  background: var(--color-muted);
  color: var(--color-muted-foreground);
}

.w-full {
  width: 100%;
}
</style>
