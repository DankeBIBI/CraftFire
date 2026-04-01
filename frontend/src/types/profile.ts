/**
 * CraftFire 玩家个人资料类型定义。
 */

import type { InventoryItem } from "./player";

/** 角色定制 */
export interface Customization {
	skinColor?: string;
	clothingStyle?: string;
	accessories?: string[];
}

/** 玩家资料 */
export interface PlayerProfile {
	playerId: string;
	nickname: string;
	avatar?: string;
	characterModel: string;
	joinedAt: number;
	lastSeenAt: number;
	totalPlayTime: number;
	level: number;
	experience: number;
	nextLevelExp: number;
	customization: Customization;
	equipment: {
		weapon?: string;
		armor?: string;
		ammo: number;
	};
	inventory: InventoryItem[];
}

/** 玩家统计 */
export interface PlayerStatistics {
	playerId: string;
	totalBlocksPlaced: number;
	totalBlocksRemoved: number;
	totalKills: number;
	totalDeaths: number;
	distanceTraveled: number;
	gameTime: number;
	roomsVisited: number;
	achievements: string[];
	lastUpdated: number;
}

/** 资料更新请求 */
export interface PlayerProfileUpdate {
	nickname?: string;
	skinColor?: string;
	clothingStyle?: string;
	accessories?: string[];
	customization?: Customization;
}
