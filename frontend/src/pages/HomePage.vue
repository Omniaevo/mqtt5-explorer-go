<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAppStore } from '../stores/app'
import { useRouter } from 'vue-router'
import type { Connection } from '../types'
import {
  DialogRoot,
  DialogTrigger,
  DialogContent,
  DialogTitle,
  DialogClose,
  DialogDescription
} from 'radix-vue'

const store = useAppStore()
const router = useRouter()

const searchQuery = ref('')
const showDialog = ref(false)
const editingConnection = ref<Connection | null>(null)
const showInfoDialog = ref(false)
const infoConnection = ref<Connection | null>(null)
const showDeleteDialog = ref(false)
const connectionToDelete = ref<Connection | null>(null)
const showPassword = ref(false)
const showInfoPassword = ref(false)

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
  defaultSubscriptions: '#,$SYS/#'
})

const filteredConnections = computed(() => {
  return store.connections || []
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
  if (isOpen && !editingConnection.value) {
    openNewConnection()
  }
})

function openNewConnection() {
  editingConnection.value = null
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
    defaultSubscriptions: '#,$SYS/#'
  }
  showDialog.value = true
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
    defaultSubscriptions: conn.defaultSubscriptions || ''
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
    await store.deleteConnection(connectionToDelete.value.id)
    showDeleteDialog.value = false
    connectionToDelete.value = null
  }
}

