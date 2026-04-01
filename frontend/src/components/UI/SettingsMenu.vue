<script setup lang="ts">
/**
 * SettingsMenu - 统一设置菜单。
 * 支持主界面按钮与游戏内 ESC 唤出。
 * 包含两个标签页：基本设置 / 房间管理（房主）
 */
import { ref, computed } from 'vue'
import { useUIStore } from '@/stores/ui'
import { useSettingsStore } from '@/stores/settings'
import { useGameStateStore } from '@/stores/gameState'
import { useRoomStore } from '@/stores/room'
import { usePlayerStore } from '@/stores/player'
import { wsService } from '@/services/WebSocketService'

defineOptions({ name: 'SettingsMenu' })

const ui = useUIStore()
const settings = useSettingsStore()
const gameState = useGameStateStore()
const roomStore = useRoomStore()
const playerStore = usePlayerStore()

/** 标签页类型 */
type TabType = 'basic' | 'room' | 'superadmin'
const activeTab = ref<TabType>('basic')

/** 是否超级管理员（简化判断：本地开发或特定条件） */
const isSuperAdmin = ref(true)

/** 房间是否已锁定 */
const isRoomLocked = ref(false)

/** 在线玩家列表（房主视角） */
const onlinePlayers = ref<Array<{
  id: string
  name: string
  health: number
  isMuted: boolean
}>>([])

// ─── 基本设置 ────────────────────────────────
const brightnessLabel = computed(() => `${settings.video.brightness}%`)
const sensitivityLabel = computed(() => settings.controls.mouseSensitivity.toFixed(2))
const isInGameSettings = computed(() => gameState.currentView === 'game')

function closeMenu() {
  ui.showSettings = false
}

function exitToMenu() {
  closeMenu()
  gameState.exitGame()
  void roomStore.leaveRoom()
}

function onBrightnessInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  settings.updateVideoSettings({ brightness: Number(target.value) })
}

function onSensitivityInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  settings.updateControlSettings({ mouseSensitivity: Number(target.value) })
}

function onMasterVolumeInput(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  settings.updateAudioSettings({ masterVolume: Number(target.value) })
}

function onShowFPSChange(event: Event) {
  const target = event.target as HTMLInputElement | null
  if (!target) return
  settings.updateVideoSettings({ showFPS: target.checked })
}

// ─── 房间管理 ────────────────────────────────
/** 刷新玩家列表 */
function refreshRoomPlayers() {
  const players: typeof onlinePlayers.value = []

  if (playerStore.localPlayer) {
    players.push({
      id: playerStore.localPlayer.id,
      name: playerStore.localPlayer.name + ' ★',
      health: playerStore.localPlayer.health,
      isMuted: false,
    })
  }

  playerStore.remotePlayers.forEach((p) => {
    players.push({
      id: p.id,
      name: p.name,
      health: p.health,
      isMuted: false,
    })
  })

  onlinePlayers.value = players
}

/** 切换标签页时刷新玩家 */
function switchToRoomTab() {
  activeTab.value = 'room'
  refreshRoomPlayers()
}

/** 踢出玩家 */
async function kickPlayer(playerId: string, playerName: string) {
  const confirmed = await ui.confirm(`确定要踢出玩家 "${playerName}" 吗？`)
  if (!confirmed) return

  wsService.send('admin_kick', {
    targetPlayerId: playerId,
    reason: '违反游戏规则',
  })

  onlinePlayers.value = onlinePlayers.value.filter(p => p.id !== playerId)
  playerStore.removeRemotePlayer(playerId)
  ui.showToast(`已踢出玩家 ${playerName}`, 'success')
}

/** 切换房间锁定状态 */
function toggleRoomLock() {
  isRoomLocked.value = !isRoomLocked.value
  if (isRoomLocked.value) {
    ui.showToast('房间已锁定', 'success')
  } else {
    ui.showToast('房间已解锁', 'success')
  }
}

/** 切换游戏模式 */
function changeGameMode(mode: string) {
  if (!roomStore.currentRoom) return
  wsService.send('admin_change_mode', { mode })
  if (roomStore.currentRoom) {
    roomStore.currentRoom.gameMode = mode as 'sandbox' | 'survival' | 'pvp'
  }
  ui.showToast(`游戏模式已切换为 ${mode}`, 'success')
}

/** 加血 */
function healPlayer(playerId: string, playerName: string) {
  const player = onlinePlayers.value.find(p => p.id === playerId)
  if (player) {
    player.health = 100
    wsService.send('admin_heal', { targetPlayerId: playerId })
    ui.showToast(`已为 ${playerName} 恢复满血`, 'success')
  }
}

