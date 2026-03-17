<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '../stores/app'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  TimeScale
} from 'chart.js'
import type { Message } from '../types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  TimeScale
)

const store = useAppStore()
const route = useRoute()

const selectedChartTopic = ref('')
const chartData = ref<number[]>([])
const chartLabels = ref<string[]>([])
const numericMessages = ref<Message[]>([])
const topicSearchQuery = ref('')

const topicList = computed(() => {
  if (!store.topicTree) return []
  return flattenTree(store.topicTree)
})

const filteredTopicList = computed(() => {
  if (!topicSearchQuery.value) return topicList.value
  const query = topicSearchQuery.value.toLowerCase()
  return topicList.value.filter(t => t.topic.toLowerCase().includes(query))
})

const showSuggestions = ref(false)
const selectedIndex = ref(-1)

function hideSuggestions() {
  setTimeout(() => {
    showSuggestions.value = false
    selectedIndex.value = -1
  }, 200)
}

function selectFirstSuggestion() {
  if (filteredTopicList.value.length > 0) {
    selectTopic(filteredTopicList.value[0].topic)
    showSuggestions.value = false
  }
}

function moveSelection(delta: number) {
  const max = filteredTopicList.value.length - 1
  selectedIndex.value = Math.max(-1, Math.min(max, selectedIndex.value + delta))
}

function selectTopic(topic: string) {
  selectedChartTopic.value = topic
  topicSearchQuery.value = topic
  showSuggestions.value = false
  selectedIndex.value = -1
  loadChartData()
}

function flattenTree(node: any, result: { topic: string; count: number }[] = []) {
  if (node.fullTopic) {
    result.push({ topic: node.fullTopic, count: node.messageCount })
  }
  if (node.children) {
    Object.values(node.children).forEach((child: any) => {
      flattenTree(child, result)
    })
  }
  return result
}

