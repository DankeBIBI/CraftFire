<script setup lang="ts">
/**
 * RoomConnect - 统一联机入口。
 * 同时支持手动房间号加入与局域网房间发现。
 */
import { onMounted, onUnmounted, ref } from 'vue'
import { useRoomStore } from '@/stores/room'
import { useGameStateStore } from '@/stores/gameState'
import { usePlayerStore } from '@/stores/player'
import { useUIStore } from '@/stores/ui'

const emit = defineEmits<{
  (e: 'back'): void
}>()

const roomStore = useRoomStore()
const gameState = useGameStateStore()
const playerStore = usePlayerStore()
const ui = useUIStore()

const playerName = ref(playerStore.playerName || 'Player')
const roomId = ref('')
const targetIP = ref('127.0.0.1')
const isSubmitting = ref(false)
const autoRefreshEnabled = ref(true)

let refreshInterval: ReturnType<typeof setInterval> | null = null

function onRoomIdInput(event: Event) {
  const input = event.target as HTMLInputElement
  roomId.value = input.value.replace(/\D/g, '').slice(0, 6)
}

async function refresh() {
  await roomStore.searchLANServers()
}

function startAutoRefresh() {
  stopAutoRefresh()
  refreshInterval = setInterval(() => {
    if (autoRefreshEnabled.value) {
      refresh()
    }
  }, 5000)
}

function stopAutoRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

async function joinRoom(roomIdStr: string, ip: string) {
  if (!playerName.value.trim()) {
    ui.showToast('请输入玩家名称', 'warning')
    return
  }

  isSubmitting.value = true
  try {
    const success = await roomStore.joinRoom(roomIdStr, ip)
    if (!success) {
      return
    }

    const id = crypto.randomUUID?.() ?? `player-${Math.random().toString(36).slice(2)}-${Date.now()}`
    playerStore.initLocalPlayer(id, playerName.value.trim())
    gameState.startGame('sandbox')

    setTimeout(() => {
      gameState.setLoadingProgress(100, '加载完成！')
      gameState.enterGame()
    }, 1000)

    ui.showToast(`已加入房间 ${roomIdStr}`, 'success')
  } finally {
    isSubmitting.value = false
  }
}

async function handleManualJoin() {
  if (roomId.value.length !== 6) {
    ui.showToast('请输入 6 位房间号', 'warning')
    return
  }
  const ip = targetIP.value.trim() || '127.0.0.1'
  await joinRoom(roomId.value, ip)
}

async function handleLANJoin(targetRoomId: string, ip: string) {
  await joinRoom(targetRoomId, ip)
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
  <div class="room-connect panel-lowpoly p-6 w-[620px] mx-auto">
    <h2 class="text-xl text-craft-primary font-game mb-5 text-center">加入房间</h2>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
      <div>
        <label class="block text-craft-text text-sm mb-1 font-game">玩家名称</label>
        <input
          v-model="playerName"
          type="text"
          maxlength="16"
          class="input-lowpoly w-full"
          placeholder="输入你的名字"
        />
      </div>

      <div>
        <label class="block text-craft-text text-sm mb-1 font-game">房间号（6位数字）</label>
        <input
          :value="roomId"
          @input="onRoomIdInput"
          type="text"
          inputmode="numeric"
          maxlength="6"
          class="input-lowpoly w-full text-center text-xl tracking-[0.4em] font-game"
          placeholder="______"
        />
      </div>

      <div>
        <label class="block text-craft-text text-sm mb-1 font-game">服务器IP</label>
        <input
          v-model="targetIP"
          type="text"
          class="input-lowpoly w-full"
          placeholder="例如 192.168.1.10"
        />
      </div>
    </div>

    <div class="flex gap-3 mb-5">
      <button
        class="btn-primary flex-1"
        :disabled="isSubmitting || roomStore.isJoining || roomId.length !== 6"
        @click="handleManualJoin"
      >
        {{ isSubmitting ? '加入中...' : '通过房间号加入' }}
      </button>
      <button
        class="btn-lowpoly px-4"
        :disabled="roomStore.isSearching"
        @click="refresh"
      >
        {{ roomStore.isSearching ? '搜索中...' : '刷新局域网' }}
      </button>
    </div>

    <div class="flex items-center justify-between mb-2">
      <h3 class="text-craft-secondary text-xs font-game">局域网房间</h3>
      <label class="flex items-center gap-1 text-xs text-craft-text cursor-pointer">
        <input v-model="autoRefreshEnabled" type="checkbox" class="accent-craft-primary" />
        自动刷新
      </label>
    </div>

    <div class="space-y-2 max-h-72 overflow-y-auto rounded border border-white/10 p-2 bg-craft-dark/50">
      <div
        v-if="roomStore.lanServers.length === 0"
        class="text-center text-craft-text/60 py-8 text-sm font-game"
      >
        {{ roomStore.isSearching ? '正在搜索局域网内的房间...' : '未发现局域网房间' }}
      </div>

      <div
        v-for="server in roomStore.lanServers"
        :key="`${server.roomId}-${server.ip}`"
        class="flex items-center justify-between p-3 rounded border border-white/10 hover:border-craft-secondary/60 bg-craft-surface/70 transition-colors"
      >
        <div>
          <div class="text-craft-text font-game text-sm">房间 #{{ server.roomId }}</div>
          <div class="text-craft-text/60 text-xs mt-1">
            {{ server.ip }} · {{ server.gameMode }} · {{ server.playerCount }}/{{ server.maxPlayers }} 玩家
          </div>
        </div>
        <button
          class="btn-secondary text-xs px-4 py-1"
          :disabled="isSubmitting || server.playerCount >= server.maxPlayers"
          @click="handleLANJoin(server.roomId, server.ip)"
        >
          {{ server.playerCount >= server.maxPlayers ? '已满' : '加入' }}
        </button>
      </div>
    </div>

    <p v-if="roomStore.joinError" class="text-red-400 text-xs mt-4 font-game text-center">
      {{ roomStore.joinError }}
    </p>

    <div class="mt-5">
      <button class="btn-lowpoly w-full" @click="emit('back')">返回</button>
    </div>
  </div>
</template>
