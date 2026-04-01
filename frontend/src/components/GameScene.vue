<script setup lang="ts">
/**
 * GameScene - 游戏 3D 场景主容器。
 * 使用 TresJS 构建 Three.js 声明式 3D 场景。
 * 包含：天空、光照、体素世界、玩家实体、特效。
 */
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { TresCanvas, useRenderLoop } from '@tresjs/core'
import { useGameStateStore } from '@/stores/gameState'
import { usePlayerStore } from '@/stores/player'
import { useSettingsStore } from '@/stores/settings'
import { useRoomStore } from '@/stores/room'
import { useWorldStore } from '@/stores/world'
import { useWebSocketStore } from '@/stores/websocket'
import { wsService } from '@/services/WebSocketService'
import { logger } from '@/utils/logger'
import type { WSMessage, PlayerMovePayload, PlayerJoinPayload, WorldUpdatePayload } from '@/types/websocket'
import VoxelWorld from './World/VoxelWorld.vue'
import PlayerEntity from './Player/PlayerEntity.vue'
import FirstPersonController from './Player/FirstPersonController.vue'
import RemotePlayer from './Player/RemotePlayer.vue'
import BlockPlacer from './World/BlockPlacer.vue'
import ScenePistol from './World/ScenePistol.vue'
import SparkEffects from './Effects/SparkEffects.vue'
import { botManager } from '@/composables/botManager'

const props = defineProps<{
  roomId: string
}>()

const gameState = useGameStateStore()
const playerStore = usePlayerStore()
const settings = useSettingsStore()
const roomStore = useRoomStore()
const worldStore = useWorldStore()
const wsStore = useWebSocketStore()

const canvasRef = ref<InstanceType<typeof TresCanvas> | null>(null)

// 帧率统计
let frameCount = 0
let fpsAccumulator = 0

function gameLoop(dt: number) {
	if (!gameState.isRunning || gameState.isPaused) return

	gameState.updateDeltaTime(dt)

	// 更新机器人
	botManager.updateBots(dt)

	// FPS 计算
	frameCount++
	fpsAccumulator += dt
	if (fpsAccumulator >= 1.0) {
		gameState.updateFPS(Math.round(frameCount / fpsAccumulator))
		frameCount = 0
		fpsAccumulator = 0
	}
}

let _syncInterval: ReturnType<typeof setInterval> | null = null
const { onLoop } = useRenderLoop()

const canvasDpr = computed(() => {
  if (typeof window === 'undefined') return 1
  return Math.min(window.devicePixelRatio || 1, 1.5)
})

const currentMapId = computed(() => worldStore.worldSeed || 'dust2')

function ensureRemotePlayer(playerId: string, name = 'Remote'): void {
  if (!playerStore.remotePlayers.has(playerId)) {
    playerStore.upsertRemotePlayer({
      id: playerId,
      name,
      position: { x: 0, y: 10, z: 0 },
      velocity: { x: 0, y: 0, z: 0 },
      rotation: { pitch: 0, yaw: 0, roll: 0 },
      health: 100,
      ammo: 30,
      equipment: 'pistol',
      isAlive: true,
      lastUpdateTime: Date.now(),
    })
  }
}

function onWSMessage(message: WSMessage): void {
  if (!message?.type) return

  if (message.type === 'player_join') {
    const payload = message.payload as PlayerJoinPayload
    if (!payload?.playerId || payload.playerId === playerStore.playerId) return
    playerStore.upsertRemotePlayer({
      id: payload.playerId,
      name: payload.playerName || 'Remote',
      position: { x: payload.x ?? 0, y: payload.y ?? 10, z: payload.z ?? 0 },
      velocity: { x: 0, y: 0, z: 0 },
      rotation: { pitch: 0, yaw: 0, roll: 0 },
      health: 100,
      ammo: 30,
      equipment: 'pistol',
      isAlive: true,
      lastUpdateTime: Date.now(),
    })
    return
  }

  if (message.type === 'player_leave') {
    const payload = message.payload as { playerId?: string }
    if (payload?.playerId) {
      playerStore.removeRemotePlayer(payload.playerId)
    }
    return
  }

  if (message.type === 'player_move') {
    const payload = message.payload as PlayerMovePayload
    if (!message.playerId || message.playerId === playerStore.playerId) return
    ensureRemotePlayer(message.playerId)
    const prev = playerStore.remotePlayers.get(message.playerId)
    if (!prev) return
    playerStore.upsertRemotePlayer({
      ...prev,
      position: { x: payload.x, y: payload.y, z: payload.z },
      rotation: {
        pitch: payload.rotation?.pitch ?? 0,
        yaw: payload.rotation?.yaw ?? 0,
        roll: 0,
      },
      lastUpdateTime: Date.now(),
    })
    return
  }

  if (message.type === 'world_update') {
    const payload = message.payload as WorldUpdatePayload
    if (payload?.changes?.length) {
      worldStore.applyWorldUpdate(payload.changes)
    }
  }
}

