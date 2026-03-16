<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '../stores/app'
import type { Message, TopicNode } from '../types'
import {
  SplitterGroup,
  SplitterPanel,
  SplitterResizeHandle
} from 'radix-vue'

const store = useAppStore()

type SearchMode = 'substring' | 'case-sensitive' | 'exact' | 'regex'

const topicSearchQuery = ref('')
const topicSearchMode = ref<SearchMode>('substring')
const valueSearchQuery = ref('')
const valueSearchMode = ref<SearchMode>('substring')
const expandedNodes = ref<Set<string>>(new Set())
const flashingTopics = ref<Set<string>>(new Set())
const sidebarWidth = ref(280)

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

const sendForm = ref({
  topic: '',
  payload: '',
  qos: 0,
  retain: false,
  contentType: '',
  userProperties: [] as { key: string; value: string }[],
  responseTopic: ''
})

const sendPanelRef = ref<InstanceType<typeof SplitterPanel> | null>(null)
const isSendPanelCollapsed = ref(true)

function toggleSendPanel() {
  if (sendPanelRef.value) {
    if (isSendPanelCollapsed.value) {
      sendPanelRef.value.resize(55)
    } else {
      sendPanelRef.value.resize(8)
    }
    isSendPanelCollapsed.value = !isSendPanelCollapsed.value
  }
}

function initSendFormFromLastMessage() {
  if (store.messages && store.messages.length > 0) {
    const lastMsg = store.messages[0]
    sendForm.value.topic = lastMsg.topic
    try {
      sendForm.value.payload = new TextDecoder().decode(new Uint8Array(lastMsg.payload))
    } catch {
      sendForm.value.payload = ''
    }
    sendForm.value.qos = lastMsg.qos
    sendForm.value.retain = lastMsg.retain
  }
}

const uniqueTopics = computed(() => {
  const topics: { topic: string; count: number; level: number }[] = []

  function collectTopics(node: TopicNode, level: number = 0) {
    if (node.fullTopic) {
      topics.push({ topic: node.fullTopic, count: node.messageCount || 0, level })
    }
    if (node.children) {
      Object.values(node.children).forEach(child => collectTopics(child, level + 1))
    }
  }

  if (store.topicTree && store.topicTree.children) {
    Object.values(store.topicTree.children).forEach(child => collectTopics(child, 0))
  }

  return topics.sort((a, b) => a.topic.localeCompare(b.topic))
})

const filteredTopics = computed(() => {
  let topics = uniqueTopics.value

  if (topicSearchQuery.value) {
    const query = topicSearchMode.value === 'substring' || topicSearchMode.value === 'case-sensitive'
      ? topicSearchQuery.value.toLowerCase()
      : topicSearchQuery.value

    topics = topics.filter(item => {
      const topic = item.topic
      const searchIn = topicSearchMode.value === 'case-sensitive' ? topic : topic.toLowerCase()

      if (topicSearchMode.value === 'exact') {
        return searchIn === query
      } else if (topicSearchMode.value === 'regex') {
        try {
          return new RegExp(topicSearchQuery.value).test(topic)
        } catch {
          return false
        }
      }
      return searchIn.includes(query)
    })
  }

  return topics
})

const filteredMessages = computed(() => {
  if (!store.messages) return []

  if (!valueSearchQuery.value) return store.messages

  const query = valueSearchMode.value === 'substring' || valueSearchMode.value === 'case-sensitive'
    ? valueSearchQuery.value.toLowerCase()
    : valueSearchQuery.value

  return store.messages.filter(msg => {
    let payload = ''
    try {
      payload = new TextDecoder().decode(new Uint8Array(msg.payload))
    } catch {
      payload = ''
    }

    const searchIn = valueSearchMode.value === 'case-sensitive' ? payload : payload.toLowerCase()

    if (valueSearchMode.value === 'exact') {
      return searchIn === query
    } else if (valueSearchMode.value === 'regex') {
      try {
        return new RegExp(valueSearchQuery.value).test(payload)
      } catch {
        return false
      }
    }
    return searchIn.includes(query)
  })
})

function selectTopic(topic: string) {
  store.selectedTopic = topic
  store.loadMessages(topic)
  valueSearchQuery.value = ''
}

function closeTopic() {
  store.selectedTopic = null
  valueSearchQuery.value = ''
}

