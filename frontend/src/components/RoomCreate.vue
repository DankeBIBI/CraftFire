<script setup lang="ts">
/**
 * RoomCreate - 创建房间面板。
 * 房主设置玩家名、最大人数、游戏模式、地图信息后创建房间。
 */
import { ref, computed } from 'vue'
import { useRoomStore } from '@/stores/room'
import { useGameStateStore } from '@/stores/gameState'
import { usePlayerStore } from '@/stores/player'
import { useWorldStore } from '@/stores/world'
import { useUIStore } from '@/stores/ui'
import { getMapById, DEFAULT_MAP } from '@/maps/index'

const props = defineProps<{
  mapId?: string
}>()

const emit = defineEmits<{
  (e: 'back'): void
}>()

const roomStore = useRoomStore()
const gameState = useGameStateStore()
const playerStore = usePlayerStore()
const worldStore = useWorldStore()
const ui = useUIStore()

const playerName = ref('Player')
const maxPlayers = ref(10)
const gameMode = ref<'sandbox' | 'survival' | 'pvp'>('sandbox')
const isSubmitting = ref(false)

const selectedMap = computed(() => getMapById(props.mapId ?? 'dust2') ?? DEFAULT_MAP)

async function handleCreate() {
  if (!playerName.value.trim()) {
    ui.showToast('请输入玩家名称', 'warning')
    return
  }

  isSubmitting.value = true
  try {
    const success = await roomStore.createRoom(playerName.value.trim(), maxPlayers.value, gameMode.value)
    if (success) {
      // 初始化本地玩家
      const id = crypto.randomUUID?.() ?? `player-${Date.now()}`
      playerStore.initLocalPlayer(id, playerName.value.trim())

      // 初始化世界地图
      if (selectedMap.value) {
        const blocks = selectedMap.value.generate()
        worldStore.initializeWorld(selectedMap.value.id, blocks)
      }

      gameState.startGame(gameMode.value)

      setTimeout(() => {
        gameState.setLoadingProgress(100, `${selectedMap.value?.name} 加载完成！`)
        gameState.enterGame()
      }, 1500)

      ui.showToast(`房间 ${roomStore.roomId} 已创建`, 'success')
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="room-create panel-lowpoly p-6 w-96 mx-auto">
    <h2 class="text-xl text-craft-primary font-game mb-6 text-center">创建房间</h2>

    <!-- 地图信息卡 -->
    <div
      v-if="selectedMap"
      class="mb-4 p-3 rounded border border-craft-primary/40 bg-craft-surface/60 flex items-center gap-3"
    >
      <div
        class="w-10 h-10 rounded flex items-center justify-center text-xl flex-shrink-0"
        :style="{ backgroundColor: selectedMap.environment.skyColor + '44' }"
      >
        🗺️
      </div>
      <div class="flex-1 min-w-0">
        <div class="text-craft-light text-xs font-game truncate">{{ selectedMap.name }}</div>
        <div class="text-craft-text/50 text-[10px] font-game truncate">{{ selectedMap.description }}</div>
      </div>
      <button class="btn-lowpoly px-2 py-1 text-[10px]" @click="emit('back')">
        换地图
      </button>
    </div>

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

    <!-- 最大人数 -->
    <div class="mb-4">
      <label class="block text-craft-text text-sm mb-1 font-game">最大人数</label>
      <select v-model.number="maxPlayers" class="input-lowpoly w-full">
        <option :value="2">2 人</option>
        <option :value="5">5 人</option>
        <option :value="10">10 人</option>
        <option :value="20">20 人</option>
        <option :value="50">50 人</option>
      </select>
    </div>

    <!-- 游戏模式 -->
    <div class="mb-6">
      <label class="block text-craft-text text-sm mb-1 font-game">游戏模式</label>
      <div class="flex gap-2">
        <button
          v-for="mode in (['sandbox', 'survival', 'pvp'] as const)"
          :key="mode"
          class="btn-lowpoly flex-1 text-xs"
          :class="{ 'btn-primary': gameMode === mode }"
          @click="gameMode = mode"
        >
          {{ mode === 'sandbox' ? '沙盒' : mode === 'survival' ? '生存' : 'PVP' }}
        </button>
      </div>
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
        :disabled="isSubmitting || roomStore.isCreating"
        @click="handleCreate"
      >
        {{ isSubmitting ? '创建中...' : '创建房间' }}
      </button>
    </div>
  </div>
</template>