function openInfoDialog(conn: Connection) {
  infoConnection.value = conn
  showInfoDialog.value = true
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

onMounted(() => {
  store.loadConnections()
})
</script>

<template>
  <div class="home-page">
    <header class="page-header">
      <div class="header-left">
        <h1>Connections</h1>
        <span class="text-muted text-sm">{{ store.connections?.length || 0 }} connections</span>
      </div>
      <div class="header-right">
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
          <DialogTrigger as-child>
            <button class="btn btn-primary">
              <span class="mdi mdi-plus"></span>
              New Connection
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
                <div class="form-group">
                  <label class="label">Username</label>
                  <input v-model="formData.username" type="text" class="input w-full" />
                </div>
                <div class="form-group">
                  <label class="label">Password</label>
                  <div class="password-input">
                    <input v-model="formData.password" :type="showPassword ? 'text' : 'password'" class="input w-full" />
                    <button type="button" class="password-toggle" @click="showPassword = !showPassword">
                      <span class="mdi" :class="showPassword ? 'mdi-eye-off' : 'mdi-eye'"></span>
                    </button>
                  </div>
                </div>
              </div>

              <div class="divider"></div>

              <details class="advanced-section">
                <summary class="advanced-toggle">Advanced Settings</summary>
                
                <div class="form-group">
                  <label class="checkbox-label">
                    <input v-model="formData.validateCert" type="checkbox" />
                    Validate Certificate
                  </label>
                </div>

                <div class="form-group">
                  <label class="label">CA File</label>
                  <input v-model="formData.caFile" type="text" class="input w-full" placeholder="Path to CA certificate" />
                </div>

                <div class="form-row">
                  <div class="form-group">
                    <label class="label">Client Certificate</label>
                    <input v-model="formData.clientCert" type="text" class="input w-full" placeholder="Path to client cert" />
                  </div>
                  <div class="form-group">
                    <label class="label">Client Key</label>
                    <input v-model="formData.clientKey" type="text" class="input w-full" placeholder="Path to client key" />
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

        <DialogRoot v-model:open="showInfoDialog">
          <DialogContent class="dialog-content" v-if="infoConnection">
            <DialogTitle class="dialog-title">
              Connection Details
            </DialogTitle>
            <DialogDescription class="dialog-description">
              {{ infoConnection.name }}
            </DialogDescription>

            <div class="info-grid">
              <div class="info-row">
                <span class="info-label">Protocol</span>
                <span class="info-value">{{ infoConnection.protocol.toUpperCase() }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">MQTT Version</span>
                <span class="info-value">{{ infoConnection.mqttVersion === 3 ? '3.1' : infoConnection.mqttVersion === 4 ? '3.1.1' : '5.0' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Host</span>
                <span class="info-value mono">{{ infoConnection.host }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Port</span>
                <span class="info-value">{{ infoConnection.port }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Username</span>
                <span class="info-value">{{ infoConnection.username || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Password</span>
                <div class="info-password">
                  <span class="info-value mono">{{ infoConnection.password ? (showInfoPassword ? infoConnection.password : '••••••••') : '-' }}</span>
                  <button v-if="infoConnection.password" type="button" class="password-toggle" @click="showInfoPassword = !showInfoPassword">
                    <span class="mdi" :class="showInfoPassword ? 'mdi-eye-off' : 'mdi-eye'"></span>
                  </button>
                </div>
              </div>
              <div class="info-row">
                <span class="info-label">Validate Certificate</span>
                <span class="info-value">{{ infoConnection.validateCert ? 'Yes' : 'No' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">CA File</span>
                <span class="info-value mono">{{ infoConnection.caFile || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Client Certificate</span>
                <span class="info-value mono">{{ infoConnection.clientCert || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Client Key</span>
                <span class="info-value mono">{{ infoConnection.clientKey || '-' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">Default Subscriptions</span>
                <span class="info-value mono">{{ infoConnection.defaultSubscriptions || '-' }}</span>
              </div>
            </div>

            <div class="form-actions">
              <DialogClose as-child>
                <button type="button" class="btn btn-primary">Close</button>
              </DialogClose>
            </div>

            <DialogClose class="dialog-close">
              <span class="mdi mdi-close"></span>
            </DialogClose>
          </DialogContent>
        </DialogRoot>

        <DialogRoot v-model:open="showDeleteDialog">
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
    </header>

    <div class="connections-grid">
      <div
        v-for="conn in filteredConnections || []"
        :key="conn.id"
        class="connection-card"
      >
        <div class="card-header">
          <div class="card-title-area">
            <h3>{{ conn.name }}</h3>
            <span class="badge" :class="getStatusClass(conn)">{{ getStatusText(conn) }}</span>
          </div>
        </div>

        <div class="card-body">
          <div class="connection-info">
            <span class="info-label">Protocol:</span>
            <span class="info-value">{{ conn.protocol.toUpperCase() }}</span>
          </div>
          <div class="connection-info">
            <span class="info-label">Version:</span>
            <span class="info-value">{{ conn.mqttVersion === 3 ? '3.1' : conn.mqttVersion === 4 ? '3.1.1' : '5.0' }}</span>
          </div>
          <div class="connection-info">
            <span class="info-label">Address:</span>
            <span class="info-value mono">{{ conn.host }}:{{ conn.port }}</span>
          </div>
        </div>

        <div class="card-footer">
          <div class="card-actions-footer">
            <button 
              class="btn btn-secondary flex-1"
              @click="openInfoDialog(conn)"
            >
              <span class="mdi mdi-information"></span>
              Info
            </button>
            <button 
              class="btn btn-secondary flex-1"
              @click="openEditConnection(conn)"
            >
              <span class="mdi mdi-pencil"></span>
              Edit
            </button>
            <button 
              class="btn btn-secondary flex-1"
              @click="confirmDeleteConnection(conn)"
            >
              <span class="mdi mdi-delete"></span>
              Delete
            </button>
          </div>
          <button
            v-if="store.connectionStatuses[conn.id]?.status === 'connected'"
            class="btn btn-secondary w-full"
            @click="disconnectFrom(conn)"
          >
            <span class="mdi mdi-link-off"></span>
            Disconnect
          </button>
          <button
            v-else
            class="btn btn-primary w-full"
            :disabled="store.hasActiveConnection"
            @click="connectTo(conn)"
          >
            <span class="mdi mdi-link"></span>
            Connect
          </button>
        </div>
      </div>

      <div v-if="!filteredConnections?.length" class="empty-state">
        <span class="mdi mdi-connection icon"></span>
        <h3>No connections yet</h3>
        <p class="text-muted">Create your first MQTT connection to get started</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-background);
}

.header-left h1 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 4px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
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
  width: 240px;
}

.connections-grid {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
  align-content: start;
}

.connection-card {
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
}

.card-header {
  margin-bottom: 12px;
  overflow: hidden;
}

.card-title-area {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  overflow: hidden;
}

.card-title-area h3 {
  font-size: 16px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}

.card-actions {
  display: flex;
  gap: 4px;
}

.card-body {
  margin-bottom: 16px;
}

.card-footer {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card-actions-footer {
  display: flex;
  gap: 8px;
}

.card-actions-footer .btn {
  flex: 1;
}

.connection-info {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  font-size: 13px;
}

.info-label {
  color: var(--color-muted-foreground);
}

.info-value {
  color: var(--color-foreground);
  font-weight: 500;
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  text-align: center;
}

.empty-state .icon {
  font-size: 64px;
  color: var(--color-muted-foreground);
  margin-bottom: 16px;
}

.empty-state h3 {
  font-size: 18px;
  margin-bottom: 8px;
}

.dialog-content {
  background: var(--color-card);
  border-radius: 8px;
  padding: 24px;
  width: 480px;
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

.info-password {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-password .password-toggle {
  position: static;
  padding: 2px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.advanced-section {
  margin-top: 8px;
}

.advanced-toggle {
  cursor: pointer;
  color: var(--color-primary);
  font-weight: 500;
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

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid var(--color-border);
}

.info-row:last-child {
  border-bottom: none;
}

.info-row .info-label {
  color: var(--color-muted-foreground);
  font-size: 13px;
}

.info-row .info-value {
  color: var(--color-foreground);
  font-size: 13px;
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.info-row .info-value.mono {
  font-family: monospace;
}
</style>
