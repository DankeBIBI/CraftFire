/**
 * CraftFire 游戏核心类型定义。
 * 包含向量、旋转和基础游戏数据结构。
 */

/** 三维向量 */
export interface Vector3 {
	x: number;
	y: number;
	z: number;
}

/** 欧拉角旋转 */
export interface Rotation {
	pitch: number;
	yaw: number;
	roll: number;
}

/** 方块类型 */
export type BlockType =
	| "stone"
	| "wood"
	| "glass"
	| "dirt"
	| "grass"
	| "sand"
	| "water"
	| "air"
	| "sandDark"
	| "crate"
	| "stoneDark"
	| "metal";

/** 方块数据 */
export interface BlockData {
	x: number;
	y: number;
	z: number;
	type: BlockType;
	metadata?: number;
}

/** 游戏模式 */
export type GameMode = "sandbox" | "survival" | "pvp";

/** 设备信息 */
export interface DeviceInfo {
	platform: string;
	screenWidth: number;
	screenHeight: number;
	gpu: string;
}

/** 游戏设置 */
export interface GameSettings {
	targetFPS: number;
	renderDistance: number;
	pointerLockRequired: boolean;
	showDebugInfo: boolean;
	language: string;
	uiScale: number;
	volume: number;
	mouseSensitivity: number;
}

/** 默认游戏设置 */
export const DEFAULT_GAME_SETTINGS: GameSettings = {
	targetFPS: 60,
	renderDistance: 6,
	pointerLockRequired: true,
	showDebugInfo: false,
	language: "zh-CN",
	uiScale: 1.0,
	volume: 0.7,
	mouseSensitivity: 0.5,
};
