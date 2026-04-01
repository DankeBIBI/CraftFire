/**
 * 输入处理服务。
 * 处理键盘和鼠标输入，支持第一人称视角控制和游戏操作。
 * 使用 VueUse 的 useEventListener 进行事件绑定。
 */

import { logger } from "@/utils/logger";

/** 按键绑定映射 */
export const KEY_BINDINGS = {
	MOVE_FORWARD: "KeyW",
	MOVE_BACKWARD: "KeyS",
	MOVE_LEFT: "KeyA",
	MOVE_RIGHT: "KeyD",
	JUMP: "Space",
	CROUCH: "ShiftLeft",
	SPRINT: "ControlLeft",
	INTERACT: "KeyE",
	INVENTORY: "KeyI",
	PAUSE: "Escape",
	WEAPON_SWITCH: "KeyQ",
	PLACE_BLOCK: 0, // 鼠标左键
	REMOVE_BLOCK: 2, // 鼠标右键
	MIDDLE_CLICK: 1, // 鼠标中键（拾取方块）
} as const;

/** 输入状态 */
export interface InputState {
	forward: boolean;
	backward: boolean;
	left: boolean;
	right: boolean;
	jump: boolean;
	crouch: boolean;
	sprint: boolean;
	mouseX: number;
	mouseY: number;
	mouseDeltaX: number;
	mouseDeltaY: number;
}

/**
 * InputHandler 输入处理器。
 * 追踪当前按键状态，并在每帧提供输入快照。
 */
export class InputHandler {
	private keys: Set<string> = new Set();
	private mouseDeltaX = 0;
	private mouseDeltaY = 0;
	private sensitivity = 0.002;
	private isPointerLocked = false;

	/** 当前输入状态快照 */
	getState(): InputState {
		const state: InputState = {
			forward: this.keys.has(KEY_BINDINGS.MOVE_FORWARD),
			backward: this.keys.has(KEY_BINDINGS.MOVE_BACKWARD),
			left: this.keys.has(KEY_BINDINGS.MOVE_LEFT),
			right: this.keys.has(KEY_BINDINGS.MOVE_RIGHT),
			jump: this.keys.has(KEY_BINDINGS.JUMP),
			crouch: this.keys.has(KEY_BINDINGS.CROUCH),
			sprint: this.keys.has(KEY_BINDINGS.SPRINT),
			mouseX: 0,
			mouseY: 0,
			mouseDeltaX: this.mouseDeltaX * this.sensitivity,
			mouseDeltaY: this.mouseDeltaY * this.sensitivity,
		};

		// 重置增量（每帧读取一次后归零）
		this.mouseDeltaX = 0;
		this.mouseDeltaY = 0;

		return state;
	}

	/** 开始监听输入事件 */
	start(): void {
		document.addEventListener("keydown", this.onKeyDown);
		document.addEventListener("keyup", this.onKeyUp);
		document.addEventListener("mousemove", this.onMouseMove);
		document.addEventListener("pointerlockchange", this.onPointerLockChange);
		logger.info("Input", "输入处理器已启动");
	}

	/** 停止监听输入事件 */
	stop(): void {
		document.removeEventListener("keydown", this.onKeyDown);
		document.removeEventListener("keyup", this.onKeyUp);
		document.removeEventListener("mousemove", this.onMouseMove);
		document.removeEventListener("pointerlockchange", this.onPointerLockChange);
		this.keys.clear();
		logger.info("Input", "输入处理器已停止");
	}

	/** 请求指针锁定（第一人称控制必需） */
	requestPointerLock(): void {
		document.body.requestPointerLock();
	}

	/** 退出指针锁定 */
	exitPointerLock(): void {
		document.exitPointerLock();
	}

	/** 设置鼠标灵敏度 */
	setSensitivity(value: number): void {
		this.sensitivity = value;
	}

	get pointerLocked(): boolean {
		return this.isPointerLocked;
	}

	private onKeyDown = (e: KeyboardEvent): void => {
		this.keys.add(e.code);
	};

	private onKeyUp = (e: KeyboardEvent): void => {
		this.keys.delete(e.code);
	};

	private onMouseMove = (e: MouseEvent): void => {
		if (this.isPointerLocked) {
			this.mouseDeltaX += e.movementX;
			this.mouseDeltaY += e.movementY;
		}
	};

	private onPointerLockChange = (): void => {
		this.isPointerLocked = document.pointerLockElement === document.body;
	};
}

/** 全局单例 */
export const inputHandler = new InputHandler();
