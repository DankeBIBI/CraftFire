/**
 * CraftFire 玩家相关类型定义。
 */

import type { Vector3, Rotation } from "./game";

/** 玩家状态 */
export interface PlayerState {
	/** 玩家唯一 ID（通常为 UUID） */
	id: string;
	/** 玩家显示名称 */
	name: string;
	/** 世界坐标位置（x, y, z） */
	position: Vector3;
	/** 当前速度向量 */
	velocity: Vector3;
	/** 视角旋转（俯仰/偏航/滚转） */
	rotation: Rotation;
	/** 当前生命值（0 表示死亡） */
	health: number;
	/** 当前弹药数量 */
	ammo: number;
	/** 装备标识或描述 */
	equipment: string;
	/** 是否存活 */
	isAlive: boolean;
	/** 上次更新的时间戳（毫秒） */
	lastUpdateTime: number;
}

/** 背包物品 */
export interface InventoryItem {
	/** 物品唯一标识 */
	itemId: string;
	/** 物品类型（用于区分具体类别） */
	itemType: string;
	/** 数量 */
	quantity: number;
	/** 可选的额外元数据 */
	metadata?: unknown;
}

/** 装备信息 */
export interface EquipmentInfo {
	/** 装备中的武器 ID 或标识 */
	weapon: string;
	/** 装备中的护甲 ID 或标识 */
	armor: string;
	/** 当前弹药量 */
	ammo: number;
}

/** 玩家状态标识 */
export type PlayerStatus = "online" | "idle" | "dead";
