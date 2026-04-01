<script setup lang="ts">
/**
 * MainMenu - 游戏主菜单。
 * Low-Poly 风格主界面：地图选择、创建房间、加入房间、设置等。
 */
import { ref } from 'vue'
import { useUIStore } from '@/stores/ui'
import MapSelector from '@/components/MapSelector.vue'
import RoomCreate from '@/components/RoomCreate.vue'
import RoomConnect from '@/components/RoomConnect.vue'

type MenuView = 'main' | 'map' | 'create' | 'connect'

const ui = useUIStore()
const currentView = ref<MenuView>('main')
const selectedMapId = ref('dust2')

function navigate(view: MenuView) {
  currentView.value = view
}

function onMapSelected(mapId: string) {
  selectedMapId.value = mapId
  currentView.value = 'create'
}

function reloadGame() {
  window.location.reload()
}

// 管理员入口：连续点击版本号 5 次
const versionClickCount = ref(0)
const versionClickTimer = ref<ReturnType<typeof setTimeout> | null>(null)

function handleVersionClick() {
  versionClickCount.value++
  if (versionClickTimer.value) clearTimeout(versionClickTimer.value)
  versionClickTimer.value = setTimeout(() => { versionClickCount.value = 0 }, 3000)
  if (versionClickCount.value >= 5) {
    versionClickCount.value = 0
    ui.showAdminPanel = true
    ui.showToast('管理员面板已打开', 'info')
  }
}
</script>

<template>
  <div class="main-menu flex flex-col items-center justify-center min-h-screen">
    <!-- 地图选择 -->
    <MapSelector
      v-if="currentView === 'map'"
      @select="onMapSelected"
      @back="navigate('main')"
    />

    <!-- 创建房间（带地图信息） -->
    <RoomCreate
      v-else-if="currentView === 'create'"
      :map-id="selectedMapId"
      @back="navigate('map')"
    />

    <!-- 加入房间 -->
    <RoomConnect
      v-else-if="currentView === 'connect'"
      @back="navigate('main')"
    />

    <!-- 主菜单 -->
    <template v-else>
      <!-- Logo -->
      <div class="mb-12">
        <h1 class="text-5xl font-game text-craft-primary tracking-widest drop-shadow-lg">
          CRAFTFIRE
        </h1>
        <p class="text-center text-craft-text/60 text-xs font-game mt-2">
          Sandbox × FPS · Low-Poly Edition
        </p>
      </div>

      <!-- 菜单按钮 -->
      <div class="flex flex-col gap-3 w-72">
        <button class="btn-primary py-3 text-sm font-game" @click="navigate('map')">
          🗺️ 创建房间
        </button>
        <button class="btn-secondary py-3 text-sm font-game" @click="navigate('connect')">
          加入房间 / 局域网
        </button>

        <div class="border-t border-white/10 my-2" />

        <button class="btn-info py-3 text-sm font-game" @click="ui.togglePlayerProfile()">
          个人资料
        </button>
        <button class="btn-info py-3 text-sm font-game" @click="ui.toggleSettings()">
          设置
        </button>
        <button class="btn-secondary py-3 text-sm font-game" @click="reloadGame()">
          重载资源
        </button>
      </div>

      <!-- 版本号 -->
      <p
        class="absolute bottom-4 text-craft-text/30 text-xs font-game cursor-pointer hover:text-craft-primary/50 transition-colors select-none"
        title="点击5次打开管理员面板"
        @click="handleVersionClick"
      >
        CraftFire v1.0.0
      </p>
    </template>
  </div>
</template>

<style scoped>
.main-menu {
  background:
    radial-gradient(circle at 50% 45%, rgba(243, 55, 140, 0.82) 0%, rgba(243,55,140,0.46) 6%, rgba(243,55,140,0.18) 12%, transparent 32%),
    radial-gradient(ellipse at 20% 25%, rgba(234, 77, 255, 0.42) 0%, transparent 55%),
    radial-gradient(ellipse at 82% 18%, rgba(182, 165, 251, 0.36) 0%, transparent 52%),
    radial-gradient(ellipse at 50% 85%, rgba(240, 127, 47, 0.6) 0%, transparent 54%);
  background-size: 100% 120%, 110% 160%, 180% 180%, 170% 170%;
  animation: menuGradientFlow 6s ease-in-out infinite alternate;
  background-repeat: no-repeat;
  background-blend-mode: screen;
}

@keyframes menuGradientFlow {
  0% {
    background-position: 50% 45%, 0% 0%, 100% 0%, 50% 100%;
    background-size: 120% 120%, 160% 160%, 180% 180%, 170% 170%;
  }
  50% {
    background-position: 50% 45%, 30% 24%, 70% 30%, 36% 70%;
    background-size: 110% 110%, 170% 170%, 170% 170%, 160% 160%;
  }
  100% {
    background-position: 50% 45%, 16% 34%, 88% 18%, 66% 92%;
    background-size: 130% 130%, 160% 160%, 190% 190%, 170% 170%;
  }
}
</style>
