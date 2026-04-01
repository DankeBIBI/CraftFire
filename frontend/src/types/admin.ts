/**
 * CraftFire 管理员面板类型定义。
 */

import type { Vector3, Rotation } from "./game";
import type { InventoryItem, PlayerStatus } from "./player";

/** 管理员面板玩家信息（列表用） */
export interface PlayerInfo {
	id: string;
	name: string;
	position: Vector3;
	health: number;
	status: PlayerStatus;
	connectedAt: number;
	lastActivityAt: number;
	ping: number;
	equipment: string;
}

/** 玩家详细信息（详情弹窗用） */
export interface PlayerDetails {
	id: string;
	name: string;
	position: Vector3;
	velocity: Vector3;
	rotation: Rotation;
	health: number;
	maxHealth: number;
	isAlive: boolean;
	status: PlayerStatus;
	equipment: {
		weapon: string;
		armor: string;
		ammo: number;
	};
	inventory: InventoryItem[];
	connectedAt: number;
	lastActivityAt: number;
	remoteIP: string;
	ping: number;
	packetLoss: number;
	statistics: {
		blocksPlaced: number;
		blocksRemoved: number;
		killCount: number;
		deathCount: number;
		distanceTraveled: number;
	};
	isMuted: boolean;
	muteEndTime?: number;
	joinedAt: number;
}

/** 房间统计数据 */
export interface RoomStatistics {
	roomId: string;
	totalPlayers: number;
	maxPlayers: number;
	totalPlayersJoined: number;
	uptime: number;
	totalBlocksPlaced: number;
	totalBlocksRemoved: number;
	averagePing: number;
	peakPlayerCount: number;
	createdAt: number;
	lastUpdated: number;
}

/** 管理员状态 */
export interface AdminState {
	isAuthenticated: boolean;
	sessionToken: string | null;
	tokenExpiresAt: number;
	players: PlayerInfo[];
	selectedPlayer: PlayerDetails | null;
	roomStats: RoomStatistics | null;
	lastUpdateTime: number;
	isLoading: boolean;
	error: string | null;
}
