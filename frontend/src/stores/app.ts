import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { Connection, ConnectionStatus, Message, Settings, TopicNode, SendMessageRequest } from '../types'

declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetSettings: () => Promise<Record<string, string>>
          SetSetting: (key: string, value: string) => Promise<void>
          GetConnections: () => Promise<Connection[]>
          GetConnection: (id: number) => Promise<Connection>
          CreateConnection: (conn: Omit<Connection, 'id' | 'createdAt' | 'updatedAt'>) => Promise<number>
          UpdateConnection: (conn: Connection) => Promise<void>
          DeleteConnection: (id: number) => Promise<void>
          SearchConnections: (query: string) => Promise<Connection[]>
          ExportConnections: (ids: number[]) => Promise<string>
          ExportConnectionsToFile: (ids: number[]) => Promise<string>
          ImportConnections: (jsonData: string) => Promise<number[]>
          ImportFromOldVersion: (jsonData: string) => Promise<number>
          Connect: (id: number) => Promise<void>
          Disconnect: (id: number) => Promise<void>
          GetConnectionStatus: (id: number) => Promise<[boolean, string]>
          GetClientID: (id: number) => Promise<string>
          GetMessages: (connectionId: number, topic: string, limit: number) => Promise<Message[]>
          SearchMessages: (connectionId: number, pattern: string, useRegex: boolean) => Promise<Message[]>
          ClearMessages: (connectionId: number) => Promise<void>
          SendMessage: (req: SendMessageRequest) => Promise<void>
          Subscribe: (connectionId: number, topic: string, qos: number) => Promise<void>
          Unsubscribe: (connectionId: number, topic: string) => Promise<void>
          GetTopicTree: (connectionId: number) => Promise<TopicNode>
          GetNumericMessages: (connectionId: number, topic: string, limit: number) => Promise<Message[]>
        }
      }
    }
  }
}