function deleteTopic() {
  if (store.selectedTopic && store.currentConnectionId) {
    window.go.main.App.Unsubscribe(store.currentConnectionId, store.selectedTopic)
    store.selectedTopic = null
    valueSearchQuery.value = ''
  }
}

function toggleTopicSearchMode() {
  const modes: SearchMode[] = ['substring', 'case-sensitive', 'exact', 'regex']
  const currentIndex = modes.indexOf(topicSearchMode.value)
  topicSearchMode.value = modes[(currentIndex + 1) % modes.length]
}

function toggleValueSearchMode() {
  const modes: SearchMode[] = ['substring', 'case-sensitive', 'exact', 'regex']
  const currentIndex = modes.indexOf(valueSearchMode.value)
  valueSearchMode.value = modes[(currentIndex + 1) % modes.length]
}

function getSearchModeIcon(mode: SearchMode): string {
  switch (mode) {
    case 'substring': return 'mdi-magnify'
    case 'case-sensitive': return 'mdi-format-text'
    case 'exact': return 'mdi-equal'
    case 'regex': return 'mdi-regex'
    default: return 'mdi-magnify'
  }
}

function getSearchModeTooltip(mode: SearchMode): string {
  switch (mode) {
    case 'substring': return 'Substring match (Aa for case sensitive)'
    case 'case-sensitive': return 'Case sensitive'
    case 'exact': return 'Exact match'
    case 'regex': return 'Regex'
    default: return ''
  }
}

async function sendMessage() {
  if (!store.currentConnectionId) return
  try {
    await store.sendMessage({
      connectionId: store.currentConnectionId,
      topic: sendForm.value.topic,
      payload: sendForm.value.payload,
      qos: sendForm.value.qos,
      retain: sendForm.value.retain,
      contentType: sendForm.value.contentType || undefined,
      userProperties: sendForm.value.userProperties.reduce((acc, prop) => {
        if (prop.key) acc[prop.key] = prop.value
        return acc
      }, {} as Record<string, string>),
      responseTopic: sendForm.value.responseTopic || undefined
    })
  } catch (error) {
    console.error('Failed to send message:', error)
  }
}

function addUserProperty() {
  sendForm.value.userProperties.push({ key: '', value: '' })
}

function removeUserProperty(index: number) {
  sendForm.value.userProperties.splice(index, 1)
}

function toggleNode(topic: string) {
  if (expandedNodes.value.has(topic)) {
    expandedNodes.value.delete(topic)
  } else {
    expandedNodes.value.add(topic)
  }
}

function isExpanded(topic: string): boolean {
  return expandedNodes.value.has(topic)
}

function getLastSegment(topic: string): string {
  const segments = topic.split('/')
  return segments[segments.length - 1]
}

function hasChildren(topic: string): boolean {
  const topicParts = topic.split('/')
  return uniqueTopics.value.some(t => {
    const tParts = t.topic.split('/')
    return tParts.length > topicParts.length &&
           t.topic.startsWith(topic + '/')
  })
}

function getTopicIcon(topic: string, isBranch: boolean, isSelected: boolean): string {
  if (isBranch) {
    if (isSelected || isExpanded(topic)) {
      return 'mdi-folder-open'
    }
    return 'mdi-folder-outline'
  }
  if (isSelected) {
    return 'mdi-file'
  }
  return 'mdi-file-outline'
}

function isVisible(topic: string, level: number): boolean {
  if (level === 0) return true
  const parents = topic.split('/').slice(0, -1)
  let current = ''
  for (const p of parents) {
    current = current ? current + '/' + p : p
    if (!expandedNodes.value.has(current)) {
      return false
    }
  }
  return true
}

function showMessageDetail(msg: Message) {
  store.selectedMessage = msg
}

function formatPayload(payload: number[] | Uint8Array): string {
  try {
    const text = new TextDecoder().decode(new Uint8Array(payload))
    const parsed = JSON.parse(text)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return new TextDecoder().decode(new Uint8Array(payload))
  }
}

function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleTimeString()
}

let messageInterval: number
let previousMessageCount = 0

function countAllMessages(node: TopicNode): number {
  let count = node.messageCount || 0
  if (node.children) {
    Object.values(node.children).forEach(child => {
      count += countAllMessages(child)
    })
  }
  return count
}