/** 传送玩家到身边 */
function teleportToMe(playerId: string, playerName: string) {
  wsService.send('admin_teleport', {
    targetPlayerId: playerId,
    toPosition: playerStore.localPlayer?.position || { x: 0, y: 10, z: 0 },
  })
  ui.showToast(`已传送 ${playerName} 到身边`, 'success')
}

/** 获取血量颜色 */
function getHealthColor(health: number): string {
  if (health > 60) return '#22c55e'
  if (health > 30) return '#eab308'
  return '#ef4444'
}

// ─── 超级管理员 ────────────────────────────────
/** 全部房间列表（超管） */
const allRooms = ref<Array<{
  id: string
  name: string
  players: number
  maxPlayers: number
}>>([])

/** 全服公告 */
const broadcastMessage = ref('')
const showBroadcastModal = ref(false)

/** 刷新全部房间 */
function refreshAllRooms() {
  // 模拟数据，实际应从服务端获取
  allRooms.value = [
    { id: '123456', name: '沙盒房间', players: 3, maxPlayers: 8 },
    { id: '654321', name: 'PVP战场', players: 5, maxPlayers: 10 },
  ]
  ui.showToast('已刷新房间列表', 'success')
}

/** 关闭房间 */
async function closeRoom(roomId: string) {
  const confirmed = await ui.confirm(`确定要关闭房间 ${roomId} 吗？`)
  if (!confirmed) return

  wsService.send('superadmin_close_room', { roomId })
  allRooms.value = allRooms.value.filter(r => r.id !== roomId)
  ui.showToast(`已关闭房间 ${roomId}`, 'success')
}

/** 发送全服公告 */
function sendBroadcast() {
  if (!broadcastMessage.value.trim()) {
    ui.showToast('公告内容不能为空', 'warning')
    return
  }
  wsService.send('superadmin_broadcast', { message: broadcastMessage.value.trim() })
  ui.showToast('全服公告已发送', 'success')
  broadcastMessage.value = ''
  showBroadcastModal.value = false
}

/** 打开公告弹窗 */
function openBroadcastModal() {
  broadcastMessage.value = ''
  showBroadcastModal.value = true
}
</script>