const chartConfig = computed(() => ({
  labels: chartLabels.value,
  datasets: [
    {
      label: selectedChartTopic.value,
      data: chartData.value,
      borderColor: store.settings?.accentColor || '#007AFF',
      backgroundColor: (store.settings?.accentColor || '#007AFF') + '20',
      tension: 0.3,
      fill: true
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    }
  },
  scales: {
    x: {
      display: true,
      grid: {
        color: 'rgba(128, 128, 128, 0.1)'
      }
    },
    y: {
      display: true,
      grid: {
        color: 'rgba(128, 128, 128, 0.1)'
      }
    }
  }
}

function parseNumericValue(payload: number[] | Uint8Array): number | null {
  try {
    const text = new TextDecoder().decode(new Uint8Array(payload))
    const num = parseFloat(text)
    if (!isNaN(num)) return num
    
    const parsed = JSON.parse(text)
    if (typeof parsed === 'number') return parsed
    if (typeof parsed === 'object' && parsed !== null) {
      const values = Object.values(parsed).filter(v => typeof v === 'number')
      return values[0] as number || null
    }
    return null
  } catch {
    return null
  }
}

async function loadChartData() {
  if (!selectedChartTopic.value || !store.currentConnectionId) return
  
  try {
    numericMessages.value = await window.go.main.App.GetNumericMessages(
      store.currentConnectionId,
      selectedChartTopic.value,
      50
    )

    const data: number[] = []
    const labels: string[] = []

    numericMessages.value.forEach(msg => {
      const value = parseNumericValue(msg.payload)
      if (value !== null) {
        data.push(value)
        labels.push(new Date(msg.timestamp).toLocaleTimeString())
      }
    })

    chartData.value = data
    chartLabels.value = labels
  } catch (error) {
    console.error('Failed to load chart data:', error)
  }
}

watch(selectedChartTopic, () => {
  loadChartData()
})

onMounted(() => {
  if (route.query.topic && typeof route.query.topic === 'string') {
    selectedChartTopic.value = route.query.topic
    topicSearchQuery.value = route.query.topic
    if (store.isConnected) {
      store.loadTopicTree().then(() => {
        loadChartData()
      })
    }
  } else if (store.isConnected) {
    store.loadTopicTree()
  }
})

watch(() => store.isConnected, (connected) => {
  if (connected) {
    store.loadTopicTree().then(() => {
      if (topicList.value.length > 0) {
        selectedChartTopic.value = topicList.value[0].topic
      }
    })
  }
})
</script>

<template>
  <div class="charts-page">
    <header class="page-header">
      <div class="header-left">
        <h1>Charts</h1>
        <span class="text-muted text-sm">Visualize numeric MQTT messages</span>
      </div>
      <div class="header-right">
        <div class="topic-autocomplete" v-if="store.isConnected">
          <input 
            v-model="topicSearchQuery" 
            type="text" 
            class="input topic-input" 
            placeholder="Search topics..."
            @focus="showSuggestions = true"
            @blur="hideSuggestions"
            @keydown.enter="selectFirstSuggestion"
            @keydown.down.prevent="moveSelection(1)"
            @keydown.up.prevent="moveSelection(-1)"
          />
          <div v-if="showSuggestions && filteredTopicList.length > 0" class="suggestions-dropdown">
            <div 
              v-for="(item, index) in filteredTopicList" 
              :key="item.topic"
              class="suggestion-item"
              :class="{ active: selectedIndex === index }"
              @mousedown="selectTopic(item.topic)"
              @mouseenter="selectedIndex = index"
            >
              <span class="suggestion-topic">{{ item.topic }}</span>
              <span class="suggestion-count">{{ item.count }} msgs</span>
            </div>
          </div>
        </div>
        <button class="btn btn-secondary" @click="loadChartData" :disabled="!selectedChartTopic">
          <span class="mdi mdi-refresh"></span>
          Refresh
        </button>
      </div>
    </header>

    <div class="charts-content">
      <div v-if="!store.isConnected" class="empty-state">
        <span class="mdi mdi-connection-off icon"></span>
        <h3>Not Connected</h3>
        <p class="text-muted">Connect to an MQTT broker to view charts</p>
      </div>

      <div v-else-if="!selectedChartTopic" class="empty-state">
        <span class="mdi mdi-chart-line icon"></span>
        <h3>Select a Topic</h3>
        <p class="text-muted">Choose a topic to visualize its numeric data</p>
      </div>

      <div v-else class="chart-container">
        <div class="chart-card">
          <h3 class="chart-title">{{ selectedChartTopic }}</h3>
          <div class="chart-wrapper">
            <Line v-if="chartData.length > 0" :data="chartConfig" :options="chartOptions" />
            <div v-else class="no-data">
              <span class="mdi mdi-chart-box-outline icon"></span>
              <p class="text-muted">No numeric data found for this topic</p>
            </div>
          </div>
        </div>

        <div class="stats-card">
          <h3 class="card-title">Statistics</h3>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-label">Data Points</span>
              <span class="stat-value">{{ chartData.length }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Min</span>
              <span class="stat-value">{{ chartData.length > 0 ? Math.min(...chartData).toFixed(2) : '-' }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Max</span>
              <span class="stat-value">{{ chartData.length > 0 ? Math.max(...chartData).toFixed(2) : '-' }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Average</span>
              <span class="stat-value">
                {{ chartData.length > 0 ? (chartData.reduce((a, b) => a + b, 0) / chartData.length).toFixed(2) : '-' }}
              </span>
            </div>
            <div class="stat-item">
              <span class="stat-label">Latest</span>
              <span class="stat-value">{{ chartData.length > 0 ? chartData[chartData.length - 1].toFixed(2) : '-' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.charts-page {
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

.topic-autocomplete {
  position: relative;
}

.topic-input {
  width: 400px;
}

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  width: max-content;
  min-width: 100%;
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  max-height: 300px;
  overflow-y: auto;
  z-index: 100;
}

.suggestion-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.suggestion-item:hover,
.suggestion-item.active {
  background: var(--color-muted);
}

.suggestion-topic {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.suggestion-count {
  font-size: 12px;
  color: var(--color-muted-foreground);
  margin-left: 12px;
}

.charts-content {
  flex: 1;
  overflow: auto;
  padding: 24px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
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

.chart-container {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 24px;
  height: 100%;
}

.chart-card {
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.chart-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

.chart-wrapper {
  flex: 1;
  min-height: 300px;
  position: relative;
}

.no-data {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--color-muted-foreground);
}

.no-data .icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.stats-card {
  background: var(--color-card);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

.stats-grid {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 12px;
  background: var(--color-secondary);
  border-radius: 6px;
}

.stat-label {
  font-size: 13px;
  color: var(--color-muted-foreground);
}

.stat-value {
  font-size: 14px;
  font-weight: 600;
  font-family: monospace;
}
</style>