const totalMessageCount = computed(() => {
  if (!store.topicTree) return 0
  return countAllMessages(store.topicTree)
})

function triggerFlash(topic: string) {
  flashingTopics.value.add(topic)
  const segments = topic.split('/')
  for (let i = 1; i < segments.length; i++) {
    const parentTopic = segments.slice(0, i).join('/')
    flashingTopics.value.add(parentTopic)
  }
  setTimeout(() => {
    flashingTopics.value.delete(topic)
    segments.forEach((_, i) => {
      if (i > 0) {
        flashingTopics.value.delete(segments.slice(0, i).join('/'))
      }
    })
  }, 500)
}

function handleNewMessage() {
  const newCount = totalMessageCount.value
  if (newCount > previousMessageCount && previousMessageCount > 0) {
    if (store.topicTree && store.topicTree.lastMessage) {
      const topic = store.topicTree.lastMessage.topic
      triggerFlash(topic)
    }
  }
  previousMessageCount = newCount
}

onMounted(() => {
  if (store.isConnected) {
    store.loadTopicTree()
  }

  messageInterval = window.setInterval(() => {
    if (store.isConnected) {
      const prevCount = totalMessageCount.value
      store.loadTopicTree().then(() => {
        if (store.topicTree) {
          const newCount = countAllMessages(store.topicTree)
          if (newCount > prevCount && store.topicTree.lastMessage) {
            triggerFlash(store.topicTree.lastMessage.topic)
          }
        }
      })

      if (store.selectedTopic) {
        store.loadMessages(store.selectedTopic)
      }
    }
  }, 5000)
})

onUnmounted(() => {
  if (messageInterval) {
    clearInterval(messageInterval)
  }
})

watch(() => store.isConnected, (connected) => {
  if (connected) {
    store.loadTopicTree()
  }
})

watch(() => store.messages, () => {
  handleNewMessage()
}, { deep: true })
</script>

