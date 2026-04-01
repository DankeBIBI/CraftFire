/**
 * 房间状态管理。
 * 管理房间的创建、加入、局域网发现等。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { RoomConfig, LANServerInfo } from "@/types/room";
import * as WailsService from "@/services/WailsService";
import { logger } from "@/utils/logger";

export const useRoomStore = defineStore("room", () => {
	// ─── 状态 ────────────────────────────────
	const currentRoom = ref<RoomConfig | null>(null);
	const isHost = ref(false);
	const lanServers = ref<LANServerInfo[]>([]);
	const isSearching = ref(false);
	const joinError = ref<string | null>(null);
	const isCreating = ref(false);
	const isJoining = ref(false);

	// ─── 计算属性 ────────────────────────────
	const isInRoom = computed(() => currentRoom.value !== null);
	const roomId = computed(() => currentRoom.value?.roomId ?? "");
	const playerCount = computed(() => currentRoom.value?.currentPlayers ?? 0);
	const maxPlayers = computed(() => currentRoom.value?.maxPlayers ?? 10);

	// ─── 操作 ────────────────────────────────

	/** 创建房间 */
	async function createRoom(
		_playerName: string,
		maxPlayers: number,
		gameMode: string,
	): Promise<boolean> {
		isCreating.value = true;
		joinError.value = null;
		try {
			logger.info("Room", "正在创建房间...");
			const room = await WailsService.CreateRoom();
			if (room) {
				logger.info("Room", "房间已创建, roomId =", room.roomId);
				currentRoom.value = {
					...(room as RoomConfig),
					maxPlayers,
					gameMode: gameMode as RoomConfig["gameMode"],
					currentPlayers: 1,
					lastActivityAt: Date.now(),
				};
				isHost.value = true;
				return true;
			}
			logger.warn("Room", "创建房间失败");
			joinError.value = "创建房间失败";
			return false;
		} catch (err: unknown) {
			logger.error("Room", "创建房间异常:", err);
			joinError.value =
				err instanceof Error ? err.message : "创建房间时发生错误";
			return false;
		} finally {
			isCreating.value = false;
		}
	}

	/** 加入房间 */
	async function joinRoom(
		roomIdStr: string,
		ip = "127.0.0.1",
	): Promise<boolean> {
		if (!/^\d{6}$/.test(roomIdStr)) {
			joinError.value = "房间号必须为 6 位数字";
			return false;
		}
		isJoining.value = true;
		joinError.value = null;
		try {
			logger.info("Room", `正在加入房间 ${roomIdStr} (ip: ${ip})`);
			const result = await WailsService.JoinRoom(roomIdStr, ip);
			if (result) {
				logger.info("Room", `已加入房间 ${roomIdStr}`);
				currentRoom.value = {
					roomId: roomIdStr,
					port: parseInt(roomIdStr, 10),
					ip,
					maxPlayers: 10,
					currentPlayers: 1,
					worldSeed: "",
					createdAt: Date.now(),
					lastActivityAt: Date.now(),
					isPublic: true,
					gameMode: "sandbox",
				};
				isHost.value = false;
				return true;
			}
			logger.warn("Room", "加入房间失败");
			joinError.value = "加入房间失败，请检查房间号";
			return false;
		} catch (err: unknown) {
			logger.error("Room", "加入房间异常:", err);
			joinError.value =
				err instanceof Error ? err.message : "加入房间时发生错误";
			return false;
		} finally {
			isJoining.value = false;
		}
	}

	/** 离开房间 */
	async function leaveRoom() {
		if (currentRoom.value) {
			try {
				await WailsService.LeaveRoom();
			} catch {
				// 最大努力清理
			}
		}
		currentRoom.value = null;
		isHost.value = false;
		joinError.value = null;
	}

	/** 搜索局域网服务器 */
	async function searchLANServers() {
		isSearching.value = true;
		try {
			const servers = await WailsService.FindLANServers();
			lanServers.value = (servers as LANServerInfo[]) || [];
		} catch {
			lanServers.value = [];
		} finally {
			isSearching.value = false;
		}
	}

	/** 更新房间内当前玩家数 */
	function updatePlayerCount(count: number) {
		if (currentRoom.value) {
			currentRoom.value.currentPlayers = count;
			currentRoom.value.lastActivityAt = Date.now();
		}
	}

	function $reset() {
		currentRoom.value = null;
		isHost.value = false;
		lanServers.value = [];
		isSearching.value = false;
		joinError.value = null;
		isCreating.value = false;
		isJoining.value = false;
	}

	return {
		// State
		currentRoom,
		isHost,
		lanServers,
		isSearching,
		joinError,
		isCreating,
		isJoining,
		// Computed
		isInRoom,
		roomId,
		playerCount,
		maxPlayers,
		// Actions
		createRoom,
		joinRoom,
		leaveRoom,
		searchLANServers,
		updatePlayerCount,
		$reset,
	};
});
