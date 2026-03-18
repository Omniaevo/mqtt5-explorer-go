<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAppStore } from '../stores/app'
import type { Settings } from '../types'

const store = useAppStore()

const localSettings = ref<Settings>({ ...store.settings })

const accentColorPresets = [
  '#007AFF', // Blue
  '#34C759', // Green
  '#FF9500', // Orange
  '#FF3B30', // Red
  '#AF52DE', // Purple
  '#FF2D55', // Pink
  '#5856D6', // Indigo
  '#00C7BE'  // Teal
]

watch(() => store.settings, (newSettings) => {
  if (newSettings) {
    localSettings.value = { ...newSettings }
  }
}, { deep: true, immediate: false })

async function saveSetting<K extends keyof Settings>(key: K, value: Settings[K]) {
  await store.saveSetting(key, value)
}

async function exportAllConnections() {
  if (!store.connections?.length) return

  try {
    const ids = store.connections.map(c => c.id)
    const filePath = await store.exportConnectionsToFile(ids)
    if (filePath) {
      console.log('Exported to:', filePath)
    }
  } catch (error) {
    console.error('Export failed:', error)
  }
}

async function importConnections(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    await store.importConnections(text)
    await store.loadConnections()
    store.showToast('Connections imported successfully')
  } catch (error) {
    console.error('Import failed:', error)
    store.showToast('Failed to import connections', 'error')
  }

  input.value = ''
}

async function importFromOldVersion(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const imported = await store.importFromOldVersion(text)
    await store.loadConnections()
    store.showToast(`Imported ${imported} connections successfully`)
  } catch (error) {
    console.error('Import failed:', error)
    store.showToast('Failed to import connections', 'error')
  }

  input.value = ''
}

function openGitHub() {
  window.open('https://github.com/anomalyco/mqtt5-explorer-go/issues', '_blank')
}

onMounted(() => {
  localSettings.value = { ...store.settings }
})
</script>

