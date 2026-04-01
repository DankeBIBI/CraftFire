/**
 * 地图 JSON 加载器。
 * 读取 RLE 压缩的地图 JSON，展开为 BlockData[]。
 */
import type { BlockData, BlockType } from '@/types/game'
import type { MapEnvironment } from './index'

// ─── JSON Schema 类型 ────────────────────────────────

export interface ScatterRule {
  on: string
  type: string
  probability: number
  yOffset: number
}

export interface MapJSON {
  id: string
  name: string
  description: string
  author: string
  difficulty: 'easy' | 'medium' | 'hard'
  recommendedPlayers: string
  size: string
  estimatedBlocks: number
  environment: MapEnvironment
  layers: LayerDef[]
  structures: StructureDef[]
  /** 散落规则（确定性，按坐标哈希决定是否放置） */
  scatters?: ScatterRule[]
}

export interface LayerDef {
  y: number
  type: string
  x1: number
  z1: number
  x2: number
  z2: number
}

export interface StructureDef {
  type: string
  x1: number
  y1: number
  z1: number
  x2: number
  y2: number
  z2: number
}

// ─── 加载器 ────────────────────────────────────────

/** 哈希确定性：同一地图同一坐标永远得到相同结果 */
function shouldScatter(mapId: string, x: number, z: number, probability: number): boolean {
  const hash = Math.abs(
    (mapId.charCodeAt(0) * 73856093) ^
    (x * 19349663) ^
    (z * 83492791)
  ) % 100
  return hash < probability * 100
}

/**
 * 将 RLE JSON 解压为 BlockData[]。
 * - 跳过 "air" 类型的空洞标记
 * - 散落规则（scatters）按坐标哈希确定性生成
 */
export function loadMapBlocks(map: MapJSON): BlockData[] {
  const blocks: BlockData[] = []
  const blockSet = new Set<string>() // 去重

  function addBlock(x: number, y: number, z: number, type: string) {
    const key = `${x},${y},${z}`
    if (!blockSet.has(key)) {
      blockSet.add(key)
      blocks.push({ x, y, z, type: type as BlockType })
    }
  }

  for (const layer of map.layers) {
    if (layer.type === 'air') continue
    const { y, type, x1, z1, x2, z2 } = layer
    for (let x = x1; x <= x2; x++) {
      for (let z = z1; z <= z2; z++) {
        addBlock(x, y, z, type)
      }
    }
  }

  for (const s of map.structures) {
    if (s.type === 'air') continue
    for (let x = s.x1; x <= s.x2; x++) {
      for (let y = s.y1; y <= s.y2; y++) {
        for (let z = s.z1; z <= s.z2; z++) {
          addBlock(x, y, z, s.type)
        }
      }
    }
  }

  // 散落规则：确定性哈希，不走随机数
  if (map.scatters) {
    for (const sc of map.scatters) {
      for (const b of blocks) {
        if (b.type === sc.on && shouldScatter(map.id, b.x, b.z, sc.probability)) {
          addBlock(b.x, b.y + sc.yOffset, b.z, sc.type)
        }
      }
    }
  }

  return blocks
}
