<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView, useRouter, useRoute } from 'vue-router'
import { useAppStore } from './stores/app'
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
const route = useRoute()

const showInfoDialog = ref(false)

const navItems = [
  { name: 'Home', path: '/', icon: 'mdi-home' },
  { name: 'Messages', path: '/messages', icon: 'mdi-message-text' },
  { name: 'Charts', path: '/charts', icon: 'mdi-chart-line' },
  { name: 'Settings', path: '/settings', icon: 'mdi-cog' }
]

onMounted(async () => {
  await store.loadSettings()
  await store.loadConnections()
})
</script>

<template>
  <div class="app-layout">
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="app-logo">
          <span class="mdi mdi-mqtt"></span>
        </div>
        <span class="app-title">MQTT Explorer</span>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: route.path === item.path }"
        >
          <span class="mdi" :class="item.icon"></span>
          <span class="nav-label">{{ item.name }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <span class="app-version">v1.0.0</span>
      </div>
    </aside>

    <main class="main-content">
      <RouterView />
      <div class="status-bar">
        <div v-if="store.currentConnection" class="connection-status">
          <span 
            class="status-dot"
            :class="store.isConnected ? 'connected' : 'disconnected'"
          ></span>
          <span class="status-text">{{ store.currentConnection?.name }}</span>
          <span v-if="store.isConnected && store.currentClientId" class="status-client-id">
            {{ store.currentClientId }}
          </span>
          <div class="status-actions">
            <button 
              class="info-btn"
              @click="showInfoDialog = true"
            >
              <span class="mdi mdi-information"></span>
              Info
            </button>
            <button 
              v-if="store.isConnected" 
              class="disconnect-btn"
              @click="store.disconnect(store.currentConnectionId!)"
            >
              <span class="mdi mdi-link-off"></span>
              Disconnect
            </button>
          </div>
        </div>
        <div v-else class="no-connection">
          <span class="mdi mdi-connection-off"></span>
          <span>No active connection</span>
        </div>
      </div>

      <DialogRoot v-model:open="showInfoDialog">
        <DialogContent class="dialog-content" v-if="store.currentConnection">
          <DialogTitle class="dialog-title">
            Connection Details
          </DialogTitle>
          <DialogDescription class="dialog-description">
            {{ store.currentConnection.name }}
          </DialogDescription>

          <div class="info-grid">
            <div class="info-row">
              <span class="info-label">Protocol</span>
              <span class="info-value">{{ store.currentConnection.protocol.toUpperCase() }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">MQTT Version</span>
              <span class="info-value">{{ store.currentConnection.mqttVersion === 3 ? '3.1' : store.currentConnection.mqttVersion === 4 ? '3.1.1' : '5.0' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Host</span>
              <span class="info-value mono">{{ store.currentConnection.host }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Port</span>
              <span class="info-value">{{ store.currentConnection.port }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Username</span>
              <span class="info-value">{{ store.currentConnection.username || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Validate Certificate</span>
              <span class="info-value">{{ store.currentConnection.validateCert ? 'Yes' : 'No' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">CA File</span>
              <span class="info-value mono">{{ store.currentConnection.caFile || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Client Certificate</span>
              <span class="info-value mono">{{ store.currentConnection.clientCert || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Client Key</span>
              <span class="info-value mono">{{ store.currentConnection.clientKey || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Default Subscriptions</span>
              <span class="info-value mono">{{ store.currentConnection.defaultSubscriptions || '-' }}</span>
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
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  min-width: 220px;
  background: var(--color-secondary);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
}

.app-logo {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary);
  border-radius: 8px;
  color: white;
  font-size: 18px;
  flex-shrink: 0;
}

.app-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-foreground);
}

.sidebar-nav {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 6px;
  color: var(--color-muted-foreground);
  text-decoration: none;
  transition: all 0.15s ease;
  margin-bottom: 2px;
}

.nav-item:hover {
  background: var(--color-muted);
  color: var(--color-foreground);
}

.nav-item.active {
  background: var(--color-primary);
  color: white;
}

.nav-item .mdi {
  font-size: 20px;
}

.nav-label {
  font-size: 14px;
  font-weight: 500;
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--color-border);
  text-align: center;
}

.app-version {
  font-size: 12px;
  color: var(--color-muted-foreground);
}

.main-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--color-background);
}

.status-bar {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  border-top: 1px solid var(--color-border);
  background: var(--color-secondary);
  min-height: 40px;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.status-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.status-text {
  font-size: 13px;
  color: var(--color-foreground);
  font-weight: 500;
}

.status-client-id {
  font-size: 12px;
  color: var(--color-muted-foreground);
  font-family: monospace;
  background: var(--color-muted);
  padding: 2px 8px;
  border-radius: 4px;
}

.disconnect-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  font-size: 12px;
  color: var(--color-muted-foreground);
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.disconnect-btn:hover {
  color: var(--color-foreground);
  background: var(--color-muted);
  border-color: var(--color-muted-foreground);
}

.disconnect-btn .mdi {
  font-size: 14px;
}

.no-connection {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-muted-foreground);
  font-size: 13px;
}

.no-connection .mdi {
  font-size: 18px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot.connected {
  background: var(--color-success);
}

.status-dot.disconnected {
  background: var(--color-muted-foreground);
}

.info-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
  font-size: 12px;
  color: var(--color-muted-foreground);
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.info-btn:hover {
  color: var(--color-foreground);
  background: var(--color-muted);
  border-color: var(--color-muted-foreground);
}

.info-btn .mdi {
  font-size: 14px;
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

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 16px;
  font-size: 14px;
  font-weight: 500;
  border-radius: 6px;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  opacity: 0.9;
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
</style>
