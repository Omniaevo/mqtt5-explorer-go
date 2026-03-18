export interface Connection {
  id: number
  name: string
  mqttVersion: number
  protocol: 'mqtt' | 'mqtts' | 'ws' | 'wss'
  host: string
  port: number
  username?: string
  password?: string
  validateCert: boolean
  caFile?: string
  clientCert?: string
  clientKey?: string
  defaultSubscriptions?: string
  favourite: boolean
  lastConnected?: string
  createdAt: string
  updatedAt: string
}

export interface Message {
  id: number
  connectionId: number
  topic: string
  payload: number[]
  qos: number
  retain: boolean
  timestamp: string
  contentType?: string
  userProperties?: Record<string, string>
  responseTopic?: string
  correlationData?: number[]
  messageExpiry?: number
  topicAlias?: number
  clientId?: string
}

export interface Settings {
  theme: 'light' | 'dark'
  accentColor: string
  closeToTray: boolean
  maxCachedMessages: number
  defaultClientId: string
  keepalive: number
  reconnectPeriod: number
  maxReconnects: number
  connectionTimeout: number
}

export interface TopicNode {
  name: string
  fullTopic: string
  children: Record<string, TopicNode>
  messageCount: number
  lastMessage?: Message
}

export interface SendMessageRequest {
  connectionId: number
  topic: string
  payload: string
  qos: number
  retain: boolean
  contentType?: string
  userProperties?: Record<string, string>
  responseTopic?: string
  correlationData?: string
  messageExpiry?: number
}

export interface ConnectionStatus {
  connectionId: number
  status: 'connected' | 'disconnected' | 'connecting' | 'error'
  error?: string
}

export type NavItem = {
  name: string
  path: string
  icon: string
}
