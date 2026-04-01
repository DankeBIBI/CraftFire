/**
 * 地图 2D 俯视预览生成器。
 * 将 3D 体素地图投影为 2D canvas 图像。
 */
import type { BlockData } from '@/types/game'

/** 俯视图配色表（与 maps/index.ts 中的 block type 对应） */
const PREVIEW_COLORS: Record<string, string> = {
  // 沙漠
  sand: '#D4A556',
  sandDark: '#B8894A',
  stone: '#8B7355',
  stoneDark: '#6B5344',
  wood: '#8B6914',
  crate: '#A0522D',
  metal: '#696969',
  // 雪地
  snow: '#F0F4F8',
  snowDark: '#C8D6E5',
  ice: '#A8D8EA',
  iceBlock: '#89CFF0',
  pineDark: '#2D5016',
  pineLight: '#4A7C23',
  // 丛林
  jungleGrass: '#3D8B37',
  jungleDirt: '#6B4226',
  jungleStone: '#5A5A5A',
  vine: '#2E7D32',
  // 霓虹
  neonPink: '#FF2D78',
  neonBlue: '#00F5FF',
  neonPurple: '#B026FF',
  neonOrange: '#FF6B35',
  neonGreen: '#39FF14',
  gridBase: '#0A0A1A',
  gridLine: '#1A1A3A',
  // 岩浆
  lavaRock: '#3D3D3D',
  lavaRockLight: '#5C4033',
  lava: '#FF4500',
  lavaGlow: '#FF6600',
  basalt: '#2C2C2C',
  obsidian: '#1A0A2E',
  // 其他
  air: 'transparent',
}

/** 获取俯视图高度（用于阴影深度） */
function getElevationColor(baseColor: string, y: number): string {
  if (y <= 0) return baseColor
  // 越高越亮
  const lift = Math.min(y * 8, 40)
  return adjustBrightness(baseColor, lift)
}

/** 调整颜色亮度 */
function adjustBrightness(hex: string, amount: number): string {
  const c = parseInt(hex.replace('#', ''), 16)
  if (isNaN(c)) return hex
  const r = Math.min(255, Math.max(0, ((c >> 16) & 0xff) + amount))
  const g = Math.min(255, Math.max(0, ((c >> 8) & 0xff) + amount))
  const b = Math.min(255, Math.max(0, (c & 0xff) + amount))
  return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`
}

/** 为地图生成俯视预览 canvas（data URL） */
export function generateMapPreview(
  blocks: BlockData[],
  canvasSize = 128,
): string {
  // 统计每个 (x, z) 的最高 y 层
  const heightMap = new Map<string, { y: number; type: string }>()
  for (const b of blocks) {
    if (b.type === 'air') continue
    const key = `${b.x},${b.z}`
    const existing = heightMap.get(key)
    if (!existing || b.y > existing.y) {
      heightMap.set(key, { y: b.y, type: b.type })
    }
  }

  if (heightMap.size === 0) return ''

  // 计算包围盒
  let minX = Infinity, maxX = -Infinity
  let minZ = Infinity, maxZ = -Infinity
  for (const key of heightMap.keys()) {
    const [x, z] = key.split(',').map(Number)
    minX = Math.min(minX, x)
    maxX = Math.max(maxX, x)
    minZ = Math.min(minZ, z)
    maxZ = Math.max(maxZ, z)
  }

  const worldW = maxX - minX + 1
  const worldH = maxZ - minZ + 1
  const scale = canvasSize / Math.max(worldW, worldH)

  const canvas = document.createElement('canvas')
  canvas.width = canvasSize
  canvas.height = canvasSize
  const ctx = canvas.getContext('2d')!
  ctx.imageSmoothingEnabled = false

  // 背景色（根据地图主题）
  ctx.fillStyle = '#1a1a2e'
  ctx.fillRect(0, 0, canvasSize, canvasSize)

  // 绘制每个区块
  for (const [key, data] of heightMap.entries()) {
    const [x, z] = key.split(',').map(Number)
    const px = (x - minX) * scale
    const pz = (z - minZ) * scale
    const size = Math.max(scale * 0.9, 1)

    const baseColor = PREVIEW_COLORS[data.type] ?? '#888888'
    const color = getElevationColor(baseColor, data.y)

    ctx.fillStyle = color
    ctx.fillRect(Math.round(px), Math.round(pz), Math.ceil(size), Math.ceil(size))
  }

  // 边框
  ctx.strokeStyle = 'rgba(255,255,255,0.15)'
  ctx.lineWidth = 1
  ctx.strokeRect(0, 0, canvasSize, canvasSize)

  return canvas.toDataURL('image/png')
}
