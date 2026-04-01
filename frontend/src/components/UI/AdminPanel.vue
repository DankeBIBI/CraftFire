<script setup lang="ts">
/**
 * AdminPanel - 管理员控制面板。
 * 提供房间管理、玩家管理、踢人、公告等功能。
 */
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useUIStore } from '@/stores/ui'
import { useRoomStore } from '@/stores/room'
import { usePlayerStore } from '@/stores/player'
import { wsService } from '@/services/WebSocketService'

const ui = useUIStore()
const roomStore = useRoomStore()
const playerStore = usePlayerStore()

// 管理员密码
const adminPassword = ref('')
const isAuthenticated = ref(false)
const authError = ref('')
const isLoading = ref(false)

// 活动标签页
type TabType = 'players' | 'actions' | 'logs'
const activeTab = ref<TabType>('players')

// 公告内容
const broadcastMessage = ref('')
const showBroadcastModal = ref(false)

// 锁定房间
const isRoomLocked = ref(false)

// 在线玩家列表
const onlinePlayers = ref<Array<{
  id: string
  name: string
  health: number
  isBot: boolean
  isMuted: boolean
  position?: { x: number; y: number; z: number }
}>>([])

// 操作日志
interface LogEntry {
  id: number
  time: string
  action: string
  target?: string
  type: 'kick' | 'mute' | 'broadcast' | 'lock' | 'unlock' | 'mode' | 'other'
}
const logs = ref<LogEntry[]>([])
let logIdCounter = 0

// 选中的玩家详情
const selectedPlayer = ref<typeof onlinePlayers.value[0] | null>(null)

// 刷新间隔
let refreshInterval: ReturnType<typeof setInterval> | null = null

// 房间信息
const roomInfo = computed(() => roomStore.currentRoom)

