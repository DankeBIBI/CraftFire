/**
 * WebSocket 客户端服务。
 * 负责与 Go 后端 WebSocket 服务器的连接、消息收发和自动重连。
 */

import type { WSMessage, WSConnectionState } from "@/types/websocket";
import { MSG_TYPES } from "@/types/websocket";
import { logger } from "@/utils/logger";

/** WebSocket 消息类型验证 */
function isWSMessage(obj: unknown): obj is WSMessage {
	if (typeof obj !== "object" || obj === null) return false;
	const msg = obj as Record<string, unknown>;
	return (
		typeof msg.type === "string" &&
		typeof msg.timestamp === "number" &&
		typeof msg.playerId === "string" &&
		typeof msg.roomId === "string" &&
		typeof msg.id === "string" &&
		"payload" in msg
	);
}

/** 最大重连次数 */
const MAX_RECONNECT_ATTEMPTS = 5;

/** 重连基础延迟（毫秒） */
const RECONNECT_BASE_DELAY = 1000;

/** 消息回调函数类型 */
type MessageHandler = (message: WSMessage) => void;

/** 连接状态变化回调 */
type StateHandler = (state: WSConnectionState) => void;

/** 移动消息节流间隔（毫秒） */
const MOVE_THROTTLE_MS = 50;

/** 移动消息节流状态 */
interface ThrottleState {
	lastSendTime: number;
}

/**
 * WebSocketService 封装 WebSocket 连接管理。
 * 支持自动重连（指数退避）和消息分发。
 */
export class WebSocketService {
	private ws: WebSocket | null = null;
	private url = "";
	private reconnectAttempts = 0;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private messageHandlers: Map<string, MessageHandler[]> = new Map();
	private stateHandlers: StateHandler[] = [];
	private _state: WSConnectionState = "disconnected";
	private playerId = "";
	private roomId = "";
	private playerName = "";
	private moveThrottle: ThrottleState = { lastSendTime: 0 };

	/** 当前连接状态 */
	get state(): WSConnectionState {
		return this._state;
	}

	/**
	 * 连接到 WebSocket 服务器。
	 *
	 * @param roomId - 房间号（同时作为端口号）
	 * @param playerName - 玩家昵称
	 * @param host - 服务器地址，默认 localhost
	 */
	/**
	 * 将 6 位房间号映射到合法端口范围 (10000-65535)。
	 */
	private roomIdToPort(roomId: string): number {
		const num = Number.parseInt(roomId, 10);
		if (Number.isNaN(num)) return 10000;
		return 10000 + (num % 55536); // 映射到 10000-65535
	}

	connect(roomId: string, playerName: string, host = "localhost"): void {
		this.roomId = roomId;
		this.playerName = playerName;
		const port = this.roomIdToPort(roomId);
		this.url = `ws://${host}:${port}/ws?name=${encodeURIComponent(playerName)}`;
		this._setState("connecting");

		logger.info("Network", `正在连接 ${this.url}`);
		try {
			this.ws = new WebSocket(this.url);
			this.setupEventHandlers();
		} catch (err) {
			logger.warn("Network", "WebSocket 连接创建失败:", err);
			this._setState("disconnected");
		}
	}

	/** 设置 WebSocket 事件处理器 */
	private setupEventHandlers(): void {
		if (!this.ws) return;

		this.ws.onopen = () => {
			logger.info("Network", "WebSocket 已连接");
			this._setState("connected");
			this.reconnectAttempts = 0;
		};

		this.ws.onclose = (event) => {
			logger.warn("Network", `WebSocket 已断开 (code: ${event.code})`);
			this._setState("disconnected");

			if (!event.wasClean) {
				this.scheduleReconnect();
			}
		};

		this.ws.onerror = (error) => {
			logger.error("Network", "WebSocket 错误:", error);
			this._setState("error");
		};

		this.ws.onmessage = (event) => {
			try {
				const parsed = JSON.parse(event.data);
				if (!isWSMessage(parsed)) {
					logger.warn("Network", "消息格式无效:", event.data);
					return;
				}
				this.dispatchMessage(parsed);
			} catch (err) {
				logger.warn("Network", "消息解析失败:", err);
			}
		};
	}

