/**
 * 玩家状态管理。
 * 管理本地玩家和远程玩家的状态数据。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { PlayerState } from "@/types/player";
import type { Vector3, Rotation } from "@/types/game";

export const usePlayerStore = defineStore("player", () => {
  const DEFAULT_MAG_SIZE = 30;
  const SAFE_SPAWN_POSITION = { x: 0, y: 10, z: 0 };

  // ─── 本地玩家状态 ────────────────────────
  const localPlayer = ref<PlayerState | null>(null);
  const playerId = ref("");
  const playerName = ref("Player");

  // ─── 远程玩家 ────────────────────────────
  const remotePlayers = ref<Map<string, PlayerState>>(new Map());

  // ─── 计算属性 ────────────────────────────
  const isAlive = computed(() => localPlayer.value?.isAlive ?? false);
  const health = computed(() => localPlayer.value?.health ?? 100);
  const position = computed<Vector3>(
    () => localPlayer.value?.position ?? { x: 0, y: 10, z: 0 },
  );
  const rotation = computed<Rotation>(
    () => localPlayer.value?.rotation ?? { pitch: 0, yaw: 0, roll: 0 },
  );
  const remotePlayerCount = computed(() => remotePlayers.value.size);
  const allPlayers = computed<PlayerState[]>(() => {
    const list: PlayerState[] = [];
    if (localPlayer.value) list.push(localPlayer.value);
    remotePlayers.value.forEach((p) => list.push(p));
    return list;
  });

  // ─── 操作 ────────────────────────────────
  function initLocalPlayer(id: string, name: string) {
    playerId.value = id;
    playerName.value = name;
    localPlayer.value = {
      id,
      name,
      position: { ...SAFE_SPAWN_POSITION }, // 避免出生点被地图方块占用
      velocity: { x: 0, y: 0, z: 0 },
      rotation: { pitch: 0, yaw: 0, roll: 0 },
      health: 100,
      ammo: 30,
      equipment: "pistol",
      isAlive: true,
      lastUpdateTime: Date.now(),
    };
  }

  function updateLocalPosition(pos: Vector3) {
    if (localPlayer.value) {
      localPlayer.value.position = { ...pos };
      localPlayer.value.lastUpdateTime = Date.now();
    }
  }

  function updateLocalVelocity(vel: Vector3) {
    if (localPlayer.value) {
      localPlayer.value.velocity = { ...vel };
    }
  }

  function updateLocalRotation(rot: Rotation) {
    if (localPlayer.value) {
      localPlayer.value.rotation = { ...rot };
    }
  }

  function updateLocalHealth(newHealth: number) {
    if (localPlayer.value) {
      localPlayer.value.health = Math.max(0, Math.min(100, newHealth));
      if (localPlayer.value.health <= 0) {
        localPlayer.value.isAlive = false;
      }
    }
  }

  function getLocalAmmo(): number {
    return localPlayer.value?.ammo ?? 0;
  }

  function canLocalShoot(cost = 1): boolean {
    return getLocalAmmo() >= cost;
  }

  function consumeLocalAmmo(cost = 1): boolean {
    if (!localPlayer.value) return false;
    if (cost <= 0) return true;
    if (localPlayer.value.ammo < cost) return false;

    localPlayer.value.ammo = Math.max(0, localPlayer.value.ammo - cost);
    return true;
  }

  function refillLocalAmmo(targetAmmo = DEFAULT_MAG_SIZE): void {
    if (!localPlayer.value) return;
    localPlayer.value.ammo = Math.max(0, targetAmmo);
  }

  function respawn() {
    if (localPlayer.value) {
      localPlayer.value.health = 100;
      localPlayer.value.isAlive = true;
      localPlayer.value.position = { ...SAFE_SPAWN_POSITION };
      localPlayer.value.velocity = { x: 0, y: 0, z: 0 };
      localPlayer.value.ammo = DEFAULT_MAG_SIZE;
    }
  }

  /** 添加或更新远程玩家 */
  function upsertRemotePlayer(state: PlayerState) {
    remotePlayers.value.set(state.id, { ...state });
  }

  /** 移除远程玩家 */
  function removeRemotePlayer(id: string) {
    remotePlayers.value.delete(id);
  }

  /** 批量同步远程玩家 */
  function syncRemotePlayers(states: PlayerState[]) {
    const ids = new Set(states.map((s) => s.id));
    // 更新/添加
    for (const s of states) {
      if (s.id !== playerId.value) {
        remotePlayers.value.set(s.id, { ...s });
      }
    }
    // 移除已离开的玩家
    for (const [id] of remotePlayers.value) {
      if (!ids.has(id)) {
        remotePlayers.value.delete(id);
      }
    }
  }

  function $reset() {
    localPlayer.value = null;
    playerId.value = "";
    playerName.value = "Player";
    remotePlayers.value.clear();
  }

  return {
    // State
    localPlayer,
    playerId,
    playerName,
    remotePlayers,
    // Computed
    isAlive,
    health,
    position,
    rotation,
    remotePlayerCount,
    allPlayers,
    // Actions
    initLocalPlayer,
    updateLocalPosition,
    updateLocalVelocity,
    updateLocalRotation,
    updateLocalHealth,
    getLocalAmmo,
    canLocalShoot,
    consumeLocalAmmo,
    refillLocalAmmo,
    respawn,
    upsertRemotePlayer,
    removeRemotePlayer,
    syncRemotePlayers,
    $reset,
  };
});
