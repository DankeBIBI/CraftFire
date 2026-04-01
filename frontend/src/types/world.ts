/**
 * CraftFire 世界相关类型定义。
 */

import type { BlockData, BlockType } from "./game";

/** 世界变化条目 */
export interface WorldChangeEntry {
	x: number;
	y: number;
	z: number;
	blockType: BlockType;
	action: "place" | "remove";
	timestamp: number;
	playerId?: string;
}

/** 世界快照数据 */
export interface WorldSnapshot {
	blocks: BlockData[];
	timestamp: number;
}

/** 分块数据 */
export interface ChunkData {
	x: number;
	z: number;
	blocks: BlockData[];
}
