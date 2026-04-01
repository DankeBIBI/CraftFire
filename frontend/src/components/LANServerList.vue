<script setup lang="ts">
/**
 * LANServerList - 局域网服务器发现列表。
 * 自动搜索局域网内的 CraftFire 房间并展示。
 */
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoomStore } from '@/stores/room'
import { useUIStore } from '@/stores/ui'

const emit = defineEmits<{
  (e: 'back'): void
  (e: 'join', roomId: string, ip: string): void
}>()

const roomStore = useRoomStore()
const ui = useUIStore()
const autoRefreshEnabled = ref(true)

let _refreshInterval: ReturnType<typeof setInterval> | null = null

async function refresh() {
  await roomStore.searchLANServers()
}

function startAutoRefresh() {
  stopAutoRefresh()
  _refreshInterval = setInterval(() => {
    if (autoRefreshEnabled.value) {
      refresh()
    }
  }, 5000) // 每 5 秒刷新
}

function stopAutoRefresh() {
  if (_refreshInterval) {
    clearInterval(_refreshInterval)
    _refreshInterval = null
  }
}

function handleJoin(roomId: string, ip: string) {
  emit('join', roomId, ip)
}

onMounted(() => {
  refresh()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<template>
  <div class="lan-server-list panel-lowpoly p-6 w-[560px] mx-auto">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl text-craft-primary font-game">局域网房间</h2>
      <div class="flex items-center gap-3">
        <label class="flex items-center gap-1 text-xs text-craft-text cursor-pointer">
          <input v-model="autoRefreshEnabled" type="checkbox" class="accent-craft-primary" />
          自动刷新
        </label>
        <button
          class="btn-lowpoly text-xs px-3 py-1"
          :disabled="roomStore.isSearching"
          @click="refresh"
        >
          {{ roomStore.isSearching ? '搜索中...' : '刷新' }}
        </button>
      </div>
    </div>

    <!-- 服务器列表 -->
    <div class="space-y-2 max-h-80 overflow-y-auto">
      <div
        v-if="roomStore.lanServers.length === 0"
        class="text-center text-craft-text/60 py-8 text-sm font-game"
      >
        {{ roomStore.isSearching ? '正在搜索局域网内的房间...' : '未发现局域网房间' }}
      </div>

      <div
        v-for="server in roomStore.lanServers"
        :key="server.roomId"
        class="flex items-center justify-between p-3 bg-craft-dark/50 rounded border border-white/10 hover:border-craft-primary/50 transition-colors"
      >
        <div>
          <div class="text-craft-text font-game text-sm">
            房间 #{{ server.roomId }}
          </div>
          <div class="text-craft-text/60 text-xs mt-1">
            {{ server.ip }} · {{ server.gameMode }} · {{ server.playerCount }}/{{ server.maxPlayers }} 玩家
          </div>
        </div>
        <button
          class="btn-primary text-xs px-4 py-1"
          :disabled="server.playerCount >= server.maxPlayers"
          @click="handleJoin(server.roomId, server.ip)"
        >
          {{ server.playerCount >= server.maxPlayers ? '已满' : '加入' }}
        </button>
      </div>
    </div>

    <!-- 返回 -->
    <div class="mt-4">
      <button class="btn-lowpoly w-full" @click="emit('back')">返回</button>
    </div>
  </div>
</template>