	/**
	 * 发送 WebSocket 消息。
	 *
	 * @param type - 消息类型
	 * @param payload - 消息负载
	 */
	send(type: string, payload: unknown): void {
		if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
			logger.warn("Network", "无法发送消息: WebSocket 未就绪");
			return;
		}

		const message: WSMessage = {
			type,
			timestamp: Date.now(),
			playerId: this.playerId,
			roomId: this.roomId,
			id: crypto.randomUUID(),
			payload,
		};

		this.ws.send(JSON.stringify(message));
	}

	/** 发送玩家移动消息（带节流） */
	sendPlayerMove(
		x: number,
		y: number,
		z: number,
		pitch: number,
		yaw: number,
	): void {
		const now = Date.now();
		// 节流：超过间隔直接发送
		if (now - this.moveThrottle.lastSendTime >= MOVE_THROTTLE_MS) {
			this.send(MSG_TYPES.PLAYER_MOVE, { x, y, z, rotation: { pitch, yaw } });
			this.moveThrottle.lastSendTime = now;
		}
		// 注意：节流期间的位置更新会被跳过，这是节流的正常行为
	}

	/** 发送放置方块消息 */
	sendBlockPlace(x: number, y: number, z: number, blockType: string): void {
		this.send(MSG_TYPES.BLOCK_PLACE, { x, y, z, blockType });
	}

	/** 发送移除方块消息 */
	sendBlockRemove(x: number, y: number, z: number): void {
		this.send(MSG_TYPES.BLOCK_REMOVE, { x, y, z });
	}

	/** 发送装备切换消息 */
	sendPlayerEquip(equipment: string): void {
		this.send(MSG_TYPES.PLAYER_EQUIP, { equipment });
	}

	/**
	 * 注册消息处理器。
	 *
	 * @param type - 要监听的消息类型
	 * @param handler - 处理函数
	 */
	on(type: string, handler: MessageHandler): void {
		if (!this.messageHandlers.has(type)) {
			this.messageHandlers.set(type, []);
		}
		this.messageHandlers.get(type)!.push(handler);
	}

	/** 注册连接状态变化回调 */
	onStateChange(handler: StateHandler): void {
		this.stateHandlers.push(handler);
	}

	/** 分发收到的消息到注册的处理器 */
	private dispatchMessage(message: WSMessage): void {
		const handlers = this.messageHandlers.get(message.type);
		if (handlers) {
			handlers.forEach((handler) => handler(message));
		}

		// 通配处理器
		const allHandlers = this.messageHandlers.get("*");
		if (allHandlers) {
			allHandlers.forEach((handler) => handler(message));
		}
	}

	/** 计划自动重连（指数退避） */
	private scheduleReconnect(): void {
		if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
			logger.error("Network", "已达最大重连次数，停止重连");
			this._setState("error");
			return;
		}

		this.reconnectAttempts++;
		const delay =
			RECONNECT_BASE_DELAY * Math.pow(2, this.reconnectAttempts - 1);
		logger.info(
			"Network",
			`${delay}ms 后尝试第 ${this.reconnectAttempts} 次重连...`,
		);

		this._setState("reconnecting");

		this.reconnectTimer = setTimeout(() => {
			if (this.roomId) {
				const hostMatch = this.url.match(/ws:\/\/([^:]+)/);
				const host = hostMatch?.[1] ?? "localhost";
				this.connect(this.roomId, this.playerName, host);
			}
		}, delay);
	}

	/** 更新连接状态并通知监听器 */
	private _setState(state: WSConnectionState): void {
		this._state = state;
		this.stateHandlers.forEach((handler) => handler(state));
	}

	/** 断开连接 */
	disconnect(): void {
		logger.info("Network", "主动断开 WebSocket");
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}

		if (this.ws) {
			this.ws.close(1000, "正常关闭");
			this.ws = null;
		}

		this._setState("disconnected");
		this.messageHandlers.clear();
		this.stateHandlers = [];
	}
}

/** 全局单例 */
export const wsService = new WebSocketService();
