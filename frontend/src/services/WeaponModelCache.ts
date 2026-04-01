/**
 * 武器模型缓存服务
 * 复用 GLTF 模型，避免每个远程玩家重复加载
 */
import { Group, Mesh, Material } from 'three'
import type { WeaponSlotConfig } from '@/composables/useWeapon'
import pistolConfigs from '@/loader/weapon/pistol'
import { loadGLTF } from '@/loader/GLTFLoader'
import { logger } from '@/utils/logger'

interface CachedModel {
  group: Group
  refCount: number
}

/** 武器模型缓存 */
const modelCache = new Map<string, CachedModel>()

/** 武器配置映射 */
const weaponConfigMap: Record<string, WeaponSlotConfig> = {
  pistol: pistolConfigs.lowpoly_generic_pistol_43,
  // 可扩展更多武器...
}

/**
 * 获取武器模型（带缓存）
 * @param weaponType 武器类型标识
 * @returns 武器 Group 克隆
 */
export function getWeaponModel(weaponType: string): Group | null {
  const config = weaponConfigMap[weaponType] || weaponConfigMap.pistol
  const cacheKey = config.url

  // 检查缓存
  const cached = modelCache.get(cacheKey)
  if (cached) {
    cached.refCount++
    const clone = cached.group.clone()
    // 克隆材质（避免共享状态）
    clone.traverse((node) => {
      if ((node as Mesh).isMesh) {
        const mesh = node as Mesh
        if (mesh.material) {
          mesh.material = (mesh.material as Material).clone()
        }
      }
    })
    return clone
  }

  // 同步返回 null（模型将异步加载）
  return null
}

/**
 * 预加载武器模型到缓存
 * @param weaponType 武器类型标识
 */
export async function preloadWeaponModel(weaponType: string): Promise<void> {
  const config = weaponConfigMap[weaponType] || weaponConfigMap.pistol
  const cacheKey = config.url

  // 已有缓存，跳过
  if (modelCache.has(cacheKey)) {
    return
  }

  logger.info('[WeaponModelCache] 预加载武器:', cacheKey)

  const [err, gltf] = await loadGLTF(config)
  if (err || !gltf) {
    logger.warn('[WeaponModelCache] 武器模型加载失败:', cacheKey)
    return
  }

  // 应用网格设置
  gltf.scene.traverse((node) => {
    if ((node as Mesh).isMesh) {
      const mesh = node as Mesh
      mesh.castShadow = false
      mesh.receiveShadow = false
      mesh.frustumCulled = false
    }
  })

  modelCache.set(cacheKey, {
    group: gltf.scene,
    refCount: 0,
  })

  logger.info('[WeaponModelCache] 武器模型已缓存:', cacheKey)
}

/**
 * 释放武器模型（减少引用计数）
 * @param weaponType 武器类型标识
 */
export function releaseWeaponModel(weaponType: string): void {
  const config = weaponConfigMap[weaponType] || weaponConfigMap.pistol
  const cacheKey = config.url

  const cached = modelCache.get(cacheKey)
  if (cached) {
    cached.refCount--
    if (cached.refCount <= 0) {
      // 引用归零，可选择清理缓存
      // 这里暂时保留缓存，只打印日志
      logger.info('[WeaponModelCache] 武器模型引用归零:', cacheKey)
    }
  }
}

/**
 * 预加载所有武器模型
 */
export async function preloadAllWeapons(): Promise<void> {
  const promises = Object.keys(weaponConfigMap).map((type) =>
    preloadWeaponModel(type)
  )
  await Promise.all(promises)
  logger.info('[WeaponModelCache] 所有武器模型预加载完成')
}
