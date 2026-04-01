<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { useTresContext, useRenderLoop } from '@tresjs/core'
import {
  AdditiveBlending,
  Group,
  Mesh,
  MeshBasicMaterial,
  Object3D,
  SphereGeometry,
  Vector3,
} from 'three'
import { useSettingsStore } from '@/stores/settings'
import { SPARK_EFFECT_EVENT } from '@/utils/sparkEffects'
import type { SparkEffectPayload } from '@/utils/sparkEffects'

interface SparkParticle {
  mesh: Mesh
  velocity: Vector3
  life: number
  maxLife: number
  baseScale: number
}

const { scene } = useTresContext()
const { onLoop } = useRenderLoop()
const settings = useSettingsStore()

const sparkGroup = new Group()
sparkGroup.name = 'SparkEffects'

const sparkGeometry = new SphereGeometry(1, 4, 4)
const muzzleMaterial = new MeshBasicMaterial({
  color: '#ffb347',
  transparent: true,
  opacity: 0.9,
  depthWrite: false,
  blending: AdditiveBlending,
})
const impactMaterial = new MeshBasicMaterial({
  color: '#ffd966',
  transparent: true,
  opacity: 0.85,
  depthWrite: false,
  blending: AdditiveBlending,
})

const particles: SparkParticle[] = []

let attachedScene: Object3D | null = null
let unwatchScene: (() => void) | null = null

function createSparkParticle(
  position: { x: number; y: number; z: number },
  velocity: Vector3,
  size: number,
  maxLife: number,
  material: MeshBasicMaterial,
): void {
  const mesh = new Mesh(sparkGeometry, material)
  mesh.position.set(position.x, position.y, position.z)
  mesh.scale.setScalar(size)
  mesh.frustumCulled = false
  sparkGroup.add(mesh)

  particles.push({
    mesh,
    velocity,
    life: 0,
    maxLife,
    baseScale: size,
  })
}

function spawnMuzzleSpark(payload: SparkEffectPayload): void {
  const normal = payload.normal
    ? new Vector3(payload.normal.x, payload.normal.y, payload.normal.z).normalize()
    : new Vector3(0, 0, -1)

  for (let i = 0; i < 10; i++) {
    const spread = new Vector3(
      (Math.random() - 0.5) * 0.6,
      (Math.random() - 0.5) * 0.5,
      (Math.random() - 0.5) * 0.6,
    )
    const velocity = normal
      .clone()
      .add(spread)
      .normalize()
      .multiplyScalar(4.5 + Math.random() * 2.5)

    createSparkParticle(payload.position, velocity, 0.016 + Math.random() * 0.018, 0.08 + Math.random() * 0.06, muzzleMaterial)
  }
}

function spawnImpactSpark(payload: SparkEffectPayload): void {
  const normal = payload.normal
    ? new Vector3(payload.normal.x, payload.normal.y, payload.normal.z).normalize()
    : new Vector3(0, 1, 0)

  for (let i = 0; i < 16; i++) {
    const spread = new Vector3(
      (Math.random() - 0.5) * 1.8,
      (Math.random() - 0.5) * 1.6,
      (Math.random() - 0.5) * 1.8,
    )
    const velocity = normal
      .clone()
      .multiplyScalar(0.8)
      .add(spread)
      .normalize()
      .multiplyScalar(2.2 + Math.random() * 2.8)

    createSparkParticle(payload.position, velocity, 0.022 + Math.random() * 0.02, 0.14 + Math.random() * 0.12, impactMaterial)
  }
}

function onSparkEvent(event: Event): void {
  if (!settings.video.particleEffects) return

  const customEvent = event as CustomEvent<SparkEffectPayload>
  if (!customEvent.detail?.position) return

  if (customEvent.detail.kind === 'muzzle') {
    spawnMuzzleSpark(customEvent.detail)
  } else {
    spawnImpactSpark(customEvent.detail)
  }
}

function attachToScene(sceneObject: Object3D | null): void {
  if (!sceneObject || attachedScene === sceneObject) return

  if (attachedScene) {
    attachedScene.remove(sparkGroup)
  }

  sceneObject.add(sparkGroup)
  attachedScene = sceneObject
}

function updateParticles(dt: number): void {
  const frameDt = Math.min(dt, 0.05)

  for (let i = particles.length - 1; i >= 0; i--) {
    const particle = particles[i]
    particle.life += frameDt

    if (particle.life >= particle.maxLife) {
      sparkGroup.remove(particle.mesh)
      particles.splice(i, 1)
      continue
    }

    particle.velocity.y -= 10 * frameDt
    particle.velocity.multiplyScalar(1 - frameDt * 2.4)
    particle.mesh.position.addScaledVector(particle.velocity, frameDt)

    const lifeRatio = 1 - particle.life / particle.maxLife
    particle.mesh.scale.setScalar(Math.max(0.001, particle.baseScale * lifeRatio))
  }
}

onLoop(({ delta }) => {
  updateParticles(delta)
})

onMounted(() => {
  unwatchScene = watch(
    () => scene.value,
    (sceneObject) => {
      attachToScene((sceneObject as unknown as Object3D | null) ?? null)
    },
    { immediate: true },
  )

  window.addEventListener(SPARK_EFFECT_EVENT, onSparkEvent as EventListener)
})

onUnmounted(() => {
  if (unwatchScene) {
    unwatchScene()
    unwatchScene = null
  }

  window.removeEventListener(SPARK_EFFECT_EVENT, onSparkEvent as EventListener)

  for (const particle of particles) {
    sparkGroup.remove(particle.mesh)
  }
  particles.length = 0

  if (attachedScene) {
    attachedScene.remove(sparkGroup)
    attachedScene = null
  }

  sparkGeometry.dispose()
  muzzleMaterial.dispose()
  impactMaterial.dispose()
})
</script>

<template />