<template>
  <div class="messages-page">
    <aside class="topics-sidebar" :style="{ width: sidebarWidth + 'px', minWidth: sidebarWidth + 'px' }">
      <div class="sidebar-header">
        <h3>Topics</h3>
        <span class="topic-count">{{ totalMessageCount }}</span>
      </div>
      <div class="topics-search">
        <div class="search-field">
          <span class="mdi mdi-magnify"></span>
          <input
            v-model="topicSearchQuery"
            type="text"
            class="input"
            placeholder="Search topics..."
          />
          <button class="search-mode-btn" @click="toggleTopicSearchMode" :title="getSearchModeTooltip(topicSearchMode)">
            <span class="mdi" :class="getSearchModeIcon(topicSearchMode)"></span>
          </button>
        </div>
      </div>
      <div class="topics-tree">
        <template v-for="item in filteredTopics" :key="item.topic">
          <div
            v-if="isVisible(item.topic, item.level)"
            class="topic-row"
            :class="{
              active: store.selectedTopic === item.topic,
              flashing: flashingTopics.has(item.topic)
            }"
            :style="{ paddingLeft: (item.level * 16 + 8) + 'px' }"
          >
            <span
              v-if="hasChildren(item.topic)"
              class="expand-icon"
              :class="{ expanded: isExpanded(item.topic) }"
              @click.stop="toggleNode(item.topic)"
            >
              <span class="mdi" :class="isExpanded(item.topic) ? 'mdi-chevron-down' : 'mdi-chevron-right'"></span>
            </span>
            <span
              class="mdi topic-icon"
              :class="getTopicIcon(item.topic, hasChildren(item.topic), store.selectedTopic === item.topic)"
            ></span>
            <span class="topic-name truncate" @click="selectTopic(item.topic)">{{ getLastSegment(item.topic) }}</span>
            <span class="badge badge-default">{{ item.count }}</span>
          </div>
        </template>
        <div v-if="filteredTopics.length === 0" class="empty-topics">
          <span class="text-muted text-sm">No topics found</span>
        </div>
      </div>
    </aside>

    <div class="resize-handle" @mousedown="startResize"></div>

    <main class="messages-main">
      <SplitterGroup direction="vertical" class="messages-splitter-group">
        <SplitterPanel :minSize="20" :defaultSize="100" class="messages-panel">
          <div class="messages-content">
            <header class="messages-header">
              <div class="header-top">
                <div class="header-left">
                  <h2 class="truncate">{{ store.selectedTopic || 'Select a topic' }}</h2>
                  <span v-if="store.selectedTopic" class="text-muted text-sm">
                    {{ store.messages?.length || 0 }} messages
                  </span>
                </div>
                <button v-if="store.selectedTopic" class="btn btn-ghost btn-icon" @click="closeTopic" title="Close">
                  <span class="mdi mdi-close"></span>
                </button>
              </div>
              <div v-if="store.selectedTopic" class="header-filters">
                <div class="search-field">
                  <span class="mdi mdi-magnify"></span>
                  <input
                    v-model="valueSearchQuery"
                    type="text"
                    class="input"
                    placeholder="Search in payload..."
                  />
                  <button class="search-mode-btn" @click="toggleValueSearchMode" :title="getSearchModeTooltip(valueSearchMode)">
                    <span class="mdi" :class="getSearchModeIcon(valueSearchMode)"></span>
                  </button>
                </div>
                <button class="btn btn-secondary" @click="deleteTopic">
                  <span class="mdi mdi-delete"></span>
                  Delete
                </button>
              </div>
            </header>

            <div class="messages-list">
              <template v-if="store.selectedTopic">
                <div
                  v-for="(msg, index) in filteredMessages"
                  :key="index"
                  class="message-item"
                  @click="showMessageDetail(msg)"
                >
                  <div class="message-header">
                    <span class="message-topic truncate">{{ msg.topic }}</span>
                    <span class="message-time text-muted text-sm">{{ formatTime(msg.timestamp) }}</span>
                  </div>
                  <div class="message-preview mono truncate">{{ formatPayload(msg.payload).slice(0, 100) }}</div>
                </div>
                <div v-if="!filteredMessages.length" class="empty-messages">
                  <span class="text-muted">No messages found</span>
                </div>
              </template>
              <div v-else class="empty-messages select-topic">
                <span class="mdi mdi-message-text-outline icon"></span>
                <h3>Select a topic</h3>
                <p class="text-muted">Choose a topic from the sidebar to view messages</p>
              </div>
            </div>
          </div>
        </SplitterPanel>

        <SplitterResizeHandle class="splitter-handle-v" :disabled="!store.isConnected" />

        <SplitterPanel ref="sendPanelRef" :minSize="8" class="send-panel-container">
          <div class="send-panel" :class="{ disabled: !store.isConnected }">
            <div class="send-panel-header" @click="store.isConnected && toggleSendPanel()">
              <h3>Send Message</h3>
              <button class="btn btn-ghost btn-icon" :disabled="!store.isConnected">
                <span class="mdi" :class="isSendPanelCollapsed ? 'mdi-chevron-up' : 'mdi-chevron-down'"></span>
              </button>
            </div>
            <div class="send-panel-content">
              <div class="form-group">
                <label class="label">Topic</label>
                <input v-model="sendForm.topic" type="text" class="input w-full" required />
              </div>
              <div class="form-group">
                <label class="label">Payload</label>
                <textarea v-model="sendForm.payload" class="input textarea w-full" rows="3" required></textarea>
              </div>
              <div class="form-row">
                <div class="form-group">
                  <label class="label">QoS</label>
                  <select v-model.number="sendForm.qos" class="input w-full">
                    <option :value="0">0 - At most once</option>
                    <option :value="1">1 - At least once</option>
                    <option :value="2">2 - Exactly once</option>
                  </select>
                </div>
                <div class="form-group">
                  <label class="checkbox-label">
                    <input v-model="sendForm.retain" type="checkbox" />
                    Retain
                  </label>
                </div>
              </div>
              <button class="btn btn-primary w-full" @click="sendMessage" :disabled="!store.isConnected">
                <span class="mdi mdi-send"></span>
                Send
              </button>
            </div>
          </div>
        </SplitterPanel>
      </SplitterGroup>
    </main>
  </div>
</template>

<style scoped>
.messages-page {
  height: 100%;
  display: flex;
  overflow: hidden;
}

.topics-sidebar {
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
  border-bottom: 1px solid var(--color-border);
}

.sidebar-header h3 {
  font-size: 14px;
  font-weight: 600;
}

.topic-count {
  background: var(--color-primary);
  color: white;
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
}