// 添加日志
function addLog(action: string, target?: string, type: LogEntry['type'] = 'other') {
  const now = new Date()
  const time = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`
  logs.value.unshift({
    id: ++logIdCounter,
    time,
    action,
    target,
    type,
  })
  // 只保留最近50条日志
  if (logs.value.length > 50) {
    logs.value.pop()
  }
}

// 管理员密码验证
async function handleAuth() {
  if (!adminPassword.value.trim()) {
    authError.value = '请输入管理员密码'
    return
  }

  isLoading.value = true
  authError.value = ''

  try {
    await new Promise(resolve => setTimeout(resolve, 500))

    if (adminPassword.value === 'admin') {
      isAuthenticated.value = true
      ui.showToast('管理员验证成功', 'success')
      startPlayerRefresh()
      addLog('管理员登录')
    } else {
      authError.value = '密码错误'
      addLog('登录失败 - 密码错误', undefined, 'other')
    }
  } catch {
    authError.value = '验证失败，请重试'
  } finally {
    isLoading.value = false
  }
}

// 开始刷新玩家列表
function startPlayerRefresh() {
  refreshOnlinePlayers()
  refreshInterval = setInterval(refreshOnlinePlayers, 3000)
}

// 刷新在线玩家
function refreshOnlinePlayers() {
  const players: typeof onlinePlayers.value = []

  // 添加本地玩家 (房主)
  if (playerStore.localPlayer) {
    players.push({
      id: playerStore.localPlayer.id,
      name: playerStore.localPlayer.name + ' ★',
      health: playerStore.localPlayer.health,
      isBot: false,
      isMuted: false,
      position: playerStore.localPlayer.position,
    })
  }

  // 添加远程玩家
  playerStore.remotePlayers.forEach((p) => {
    players.push({
      id: p.id,
      name: p.name,
      health: p.health,
      isBot: false,
      isMuted: false,
      position: p.position,
    })
  })

  onlinePlayers.value = players
}

// 选中玩家查看详情
function selectPlayer(player: typeof onlinePlayers.value[0]) {
  selectedPlayer.value = selectedPlayer.value?.id === player.id ? null : player
}

// 踢出玩家
async function kickPlayer(playerId: string, playerName: string) {
  if (!confirm(`确定要踢出玩家 "${playerName}" 吗？`)) return

  isLoading.value = true
  try {
    wsService.send('admin_kick', {
      targetPlayerId: playerId,
      reason: '违反游戏规则',
    })

    onlinePlayers.value = onlinePlayers.value.filter(p => p.id !== playerId)
    playerStore.removeRemotePlayer(playerId)

    addLog('踢出玩家', playerName, 'kick')
    ui.showToast(`已踢出玩家 ${playerName}`, 'success')

    if (selectedPlayer.value?.id === playerId) {
      selectedPlayer.value = null
    }
  } catch {
    ui.showToast('踢出玩家失败', 'error')
  } finally {
    isLoading.value = false
  }
}

// 静音玩家
function mutePlayer(playerId: string, playerName: string) {
  const player = onlinePlayers.value.find(p => p.id === playerId)
  if (!player) return

  if (player.isMuted) {
    player.isMuted = false
    addLog('取消静音', playerName, 'mute')
    ui.showToast(`已取消静音 ${playerName}`, 'success')
  } else {
    player.isMuted = true
    wsService.send('admin_mute', {
      targetPlayerId: playerId,
      durationSeconds: 300,
    })
    addLog('静音玩家', playerName, 'mute')
    ui.showToast(`已静音玩家 ${playerName}`, 'success')
  }
}

// 发送公告
function sendBroadcast() {
  if (!broadcastMessage.value.trim()) {
    ui.showToast('公告内容不能为空', 'warning')
    return
  }

  wsService.send('admin_broadcast', {
    message: broadcastMessage.value.trim(),
  })

  addLog('发送公告', broadcastMessage.value.trim(), 'broadcast')
  ui.showToast('公告已发送', 'success')
  broadcastMessage.value = ''
  showBroadcastModal.value = false
}

// 锁定/解锁房间
function toggleRoomLock() {
  isRoomLocked.value = !isRoomLocked.value
  if (isRoomLocked.value) {
    addLog('锁定房间', undefined, 'lock')
    ui.showToast('房间已锁定', 'success')
  } else {
    addLog('解锁房间', undefined, 'unlock')
    ui.showToast('房间已解锁', 'success')
  }
}

// 切换游戏模式
function changeGameMode(mode: string) {
  if (!roomInfo.value) return

  wsService.send('admin_change_mode', { mode })

  addLog('切换模式', mode, 'mode')
  ui.showToast(`游戏模式已切换为 ${mode}`, 'success')
}

// 给玩家加血
function healPlayer(playerId: string, playerName: string) {
  const player = onlinePlayers.value.find(p => p.id === playerId)
  if (player) {
    player.health = 100
    wsService.send('admin_heal', { targetPlayerId: playerId })
    addLog('治疗玩家', playerName)
    ui.showToast(`已为 ${playerName} 恢复满血`, 'success')
  }
}

// 传送玩家到自己位置
function teleportToMe(playerId: string, playerName: string) {
  wsService.send('admin_teleport', {
    targetPlayerId: playerId,
    toPosition: playerStore.localPlayer?.position || { x: 0, y: 10, z: 0 },
  })
  addLog('传送玩家', playerName)
  ui.showToast(`已传送 ${playerName} 到身边`, 'success')
}

// 关闭面板
function closePanel() {
  ui.showAdminPanel = false
  selectedPlayer.value = null
}

// 退出登录
function logout() {
  isAuthenticated.value = false
  adminPassword.value = ''
  selectedPlayer.value = null
  if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
  addLog('管理员登出')
}

// 获取血量颜色
function getHealthColor(health: number): string {
  if (health > 60) return '#22c55e'
  if (health > 30) return '#eab308'
  return '#ef4444'
}

// 获取日志图标
function getLogIcon(type: LogEntry['type']): string {
  switch (type) {
    case 'kick': return '👢'
    case 'mute': return '🔇'
    case 'broadcast': return '📢'
    case 'lock': return '🔒'
    case 'unlock': return '🔓'
    case 'mode': return '⚔️'
    default: return '⚙️'
  }
}

onMounted(() => {
  if (isAuthenticated.value) {
    startPlayerRefresh()
  }
  // 禁用 Ctrl+A 全选
  document.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'a') {
      e.preventDefault()
    }
  })
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<template>
  <div v-if="ui.showAdminPanel" class="overlay flex items-center justify-center p-4">
    <div class="panel-lowpoly w-full max-w-3xl max-h-[85vh] overflow-hidden flex flex-col pointer-events-auto">

      <!-- 标题栏 -->
      <div class="flex items-center justify-between mb-4 pb-3 border-b border-white/10">
        <h2 class="text-craft-primary text-lg font-game flex items-center gap-2">
          <span>⚙️</span>
          <span>管理员面板</span>
        </h2>
        <button class="btn-lowpoly px-3 py-1 text-xs" @click="closePanel">关闭</button>
      </div>

      <!-- 未验证状态：显示登录表单 -->
      <div v-if="!isAuthenticated" class="flex flex-col items-center py-8">
        <div class="text-5xl mb-6">🔐</div>
        <h3 class="text-craft-light text-sm font-game mb-6">请输入管理员密码</h3>

        <form @submit.prevent="handleAuth" class="w-full max-w-xs">
          <input
            v-model="adminPassword"
            type="password"
            class="input-lowpoly w-full mb-3 text-center"
            placeholder="管理员密码"
            autocomplete="off"
          />

          <p v-if="authError" class="text-red-400 text-xs mb-3 font-game text-center">
            {{ authError }}
          </p>

          <button type="submit" class="btn-primary w-full" :disabled="isLoading">
            {{ isLoading ? '验证中...' : '进入管理面板' }}
          </button>
        </form>

        <p class="text-craft-text/40 text-xs mt-6 font-game text-center">
          提示：在房间中连续点击版本号 5 次可快速打开
        </p>
      </div>

      <!-- 已验证状态：显示管理面板 -->
      <div v-else class="flex-1 overflow-hidden flex flex-col">

        <!-- 标签页 -->
        <div class="flex gap-1 mb-4 border-b border-white/10 pb-2">
          <button
            class="px-4 py-2 text-xs font-game transition-colors"
            :class="activeTab === 'players' ? 'text-craft-primary bg-craft-surface/50' : 'text-craft-text/60 hover:text-craft-light'"
            @click="activeTab = 'players'"
          >
            👥 玩家 ({{ onlinePlayers.length }})
          </button>
          <button
            class="px-4 py-2 text-xs font-game transition-colors"
            :class="activeTab === 'actions' ? 'text-craft-primary bg-craft-surface/50' : 'text-craft-text/60 hover:text-craft-light'"
            @click="activeTab = 'actions'"
          >
            ⚡ 操作
          </button>
          <button
            class="px-4 py-2 text-xs font-game transition-colors"
            :class="activeTab === 'logs' ? 'text-craft-primary bg-craft-surface/50' : 'text-craft-text/60 hover:text-craft-light'"
            @click="activeTab = 'logs'"
          >
            📋 日志 ({{ logs.length }})
          </button>
        </div>

        <!-- 玩家列表 -->
        <div v-if="activeTab === 'players'" class="flex-1 overflow-y-auto">
          <!-- 房间信息栏 -->
          <div class="flex items-center justify-between mb-3 p-2 bg-craft-dark/30 rounded border border-white/10">
            <div class="flex items-center gap-4 text-xs font-game">
              <span class="text-craft-text/60">房间:</span>
              <span class="text-craft-secondary">{{ roomInfo?.roomId || '未加入' }}</span>
              <span class="text-craft-text/40">|</span>
              <span class="text-craft-text/60">人数:</span>
              <span class="text-craft-light">{{ onlinePlayers.length }}/{{ roomInfo?.maxPlayers || 0 }}</span>
              <span class="text-craft-text/40">|</span>
              <span class="text-craft-text/60">模式:</span>
              <span class="text-craft-light">{{ roomInfo?.gameMode || '-' }}</span>
            </div>
            <button
              class="btn-lowpoly px-2 py-1 text-xs"
              @click="refreshOnlinePlayers"
            >
              🔄 刷新
            </button>
          </div>

          <!-- 玩家卡片列表 -->
          <div class="space-y-2">
            <div
              v-for="player in onlinePlayers"
              :key="player.id"
              class="p-3 rounded border transition-all cursor-pointer"
              :class="selectedPlayer?.id === player.id ? 'bg-craft-surface/50 border-craft-primary' : 'bg-craft-dark/30 border-white/10 hover:border-white/20'"
              @click="selectPlayer(player)"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded bg-craft-surface flex items-center justify-center text-lg">
                    {{ player.isBot ? '🤖' : '🎮' }}
                  </div>
                  <div>
                    <div class="text-craft-light text-xs font-game flex items-center gap-2">
                      {{ player.name }}
                      <span v-if="player.isMuted" class="text-red-400 text-xs">🔇</span>
                    </div>
                    <div class="flex items-center gap-2 mt-1">
                      <div class="w-16 h-1.5 bg-craft-dark rounded-full overflow-hidden">
                        <div
                          class="h-full transition-all"
                          :style="{ width: `${player.health}%`, backgroundColor: getHealthColor(player.health) }"
                        />
                      </div>
                      <span class="text-craft-text/40 text-xs">{{ player.health }}%</span>
                    </div>
                  </div>
                </div>

                <div class="flex gap-1" @click.stop>
                  <button
                    class="btn-lowpoly px-2 py-1 text-xs"
                    :class="player.isMuted ? 'text-red-400' : ''"
                    @click="mutePlayer(player.id, player.name)"
                    title="静音"
                  >
                    {{ player.isMuted ? '🔊' : '🔇' }}
                  </button>
                  <button
                    class="btn-danger px-2 py-1 text-xs"
                    @click="kickPlayer(player.id, player.name)"
                    title="踢出"
                  >
                    踢
                  </button>
                </div>
              </div>

              <!-- 玩家详情展开 -->
              <div v-if="selectedPlayer?.id === player.id" class="mt-3 pt-3 border-t border-white/10" @click.stop>
                <div class="grid grid-cols-2 gap-2 text-xs font-game mb-3">
                  <div class="text-craft-text/60">ID:</div>
                  <div class="text-craft-light truncate">{{ player.id }}</div>
                  <div class="text-craft-text/60">位置:</div>
                  <div class="text-craft-light">
                    {{ player.position ? `${player.position.x.toFixed(1)}, ${player.position.y.toFixed(1)}, ${player.position.z.toFixed(1)}` : '未知' }}
                  </div>
                </div>

                <div class="flex gap-2">
                  <button
                    class="btn-info px-2 py-1 text-xs flex-1"
                    @click="healPlayer(player.id, player.name)"
                  >
                    ❤️ 加血
                  </button>
                  <button
                    class="btn-info px-2 py-1 text-xs flex-1"
                    @click="teleportToMe(player.id, player.name)"
                  >
                    📍 传送
                  </button>
                </div>
              </div>
            </div>

            <div v-if="onlinePlayers.length === 0" class="text-center py-12 text-craft-text/40 text-xs font-game">
              暂无在线玩家
            </div>
          </div>
        </div>

        <!-- 管理操作 -->
        <div v-if="activeTab === 'actions'" class="flex-1 overflow-y-auto">
          <div class="grid grid-cols-2 gap-3">
            <!-- 发送公告 -->
            <button
              class="p-4 rounded border border-white/10 bg-craft-dark/30 hover:bg-craft-surface/30 transition-colors text-left"
              @click="showBroadcastModal = true"
            >
              <div class="text-2xl mb-2">📢</div>
              <div class="text-craft-light text-xs font-game">发送公告</div>
              <div class="text-craft-text/40 text-xs mt-1">向所有玩家广播消息</div>
            </button>

            <!-- 锁定房间 -->
            <button
              class="p-4 rounded border border-white/10 bg-craft-dark/30 hover:bg-craft-surface/30 transition-colors text-left"
              @click="toggleRoomLock"
            >
              <div class="text-2xl mb-2">{{ isRoomLocked ? '🔓' : '🔒' }}</div>
              <div class="text-craft-light text-xs font-game">{{ isRoomLocked ? '解锁房间' : '锁定房间' }}</div>
              <div class="text-craft-text/40 text-xs mt-1">{{ isRoomLocked ? '允许玩家加入' : '禁止新玩家加入' }}</div>
            </button>

            <!-- 切换模式 -->
            <div class="p-4 rounded border border-white/10 bg-craft-dark/30">
              <div class="text-2xl mb-2">⚔️</div>
              <div class="text-craft-light text-xs font-game mb-2">游戏模式</div>
              <div class="flex gap-1">
                <button
                  v-for="mode in ['sandbox', 'survival', 'pvp']"
                  :key="mode"
                  class="btn-lowpoly px-2 py-1 text-xs flex-1"
                  :class="roomInfo?.gameMode === mode ? 'btn-primary' : ''"
                  @click="changeGameMode(mode)"
                >
                  {{ mode === 'sandbox' ? '沙盒' : mode === 'survival' ? '生存' : 'PVP' }}
                </button>
              </div>
            </div>

            <!-- 清空日志 -->
            <button
              class="p-4 rounded border border-white/10 bg-craft-dark/30 hover:bg-craft-surface/30 transition-colors text-left"
              @click="logs = []"
            >
              <div class="text-2xl mb-2">🗑️</div>
              <div class="text-craft-light text-xs font-game">清空日志</div>
              <div class="text-craft-text/40 text-xs mt-1">清除所有操作记录</div>
            </button>
          </div>
        </div>

        <!-- 操作日志 -->
        <div v-if="activeTab === 'logs'" class="flex-1 overflow-y-auto">
          <div v-if="logs.length > 0" class="space-y-1">
            <div
              v-for="log in logs"
              :key="log.id"
              class="flex items-start gap-2 p-2 rounded bg-craft-dark/30 text-xs font-game"
            >
              <span class="text-lg">{{ getLogIcon(log.type) }}</span>
              <div class="flex-1">
                <div class="text-craft-light">{{ log.action }}</div>
                <div v-if="log.target" class="text-craft-text/60 text-xs mt-0.5 truncate">
                  {{ log.target }}
                </div>
              </div>
              <span class="text-craft-text/40 text-xs">{{ log.time }}</span>
            </div>
          </div>
          <div v-else class="text-center py-12 text-craft-text/40 text-xs font-game">
            暂无操作记录
          </div>
        </div>
      </div>

      <!-- 底部退出登录 -->
      <div v-if="isAuthenticated" class="mt-4 pt-3 border-t border-white/10">
        <button class="btn-secondary w-full py-2 text-xs" @click="logout">
          退出登录
        </button>
      </div>
    </div>

    <!-- 发送公告弹窗 -->
    <div v-if="showBroadcastModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="showBroadcastModal = false">
      <div class="panel-lowpoly w-full max-w-md p-6">
        <h3 class="text-craft-primary text-sm font-game mb-4">📢 发送公告</h3>
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
</template>
