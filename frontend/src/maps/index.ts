/**
 * CraftFire 地图注册表。
 * 所有地图定义从 JSON 文件加载，解压后供游戏使用。
 */
import type { BlockData } from '@/types/game'
import { loadMapBlocks, type MapJSON } from './loader'
import { generateRandomMapBlocks } from './randomGenerator'

// ─── 类型定义 ────────────────────────────────────────

export interface MapEnvironment {
  skyColor: string
  fogColor: string
  fogNear: number
  fogFar: number
  ambientIntensity: number
  ambientColor: string
  directionalIntensity: number
  directionalColor: string
  directionalPosition: [number, number, number]
}

export interface MapDefinition {
  id: string
  name: string
  description: string
  author: string
  difficulty: 'easy' | 'medium' | 'hard'
  recommendedPlayers: string
  size: string
  estimatedBlocks: number
  environment: MapEnvironment
  /** 解压后的方块数据 */
  generate: () => BlockData[]
}

// ─── 加载所有地图 JSON ──────────────────────────────

// Vite glob 导入（静态分析，build 时打包）
const mapModules = import.meta.glob<{ default: MapJSON }>(
  '@/assets/maps/*.json',
  { eager: true }
)

// ─── 转为 MapDefinition ─────────────────────────────

function buildMapDef(json: MapJSON): MapDefinition {
  return {
    id: json.id,
    name: json.name,
    description: json.description,
    author: json.author,
    difficulty: json.difficulty,
    recommendedPlayers: json.recommendedPlayers,
    size: json.size,
    estimatedBlocks: json.estimatedBlocks,
    environment: json.environment,
    generate: () => loadMapBlocks(json),
  }
}

// 按稳定顺序排列（与原 MAPS 一致）
const MAP_ID_ORDER = ['dust2', 'frozen_peak', 'jungle_temple', 'neon_grid', 'molten_core']

const STATIC_MAPS: MapDefinition[] = MAP_ID_ORDER
  .map((id) => {
    const entry = Object.entries(mapModules).find(([path]) =>
      path.includes(`/${id}.json`)
    )
    if (!entry) { console.warn(`[maps] JSON not found: ${id}`); return null }
    const json: MapJSON = entry[1].default
    return buildMapDef(json)
  })
  .filter(Boolean) as MapDefinition[]

const RANDOM_MAP: MapDefinition = {
  id: 'random_world',
  name: '随机世界 · Random World',
  description: '完全随机生成山地、树木与建筑。每次开局地图都不同。',
  author: 'Procedural Generator',
  difficulty: 'medium',
  recommendedPlayers: '2-12',
  size: '80x80',
  estimatedBlocks: 9000,
  environment: {
    skyColor: '#9CCBFF',
    fogColor: '#AFC9B8',
    fogNear: 70,
    fogFar: 240,
    ambientIntensity: 0.52,
    ambientColor: '#EAF6D7',
    directionalIntensity: 0.85,
    directionalColor: '#FFF0C8',
    directionalPosition: [50, 95, 40],
  },
  generate: () => generateRandomMapBlocks(),
}

export const MAPS: MapDefinition[] = [...STATIC_MAPS, RANDOM_MAP]

/** 根据 ID 查找地图 */
export function getMapById(id: string): MapDefinition | undefined {
  return MAPS.find((m) => m.id === id)
}

/** 默认地图 */
export const DEFAULT_MAP = MAPS[0]