export const useAppStore = defineStore('app', () => {
  const router = useRouter()
  
  const settings = ref<Settings>({
    theme: 'light',
    accentColor: '#007AFF',
    closeToTray: true,
    maxCachedMessages: 7,
    defaultClientId: '',
    keepalive: 120,
    reconnectPeriod: 2,
    maxReconnects: 2,
    connectionTimeout: 20
  })

  const connections = ref<Connection[]>([])
  const currentConnectionId = ref<number | null>(null)
  const connectionStatuses = ref<Record<number, ConnectionStatus>>({})
  const messages = ref<Message[]>([])
  const topicTree = ref<TopicNode | null>(null)
  const selectedMessage = ref<Message | null>(null)
  const selectedTopic = ref<string | null>(null)
  const currentClientId = ref<string>('')
  const toast = ref<{ message: string; type: 'success' | 'error' } | null>(null)

  function showToast(message: string, type: 'success' | 'error' = 'success') {
    toast.value = { message, type }
    setTimeout(() => {
      toast.value = null
    }, 3000)
  }

  EventsOn('mqtt-message', (data: string) => {
    try {
      const msg: Message = JSON.parse(data)
      if (msg.connectionId === currentConnectionId.value) {
        if (!selectedTopic.value || msg.topic === selectedTopic.value) {
          messages.value.unshift(msg)
          const maxMessages = parseInt(String(settings.value.maxCachedMessages)) || 100
          if (messages.value.length > maxMessages) {
            messages.value = messages.value.slice(0, maxMessages)
          }
        }
      }
    } catch (e) {
      console.error('Failed to parse message:', e)
    }
  })

  EventsOn('connection-status', (data: { connectionId: number; status: string }) => {
    if (data.connectionId === currentConnectionId.value) {
      connectionStatuses.value[data.connectionId] = {
        connectionId: data.connectionId,
        status: data.status as any
      }
      if (data.status === 'connected') {
        loadClientId()
      } else if (data.status === 'disconnected') {
        currentClientId.value = ''
      }
    }
  })

  const currentConnection = computed(() => {
    if (!connections.value || !currentConnectionId.value) return null
    return connections.value.find(c => c.id === currentConnectionId.value) || null
  })

  const isConnected = computed(() => {
    if (!currentConnectionId.value) return false
    return connectionStatuses.value[currentConnectionId.value]?.status === 'connected'
  })

  const hasActiveConnection = computed(() => {
    return Object.values(connectionStatuses.value).some(
      status => status.status === 'connected' || status.status === 'connecting'
    )
  })

  async function loadSettings() {
    try {
      const data = await window.go.main.App.GetSettings()
      settings.value = {
        theme: (data.theme as 'light' | 'dark') || 'light',
        accentColor: data.accentColor || '#007AFF',
        closeToTray: data.closeToTray === 'true',
        maxCachedMessages: parseInt(data.maxCachedMessages) || 7,
        defaultClientId: data.defaultClientId || '',
        keepalive: parseInt(data.keepalive) || 120,
        reconnectPeriod: parseInt(data.reconnectPeriod) || 2,
        maxReconnects: parseInt(data.maxReconnects) || 2,
        connectionTimeout: parseInt(data.connectionTimeout) || 20
      }
      if (!settings.value.defaultClientId) {
        settings.value.defaultClientId = generateClientId()
      }
      applyTheme()
    } catch (error) {
      console.error('Failed to load settings:', error)
    }
  }

  async function saveSetting(key: keyof Settings, value: any) {
    try {
      await window.go.main.App.SetSetting(key, String(value))
      settings.value[key] = value as never
      if (key === 'theme' || key === 'accentColor') {
        applyTheme()
      }
    } catch (error) {
      console.error('Failed to save setting:', error)
    }
  }

  function applyTheme() {
    const root = document.documentElement
    if (settings.value.theme === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
    root.style.setProperty('--color-primary', settings.value.accentColor)
    root.style.setProperty('--color-accent', settings.value.accentColor)
    root.style.setProperty('--color-ring', settings.value.accentColor)
  }

  async function loadConnections() {
    try {
      connections.value = await window.go.main.App.GetConnections()
    } catch (error) {
      console.error('Failed to load connections:', error)
    }
  }

  async function searchConnections(query: string) {
    try {
      connections.value = await window.go.main.App.SearchConnections(query)
    } catch (error) {
      console.error('Failed to search connections:', error)
    }
  }

  async function createConnection(conn: Omit<Connection, 'id' | 'createdAt' | 'updatedAt'>) {
    try {
      const id = await window.go.main.App.CreateConnection(conn)
      await loadConnections()
      return id
    } catch (error) {
      console.error('Failed to create connection:', error)
      throw error
    }
  }

  async function updateConnection(conn: Connection) {
    try {
      await window.go.main.App.UpdateConnection(conn)
      await loadConnections()
    } catch (error) {
      console.error('Failed to update connection:', error)
      throw error
    }
  }

  async function deleteConnection(id: number) {
    try {
      await window.go.main.App.DeleteConnection(id)
      if (currentConnectionId.value === id) {
        currentConnectionId.value = null
      }
      await loadConnections()
    } catch (error) {
      console.error('Failed to delete connection:', error)
      throw error
    }
  }

  async function connect(id: number) {
    try {
      connectionStatuses.value[id] = { connectionId: id, status: 'connecting' }
      await window.go.main.App.Connect(id)
      connectionStatuses.value[id] = { connectionId: id, status: 'connected' }
      currentConnectionId.value = id
      selectedTopic.value = null
      messages.value = []
      await loadClientId()
      await loadTopicTree()
    } catch (error: any) {
      connectionStatuses.value[id] = { 
        connectionId: id, 
        status: 'error', 
        error: error.message || 'Connection failed' 
      }
      throw error
    }
  }

  async function loadClientId() {
    if (!currentConnectionId.value) return
    try {
      currentClientId.value = await window.go.main.App.GetClientID(currentConnectionId.value)
    } catch (e) {
      currentClientId.value = ''
    }
  }

  async function disconnect(id: number) {
    try {
      await window.go.main.App.Disconnect(id)
      connectionStatuses.value[id] = { connectionId: id, status: 'disconnected' }
      if (currentConnectionId.value === id) {
        currentConnectionId.value = null
        messages.value = []
        topicTree.value = null
        currentClientId.value = ''
        router.push('/')
      }
    } catch (error) {
      console.error('Failed to disconnect:', error)
    }
  }

  async function checkConnectionStatus(id: number) {
    try {
      const [connected, status] = await window.go.main.App.GetConnectionStatus(id)
      connectionStatuses.value[id] = { 
        connectionId: id, 
        status: connected ? 'connected' : (status as any) || 'disconnected' 
      }
    } catch (error) {
      console.error('Failed to check connection status:', error)
    }
  }

  async function loadMessages(topic?: string) {
    if (!currentConnectionId.value) return
    try {
      const msgs = await window.go.main.App.GetMessages(
        currentConnectionId.value,
        topic || '',
        100
      )
      console.log('Loaded messages:', msgs)
      messages.value = msgs
    } catch (error) {
      console.error('Failed to load messages:', error)
    }
  }

  async function searchMessages(pattern: string, useRegex: boolean) {
    if (!currentConnectionId.value) return
    try {
      messages.value = await window.go.main.App.SearchMessages(
        currentConnectionId.value,
        pattern,
        useRegex
      )
    } catch (error) {
      console.error('Failed to search messages:', error)
    }
  }

  async function clearMessages() {
    if (!currentConnectionId.value) return
    try {
      await window.go.main.App.ClearMessages(currentConnectionId.value)
      messages.value = []
      topicTree.value = null
    } catch (error) {
      console.error('Failed to clear messages:', error)
    }
  }

  async function loadTopicTree() {
    if (!currentConnectionId.value) return
    try {
      topicTree.value = await window.go.main.App.GetTopicTree(currentConnectionId.value)
    } catch (error) {
      console.error('Failed to load topic tree:', error)
    }
  }

  async function sendMessage(req: SendMessageRequest) {
    try {
      await window.go.main.App.SendMessage(req)
    } catch (error) {
      console.error('Failed to send message:', error)
      throw error
    }
  }

  async function subscribe(topic: string, qos: number = 0) {
    if (!currentConnectionId.value) return
    try {
      await window.go.main.App.Subscribe(currentConnectionId.value, topic, qos)
      selectedTopic.value = topic
      await loadMessages(topic)
    } catch (error) {
      console.error('Failed to subscribe:', error)
      throw error
    }
  }

  async function unsubscribe(topic: string) {
    if (!currentConnectionId.value) return
    try {
      await window.go.main.App.Unsubscribe(currentConnectionId.value, topic)
      selectedTopic.value = null
      messages.value = []
    } catch (error) {
      console.error('Failed to unsubscribe:', error)
      throw error
    }
  }

  async function exportConnections(ids: number[]) {
    try {
      return await window.go.main.App.ExportConnections(ids)
    } catch (error) {
      console.error('Failed to export connections:', error)
      throw error
    }
  }

  async function exportConnectionsToFile(ids: number[]) {
    try {
      return await window.go.main.App.ExportConnectionsToFile(ids)
    } catch (error) {
      console.error('Failed to export connections:', error)
      throw error
    }
  }

  async function importConnections(jsonData: string) {
    try {
      await window.go.main.App.ImportConnections(jsonData)
      await loadConnections()
    } catch (error) {
      console.error('Failed to import connections:', error)
      throw error
    }
  }

  async function importFromOldVersion(jsonData: string): Promise<number> {
    try {
      const imported = await window.go.main.App.ImportFromOldVersion(jsonData)
      return imported as number
    } catch (error) {
      console.error('Failed to import connections from old version:', error)
      throw error
    }
  }

  function generateClientId(): string {
    return 'm5g-' + crypto.randomUUID()
  }

  function regenerateClientId() {
    settings.value.defaultClientId = generateClientId()
  }

  return {
    settings,
    connections,
    currentConnectionId,
    connectionStatuses,
    messages,
    topicTree,
    selectedMessage,
    selectedTopic,
    currentClientId,
    currentConnection,
    isConnected,
    hasActiveConnection,
    loadSettings,
    saveSetting,
    regenerateClientId,
    loadConnections,
    searchConnections,
    createConnection,
    updateConnection,
    deleteConnection,
    connect,
    disconnect,
    checkConnectionStatus,
    loadMessages,
    searchMessages,
    clearMessages,
    loadTopicTree,
    sendMessage,
    subscribe,
    unsubscribe,
    exportConnections,
    exportConnectionsToFile,
    importConnections,
    importFromOldVersion,
    toast,
    showToast
  }
})
