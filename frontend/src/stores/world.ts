/**
 * 世界（体素）状态管理。
 * 管理方块数据、区块加载和世界变化。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { BlockData, BlockType } from "@/types/game";
import type { ChunkData, WorldChangeEntry } from "@/types/world";

/** 方块 key：用坐标作为唯一标识 */
function blockKey(x: number, y: number, z: number): string {
  return `${x},${y},${z}`;
}

export const useWorldStore = defineStore("world", () => {
  // ─── 状态 ────────────────────────────────
  const blocks = ref<Map<string, BlockData>>(new Map());
  const loadedChunks = ref<Map<string, ChunkData>>(new Map());
  const pendingChanges = ref<WorldChangeEntry[]>([]);
  const worldSeed = ref("");
  const isWorldLoaded = ref(false);

  // ─── 计算属性 ────────────────────────────
  const totalBlocks = computed(() => blocks.value.size);
  const loadedChunkCount = computed(() => loadedChunks.value.size);

  // ─── 操作 ────────────────────────────────

  /** 放置方块（本地预测） */
  function placeBlock(x: number, y: number, z: number, type: BlockType, playerId = "local") {
    const key = blockKey(x, y, z);
    const block: BlockData = { x, y, z, type };
    blocks.value.set(key, block);
    pendingChanges.value.push({
      x,
      y,
      z,
      blockType: type,
      action: "place",
      timestamp: Date.now(),
      playerId,
    });
  }

  /** 移除方块（本地预测） */
  function removeBlock(x: number, y: number, z: number) {
    const key = blockKey(x, y, z);
    const existing = blocks.value.get(key);
    if (existing) {
      blocks.value.delete(key);
      pendingChanges.value.push({
        x,
        y,
        z,
        blockType: existing.type,
        action: "remove",
        timestamp: Date.now(),
      });
    }
  }

  /** 获取某位置的方块 */
  function getBlock(x: number, y: number, z: number): BlockData | undefined {
    return blocks.value.get(blockKey(x, y, z));
  }

  /** 应用服务器下发的世界更新 */
  function applyWorldUpdate(changes: WorldChangeEntry[]) {
    for (const change of changes) {
      const key = blockKey(change.x, change.y, change.z);
      if (change.action === "place") {
        blocks.value.set(key, {
          x: change.x,
          y: change.y,
          z: change.z,
          type: change.blockType,
        });
      } else if (change.action === "remove") {
        blocks.value.delete(key);
      }
    }
  }

  /** 加载区块数据 */
  function loadChunk(chunk: ChunkData) {
    const chunkKey = `${chunk.x},${chunk.z}`;
    loadedChunks.value.set(chunkKey, chunk);
    // 将区块内方块加入 blocks 索引
    for (const block of chunk.blocks) {
      blocks.value.set(blockKey(block.x, block.y, block.z), block);
    }
  }

  /** 卸载区块 */
  function unloadChunk(chunkX: number, chunkZ: number) {
    const chunkKey = `${chunkX},${chunkZ}`;
    const chunk = loadedChunks.value.get(chunkKey);
    if (chunk) {
      for (const block of chunk.blocks) {
        blocks.value.delete(blockKey(block.x, block.y, block.z));
      }
      loadedChunks.value.delete(chunkKey);
    }
  }

  /** 初始化世界（加载快照） */
  function initializeWorld(seed: string, initialBlocks: BlockData[]) {
    worldSeed.value = seed;
    blocks.value.clear();
    for (const block of initialBlocks) {
      blocks.value.set(blockKey(block.x, block.y, block.z), block);
    }
    isWorldLoaded.value = true;
  }

  /** 清空待处理变更 */
  function clearPendingChanges() {
    pendingChanges.value = [];
  }

  function $reset() {
    blocks.value.clear();
    loadedChunks.value.clear();
    pendingChanges.value = [];
    worldSeed.value = "";
    isWorldLoaded.value = false;
  }

  return {
    // State
    blocks,
    loadedChunks,
    pendingChanges,
    worldSeed,
    isWorldLoaded,
    // Computed
    totalBlocks,
    loadedChunkCount,
    // Actions
    placeBlock,
    removeBlock,
    getBlock,
    applyWorldUpdate,
    loadChunk,
    unloadChunk,
    initializeWorld,
    clearPendingChanges,
    $reset,
  };
});
