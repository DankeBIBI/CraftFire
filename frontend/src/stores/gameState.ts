/**
 * 游戏全局状态管理。
 * 控制游戏的核心运行时状态，不包含细分领域的子状态。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { GameMode } from "@/types/game";

/** 游戏视图 */
export type GameView = "menu" | "game" | "loading";

export const useGameStateStore = defineStore("gameState", () => {
	// ─── 核心状态 ────────────────────────────────
	const currentView = ref<GameView>("menu");
	const isRunning = ref(false);
	const isPaused = ref(false);
	const fps = ref(0);
	const deltaTime = ref(0);
	const gameMode = ref<GameMode>("sandbox");
	const loadingProgress = ref(0);
	const loadingMessage = ref("");
	const isDebugMode = ref(false);

	// ─── 计算属性 ────────────────────────────────
	const isInGame = computed(
		() => currentView.value === "game" && isRunning.value,
	);
	const isLoading = computed(() => currentView.value === "loading");

	// ─── 操作 ────────────────────────────────────
	function startGame(mode: GameMode = "sandbox") {
		gameMode.value = mode;
		currentView.value = "loading";
		loadingProgress.value = 0;
		loadingMessage.value = "正在初始化游戏...";
	}

	function enterGame() {
		currentView.value = "game";
		isRunning.value = true;
		isPaused.value = false;
		loadingProgress.value = 100;
	}

	function pauseGame() {
		if (isRunning.value) {
			isPaused.value = true;
		}
	}

	function resumeGame() {
		isPaused.value = false;
	}

	function exitGame() {
		isRunning.value = false;
		isPaused.value = false;
		currentView.value = "menu";
		loadingProgress.value = 0;
		loadingMessage.value = "";
	}

	function updateFPS(newFps: number) {
		fps.value = newFps;
	}

	function updateDeltaTime(dt: number) {
		deltaTime.value = dt;
	}

	function setLoadingProgress(progress: number, message?: string) {
		loadingProgress.value = Math.min(100, Math.max(0, progress));
		if (message) {
			loadingMessage.value = message;
		}
	}

	function toggleDebugMode() {
		isDebugMode.value = !isDebugMode.value;
	}

	return {
		// State
		currentView,
		isRunning,
		isPaused,
		fps,
		deltaTime,
		gameMode,
		loadingProgress,
		loadingMessage,
		isDebugMode,
		// Computed
		isInGame,
		isLoading,
		// Actions
		startGame,
		enterGame,
		pauseGame,
		resumeGame,
		exitGame,
		updateFPS,
		updateDeltaTime,
		setLoadingProgress,
		toggleDebugMode,
	};
});
