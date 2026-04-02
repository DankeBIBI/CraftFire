/**
 * 物理服务（前端客户端预测用）。
 * 在客户端预测玩家运动，然后由后端权威校正。
 */

import type { Vector3 } from "@/types/game";

/** 重力加速度 */
const GRAVITY = 30.0; // 稍微加大重力手感

/** 玩家移动速度 */
const MOVE_SPEED = 5.0;

/** 冲刺倍率 */
const SPRINT_MULTIPLIER = 1.6;

/** 跳跃初速度 */
const JUMP_VELOCITY = 10.0;

/** 玩家尺寸 */
const PLAYER_WIDTH = 0.6;
const PLAYER_HEIGHT = 1.8;

type GetBlockFn = (x: number, y: number, z: number) => boolean;

/**
 * PhysicsService 客户端物理预测。
 * 用于在低服务器 Tick Rate 下保持玩家操作的即时响应。
 */
export class PhysicsService {
	public velocity = { x: 0, y: 0, z: 0 };
	private isGrounded = false;

	/**
	 * 更新玩家位置（包含碰撞检测）
	 */
	updatePlayerPosition(
		currentPos: Vector3,
		deltaTime: number,
		inputs: {
			forward: boolean;
			backward: boolean;
			left: boolean;
			right: boolean;
			jump: boolean;
			sprint: boolean;
			yaw: number;
		},
		hasBlock: GetBlockFn,
	): Vector3 {
		// 1. 计算水平目标速度
		let speed = MOVE_SPEED;
		if (inputs.sprint) speed *= SPRINT_MULTIPLIER;

		const targetVelX =
			Math.sin(inputs.yaw) *
				(Number(inputs.backward) - Number(inputs.forward)) +
			Math.cos(inputs.yaw) * (Number(inputs.right) - Number(inputs.left));
		const targetVelZ =
			Math.cos(inputs.yaw) *
				(Number(inputs.backward) - Number(inputs.forward)) -
			Math.sin(inputs.yaw) * (Number(inputs.right) - Number(inputs.left));

		// 归一化并应用速度
		const length = Math.sqrt(targetVelX * targetVelX + targetVelZ * targetVelZ);
		if (length > 0) {
			this.velocity.x = (targetVelX / length) * speed;
			this.velocity.z = (targetVelZ / length) * speed;
		} else {
			this.velocity.x = 0; // 停止移动时阻尼
			this.velocity.z = 0;
		}

		// 2. 垂直速度（重力 + 跳跃）
		if (inputs.jump && this.isGrounded) {
			this.velocity.y = JUMP_VELOCITY;
			this.isGrounded = false;
		}
		this.velocity.y -= GRAVITY * deltaTime;

		// 3. 分轴移动与碰撞检测
		const newPos = { ...currentPos };

		// X 轴移动
		newPos.x += this.velocity.x * deltaTime;
		if (this.checkCollision(newPos, hasBlock)) {
			// 简单的回弹/阻挡：退回到旧的 X
			newPos.x -= this.velocity.x * deltaTime;
			this.velocity.x = 0;
		}

		// Z 轴移动
		newPos.z += this.velocity.z * deltaTime;
		if (this.checkCollision(newPos, hasBlock)) {
			newPos.z -= this.velocity.z * deltaTime;
			this.velocity.z = 0;
		}

		// Y 轴移动
		newPos.y += this.velocity.y * deltaTime;
		if (this.checkCollision(newPos, hasBlock)) {
			const wasFalling = this.velocity.y < 0;
			newPos.y -= this.velocity.y * deltaTime;
			this.velocity.y = 0;

			if (wasFalling) {
				this.isGrounded = true;
			} else {
				// 顶到头了
			}
		} else {
			this.isGrounded = false;
		}

		// 地面边界保护 (y < 0)
		if (newPos.y < 0) {
			newPos.y = 0;
			this.velocity.y = 0;
			this.isGrounded = true;
		}

		return newPos;
	}

	/** 检测玩家包围盒通过是否与方块通过 */
	private checkCollision(pos: Vector3, hasBlock: GetBlockFn): boolean {
		// 玩家 AABB
		const minX = pos.x - PLAYER_WIDTH / 2;
		const maxX = pos.x + PLAYER_WIDTH / 2;
		const minY = pos.y;
		const maxY = pos.y + PLAYER_HEIGHT;
		const minZ = pos.z - PLAYER_WIDTH / 2;
		const maxZ = pos.z + PLAYER_WIDTH / 2;

		// 检查覆盖的方块坐标范围（稍微缩小一点边界防止卡墙）
		const epsilon = 0.001;
		const startX = Math.floor(minX + epsilon);
		const endX = Math.floor(maxX - epsilon);
		const startY = Math.floor(minY + epsilon);
		const endY = Math.floor(maxY - epsilon);
		const startZ = Math.floor(minZ + epsilon);
		const endZ = Math.floor(maxZ - epsilon);

		for (let x = startX; x <= endX; x++) {
			for (let y = startY; y <= endY; y++) {
				for (let z = startZ; z <= endZ; z++) {
					if (hasBlock(x, y, z)) {
						return true;
					}
				}
			}
		}
		return false;
	}

	setGrounded(val: boolean) {
		this.isGrounded = val;
	}
}

export const physicsService = new PhysicsService();
