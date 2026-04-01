<script setup lang="ts">
/**
 * CraftFire 根组件。
 * 管理全局游戏视图状态，切换主菜单、游戏场景和管理面板。
 */
import { computed, watch, onMounted, onUnmounted } from 'vue'
import { logger } from '@/utils/logger'
import { useGameStateStore } from '@/stores/gameState'
import { useRoomStore } from '@/stores/room'
import { useUIStore } from '@/stores/ui'
import { useSettingsStore } from '@/stores/settings'
import MainMenu from './components/UI/MainMenu.vue'
import GameScene from './components/GameScene.vue'
import GameHUD from './components/GameHUD.vue'
import SettingsMenu from './components/UI/SettingsMenu.vue'
import AdminPanel from './components/UI/AdminPanel.vue'
import Toast from './components/UI/Toast.vue'
import Confirm from './components/UI/Confirm.vue'

const gameState = useGameStateStore()
const roomStore = useRoomStore()
const uiStore = useUIStore()
const settings = useSettingsStore()

// 全局键盘处理（统一 ESC / F3 行为，避免分散的监听器造成竞态）
function onGlobalKeyDown(event: KeyboardEvent) {
  // F3 在任何视图下切换调试（保留兼容以前的实现）
  if (event.key === 'F3') {
    event.preventDefault()
    gameState.toggleDebugMode()
    return
  }

  if (event.key === 'Escape') {
    // 如果当前在游戏视图：切换设置面板（由 App 的 watcher 处理 pause / pointer lock）
    if (gameState.currentView === 'game') {
      event.preventDefault()
      const willOpenSettings = !uiStore.showSettings
      uiStore.showSettings = willOpenSettings
      if (willOpenSettings && document.pointerLockElement) {
        document.exitPointerLock()
      }
      return
    }

    // 非游戏视图：若设置面板打开则关闭它
    if (uiStore.showSettings) {
      event.preventDefault()
      uiStore.showSettings = false
    }
  }
}

onMounted(() => document.addEventListener('keydown', onGlobalKeyDown))
onUnmounted(() => document.removeEventListener('keydown', onGlobalKeyDown))
const currentView = computed(() => gameState.currentView)
const currentRoomId = computed(() => roomStore.roomId)
const gameBrightnessStyle = computed(() => ({
  filter: `brightness(${settings.video.brightness}%)`
}))

let loadingFallbackTimer: ReturnType<typeof setTimeout> | null = null

watch(
  () => gameState.currentView,
  (view) => {
    logger.info('App', '视图切换 →', view)
    if (loadingFallbackTimer) {
      clearTimeout(loadingFallbackTimer)
      loadingFallbackTimer = null
    }

    if (view === 'loading') {
      loadingFallbackTimer = setTimeout(() => {
        if (gameState.currentView === 'loading') {
          gameState.setLoadingProgress(100, '加载完成！')
          gameState.enterGame()
        }
      }, 500)
    }
  },
)

watch(
  () => uiStore.showSettings,
  (show) => {
    if (gameState.currentView !== 'game') return

    if (show) {
      gameState.pauseGame()
      if (document.pointerLockElement) {
        document.exitPointerLock()
      }
      return
    }

    if (gameState.isInGame) {
      gameState.resumeGame()
      void document.body.requestPointerLock?.()
    }
  },
)

/** 返回主菜单 */
const backToMenu = () => {
  logger.info('App', '返回主菜单')
  gameState.exitGame()
  void roomStore.leaveRoom()
}
</script>

<template>
  <div class="w-screen h-screen bg-craft-dark overflow-hidden">
    <!-- 主菜单 -->
    <MainMenu v-if="currentView === 'menu'" />

    <!-- 加载中 -->
    <template v-if="currentView === 'loading'">
      <div class="w-full h-full flex items-center justify-center text-craft-text font-game">
        {{ gameState.loadingMessage || '正在加载游戏场景...' }}
      </div>
    </template>

    <!-- 游戏场景 -->
    <template v-else-if="currentView === 'game'">
      <div class="w-full h-full" :style="gameBrightnessStyle">
        <GameScene :room-id="currentRoomId" />
      </div>
      <GameHUD
        :room-id="currentRoomId"
        @back-to-menu="backToMenu"
      />
    </template>

    <!-- 设置菜单 -->
    <SettingsMenu />

    <!-- 管理面板 -->
    <AdminPanel />

    <!-- 全局 Toast 通知 -->
    <Toast />

    <!-- 全局 Confirm 弹窗 -->
    <Confirm />
  </div>
</template>
