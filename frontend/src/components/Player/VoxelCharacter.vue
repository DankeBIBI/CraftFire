<script setup lang="ts">
/**
 * VoxelCharacter - 体素风格角色模型组件。
 *
 * 设计参考 Minecraft Steve 的方块比例感，但加入战术原创设计：
 * - 战术护目镜 + 口罩（区别于 Steve）
 * - 战术背心（原创）
 * - 独立四肢组 → 支持行走动画
 * - 玩家专属配色（从 ID 派生，同一玩家颜色固定）
 *
 * 构成（站立姿态）：
 *   头部 y=1.62  (0.4×0.4×0.4)
 *   颈部 y=1.42
 *   躯干 y=1.10  (0.5×0.65×0.3)
 *   左/右臂挂点 y=1.20
 *   髋部 y=0.78
 *   左/右腿挂点 y=0.78
 */
import { ref, computed, onUnmounted } from 'vue'
import { useRenderLoop } from '@tresjs/core'
import type { PlayerState } from '@/types/player'

// ─── 颜色工具 ───────────────────────────────────────

/** 从字符串生成 HSL 颜色（同一 ID 永远相同颜色） */
function idToColor(id: string, saturation = 65, lightness = 55): string {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash)
    hash |= 0
  }
  const h = Math.abs(hash) % 360
  return `hsl(${h}, ${saturation}%, ${lightness}%)`
}

// ─── 材质颜色常量 ────────────────────────────────────

const SKIN_COLOR = '#D4A574'
const MASK_COLOR = '#5B8DB8'
const VEST_COLOR = '#4A5568'
const VEST_ACCENT_COLOR = '#2D3748'
const GOGGLES_COLOR = '#48BB78'
const GOGGLES_FRAME = '#1A202C'
const BOOT_COLOR = '#3D2914'
const BELT_COLOR = '#5C4A1E'
const GEAR_COLOR = '#2D3748'

// ─── 组件 ────────────────────────────────────────────

const props = defineProps<{
  player: PlayerState
  isLocal?: boolean
}>()

// ─── 动画状态 ────────────────────────────────────────

const walkCycle = ref(0)

// ─── 玩家配色 ────────────────────────────────────────

const primaryColor = computed(() => idToColor(props.player.id, 60, 52))
const pantsColor = computed(() => idToColor(props.player.id, 40, 35))

// ─── 死亡状态 ────────────────────────────────────────

const isDead = computed(() => !props.player.isAlive)

// ─── 行走动画（响应式驱动） ────────────────────────────

const leftArmRotX = ref(0)
const rightArmRotX = ref(0)
const leftLegRotX = ref(0)
const rightLegRotX = ref(0)
const bodyBobY = ref(0)

const { onLoop } = useRenderLoop()

onLoop(({ delta }) => {
  const moving =
    Math.abs(props.player.velocity.x) > 0.05 ||
    Math.abs(props.player.velocity.z) > 0.05

  if (moving) {
    walkCycle.value += delta * 8
  } else {
    walkCycle.value *= 0.88
    if (walkCycle.value < 0.05) walkCycle.value = 0
  }

  const wc = walkCycle.value
  leftArmRotX.value = Math.sin(wc) * 0.45
  rightArmRotX.value = -Math.sin(wc) * 0.45
  leftLegRotX.value = -Math.sin(wc) * 0.35
  rightLegRotX.value = Math.sin(wc) * 0.35
  bodyBobY.value = Math.abs(Math.sin(wc * 2)) * 0.025
})

onUnmounted(() => {
  walkCycle.value = 0
})
</script>

