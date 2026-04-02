<script setup lang="ts">
/**
 * FirstPersonController - 第一人称摄像机控制器。
 * 管理指针锁定、鼠标转向、WASD 移动。
 * 在客户端做移动预测，将结果通过 WebSocket 发送到服务器。
 */
import { ref, onMounted, onUnmounted } from 'vue'
import { useTresContext, useRenderLoop } from '@tresjs/core'
import { usePlayerStore } from '@/stores/player'
import { useGameStateStore } from '@/stores/gameState'
import { useUIStore } from '@/stores/ui'
import { useSettingsStore } from '@/stores/settings'
import { useWorldStore } from '@/stores/world'
import { InputHandler, KEY_BINDINGS } from '@/services/InputHandler'
import { physicsService } from '@/services/PhysicsService'
import { logger } from '@/utils/logger'
import { weaponManager } from '@/composables/weaponManager'

const playerStore = usePlayerStore()
const gameState = useGameStateStore()
const uiStore = useUIStore()
const settings = useSettingsStore()
const worldStore = useWorldStore()

const { camera } = useTresContext()
const { onLoop } = useRenderLoop()

const STORE_SYNC_INTERVAL = 1 / 30

// 输入处理
let inputHandler: InputHandler | null = null

// 欧拉角
const yaw = ref(0)
const pitch = ref(0)

let simulatedPosition = { x: 0, y: 10, z: 0 }
let hasSimulatedPosition = false
let positionSyncAccumulator = 0
let rotationSyncAccumulator = 0

function onMouseMove(event: MouseEvent) {
  if (!document.pointerLockElement) return
  if (gameState.isPaused || uiStore.hasActiveOverlay) return

  const sensitivity = settings.controls.mouseSensitivity * 0.002
  yaw.value -= event.movementX * sensitivity
  pitch.value -= event.movementY * sensitivity * (settings.controls.invertY ? -1 : 1)
  // 限制俯仰角
  pitch.value = Math.max(-Math.PI / 2 + 0.01, Math.min(Math.PI / 2 - 0.01, pitch.value))
}

function onClick() {
  if (!document.pointerLockElement && gameState.isInGame && !uiStore.hasActiveOverlay) {
    document.body.requestPointerLock()
  }
}

function onPointerLockChange() {
  // 暂停/恢复统一由 App 中对 ui.showSettings 的监听处理，避免状态竞态导致卡住
}

function onKeyDown(e: KeyboardEvent) {
  if (gameState.isPaused || uiStore.hasActiveOverlay) return
  if (e.code === KEY_BINDINGS.WEAPON_SWITCH) {
    weaponManager.toggleWeapon()
  }
}
function updateMovement(dt: number) {
  if (!gameState.isRunning || gameState.isPaused) return
  if (!inputHandler) return

  if (!hasSimulatedPosition) {
    const pos = playerStore.position
    simulatedPosition = { x: pos.x, y: pos.y, z: pos.z }
    hasSimulatedPosition = true
  }

  const state = inputHandler.getState()

  const newPos = physicsService.updatePlayerPosition(
    simulatedPosition,
    dt,
    {
      ...state,
      yaw: yaw.value
    },
    // 方块检测回调
    (x, y, z) => {
      // 简单优化：如果 y < 0 或 y > 256 通常没有方块 (除了基岩层)
      if (y < 0) return true // 防止掉出世界
      return !!worldStore.getBlock(x, y, z)
    }
  )

  simulatedPosition = newPos

  positionSyncAccumulator += dt
  rotationSyncAccumulator += dt

  if (positionSyncAccumulator >= STORE_SYNC_INTERVAL) {
    playerStore.updateLocalPosition(simulatedPosition)
    // 同步 velocity 到 playerStore，供 VoxelCharacter 行走动画判断
    playerStore.updateLocalVelocity(physicsService.velocity)
    positionSyncAccumulator = 0
  }

  if (rotationSyncAccumulator >= STORE_SYNC_INTERVAL) {
    playerStore.updateLocalRotation({ pitch: pitch.value, yaw: yaw.value, roll: 0 })
    rotationSyncAccumulator = 0
  }

  // 更新相机
  if (camera.value) {
    camera.value.position.set(simulatedPosition.x, simulatedPosition.y + 1.6, simulatedPosition.z)
    camera.value.rotation.set(pitch.value, yaw.value, 0, 'YXZ')
  }
}

onLoop(({ delta }) => {
  updateMovement(delta)
})

onMounted(() => {
  logger.info('FPSController', '第一人称控制器已挂载')
  const pos = playerStore.position
  simulatedPosition = { x: pos.x, y: pos.y, z: pos.z }
  hasSimulatedPosition = true
  inputHandler = new InputHandler()
  inputHandler.start()
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('click', onClick)
  document.addEventListener('pointerlockchange', onPointerLockChange)
  document.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  if (hasSimulatedPosition) {
    playerStore.updateLocalPosition(simulatedPosition)
  }
  playerStore.updateLocalRotation({ pitch: pitch.value, yaw: yaw.value, roll: 0 })
  inputHandler?.stop()
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('click', onClick)
  document.removeEventListener('pointerlockchange', onPointerLockChange)
  document.removeEventListener('keydown', onKeyDown)
})
</script>

<template>
  <!-- 第一人称控制器不渲染可视元素，通过脚本操控 camera -->
  <TresPerspectiveCamera :position="[0, 10, 0]" :rotation="[-0.15, 0, 0]" :fov="75" :near="0.1" :far="500" />
</template>