<template>
  <div v-if="ui.showSettings" class="overlay flex items-center justify-center p-4">
    <div class="panel-lowpoly w-full max-w-2xl pointer-events-auto">
      <!-- 标题栏 -->
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-craft-light text-lg font-game">游戏设置</h2>
        <button class="btn-lowpoly px-4 py-2 text-xs" @click="closeMenu">关闭</button>
      </div>

      <!-- 标签页切换 -->
      <div class="flex gap-1 mb-4 border-b border-white/10 pb-2">
        <button
          class="px-4 py-2 text-xs font-game transition-colors"
          :class="activeTab === 'basic' ? 'text-craft-primary bg-craft-surface/50' : 'text-craft-text/60 hover:text-craft-light'"
          @click="activeTab = 'basic'"
        >
          基本设置
        </button>
        <button
          v-if="isInGameSettings && roomStore.isHost"
          class="px-4 py-2 text-xs font-game transition-colors"
          :class="activeTab === 'room' ? 'text-craft-primary bg-craft-surface/50' : 'text-craft-text/60 hover:text-craft-light'"
          @click="switchToRoomTab"
        >
          房间管理
        </button>
        <button
          v-if="isSuperAdmin"
          class="px-4 py-2 text-xs font-game transition-colors"
          :class="activeTab === 'superadmin' ? 'text-craft-primary bg-craft-surface/50' : 'text-craft-text/60 hover:text-craft-light'"
          @click="activeTab = 'superadmin'"
        >
          超管
        </button>
      </div>

      <!-- 基本设置 -->
      <div v-if="activeTab === 'basic'" class="space-y-5">
        <div>
          <div class="flex items-center justify-between text-craft-light text-xs mb-2">
            <span>亮度</span>
            <span>{{ brightnessLabel }}</span>
          </div>
          <input
            class="w-full accent-craft-primary"
            type="range"
            min="50"
            max="150"
            step="1"
            :value="settings.video.brightness"
            @input="onBrightnessInput"
          />
        </div>

        <div>
          <div class="flex items-center justify-between text-craft-light text-xs mb-2">
            <span>鼠标灵敏度</span>
            <span>{{ sensitivityLabel }}</span>
          </div>
          <input
            class="w-full accent-craft-secondary"
            type="range"
            min="0.1"
            max="2"
            step="0.01"
            :value="settings.controls.mouseSensitivity"
            @input="onSensitivityInput"
          />
        </div>

        <div>
          <div class="flex items-center justify-between text-craft-light text-xs mb-2">
            <span>主音量</span>
            <span>{{ settings.audio.masterVolume }}%</span>
          </div>
          <input
            class="w-full accent-craft-accent"
            type="range"
            min="0"
            max="100"
            step="1"
            :value="settings.audio.masterVolume"
            @input="onMasterVolumeInput"
          />
        </div>

        <label class="flex items-center justify-between text-craft-light text-xs border-t border-white/10 pt-4">
          <span>显示 FPS</span>
          <input
            type="checkbox"
            class="h-4 w-4 accent-craft-primary"
            :checked="settings.video.showFPS"
            @change="onShowFPSChange"
          />
        </label>

        <div class="border-t border-white/10 pt-4 flex justify-end">
          <button
            v-if="isInGameSettings"
            class="btn-danger px-4 py-2 text-xs"
            @click="exitToMenu"
          >
            退出到主菜单
          </button>
        </div>
      </div>

      <!-- 房间管理（仅房主） -->
      <div v-if="activeTab === 'room' && isInGameSettings && roomStore.isHost" class="space-y-4">
        <!-- 房间信息 -->
        <div class="flex items-center justify-between p-2 bg-craft-dark/30 rounded border border-white/10">
          <div class="flex items-center gap-4 text-xs font-game">
            <span class="text-craft-text/60">房间:</span>
            <span class="text-craft-secondary">{{ roomStore.currentRoom?.roomId || '未加入' }}</span>
            <span class="text-craft-text/40">|</span>
            <span class="text-craft-text/60">人数:</span>
            <span class="text-craft-light">{{ onlinePlayers.length }}/{{ roomStore.currentRoom?.maxPlayers || 0 }}</span>
            <span class="text-craft-text/40">|</span>
            <span class="text-craft-text/60">模式:</span>
            <span class="text-craft-light">{{ roomStore.currentRoom?.gameMode || '-' }}</span>
          </div>
          <button class="btn-lowpoly px-2 py-1 text-xs" @click="refreshRoomPlayers">🔄 刷新</button>
        </div>

        <!-- 房间操作 -->
        <div class="grid grid-cols-3 gap-2">
          <button
            class="p-3 rounded border border-white/10 bg-craft-dark/30 hover:bg-craft-surface/30 transition-colors text-left"
            @click="toggleRoomLock"
          >
            <div class="text-xl mb-1">{{ isRoomLocked ? '🔓' : '🔒' }}</div>
            <div class="text-craft-light text-xs font-game">{{ isRoomLocked ? '解锁房间' : '锁定房间' }}</div>
          </button>

          <div class="p-3 rounded border border-white/10 bg-craft-dark/30 col-span-2">
            <div class="text-craft-light text-xs font-game mb-2">游戏模式</div>
            <div class="flex gap-1">
              <button
                v-for="mode in ['sandbox', 'survival', 'pvp']"
                :key="mode"
                class="btn-lowpoly px-2 py-1 text-xs flex-1"
                :class="roomStore.currentRoom?.gameMode === mode ? 'btn-primary' : ''"
                @click="changeGameMode(mode)"
              >
                {{ mode === 'sandbox' ? '沙盒' : mode === 'survival' ? '生存' : 'PVP' }}
              </button>
            </div>
          </div>
        </div>

        <!-- 玩家列表 -->
        <div class="border-t border-white/10 pt-4">
          <div class="text-craft-text/60 text-xs font-game mb-2">在线玩家</div>
          <div class="space-y-2 max-h-48 overflow-y-auto">
            <div
              v-for="player in onlinePlayers"
              :key="player.id"
              class="flex items-center justify-between p-2 rounded bg-craft-dark/30 border border-white/10"
            >
              <div class="flex items-center gap-2">
                <div class="w-8 h-8 rounded bg-craft-surface flex items-center justify-center text-sm">🎮</div>
                <div>
                  <div class="text-craft-light text-xs font-game flex items-center gap-2">
                    {{ player.name }}
                    <span v-if="player.isMuted" class="text-red-400 text-xs">🔇</span>
                  </div>
                  <div class="flex items-center gap-2 mt-0.5">
                    <div class="w-12 h-1 bg-craft-dark rounded-full overflow-hidden">
                      <div
                        class="h-full transition-all"
                        :style="{ width: `${player.health}%`, backgroundColor: getHealthColor(player.health) }"
                      />
                    </div>
                    <span class="text-craft-text/40 text-xs">{{ player.health }}%</span>
                  </div>
                </div>
              </div>
              <div class="flex gap-1">
                <button
                  class="btn-info px-2 py-1 text-xs"
                  @click="healPlayer(player.id, player.name)"
                >
                  ❤️
                </button>
                <button
                  class="btn-info px-2 py-1 text-xs"
                  @click="teleportToMe(player.id, player.name)"
                >
                  📍
                </button>
                <button
                  v-if="player.name !== playerStore.localPlayer?.name + ' ★'"
                  class="btn-danger px-2 py-1 text-xs"
                  @click="kickPlayer(player.id, player.name)"
                >
                  踢
                </button>
              </div>
            </div>

            <div v-if="onlinePlayers.length === 0" class="text-center py-6 text-craft-text/40 text-xs font-game">
              暂无在线玩家
            </div>
          </div>
        </div>
      </div>

      <!-- 非房主提示 -->
      <div v-if="activeTab === 'room' && isInGameSettings && !roomStore.isHost" class="text-center py-12">
        <div class="text-4xl mb-4">🏠</div>
        <div class="text-craft-text/60 text-xs font-game">你不是房主，无法管理房间</div>
      </div>

      <!-- 超级管理员面板 -->
      <div v-if="activeTab === 'superadmin' && isSuperAdmin" class="space-y-4">
        <div class="text-center py-8">
          <div class="text-4xl mb-4">🔑</div>
          <div class="text-craft-light text-sm font-game mb-2">超级管理员模式</div>
          <div class="text-craft-text/60 text-xs font-game mb-6">可管理全部房间和玩家</div>

          <!-- 操作按钮 -->
          <div class="grid grid-cols-2 gap-3">
            <button
              class="p-4 rounded border border-white/10 bg-craft-dark/30 hover:bg-craft-surface/30 transition-colors text-left"
              @click="openBroadcastModal"
            >
              <div class="text-2xl mb-2">📢</div>
              <div class="text-craft-light text-xs font-game">全服公告</div>
              <div class="text-craft-text/40 text-xs mt-1">向所有房间发送公告</div>
            </button>

            <button
              class="p-4 rounded border border-white/10 bg-craft-dark/30 hover:bg-craft-surface/30 transition-colors text-left"
              @click="refreshAllRooms"
            >
              <div class="text-2xl mb-2">🔄</div>
              <div class="text-craft-light text-xs font-game">刷新房间</div>
              <div class="text-craft-text/40 text-xs mt-1">刷新所有房间列表</div>
            </button>
          </div>
        </div>

        <!-- 全部房间列表（模拟数据） -->
        <div class="border-t border-white/10 pt-4">
          <div class="text-craft-text/60 text-xs font-game mb-2">全部房间</div>
          <div class="space-y-2 max-h-60 overflow-y-auto">
            <div
              v-for="room in allRooms"
              :key="room.id"
              class="p-3 rounded bg-craft-dark/30 border border-white/10"
            >
              <div class="flex items-center justify-between">
                <div>
                  <div class="text-craft-light text-xs font-game">{{ room.name }}</div>
                  <div class="text-craft-text/40 text-xs">ID: {{ room.id }} | {{ room.players }}/{{ room.maxPlayers }}人</div>
                </div>
                <div class="flex gap-1">
                  <button
                    class="btn-danger px-2 py-1 text-xs"
                    @click="closeRoom(room.id)"
                  >
                    关闭
                  </button>
                </div>
              </div>
            </div>

            <div v-if="allRooms.length === 0" class="text-center py-6 text-craft-text/40 text-xs font-game">
              暂无房间数据
            </div>
          </div>
        </div>
      </div>

      <!-- 全服公告弹窗 -->
      <div v-if="showBroadcastModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showBroadcastModal = false">
        <div class="panel-lowpoly w-full max-w-md p-6">
          <h3 class="text-craft-primary text-sm font-game mb-4">📢 全服公告</h3>
          <textarea
            v-model="broadcastMessage"
            class="input-lowpoly w-full h-24 resize-none mb-4"
            placeholder="输入公告内容..."
            maxlength="100"
          />
          <div class="flex gap-2">
            <button class="btn-lowpoly flex-1" @click="showBroadcastModal = false">取消</button>
            <button class="btn-primary flex-1" @click="sendBroadcast">发送</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
