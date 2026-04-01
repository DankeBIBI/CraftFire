/**
 * 玩家个人信息服务。
 * 封装获取和更新玩家资料、统计数据的后端接口。
 */

import type {
  PlayerProfile,
  PlayerStatistics,
  PlayerProfileUpdate,
} from "@/types/profile";

/** 获取玩家资料 */
export async function getPlayerProfile(playerId = ""): Promise<PlayerProfile> {
  if (window.go?.main?.App) {
    return window.go.main.App.GetPlayerProfile(
      playerId,
    ) as Promise<PlayerProfile>;
  }
  // 模拟数据
  return {
    playerId: "local-player",
    nickname: "玩家",
    characterModel: "default_player.glb",
    joinedAt: Date.now(),
    lastSeenAt: Date.now(),
    totalPlayTime: 0,
    level: 1,
    experience: 0,
    nextLevelExp: 1000,
    customization: {},
    equipment: { ammo: 30 },
    inventory: [],
  };
}

/** 更新玩家资料 */
export async function updatePlayerProfile(
  update: PlayerProfileUpdate,
): Promise<void> {
  if (window.go?.main?.App) {
    return window.go.main.App.UpdatePlayerProfile(
      update.nickname ?? "",
      update.skinColor ?? "",
    );
  }
}

/** 获取玩家统计数据 */
export async function getPlayerStatistics(
  playerId = "",
): Promise<PlayerStatistics> {
  if (window.go?.main?.App) {
    return window.go.main.App.GetPlayerStatistics(
      playerId,
    ) as Promise<PlayerStatistics>;
  }
  return {
    playerId: "local-player",
    totalBlocksPlaced: 0,
    totalBlocksRemoved: 0,
    totalKills: 0,
    totalDeaths: 0,
    distanceTraveled: 0,
    gameTime: 0,
    roomsVisited: 0,
    achievements: [],
    lastUpdated: Date.now(),
  };
}
