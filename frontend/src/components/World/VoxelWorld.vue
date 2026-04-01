<script setup lang="ts">
/**
 * VoxelWorld - 体素世界渲染器。
 * 根据 worldStore 中的方块数据渲染所有可见方块（InstancedMesh 批渲染）。
 * 根据传入 mapId 应用对应的环境配置（天空色、雾效、光照）。
 */
import { watch, onMounted, onUnmounted, ref } from 'vue'
import {
  BoxGeometry,
  Color,
  Fog,
  Group,
  InstancedMesh,
  Matrix4,
  MeshLambertMaterial,
  Object3D,
} from 'three'
import { useWorldStore } from '@/stores/world'
import { getMapById } from '@/maps/index'
import { logger } from '@/utils/logger'
import type { BlockData } from '@/types/game'

const props = defineProps<{
  mapId?: string
}>()

const worldStore = useWorldStore()
const { scene, camera } = useTresContext()

const voxelGroup = new Group()
voxelGroup.name = 'voxel-world'

const boxGeometry = new BoxGeometry(1, 1, 1)
const tempMatrix = new Matrix4()
const materialCache = new Map<string, MeshLambertMaterial>()

let attachedScene: Object3D | null = null
let unwatchScene: (() => void) | null = null
let unwatchBlocks: (() => void) | null = null
let rebuildQueued = false

// ─── 全地图方块颜色表 ────────────────────────────────

const BLOCK_COLORS: Record<string, string> = {
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
}

function getBlockColor(type: string): string {
  return BLOCK_COLORS[type] ?? '#888888'
}

function getMaterial(type: string): MeshLambertMaterial {
  const cached = materialCache.get(type)
  if (cached) return cached
  const material = new MeshLambertMaterial({ color: getBlockColor(type) })
  materialCache.set(type, material)
  return material
}

// ─── 环境配置 ───────────────────────────────────────

const currentEnv = ref(worldStore.worldSeed)

function applyEnvironment(mapId: string) {
  if (!scene.value) return

  const map = getMapById(mapId)
  if (!map) return

  const env = map.environment
  currentEnv.value = mapId

  const sc = scene.value as unknown as { background?: Color; fog?: Fog }

  // 天空颜色
  if (sc.background instanceof Color) {
    sc.background.set(env.skyColor)
  } else {
    sc.background = new Color(env.skyColor)
  }

  // 雾效
  sc.fog = new Fog(env.fogColor, env.fogNear, env.fogFar)

  // 方向光
  const dirLightObj = scene.value.getObjectByName('sun-light')
  if (dirLightObj) {
    const dl = dirLightObj as unknown as { color?: Color; intensity?: number }
    dl.color?.set(env.directionalColor)
    if (dl.intensity !== undefined) dl.intensity = env.directionalIntensity
  }

  // 环境光
  const ambLightObj = scene.value.getObjectByName('ambient-light')
  if (ambLightObj) {
    const al = ambLightObj as unknown as { color?: Color; intensity?: number }
    al.color?.set(env.ambientColor)
    if (al.intensity !== undefined) al.intensity = env.ambientIntensity
  }

  logger.info('VoxelWorld', `环境已切换: ${map.name}`)
}

// ─── 体素渲染 ───────────────────────────────────────

function isBlockVisible(blockMap: Map<string, BlockData>, block: BlockData): boolean {
  return !(
    blockMap.has(`${block.x + 1},${block.y},${block.z}`) &&
    blockMap.has(`${block.x - 1},${block.y},${block.z}`) &&
    blockMap.has(`${block.x},${block.y + 1},${block.z}`) &&
    blockMap.has(`${block.x},${block.y - 1},${block.z}`) &&
    blockMap.has(`${block.x},${block.y},${block.z + 1}`) &&
    blockMap.has(`${block.x},${block.y},${block.z - 1}`)
  )
}

function clearVoxelMeshes(): void {
  for (let i = voxelGroup.children.length - 1; i >= 0; i--) {
    voxelGroup.remove(voxelGroup.children[i])
  }
}

function rebuildVoxelMeshes(): void {
  clearVoxelMeshes()

  const blockMap = worldStore.blocks as Map<string, BlockData>
  const groupedBlocks = new Map<string, BlockData[]>()

  blockMap.forEach((block) => {
    if (!isBlockVisible(blockMap, block)) return
    const list = groupedBlocks.get(block.type)
    if (list) {
      list.push(block)
    } else {
      groupedBlocks.set(block.type, [block])
    }
  })

  groupedBlocks.forEach((blocks, type) => {
    if (!blocks.length) return

    const mesh = new InstancedMesh(boxGeometry, getMaterial(type), blocks.length)
    mesh.name = `voxels-${type}`

    for (let i = 0; i < blocks.length; i++) {
      const block = blocks[i]
      tempMatrix.makeTranslation(block.x + 0.5, block.y + 0.5, block.z + 0.5)
      mesh.setMatrixAt(i, tempMatrix)
    }

    mesh.instanceMatrix.needsUpdate = true
    voxelGroup.add(mesh)
  })
}

function queueRebuild(): void {
  if (rebuildQueued) return
  rebuildQueued = true
  requestAnimationFrame(() => {
    rebuildQueued = false
    rebuildVoxelMeshes()
  })
}

function attachToScene(sceneObject: Object3D | null): void {
  if (!sceneObject || attachedScene === sceneObject) return
  if (attachedScene) attachedScene.remove(voxelGroup)
  sceneObject.add(voxelGroup)
  attachedScene = sceneObject
}

// ─── 生命周期 ───────────────────────────────────────

onMounted(() => {
  logger.info('VoxelWorld', '体素世界挂载')

  unwatchScene = watch(
    () => scene.value,
    (sceneObject) => {
      attachToScene((sceneObject as unknown as Object3D | null) ?? null)
      if (props.mapId) applyEnvironment(props.mapId)
    },
    { immediate: true },
  )

  unwatchBlocks = watch(
    () => worldStore.blocks,
    () => { queueRebuild() },
    { deep: true, immediate: true },
  )
})

// 地图切换时更新环境
watch(() => props.mapId, (newId) => {
  if (newId) applyEnvironment(newId)
}, { immediate: true })

onUnmounted(() => {
  if (unwatchScene) { unwatchScene(); unwatchScene = null }
  if (unwatchBlocks) { unwatchBlocks(); unwatchBlocks = null }
  clearVoxelMeshes()
  if (attachedScene) { attachedScene.remove(voxelGroup); attachedScene = null }
  for (const mat of materialCache.values()) mat.dispose()
  materialCache.clear()
  boxGeometry.dispose()
})
</script>

<template />
