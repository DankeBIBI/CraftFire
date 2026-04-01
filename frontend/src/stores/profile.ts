/**
 * 玩家个人资料状态管理。
 * 管理个人信息、统计数据和角色定制。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type {
	PlayerProfile,
	PlayerStatistics,
	PlayerProfileUpdate,
} from "@/types/profile";
import * as ProfileService from "@/services/ProfileService";

export const useProfileStore = defineStore("profile", () => {
	// ─── 状态 ────────────────────────────────
	const profile = ref<PlayerProfile | null>(null);
	const statistics = ref<PlayerStatistics | null>(null);
	const isLoading = ref(false);
	const error = ref<string | null>(null);

	// ─── 计算属性 ────────────────────────────
	const nickname = computed(() => profile.value?.nickname ?? "Player");
	const level = computed(() => profile.value?.level ?? 1);
	const experience = computed(() => profile.value?.experience ?? 0);
	const nextLevelExp = computed(() => profile.value?.nextLevelExp ?? 100);
	const expProgress = computed(() => {
		if (!profile.value) return 0;
		return Math.floor(
			(profile.value.experience / profile.value.nextLevelExp) * 100,
		);
	});
	const totalPlayTime = computed(() => profile.value?.totalPlayTime ?? 0);

	// ─── 操作 ────────────────────────────────

	/** 加载玩家资料 */
	async function loadProfile() {
		isLoading.value = true;
		error.value = null;
		try {
			const data = await ProfileService.getPlayerProfile();
			profile.value = data;
		} catch (err: unknown) {
			error.value =
				err instanceof Error ? err.message : "加载个人资料失败";
		} finally {
			isLoading.value = false;
		}
	}

	/** 更新玩家资料 */
	async function updateProfile(updates: PlayerProfileUpdate) {
		isLoading.value = true;
		error.value = null;
		try {
			const updated = await ProfileService.updatePlayerProfile(updates);
			if (updated !== undefined) {
				profile.value = updated;
			}
		} catch (err: unknown) {
			error.value =
				err instanceof Error ? err.message : "更新个人资料失败";
		} finally {
			isLoading.value = false;
		}
	}

	/** 加载统计数据 */
	async function loadStatistics() {
		isLoading.value = true;
		error.value = null;
		try {
			const data = await ProfileService.getPlayerStatistics();
			statistics.value = data;
		} catch (err: unknown) {
			error.value =
				err instanceof Error ? err.message : "加载统计数据失败";
		} finally {
			isLoading.value = false;
		}
	}

	/** 更新昵称 */
	async function updateNickname(name: string) {
		await updateProfile({ nickname: name });
	}

	/** 更新角色定制 */
	async function updateCustomization(
		customization: PlayerProfile["customization"],
	) {
		await updateProfile({ customization });
	}

	function $reset() {
		profile.value = null;
		statistics.value = null;
		isLoading.value = false;
		error.value = null;
	}

	return {
		// State
		profile,
		statistics,
		isLoading,
		error,
		// Computed
		nickname,
		level,
		experience,
		nextLevelExp,
		expProgress,
		totalPlayTime,
		// Actions
		loadProfile,
		updateProfile,
		loadStatistics,
		updateNickname,
		updateCustomization,
		$reset,
	};
});
