<script setup lang="ts">
/**
 * RemotePlayer - 远程玩家渲染组件。
 * 对远程玩家状态进行插值平滑后渲染，使用体素角色模型。
 */
import { ref, watch, onUnmounted } from 'vue'
import { useRenderLoop } from '@tresjs/core'
import type { PlayerState } from '@/types/player'
import { interpolationEngine } from '@/services/InterpolationEngine'
import VoxelCharacter from './VoxelCharacter.vue'

const props = defineProps<{
  player: PlayerState
}>()

const interpolatedPosition = ref({ ...props.player.position })
const interpolatedRotation = ref({ ...props.player.rotation })

// 接收到新状态时推入插值引擎
watch(
  () => [props.player.position, props.player.rotation] as const,
  ([position, rotation]) => {
    interpolationEngine.pushState(
      props.player.id,
      { ...position },
      { ...rotation },
    )
  },
  { immediate: true }
)

const { onLoop } = useRenderLoop()

onLoop(() => {
  const state = interpolationEngine.getInterpolatedState(props.player.id)
  if (state) {
    interpolatedPosition.value = state.position
    interpolatedRotation.value = state.rotation
  }
})

onUnmounted(() => {
  interpolationEngine.removeEntity(props.player.id)
})
</script>

<template>
  <!-- VoxelCharacter 直接接收插值后的状态渲染 -->
  <VoxelCharacter
    :player="{
      ...player,
      position: interpolatedPosition,
      rotation: interpolatedRotation,
    }"
  />
</template>