.topics-tree {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.topic-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  cursor: pointer;
  transition: background 0.15s ease;
  border-radius: 4px;
  margin: 1px 4px;
}

.topic-row:hover {
  background: var(--color-muted);
}

.topic-row.active {
  background: var(--color-primary);
  color: white;
}

.topic-row.active .badge {
  background: rgba(255,255,255,0.2);
  color: white;
}

.topic-row.flashing {
  animation: flash 0.5s ease;
}

@keyframes flash {
  0%, 100% { background: transparent; }
  50% { background: var(--color-primary); }
}

.topic-row.active.flashing {
  animation: flash-active 0.5s ease;
}

@keyframes flash-active {
  0%, 100% { background: var(--color-primary); }
  50% { background: var(--color-accent); }
}

.expand-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}

.expand-icon .mdi {
  font-size: 16px;
}

.topic-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.topic-name {
  flex: 1;
  font-size: 13px;
}

.badge {
  font-size: 11px;
  padding: 1px 6px;
  background: var(--color-muted);
  border-radius: 8px;
  color: var(--color-muted-foreground);
  flex-shrink: 0;
}

.empty-topics {
  padding: 24px;
  text-align: center;
}

.messages-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.messages-splitter-group {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.messages-panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.messages-content {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.splitter-handle-v {
  height: 8px;
  background: var(--color-border);
  cursor: row-resize;
  flex-shrink: 0;
  transition: background 0.15s ease;
}

.splitter-handle-v:hover:not([disabled]),
.splitter-handle-v[data-resize-handle-active] {
  background: var(--color-primary);
}

.splitter-handle-v[disabled] {
  cursor: not-allowed;
  opacity: 0.5;
}

.send-panel-container {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--color-secondary);
}

.send-panel-collapsed {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--color-primary);
  cursor: pointer;
  color: white;
  font-size: 13px;
  font-weight: 500;
  transition: background 0.15s ease;
}

.send-panel-collapsed:hover:not(.disabled) {
  background: var(--color-accent);
}

.send-panel-collapsed.disabled {
  background: var(--color-muted);
  color: var(--color-muted-foreground);
  cursor: not-allowed;
}

.send-panel.disabled {
  opacity: 0.5;
  pointer-events: none;
}

.send-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.send-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-top: 1px solid var(--color-border);
  background: var(--color-card);
}

.send-panel-header h3 {
  font-size: 14px;
  font-weight: 600;
}

.send-panel-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.messages-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-background);
}

.header-left h2 {
  font-size: 18px;
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
  width: 200px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  cursor: pointer;
}

.messages-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.message-item {
  padding: 12px;
  border: 1px solid var(--color-border);
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: border-color 0.15s ease;
}

.message-item:hover {
  border-color: var(--color-primary);
}

.message-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.message-topic {
  font-weight: 500;
  font-size: 13px;
}

.message-time {
  font-size: 12px;
}

.message-preview {
  font-size: 12px;
  color: var(--color-muted-foreground);
}

.empty-messages {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 48px;
}

.empty-messages .icon {
  font-size: 48px;
  color: var(--color-muted-foreground);
  margin-bottom: 16px;
}

.empty-messages h3 {
  font-size: 18px;
  margin-bottom: 8px;
}

.select-topic {
  color: var(--color-muted-foreground);
}

.dialog-content {
  background: var(--color-card);
  border-radius: 8px;
  padding: 24px;
  width: 600px;
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

.message-detail-dialog {
  width: 700px;
}

.dialog-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
}

.send-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.label {
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 6px;
  color: var(--color-foreground);
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.form-row .form-group {
  flex: 1;
}

.textarea {
  resize: vertical;
  min-height: 100px;
  padding: 12px;
  font-family: inherit;
}

.checkbox-label {
  padding-bottom: 8px;
}

.divider {
  height: 1px;
  background: var(--color-border);
  margin: 8px 0;
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

.message-detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-row {
  display: flex;
  gap: 12px;
}

.detail-label {
  width: 100px;
  flex-shrink: 0;
  font-weight: 500;
  color: var(--color-muted-foreground);
}

.detail-value {
  flex: 1;
  word-break: break-all;
}

.detail-payload {
  flex: 1;
  background: var(--color-muted);
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
  font-size: 12px;
}
</style>