function connectRoomSocket(): void {
  const roomId = props.roomId || roomStore.roomId
  const roomIP = roomStore.currentRoom?.ip || '127.0.0.1'
  const playerName = playerStore.playerName || 'Player'

  if (!roomId || !playerStore.playerId) return

  try {
    wsService.on('*', (message) => {
      wsStore.incrementMessagesReceived()
      onWSMessage(message)
    })
    wsService.onStateChange((state) => {
      wsStore.setConnectionState(state)
    })

    wsService.connect(roomId, playerName, roomIP)
  } catch (err) {
    logger.warn('GameScene', 'WebSocket 连接失败，单机模式运行:', err)
  }

  _syncInterval = setInterval(() => {
    if (!gameState.isInGame) return
    const local = playerStore.localPlayer
    if (!local) return
    try {
      wsService.sendPlayerMove(
        local.position.x,
        local.position.y,
        local.position.z,
        local.rotation.pitch,
        local.rotation.yaw,
      )
      wsStore.incrementMessagesSent()
    } catch { /* 忽略发送失败 */ }
  }, 125)
}

function disconnectRoomSocket(): void {
  if (_syncInterval) {
    clearInterval(_syncInterval)
    _syncInterval = null
  }
  wsService.disconnect()
  wsStore.disconnect()
}

onMounted(() => {
	logger.info('GameScene', '3D 场景已挂载, roomId =', props.roomId || roomStore.roomId)
	botManager.initBots()
	connectRoomSocket()
})

onUnmounted(() => {
	logger.info('GameScene', '3D 场景已卸载')
	disconnectRoomSocket()
	botManager.clearBots()
})

onLoop(({ delta }) => {
  gameLoop(delta)
})

/** 机器人列表 */
const botsList = computed(() => botManager.getBotStates())
</script>

<template>
  <div class="game-scene w-full h-full">
    <TresCanvas
      ref="canvasRef"
      clear-color="#E8C99B"
      :antialias="settings.video.antiAliasing"
      :dpr="canvasDpr"
      window-size
    >
      <!-- 环境光（由 VoxelWorld 根据地图配置） -->
      <TresAmbientLight name="ambient-light" :intensity="0.4" color="#ffffff" />

      <!-- 方向光（由 VoxelWorld 根据地图配置） -->
      <TresDirectionalLight name="sun-light" :position="[50, 100, 50]" :intensity="0.8" color="#FFF8E1" />

      <!-- 第一人称控制器 -->
      <FirstPersonController />

      <!-- 体素世界（接收 mapId，应用环境配置） -->
      <VoxelWorld :map-id="currentMapId" />

      <!-- 方块放置器（射线投射） -->
      <BlockPlacer />

      <!-- 场景静态模型：手枪 -->
      <ScenePistol />

      <!-- 粒子特效：枪口/方块销毁火花 -->
      <SparkEffects />

      <!-- 本地玩家实体 -->
      <PlayerEntity
        v-if="playerStore.localPlayer"
        :player="playerStore.localPlayer"
        :is-local="true"
      />

      <!-- 远程玩家实体 -->
      <RemotePlayer
        v-for="[id, player] in playerStore.remotePlayers"
        :key="id"
        :player="player"
      />

      <!-- AI 机器人实体 -->
      <RemotePlayer
        v-for="bot in botsList"
        :key="bot.id"
        :player="bot"
      />
    </TresCanvas>
  </div>
</template>

<style scoped>
.game-scene {
  position: relative;
  overflow: hidden;
}
</style>