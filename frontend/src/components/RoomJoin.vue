<script setup lang="ts">
/**
 * RoomJoin - 加入房间面板。
 * 输入 6 位房间号加入现有房间。
 */
import { ref } from 'vue'
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

const playerName = ref('Player')
const roomId = ref('')
const targetIP = ref('127.0.0.1')
const isSubmitting = ref(false)

/** 仅允许数字输入，最多 6 位 */
function onRoomIdInput(event: Event) {
  const input = event.target as HTMLInputElement
  roomId.value = input.value.replace(/\D/g, '').slice(0, 6)
}

async function handleJoin() {
  if (!playerName.value.trim()) {
    ui.showToast('请输入玩家名称', 'warning')
    return
  }
  if (roomId.value.length !== 6) {
    ui.showToast('请输入 6 位房间号', 'warning')
    return
  }

  isSubmitting.value = true
  try {
    const ip = targetIP.value.trim() || '127.0.0.1'
    const success = await roomStore.joinRoom(roomId.value, ip)
    if (success) {
      const id = crypto.randomUUID?.() ?? `player-${Date.now()}`
      playerStore.initLocalPlayer(id, playerName.value.trim())
      gameState.startGame('sandbox')

      setTimeout(() => {
        gameState.setLoadingProgress(100, '加载完成！')
        gameState.enterGame()
      }, 1500)

      ui.showToast(`已加入房间 ${roomId.value}`, 'success')
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="room-join panel-lowpoly p-6 w-96 mx-auto">
    <h2 class="text-xl text-craft-primary font-game mb-6 text-center">加入房间</h2>

    <!-- 玩家名称 -->
    <div class="mb-4">
      <label class="block text-craft-text text-sm mb-1 font-game">玩家名称</label>
      <input
        v-model="playerName"
        type="text"
        maxlength="16"
        class="input-lowpoly w-full"
        placeholder="输入你的名字"
      />
    </div>

    <!-- 房间号 -->
    <div class="mb-6">
      <label class="block text-craft-text text-sm mb-1 font-game">房间号（6位数字）</label>
      <input
        :value="roomId"
        @input="onRoomIdInput"
        type="text"
        inputmode="numeric"
        maxlength="6"
        class="input-lowpoly w-full text-center text-2xl tracking-[0.5em] font-game"
        placeholder="______"
      />
    </div>

    <!-- 服务器 IP -->
    <div class="mb-6">
      <label class="block text-craft-text text-sm mb-1 font-game">服务器IP</label>
      <input
        v-model="targetIP"
        type="text"
        class="input-lowpoly w-full"
        placeholder="例如 192.168.1.10"
      />
    </div>

    <!-- 错误提示 -->
    <p v-if="roomStore.joinError" class="text-red-400 text-xs mb-4 font-game">
      {{ roomStore.joinError }}
    </p>

    <!-- 按钮组 -->
    <div class="flex gap-3">
      <button class="btn-lowpoly flex-1" @click="emit('back')">返回</button>
      <button
        class="btn-primary flex-1"
        :disabled="isSubmitting || roomStore.isJoining || roomId.length !== 6"
        @click="handleJoin"
      >
        {{ isSubmitting ? '加入中...' : '加入房间' }}
      </button>
    </div>
  </div>
</template>
