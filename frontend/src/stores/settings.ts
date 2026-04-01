/**
 * 用户设置状态管理。
 * 管理视频、音频、控制等用户偏好设置，支持 localStorage 持久化。
 */

import { defineStore } from "pinia";
import { ref, watch } from "vue";

export interface VideoSettings {
	brightness: number;
	renderDistance: number;
	targetFPS: number;
	showFPS: boolean;
	antiAliasing: boolean;
	shadowQuality: "off" | "low" | "medium" | "high";
	particleEffects: boolean;
}

export interface AudioSettings {
	masterVolume: number;
	musicVolume: number;
	sfxVolume: number;
	ambientVolume: number;
	muted: boolean;
}

export interface ControlSettings {
	mouseSensitivity: number;
	invertY: boolean;
	toggleSprint: boolean;
	toggleCrouch: boolean;
}

export interface WeaponViewSettings {
	offsetX: number;
	offsetY: number;
	offsetZ: number;
	scale: number;
}

const STORAGE_KEY = "craftfire-settings";

/** 存储数据结构 */
interface SettingsData {
	video?: VideoSettings;
	audio?: AudioSettings;
	controls?: ControlSettings;
	weaponView?: WeaponViewSettings;
	language?: string;
	uiScale?: number;
}

function loadFromStorage(): SettingsData | null {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return null;
		const parsed = JSON.parse(raw);
		// 简单的运行时类型验证
		if (typeof parsed !== "object" || parsed === null) return null;
		return parsed as SettingsData;
	} catch {
		return null;
	}
}

export const useSettingsStore = defineStore("settings", () => {
	const saved = loadFromStorage();

	// ─── 视频设置 ────────────────────────────
	const video = ref<VideoSettings>(
		saved?.video ?? {
			brightness: 100,
			renderDistance: 6,
			targetFPS: 60,
			showFPS: false,
			antiAliasing: true,
			shadowQuality: "medium",
			particleEffects: true,
		},
	);

	// ─── 音频设置 ────────────────────────────
	const audio = ref<AudioSettings>(
		saved?.audio ?? {
			masterVolume: 80,
			musicVolume: 60,
			sfxVolume: 80,
			ambientVolume: 50,
			muted: false,
		},
	);

	// ─── 控制设置 ────────────────────────────
	const controls = ref<ControlSettings>(
		saved?.controls ?? {
			mouseSensitivity: 0.5,
			invertY: false,
			toggleSprint: false,
			toggleCrouch: false,
		},
	);

	// ─── 第一人称武器视图设置 ────────────────
	const weaponView = ref<WeaponViewSettings>(
		saved?.weaponView ?? {
			offsetX: 0.22,
			offsetY: -0.22,
			offsetZ: -0.45,
			scale: 0.34,
		},
	);

	// ─── 语言 ────────────────────────────────
	const language = ref<string>(saved?.language ?? "zh-CN");

	// ─── UI 缩放 ────────────────────────────
	const uiScale = ref<number>(saved?.uiScale ?? 1.0);

	// ─── 持久化 ──────────────────────────────
	function save() {
		try {
			localStorage.setItem(
				STORAGE_KEY,
				JSON.stringify({
					video: video.value,
					audio: audio.value,
					controls: controls.value,
					weaponView: weaponView.value,
					language: language.value,
					uiScale: uiScale.value,
				}),
			);
		} catch {
			// 存储失败时静默处理
		}
	}

	// 自动保存
	watch([video, audio, controls, weaponView, language, uiScale], save, {
		deep: true,
	});

	// ─── 操作 ────────────────────────────────
	function updateVideoSettings(partial: Partial<VideoSettings>) {
		video.value = { ...video.value, ...partial };
	}

	function updateAudioSettings(partial: Partial<AudioSettings>) {
		audio.value = { ...audio.value, ...partial };
	}

	function updateControlSettings(partial: Partial<ControlSettings>) {
		controls.value = { ...controls.value, ...partial };
	}

	function updateWeaponViewSettings(partial: Partial<WeaponViewSettings>) {
		weaponView.value = { ...weaponView.value, ...partial };
	}

	function resetToDefaults() {
		video.value = {
			brightness: 100,
			renderDistance: 6,
			targetFPS: 60,
			showFPS: false,
			antiAliasing: true,
			shadowQuality: "medium",
			particleEffects: true,
		};
		audio.value = {
			masterVolume: 80,
			musicVolume: 60,
			sfxVolume: 80,
			ambientVolume: 50,
			muted: false,
		};
		controls.value = {
			mouseSensitivity: 0.5,
			invertY: false,
			toggleSprint: false,
			toggleCrouch: false,
		};
		weaponView.value = {
			offsetX: 0.22,
			offsetY: -0.22,
			offsetZ: -0.45,
			scale: 0.34,
		};
		language.value = "zh-CN";
		uiScale.value = 1.0;
	}

	return {
		// State
		video,
		audio,
		controls,
		weaponView,
		language,
		uiScale,
		// Actions
		updateVideoSettings,
		updateAudioSettings,
		updateControlSettings,
		updateWeaponViewSettings,
		resetToDefaults,
		save,
	};
});
