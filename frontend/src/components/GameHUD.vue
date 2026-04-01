<script setup lang="ts">
/**
 * GameHUD - 游戏 HUD 抬头显示。
 * 包含：准心、血量条、弹药数、迷你地图占位、聊天、FPS。
 */
import { computed } from 'vue'
import { usePlayerStore } from '@/stores/player'
import { useGameStateStore } from '@/stores/gameState'
import { useUIStore } from '@/stores/ui'
import { useSettingsStore } from '@/stores/settings'
import { weaponManager } from '@/composables/weaponManager'

const playerStore = usePlayerStore()
const gameState = useGameStateStore()
const ui = useUIStore()
const settings = useSettingsStore()

const healthPercent = computed(() => playerStore.health)
const healthColor = computed(() => {
  const h = playerStore.health
  if (h > 60) return '#00E676' // Vivid Green
  if (h > 30) return '#FFD600' // Vivid Yellow
  return '#FF1744' // Vivid Red
})

const pos = computed(() => {
  const p = playerStore.position
  return `X: ${p.x.toFixed(1)}  Y: ${p.y.toFixed(1)}  Z: ${p.z.toFixed(1)}`
})

const weaponSlots = computed(() => weaponManager.getAllSlotsDisplayInfo())

function handleWeaponSlotClick(slot: string) {
  weaponManager.switchWeapon(slot as 'primary' | 'secondary')
}
</script>

<template>
  <div class="game-hud absolute inset-0 pointer-events-none z-10">
    <!-- 准心 -->
    <div v-if="ui.crosshairVisible"
      class="crosshair absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 mix-blend-difference">
      <div class="w-6 h-0.5 bg-white absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 shadow-sm" />
      <div class="h-6 w-0.5 bg-white absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 shadow-sm" />
    </div>

    <!-- 左下：血量条 -->
    <div class="absolute bottom-4 left-4 pointer-events-auto">
      <div class="flex items-center gap-2">
        <span class="text-white text-sm font-game drop-shadow-md">HP</span>
        <div
          class="w-48 h-6 bg-craft-surface/90 border-2 border-black shadow-lowpoly-sm skew-x-[-10deg] overflow-hidden">
          <div class="h-full transition-all duration-300 transform scale-x-110 origin-left"
            :style="{ width: healthPercent + '%', backgroundColor: healthColor }" />
        </div>
        <span class="text-white text-sm font-game drop-shadow-md">{{ playerStore.health }}</span>
      </div>

      <!-- 弹药 -->
      <div class="flex items-center gap-2 mt-2" v-if="playerStore.localPlayer">
        <span class="text-white text-sm font-game drop-shadow-md">AMMO</span>
        <span class="text-craft-secondary text-lg font-game drop-shadow-md">
          {{ playerStore.localPlayer.ammo }}
        </span>
      </div>
    </div>

    <!-- 右下：装备 & 快捷栏占位 -->
    <div class="absolute bottom-4 right-4 pointer-events-auto">
      <!-- 武器切换栏 -->
      <div class="flex gap-1 mb-1">
        <div v-for="slot in weaponSlots" :key="slot.slot" @click="handleWeaponSlotClick(slot.slot)"
          class="w-16 h-16 border-2 bg-craft-surface/80 flex flex-col items-center justify-center text-white/80 text-xs font-game shadow-lowpoly-sm transition-all cursor-pointer"
          :class="{
            'border-craft-primary': slot.isActive,
            'border-black': !slot.isActive,
            'hover:translate-y-[-2px]': !slot.isActive,
            'hover:border-craft-primary': !slot.isActive
          }">
          <span class="text-craft-secondary">{{ slot.slot === 'primary' ? '1' : '2' }}</span>
          <span class="text-[10px]">{{ slot.name }}</span>
        </div>
      </div>
      <!-- <div class="flex gap-1">
        <div v-for="i in 9" :key="i"
          class="w-12 h-12 border-2 border-black bg-craft-surface/80 flex items-center justify-center text-white/80 text-sm font-game shadow-lowpoly-sm hover:translate-y-[-2px] hover:border-craft-primary transition-all">
          {{ i }}
        </div>
      </div> -->
    </div>

    <!-- 左上：调试信息 -->
    <div v-if="settings.video.showFPS || gameState.isDebugMode"
      class="absolute top-4 left-4 text-craft-light text-xs font-game space-y-1 bg-craft-surface/80 border-2 border-black px-3 py-2 shadow-lowpoly-sm">
      <div class="text-craft-primary">FPS: {{ gameState.fps }}</div>
      <div>{{ pos }}</div>
      <div>Online: {{ playerStore.remotePlayerCount + 1 }}</div>
      <div v-if="gameState.isDebugMode" class="text-craft-secondary">Mode: {{ gameState.gameMode }}</div>
    </div>

    <!-- 右上：小地图占位 -->
    <div class="absolute top-4 right-4">
      <div class="w-32 h-32 border-2 border-black bg-craft-surface/90 flex items-center justify-center shadow-lowpoly">
        <span class="text-white/40 text-xs font-game">MAP</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.game-hud {
  font-family: 'Press Start 2P', monospace;
}
</style>
