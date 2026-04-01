/**
 * 机器人管理器
 * 负责管理 AI 机器人的生成、更新和寻路逻辑
 */
import { ref } from 'vue'
import type { PlayerState } from '@/types/player'
import type { Vector3 } from '@/types/game'
import { useWorldStore } from '@/stores/world'

export interface Bot {
  id: string
  state: PlayerState
  targetPosition: Vector3 | null
  moveSpeed: number
  changeTargetInterval: number
  lastTargetChange: number
}

const bots = ref<Map<string, Bot>>(new Map())
const isInitialized = ref(false)

/** 生成随机 ID */
function generateBotId(): string {
  return `bot_${Math.random().toString(36).substring(2, 9)}`
}

/** 生成随机位置 */
function getRandomPosition(): Vector3 {
  const range = 12
  const x = (Math.random() - 0.5) * range * 2
  const z = (Math.random() - 0.5) * range * 2
  return { x, y: 1, z }
}

/** 生成随机机器人名称 */
function getRandomBotName(): string {
  const adjectives = ['勇敢', '敏捷', '神秘', '疯狂', '冷静', '迅猛', '坚韧', '灵活']
  const nouns = ['猎手', '战士', '游侠', '刺客', '先锋', '守卫', '突击手', '观察者']
  const adj = adjectives[Math.floor(Math.random() * adjectives.length)]
  const noun = nouns[Math.floor(Math.random() * nouns.length)]
  const num = Math.floor(Math.random() * 100)
  return `${adj}${noun}${num}`
}

/** 初始化机器人 */
function initBots(): void {
  if (isInitialized.value) return

  const worldStore = useWorldStore()
  const botCount = Math.floor(Math.random() * 6) + 5 // 5-10 个随机

  for (let i = 0; i < botCount; i++) {
    const botId = generateBotId()
    const startPos = getRandomPosition()

    const botState: PlayerState = {
      id: botId,
      name: getRandomBotName(),
      position: startPos,
      velocity: { x: 0, y: 0, z: 0 },
      rotation: { pitch: 0, yaw: Math.random() * Math.PI * 2, roll: 0 },
      health: 100,
      ammo: 30,
      equipment: 'pistol',
      isAlive: true,
      lastUpdateTime: Date.now(),
    }

    const bot: Bot = {
      id: botId,
      state: botState,
      targetPosition: null,
      moveSpeed: 2 + Math.random() * 2, // 2-4 移动速度
      changeTargetInterval: 2000 + Math.random() * 3000, // 2-5 秒改变目标
      lastTargetChange: Date.now(),
    }

    bots.value.set(botId, bot)
  }

  isInitialized.value = true
}

/** 更新机器人位置和寻路 */
function updateBots(deltaTime: number): void {
  const now = Date.now()

  bots.value.forEach((bot) => {
    if (!bot.state.isAlive) return

    // 检查是否需要更新目标位置
    if (!bot.targetPosition || now - bot.lastTargetChange > bot.changeTargetInterval) {
      bot.targetPosition = getRandomPosition()
      bot.lastTargetChange = now
    }

    if (!bot.targetPosition) return

    // 计算移动方向
    const dx = bot.targetPosition.x - bot.state.position.x
    const dz = bot.targetPosition.z - bot.state.position.z
    const distance = Math.sqrt(dx * dx + dz * dz)

    if (distance > 0.1) {
      // 移动机器人
      const moveAmount = bot.moveSpeed * deltaTime
      const ratio = Math.min(moveAmount / distance, 1)

      bot.state.position.x += dx * ratio
      bot.state.position.z += dz * ratio

      // 朝向目标方向
      bot.state.rotation.yaw = Math.atan2(dx, dz)
    }

    // 更新速度（用于网络同步）
    bot.state.velocity.x = (bot.targetPosition.x - bot.state.position.x) * 0.1
    bot.state.velocity.z = (bot.targetPosition.z - bot.state.position.z) * 0.1

    bot.state.lastUpdateTime = now
  })
}

/** 获取所有机器人状态 */
function getBotStates(): PlayerState[] {
  const states: PlayerState[] = []
  bots.value.forEach((bot) => {
    states.push({ ...bot.state })
  })
  return states
}

/** 获取机器人 Map */
function getBots(): Map<string, Bot> {
  return bots.value
}

/** 移除所有机器人 */
function clearBots(): void {
  bots.value.clear()
  isInitialized.value = false
}

/** 获取机器人数量 */
function getBotCount(): number {
  return bots.value.size
}

export const botManager = {
  bots,
  isInitialized,
  initBots,
  updateBots,
  getBotStates,
  getBots,
  clearBots,
  getBotCount,
}