<script setup lang="ts">
/**
 * PlayerEntity - 玩家 3D 实体（第三人称可见身体）。
 * 使用体素角色模型 VoxelCharacter 渲染。
 */
import type { PlayerState } from '@/types/player'
import VoxelCharacter from './VoxelCharacter.vue'

const props = defineProps<{
  player: PlayerState
  isLocal?: boolean
}>()
</script>

<template>
  <!--
    本地玩家在第一人称时不渲染身体（由 ScenePistol 的手臂替代）。
    只有非本地玩家（远程玩家/Bot）才渲染完整身体。
  -->
  <VoxelCharacter
    v-if="!isLocal"
    :player="player"
    :is-local="isLocal"
  />
</template>
