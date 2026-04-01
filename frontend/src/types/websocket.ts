/**
 * CraftFire WebSocket 通信类型定义。
 */

/** WebSocket 消息信封 */
export interface WSMessage {
	type: string;
	timestamp: number;
	playerId: string;
	roomId: string;
	id: string;
	payload: unknown;
}

/** 消息类型常量 */
export const MSG_TYPES = {
	PLAYER_JOIN: "player_join",
	PLAYER_LEAVE: "player_leave",
	PLAYER_MOVE: "player_move",
	PLAYER_EQUIP: "player_equip",
	PLAYER_STATE_SYNC: "player_state_sync",
	BLOCK_PLACE: "block_place",
	BLOCK_REMOVE: "block_remove",
	WORLD_UPDATE: "world_update",
	WORLD_SNAPSHOT: "world_snapshot",
	PING: "ping",
	PONG: "pong",
	CHAT: "chat",
	ERROR: "error",
} as const;

/** 玩家移动负载 */
export interface PlayerMovePayload {
	x: number;
	y: number;
	z: number;
	rotation: { pitch: number; yaw: number };
}

/** 方块放置负载 */
export interface BlockPlacePayload {
	x: number;
	y: number;
	z: number;
	blockType: string;
}

/** 方块移除负载 */
export interface BlockRemovePayload {
	x: number;
	y: number;
	z: number;
}

/** 玩家加入负载 */
export interface PlayerJoinPayload {
	playerId: string;
	playerName: string;
	x: number;
	y: number;
	z: number;
	equipment?: string;
}

/** 玩家离开负载 */
export interface PlayerLeavePayload {
	playerId: string;
}

/** 玩家装备切换负载 */
export interface PlayerEquipPayload {
	playerId: string;
	equipment: string;
}

/** 世界更新负载 */
export interface WorldUpdatePayload {
	changes: Array<{
		x: number;
		y: number;
		z: number;
		blockType: string;
		action: "place" | "remove";
		timestamp?: number;
		playerId?: string;
	}>;
}

/** WebSocket 连接状态 */
export type WSConnectionState =
	| "connecting"
	| "connected"
	| "disconnected"
	| "reconnecting"
	| "error";
