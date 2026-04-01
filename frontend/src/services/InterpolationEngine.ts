/**
 * 插值引擎服务。
 * 在低 Tick Rate (8 TPS / 125ms) 下平滑远端玩家的运动轨迹。
 * 使用三次平滑插值 (smoothstep) 确保视觉效果流畅。
 * 支持外推（extrapolation）在网络延迟时预测位置。
 */

import type { Vector3, Rotation } from "@/types/game";

/** 默认插值持续时间（毫秒），与 8 TPS 刷新率对应 */
const DEFAULT_DURATION = 125;

/** 外推超时时间（毫秒），超过此时间不再外推 */
const EXTRAPOLATION_TIMEOUT = 500;

/** 单个实体的插值状态 */
interface InterpolationEntry {
	previousPosition: Vector3;
	targetPosition: Vector3;
	previousRotation: Rotation;
	targetRotation: Rotation;
	previousVelocity: Vector3;
	startTime: number;
	lastUpdateTime: number;
	duration: number;
}

/**
 * InterpolationEngine 插值引擎。
 * 为每个远端实体维护一个插值状态，在渲染帧之间平滑运动。
 * 支持外推：在收到新数据间隔较长时，使用速度预测位置。
 */
export class InterpolationEngine {
	private entries: Map<string, InterpolationEntry> = new Map();
	private duration: number;

	constructor(durationMs = DEFAULT_DURATION) {
		this.duration = durationMs;
	}

	/**
	 * 推送新的目标状态（通常在收到服务器 tick 更新时调用）。
	 */
	pushState(entityId: string, position: Vector3, rotation: Rotation): void {
		const now = performance.now();
		const existing = this.entries.get(entityId);

		// 计算速度（如果有上一状态）
		let velocity: Vector3 = { x: 0, y: 0, z: 0 };
		if (existing && existing.lastUpdateTime > 0) {
			const dt = (now - existing.lastUpdateTime) / 1000;
			if (dt > 0) {
				velocity = {
					x: (position.x - existing.targetPosition.x) / dt,
					y: (position.y - existing.targetPosition.y) / dt,
					z: (position.z - existing.targetPosition.z) / dt,
				};
			}
		}

		this.entries.set(entityId, {
			previousPosition: existing?.targetPosition ?? position,
			targetPosition: position,
			previousRotation: existing?.targetRotation ?? rotation,
			targetRotation: rotation,
			previousVelocity: existing?.previousVelocity ?? velocity,
			startTime: now,
			lastUpdateTime: now,
			duration: this.duration,
		});
	}

	/**
	 * 获取当前时刻的插值位置（通常在每帧渲染时调用）。
	 */
	getInterpolatedState(
		entityId: string,
	): { position: Vector3; rotation: Rotation } | null {
		const entry = this.entries.get(entityId);
		if (!entry) return null;

		const now = performance.now();
		const elapsed = now - entry.startTime;
		const timeSinceUpdate = now - entry.lastUpdateTime;

		let position: Vector3;
		let t: number;

		if (timeSinceUpdate > EXTRAPOLATION_TIMEOUT) {
			// 超过外推时间，使用目标位置静止
			position = { ...entry.targetPosition };
			t = 1.0;
		} else if (elapsed >= entry.duration) {
			// 插值完成，使用目标位置
			position = { ...entry.targetPosition };
			t = 1.0;
		} else {
			t = Math.min(elapsed / entry.duration, 1.0);
			// 三次平滑插值
			t = t * t * (3 - 2 * t);

			// 位置插值
			position = {
				x: lerp(entry.previousPosition.x, entry.targetPosition.x, t),
				y: lerp(entry.previousPosition.y, entry.targetPosition.y, t),
				z: lerp(entry.previousPosition.z, entry.targetPosition.z, t),
			};
		}

		// 旋转插值
		const rotation: Rotation = {
			pitch: lerp(entry.previousRotation.pitch, entry.targetRotation.pitch, t),
			yaw: lerpAngle(entry.previousRotation.yaw, entry.targetRotation.yaw, t),
			roll: lerp(entry.previousRotation.roll, entry.targetRotation.roll, t),
		};

		return { position, rotation };
	}

	/** 移除实体的插值状态 */
	removeEntity(entityId: string): void {
		this.entries.delete(entityId);
	}

	/** 清除所有插值状态 */
	clear(): void {
		this.entries.clear();
	}

	/** 获取实体的最后更新时间 */
	getEntityLastUpdate(entityId: string): number {
		return this.entries.get(entityId)?.lastUpdateTime ?? 0;
	}
}

/** 线性插值 */
function lerp(a: number, b: number, t: number): number {
	return a + (b - a) * t;
}

/** 角度线性插值（处理 -180° 到 180° 的环绕） */
function lerpAngle(a: number, b: number, t: number): number {
	let delta = b - a;
	while (delta > Math.PI) delta -= Math.PI * 2;
	while (delta < -Math.PI) delta += Math.PI * 2;
	return a + delta * t;
}

/** 全局单例 */
export const interpolationEngine = new InterpolationEngine();
