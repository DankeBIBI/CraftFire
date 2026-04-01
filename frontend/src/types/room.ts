/**
 * CraftFire 房间相关类型定义。
 */

import type { GameMode } from "./game";

/** 房间配置 */
export interface RoomConfig {
	roomId: string;
	port: number;
	ip?: string;
	maxPlayers: number;
	currentPlayers: number;
	worldSeed: string;
	createdAt: number;
	lastActivityAt: number;
	isPublic: boolean;
	gameMode: GameMode;
}

/** 局域网服务器信息 */
export interface LANServerInfo {
	roomId: string;
	ip: string;
	playerCount: number;
	maxPlayers: number;
	gameMode: string;
	hostName?: string;
}
