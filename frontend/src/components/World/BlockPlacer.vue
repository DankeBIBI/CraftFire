<script setup lang="ts">
/**
 * BlockPlacer - 方块放置/移除组件。
 * 使用射线投射检测鼠标指向的方块面，处理放置和移除操作。
 */
import { ref, onMounted, onUnmounted } from 'vue'
import { useWorldStore } from '@/stores/world'
import { usePlayerStore } from '@/stores/player'
import { useGameStateStore } from '@/stores/gameState'
import { useUIStore } from '@/stores/ui'
import * as WailsService from '@/services/WailsService'
import { emitSparkEffect } from '@/utils/sparkEffects'

const worldStore = useWorldStore()
const playerStore = usePlayerStore()
const gameState = useGameStateStore()
const uiStore = useUIStore()

/** 当前选择的方块类型 */
const selectedBlockType = ref('stone')

/** 高亮预览方块的位置 */
const previewPosition = ref<{ x: number; y: number; z: number } | null>(null)

/** 处理鼠标点击 */
function onMouseDown(event: MouseEvent) {
  if (!document.pointerLockElement) return
  if (gameState.isPaused || uiStore.hasActiveOverlay) return

  // 左键留给武器射击逻辑，避免开火时触发方块编辑造成卡顿
  if (event.button === 0) return

  // 仅处理中键移除与右键放置
  if (event.button !== 1 && event.button !== 2) return

  // 简单的射线投射：从玩家位置沿朝向方向投射
  const pos = playerStore.position
  const rot = playerStore.rotation
  const dir = {
    x: -Math.sin(rot.yaw) * Math.cos(rot.pitch),
    y: Math.sin(rot.pitch),
    z: -Math.cos(rot.yaw) * Math.cos(rot.pitch),
  }

  // 步进射线
  const maxDist = 8 // 最大交互距离
  const step = 0.1
  let hitBlock: { x: number; y: number; z: number } | null = null
  let hitNormal: { x: number; y: number; z: number } | null = null
  let prevX = 0, prevY = 0, prevZ = 0

  for (let t = 0; t < maxDist; t += step) {
    const rx = pos.x + dir.x * t
    const ry = (pos.y + 1.6) + dir.y * t
    const rz = pos.z + dir.z * t
    const bx = Math.floor(rx)
    const by = Math.floor(ry)
    const bz = Math.floor(rz)

    const block = worldStore.getBlock(bx, by, bz)
    if (block) {
      hitBlock = { x: bx, y: by, z: bz }
      hitNormal = { x: prevX - bx, y: prevY - by, z: prevZ - bz }
      break
    }
    prevX = bx
    prevY = by
    prevZ = bz
  }

  if (!hitBlock) return

  if (event.button === 1) {
    // 中键：移除方块
    const impactNormal = hitNormal ?? { x: 0, y: 1, z: 0 }
    emitSparkEffect({
      kind: 'impact',
      position: {
        x: hitBlock.x + 0.5 + impactNormal.x * 0.52,
        y: hitBlock.y + 0.5 + impactNormal.y * 0.52,
        z: hitBlock.z + 0.5 + impactNormal.z * 0.52,
      },
      normal: impactNormal,
    })
    worldStore.removeBlock(hitBlock.x, hitBlock.y, hitBlock.z)
    WailsService.RemoveBlock(hitBlock.x, hitBlock.y, hitBlock.z).catch(() => {})
  } else if (event.button === 2 && hitNormal) {
    // 右键：放置方块
    const placePos = {
      x: hitBlock.x + hitNormal.x,
      y: hitBlock.y + hitNormal.y,
      z: hitBlock.z + hitNormal.z,
    }
    worldStore.placeBlock(placePos.x, placePos.y, placePos.z, selectedBlockType.value)
    WailsService.PlaceBlock(placePos.x, placePos.y, placePos.z, selectedBlockType.value).catch(() => {})
  }
}

/** 禁用右键菜单 */
function onContextMenu(event: MouseEvent) {
  if (document.pointerLockElement) {
    event.preventDefault()
  }
}

/** 滚轮切换方块类型 */
const blockTypes = ['stone', 'wood', 'glass', 'dirt', 'grass', 'sand']

function onWheel(event: WheelEvent) {
  if (!document.pointerLockElement) return
  const idx = blockTypes.indexOf(selectedBlockType.value)
  const dir = event.deltaY > 0 ? 1 : -1
  const newIdx = (idx + dir + blockTypes.length) % blockTypes.length
  selectedBlockType.value = blockTypes[newIdx]
}

onMounted(() => {
  document.addEventListener('mousedown', onMouseDown)
  document.addEventListener('contextmenu', onContextMenu)
  document.addEventListener('wheel', onWheel)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', onMouseDown)
  document.removeEventListener('contextmenu', onContextMenu)
  document.removeEventListener('wheel', onWheel)
})
</script>

<template>
  <!-- 预览高亮方块 -->
  <TresMesh
    v-if="previewPosition"
    :position="[previewPosition.x + 0.5, previewPosition.y + 0.5, previewPosition.z + 0.5]"
  >
    <TresBoxGeometry :args="[1.02, 1.02, 1.02]" />
    <TresMeshBasicMaterial color="#FFFFFF" :opacity="0.3" transparent wireframe />
  </TresMesh>
</template>