<template>
  <div class="settings-page">
    <header class="page-header">
      <h1>Settings</h1>
    </header>

    <div class="settings-content">
      <section class="settings-section">
        <h2 class="section-title">Appearance</h2>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Theme</label>
            <p class="setting-description">Choose between light and dark mode</p>
          </div>
          <div class="setting-control">
            <select
              :value="localSettings.theme"
              @change="saveSetting('theme', ($event.target as HTMLSelectElement).value as 'light' | 'dark')"
              class="input"
            >
              <option value="light">Light</option>
              <option value="dark">Dark</option>
            </select>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Accent Color</label>
            <p class="setting-description">Choose your preferred accent color</p>
          </div>
          <div class="setting-control">
            <div class="color-presets">
              <button
                v-for="color in accentColorPresets"
                :key="color"
                class="color-swatch"
                :class="{ active: localSettings.accentColor === color }"
                :style="{ backgroundColor: color }"
                @click="saveSetting('accentColor', color)"
              ></button>
            </div>
          </div>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="section-title">Behavior</h2>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Close to System Tray</label>
            <p class="setting-description">Keep the app running in the system tray when closed</p>
          </div>
          <div class="setting-control">
            <label class="switch">
              <input
                type="checkbox"
                :checked="localSettings.closeToTray"
                @change="saveSetting('closeToTray', ($event.target as HTMLInputElement).checked)"
              />
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Max Cached Messages</label>
            <p class="setting-description">Maximum number of messages to cache per topic</p>
          </div>
          <div class="setting-control">
            <input
              type="number"
              :value="localSettings.maxCachedMessages"
              @change="saveSetting('maxCachedMessages', parseInt(($event.target as HTMLInputElement).value))"
              class="input number-input"
              min="1"
              max="1000"
            />
          </div>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="section-title">MQTT Defaults</h2>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Default Client ID</label>
            <p class="setting-description">Default client ID used when connecting (leave empty to auto-generate)</p>
          </div>
          <div class="setting-control client-id-control">
            <input
              type="text"
              :value="localSettings.defaultClientId"
              @change="saveSetting('defaultClientId', ($event.target as HTMLInputElement).value)"
              class="input"
              placeholder="m5g..."
            />
            <button
              class="btn btn-secondary"
              @click="store.regenerateClientId(); saveSetting('defaultClientId', store.settings.defaultClientId)"
              title="Generate new ID"
            >
              <span class="mdi mdi-refresh"></span>
            </button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Keepalive (seconds)</label>
            <p class="setting-description">Keepalive interval in seconds</p>
          </div>
          <div class="setting-control">
            <input
              type="number"
              :value="localSettings.keepalive"
              @change="saveSetting('keepalive', parseInt(($event.target as HTMLInputElement).value))"
              class="input number-input"
              min="0"
              max="65535"
            />
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Reconnect Period (seconds)</label>
            <p class="setting-description">Time between reconnection attempts</p>
          </div>
          <div class="setting-control">
            <input
              type="number"
              :value="localSettings.reconnectPeriod"
              @change="saveSetting('reconnectPeriod', parseInt(($event.target as HTMLInputElement).value))"
              class="input number-input"
              min="0"
              max="60"
            />
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Max Reconnects</label>
            <p class="setting-description">Maximum reconnection attempts (0 = unlimited)</p>
          </div>
          <div class="setting-control">
            <input
              type="number"
              :value="localSettings.maxReconnects"
              @change="saveSetting('maxReconnects', parseInt(($event.target as HTMLInputElement).value))"
              class="input number-input"
              min="0"
              max="100"
            />
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Connection Timeout (seconds)</label>
            <p class="setting-description">Maximum time to wait for connection</p>
          </div>
          <div class="setting-control">
            <input
              type="number"
              :value="localSettings.connectionTimeout"
              @change="saveSetting('connectionTimeout', parseInt(($event.target as HTMLInputElement).value))"
              class="input number-input"
              min="1"
              max="300"
            />
          </div>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="section-title">Data Management</h2>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Export Connections</label>
            <p class="setting-description">Export all connections to a JSON file</p>
          </div>
          <div class="setting-control">
            <button class="btn btn-secondary" @click="exportAllConnections">
              <span class="mdi mdi-export"></span>
              Export
            </button>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Import Connections</label>
            <p class="setting-description">Import connections from a JSON file</p>
          </div>
          <div class="setting-control">
            <label class="btn btn-secondary import-btn">
              <span class="mdi mdi-import"></span>
              Import
              <input type="file" accept=".json" @change="importConnections" hidden />
            </label>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Import from Old Version</label>
            <p class="setting-description">Import connections from MQTT5 Explorer v1.x.x</p>
          </div>
          <div class="setting-control">
            <label class="btn btn-secondary import-btn">
              <span class="mdi mdi-database-import"></span>
              Import
              <input type="file" accept=".json" @change="importFromOldVersion" hidden />
            </label>
          </div>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="section-title">Feedback</h2>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Report a Bug</label>
            <p class="setting-description">Found an issue? Let us know on GitHub</p>
          </div>
          <div class="setting-control">
            <button class="btn btn-secondary" @click="openGitHub">
              <span class="mdi mdi-bug"></span>
              Report Bug
            </button>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.settings-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.page-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--color-border);
}

.page-header h1 {
  font-size: 20px;
  font-weight: 600;
}

.settings-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.settings-section {
  margin-bottom: 32px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-muted-foreground);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 16px;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  margin-bottom: 8px;
}

.setting-info {
  flex: 1;
}

.setting-label {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 4px;
  display: block;
}

.setting-description {
  font-size: 13px;
  color: var(--color-muted-foreground);
}

.setting-control {
  margin-left: 24px;
}

.number-input {
  width: 100px;
  text-align: center;
}

.client-id-control {
  display: flex;
  gap: 8px;
}

.client-id-control .input {
  width: 360px;
}

.color-presets {
  display: flex;
  gap: 8px;
}

.color-swatch {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  transition: transform 0.15s ease, border-color 0.15s ease;
}

.color-swatch:hover {
  transform: scale(1.1);
}

.color-swatch.active {
  border-color: var(--color-foreground);
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

input:checked + .slider {
  background-color: var(--color-primary);
}

input:checked + .slider:before {
  transform: translateX(20px);
}

.import-btn {
  cursor: pointer;
}
</style>
