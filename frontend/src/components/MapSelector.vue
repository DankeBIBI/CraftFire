<script setup lang="ts">
/**
 * MapSelector - 地图选择界面。
 * 低多边形卡片风格，展示所有可用地图（俯视预览图、难度、人数）。
 */
import { ref, computed, onMounted } from 'vue'
import { MAPS, getMapById, type MapDefinition } from '@/maps/index'
import { generateMapPreview } from '@/maps/preview'

const emit = defineEmits<{
  (e: 'select', mapId: string): void
  (e: 'back'): void
}>()

const selectedId = ref('dust2')
const previews = ref<Record<string, string>>({})
const hoveredId = ref<string | null>(null)

// 难度颜色
const DIFFICULTY_COLORS: Record<string, string> = {
  easy: '#4CAF50',
  medium: '#FFC107',
  hard: '#F44336',
}

const DIFFICULTY_LABELS: Record<string, string> = {
  easy: '简单',
  medium: '中等',
  hard: '困难',
}

// 选中地图的详细信息
const selectedMap = computed(() => getMapById(selectedId.value))

onMounted(() => {
  // 生成所有地图的俯视预览图
  for (const map of MAPS) {
    try {
      const blocks = map.generate()
      previews.value[map.id] = generateMapPreview(blocks, 128)
    } catch (e) {
      console.warn(`Map preview failed for ${map.id}:`, e)
    }
  }
})

function selectMap(id: string) {
  selectedId.value = id
}

function confirmSelection() {
  emit('select', selectedId.value)
}
</script>

<template>
  <div class="map-selector panel-lowpoly w-full max-w-4xl mx-auto p-6">
    <!-- 标题 -->
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-xl text-craft-primary font-game">选择地图</h2>
      <button class="btn-lowpoly px-4 py-2 text-xs" @click="emit('back')">
        ← 返回
      </button>
    </div>

    <!-- 地图网格 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-6">
      <div
        v-for="map in MAPS"
        :key="map.id"
        class="map-card relative cursor-pointer transition-all duration-200"
        :class="{
          'ring-2 ring-craft-primary scale-[1.02]': selectedId === map.id,
          'border-2 border-transparent hover:border-craft-primary/50 hover:scale-[1.01]': selectedId !== map.id,
        }"
        @click="selectMap(map.id)"
        @mouseenter="hoveredId = map.id"
        @mouseleave="hoveredId = null"
      >
        <!-- 预览图 -->
        <div class="relative overflow-hidden rounded-t"
          :style="{ backgroundColor: map.environment.fogColor + '33' }">
          <img
            v-if="previews[map.id]"
            :src="previews[map.id]"
            :alt="map.name"
            class="w-full h-32 object-cover object-center"
            style="image-rendering: pixelated;"
          />
          <!-- 无预览占位 -->
          <div
            v-else
            class="w-full h-32 flex items-center justify-center"
            :style="{ background: `linear-gradient(135deg, ${map.environment.skyColor}44, ${map.environment.fogColor}44)` }"
          >
            <span class="text-craft-text/30 text-3xl">🗺️</span>
          </div>

          <!-- 选中标记 -->
          <div
            v-if="selectedId === map.id"
            class="absolute top-2 right-2 w-6 h-6 bg-craft-primary rounded-full flex items-center justify-center"
          >
            <span class="text-white text-sm font-bold">✓</span>
          </div>

          <!-- 难度角标 -->
          <div
            class="absolute bottom-2 left-2 px-2 py-0.5 rounded text-[10px] font-game text-white"
            :style="{ backgroundColor: DIFFICULTY_COLORS[map.difficulty] }"
          >
            {{ DIFFICULTY_LABELS[map.difficulty] }}
          </div>
        </div>

        <!-- 地图信息 -->
        <div class="bg-craft-surface p-3 rounded-b border-2 border-black">
          <h3 class="text-craft-light text-xs font-game mb-1 truncate">{{ map.name }}</h3>
          <p class="text-craft-text/50 text-[10px] font-game truncate">{{ map.description }}</p>
          <div class="flex items-center justify-between mt-2 text-[10px] font-game text-craft-text/60">
            <span>👥 {{ map.recommendedPlayers }}</span>
            <span>📐 {{ map.size }}</span>
            <span>🧊 ~{{ map.estimatedBlocks }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 选中地图详情 -->
    <div
      v-if="selectedMap"
      class="bg-craft-dark/60 border-2 border-craft-primary/40 rounded p-4 mb-4"
    >
      <div class="flex items-start gap-4">
        <!-- 俯视预览 -->
        <div class="flex-shrink-0">
          <img
            v-if="previews[selectedMap.id]"
            :src="previews[selectedMap.id]"
            :alt="selectedMap.name"
            class="w-24 h-24 rounded border-2 border-black object-cover"
            style="image-rendering: pixelated;"
          />
        </div>
        <div class="flex-1">
          <div class="flex items-center gap-3 mb-2">
            <h3 class="text-craft-primary font-game text-sm">{{ selectedMap.name }}</h3>
            <span
              class="px-2 py-0.5 rounded text-[10px] font-game text-white"
              :style="{ backgroundColor: DIFFICULTY_COLORS[selectedMap.difficulty] }"
            >
              {{ DIFFICULTY_LABELS[selectedMap.difficulty] }}
            </span>
            <span class="text-craft-text/40 text-xs font-game">by {{ selectedMap.author }}</span>
          </div>
          <p class="text-craft-text/70 text-xs font-game leading-relaxed mb-3">
            {{ selectedMap.description }}
          </p>
          <div class="flex gap-4 text-[10px] font-game text-craft-text/50">
            <span>👥 推荐 {{ selectedMap.recommendedPlayers }} 人</span>
            <span>📐 {{ selectedMap.size }}</span>
            <span>🧊 {{ selectedMap.estimatedBlocks }} 方块</span>
          </div>
          <!-- 环境预览 -->
          <div class="flex gap-2 mt-2 items-center">
            <div class="text-[10px] font-game text-craft-text/40">环境：</div>
            <div
              class="w-4 h-4 rounded border border-white/20"
              :style="{ backgroundColor: selectedMap.environment.skyColor }"
              title="天空色"
            />
            <div
              class="w-4 h-4 rounded border border-white/20"
              :style="{ backgroundColor: selectedMap.environment.fogColor }"
              title="雾色"
            />
            <div
              class="w-4 h-4 rounded border border-white/20"
              :style="{ backgroundColor: selectedMap.environment.ambientColor }"
              title="环境光"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 确认按钮 -->
    <button
      class="btn-primary w-full py-3 text-sm font-game"
      @click="confirmSelection"
    >
      使用此地图 →
    </button>
  </div>
</template>

<style scoped>
.map-selector {
  max-height: 90vh;
  overflow-y: auto;
}

.map-card {
  background: var(--craft-surface);
  border-radius: 0;
  box-shadow: 4px 4px 0 #000;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.map-card:hover {
  box-shadow: 6px 6px 0 #000;
}

.map-card.ring-2 {
  box-shadow: 6px 6px 0 var(--craft-primary);
}
</style>