<template>
  <!-- 整体旋转跟随 yaw -->
  <TresGroup
    :position="[player.position.x, player.position.y, player.position.z]"
    :rotation="[0, player.rotation.yaw, 0]"
  >
    <!-- 上下跳动 -->
    <TresGroup :position="[0, isDead ? -0.3 : -bodyBobY, 0]">

      <!-- ══════════════════════════════════════════════ -->
      <!-- 头部                                            -->
      <!-- ══════════════════════════════════════════════ -->
      <TresGroup :position="[0, 1.62, 0]">
        <!-- 主头块 -->
        <TresMesh>
          <TresBoxGeometry :args="[0.4, 0.4, 0.4]" />
          <TresMeshLambertMaterial :color="primaryColor" />
        </TresMesh>

        <!-- 头发层 -->
        <TresMesh :position="[0, 0.18, -0.01]">
          <TresBoxGeometry :args="[0.42, 0.08, 0.38]" />
          <TresMeshLambertMaterial :color="isDead ? '#888' : '#3D2314'" />
        </TresMesh>

        <!-- 护目镜左 -->
        <TresMesh :position="[-0.08, 0.04, 0.2]">
          <TresBoxGeometry :args="[0.14, 0.1, 0.06]" />
          <TresMeshLambertMaterial :color="isDead ? '#666' : GOGGLES_FRAME" />
        </TresMesh>
        <TresMesh :position="[-0.08, 0.04, 0.225]">
          <TresBoxGeometry :args="[0.1, 0.07, 0.01]" />
          <TresMeshLambertMaterial
            :color="isDead ? '#555' : GOGGLES_COLOR"
            :transparent="true"
            :opacity="isDead ? 0.3 : 0.85"
          />
        </TresMesh>

        <!-- 护目镜右 -->
        <TresMesh :position="[0.08, 0.04, 0.2]">
          <TresBoxGeometry :args="[0.14, 0.1, 0.06]" />
          <TresMeshLambertMaterial :color="isDead ? '#666' : GOGGLES_FRAME" />
        </TresMesh>
        <TresMesh :position="[0.08, 0.04, 0.225]">
          <TresBoxGeometry :args="[0.1, 0.07, 0.01]" />
          <TresMeshLambertMaterial
            :color="isDead ? '#555' : GOGGLES_COLOR"
            :transparent="true"
            :opacity="isDead ? 0.3 : 0.85"
          />
        </TresMesh>

        <!-- 护目镜横梁 -->
        <TresMesh :position="[0, 0.04, 0.2]">
          <TresBoxGeometry :args="[0.04, 0.04, 0.06]" />
          <TresMeshLambertMaterial :color="isDead ? '#555' : GOGGLES_FRAME" />
        </TresMesh>

        <!-- 口罩 -->
        <TresMesh :position="[0, -0.1, 0.2]">
          <TresBoxGeometry :args="[0.36, 0.12, 0.06]" />
          <TresMeshLambertMaterial :color="isDead ? '#555' : MASK_COLOR" />
        </TresMesh>
        <TresMesh :position="[0, -0.04, 0.2]">
          <TresBoxGeometry :args="[0.38, 0.04, 0.06]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : '#3D6B8E'" />
        </TresMesh>

        <!-- 口罩固定带 -->
        <TresMesh :position="[-0.19, 0, 0.18]">
          <TresBoxGeometry :args="[0.04, 0.28, 0.04]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : '#2C5282'" />
        </TresMesh>
        <TresMesh :position="[0.19, 0, 0.18]">
          <TresBoxGeometry :args="[0.04, 0.28, 0.04]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : '#2C5282'" />
        </TresMesh>
      </TresGroup>

      <!-- 颈部 -->
      <TresMesh :position="[0, 1.42, 0]">
        <TresBoxGeometry :args="[0.1, 0.12, 0.1]" />
        <TresMeshLambertMaterial :color="isDead ? '#666' : SKIN_COLOR" />
      </TresMesh>

      <!-- ══════════════════════════════════════════════ -->
      <!-- 躯干 + 战术背心                                -->
      <!-- ══════════════════════════════════════════════ -->
      <TresGroup :position="[0, 1.10, 0]">
        <!-- 背心底层 -->
        <TresMesh>
          <TresBoxGeometry :args="[0.5, 0.65, 0.3]" />
          <TresMeshLambertMaterial :color="isDead ? '#555' : VEST_COLOR" />
        </TresMesh>

        <!-- 背心护甲 -->
        <TresMesh :position="[0, 0.05, 0.16]">
          <TresBoxGeometry :args="[0.42, 0.48, 0.04]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : VEST_ACCENT_COLOR" />
        </TresMesh>

        <!-- 肩垫 -->
        <TresMesh :position="[-0.28, 0.28, 0.02]">
          <TresBoxGeometry :args="[0.1, 0.16, 0.3]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : VEST_ACCENT_COLOR" />
        </TresMesh>
        <TresMesh :position="[0.28, 0.28, 0.02]">
          <TresBoxGeometry :args="[0.1, 0.16, 0.3]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : VEST_ACCENT_COLOR" />
        </TresMesh>

        <!-- MOLLE 条带 -->
        <TresMesh
          v-for="i in 3"
          :key="'molle-'+i"
          :position="[0, 0.18 - (i-1)*0.14, 0.155]"
        >
          <TresBoxGeometry :args="[0.38, 0.04, 0.02]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : '#1A202C'" />
        </TresMesh>

        <!-- 口袋 -->
        <TresMesh :position="[-0.12, -0.08, 0.155]">
          <TresBoxGeometry :args="[0.16, 0.14, 0.04]" />
          <TresMeshLambertMaterial :color="isDead ? '#3a3a3a' : '#374151'" />
        </TresMesh>
        <TresMesh :position="[0.12, -0.08, 0.155]">
          <TresBoxGeometry :args="[0.16, 0.14, 0.04]" />
          <TresMeshLambertMaterial :color="isDead ? '#3a3a3a' : '#374151'" />
        </TresMesh>

        <!-- 腰带 -->
        <TresMesh :position="[0, -0.3, 0]">
          <TresBoxGeometry :args="[0.5, 0.08, 0.3]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : BELT_COLOR" />
        </TresMesh>

        <!-- 裤子 -->
        <TresGroup :position="[0, -0.34, 0]">
          <TresMesh>
            <TresBoxGeometry :args="[0.22, 0.65, 0.22]" />
            <TresMeshLambertMaterial :color="isDead ? '#444' : pantsColor" />
          </TresMesh>
          <TresMesh :position="[0.22, 0, 0]">
            <TresBoxGeometry :args="[0.22, 0.65, 0.22]" />
            <TresMeshLambertMaterial :color="isDead ? '#444' : pantsColor" />
          </TresMesh>
        </TresGroup>
      </TresGroup>

      <!-- ══════════════════════════════════════════════ -->
      <!-- 左臂（绕 X 轴旋转）                           -->
      <!-- ══════════════════════════════════════════════ -->
      <TresGroup :position="[-0.32, 1.20, 0]" :rotation="[leftArmRotX, 0, 0]">
        <!-- 大臂 -->
        <TresMesh :position="[0, -0.16, 0]">
          <TresBoxGeometry :args="[0.14, 0.32, 0.14]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : VEST_ACCENT_COLOR" />
        </TresMesh>
        <!-- 护具 -->
        <TresMesh :position="[0, -0.36, 0]">
          <TresBoxGeometry :args="[0.14, 0.1, 0.14]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : GEAR_COLOR" />
        </TresMesh>
        <!-- 手掌 -->
        <TresMesh :position="[0, -0.48, 0]">
          <TresBoxGeometry :args="[0.12, 0.14, 0.1]" />
          <TresMeshLambertMaterial :color="isDead ? '#555' : SKIN_COLOR" />
        </TresMesh>
      </TresGroup>

      <!-- ══════════════════════════════════════════════ -->
      <!-- 右臂（绕 X 轴旋转）                           -->
      <!-- ══════════════════════════════════════════════ -->
      <TresGroup :position="[0.32, 1.20, 0]" :rotation="[rightArmRotX, 0, 0]">
        <!-- 大臂 -->
        <TresMesh :position="[0, -0.16, 0]">
          <TresBoxGeometry :args="[0.14, 0.32, 0.14]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : VEST_ACCENT_COLOR" />
        </TresMesh>
        <!-- 护具 -->
        <TresMesh :position="[0, -0.36, 0]">
          <TresBoxGeometry :args="[0.14, 0.1, 0.14]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : GEAR_COLOR" />
        </TresMesh>
        <!-- 手掌 -->
        <TresMesh :position="[0, -0.48, 0]">
          <TresBoxGeometry :args="[0.12, 0.14, 0.1]" />
          <TresMeshLambertMaterial :color="isDead ? '#555' : SKIN_COLOR" />
        </TresMesh>
      </TresGroup>

      <!-- ══════════════════════════════════════════════ -->
      <!-- 左腿（绕 X 轴旋转）                           -->
      <!-- ══════════════════════════════════════════════ -->
      <TresGroup :position="[-0.14, 0.78, 0]" :rotation="[leftLegRotX, 0, 0]">
        <!-- 大腿 -->
        <TresMesh :position="[0, -0.16, 0]">
          <TresBoxGeometry :args="[0.2, 0.32, 0.2]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : pantsColor" />
        </TresMesh>
        <!-- 护膝 -->
        <TresMesh :position="[0, -0.3, 0.06]">
          <TresBoxGeometry :args="[0.2, 0.1, 0.08]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : GEAR_COLOR" />
        </TresMesh>
        <!-- 靴子 -->
        <TresMesh :position="[0, -0.46, 0]">
          <TresBoxGeometry :args="[0.18, 0.3, 0.18]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : BOOT_COLOR" />
        </TresMesh>
        <!-- 靴底 -->
        <TresMesh :position="[0, -0.62, 0.02]">
          <TresBoxGeometry :args="[0.2, 0.06, 0.24]" />
          <TresMeshLambertMaterial :color="isDead ? '#2a2a2a' : '#1A0F0A'" />
        </TresMesh>
      </TresGroup>

      <!-- ══════════════════════════════════════════════ -->
      <!-- 右腿（绕 X 轴旋转）                           -->
      <!-- ══════════════════════════════════════════════ -->
      <TresGroup :position="[0.14, 0.78, 0]" :rotation="[rightLegRotX, 0, 0]">
        <!-- 大腿 -->
        <TresMesh :position="[0, -0.16, 0]">
          <TresBoxGeometry :args="[0.2, 0.32, 0.2]" />
          <TresMeshLambertMaterial :color="isDead ? '#444' : pantsColor" />
        </TresMesh>
        <!-- 护膝 -->
        <TresMesh :position="[0, -0.3, 0.06]">
          <TresBoxGeometry :args="[0.2, 0.1, 0.08]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : GEAR_COLOR" />
        </TresMesh>
        <!-- 靴子 -->
        <TresMesh :position="[0, -0.46, 0]">
          <TresBoxGeometry :args="[0.18, 0.3, 0.18]" />
          <TresMeshLambertMaterial :color="isDead ? '#333' : BOOT_COLOR" />
        </TresMesh>
        <!-- 靴底 -->
        <TresMesh :position="[0, -0.62, 0.02]">
          <TresBoxGeometry :args="[0.2, 0.06, 0.24]" />
          <TresMeshLambertMaterial :color="isDead ? '#2a2a2a' : '#1A0F0A'" />
        </TresMesh>
      </TresGroup>

    </TresGroup>
  </TresGroup>
</template>
