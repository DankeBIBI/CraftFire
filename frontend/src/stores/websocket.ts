/**
 * WebSocket 连接状态管理。
 * 管理与游戏服务器的 WebSocket 连接生命周期。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { WSConnectionState } from "@/types/websocket";

export const useWebSocketStore = defineStore("websocket", () => {
	// ─── 状态 ────────────────────────────────
	const connectionState = ref<WSConnectionState>("disconnected");
	const serverUrl = ref<string | null>(null);
	const reconnectAttempts = ref(0);
	const maxReconnectAttempts = ref(5);
	const lastPing = ref(0);
	const latency = ref(0);
	const messagesReceived = ref(0);
	const messagesSent = ref(0);

	// ─── 计算属性 ────────────────────────────────
	const isConnected = computed(() => connectionState.value === "connected");
	const isConnecting = computed(() => connectionState.value === "connecting");
	const isReconnecting = computed(
		() => connectionState.value === "reconnecting",
	);

	// ─── 操作 ────────────────────────────────────
	function setConnectionState(state: WSConnectionState) {
		connectionState.value = state;
	}

	function setServerUrl(url: string) {
		serverUrl.value = url;
	}

	function incrementReconnectAttempts() {
		reconnectAttempts.value++;
	}

	function resetReconnectAttempts() {
		reconnectAttempts.value = 0;
	}

	function updateLatency(ping: number) {
		lastPing.value = Date.now();
		latency.value = ping;
	}

	function incrementMessagesReceived() {
		messagesReceived.value++;
	}

	function incrementMessagesSent() {
		messagesSent.value++;
	}

	function disconnect() {
		connectionState.value = "disconnected";
		serverUrl.value = null;
		reconnectAttempts.value = 0;
		latency.value = 0;
	}

	function $reset() {
		connectionState.value = "disconnected";
		serverUrl.value = null;
		reconnectAttempts.value = 0;
		lastPing.value = 0;
		latency.value = 0;
		messagesReceived.value = 0;
		messagesSent.value = 0;
	}

	return {
		// State
		connectionState,
		serverUrl,
		reconnectAttempts,
		maxReconnectAttempts,
		lastPing,
		latency,
		messagesReceived,
		messagesSent,
		// Computed
		isConnected,
		isConnecting,
		isReconnecting,
		// Actions
		setConnectionState,
		setServerUrl,
		incrementReconnectAttempts,
		resetReconnectAttempts,
		updateLatency,
		incrementMessagesReceived,
		incrementMessagesSent,
		disconnect,
		$reset,
	};
});
