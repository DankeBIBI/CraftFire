/**
 * 3D 模型管理状态。
 * 管理模型导入、列表、局域网同步等。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { ModelInfo, ImportStatus } from "@/types/model";
import * as ModelService from "@/services/ModelService";

export const useModelStore = defineStore("model", () => {
	// ─── 状态 ────────────────────────────────
	const models = ref<ModelInfo[]>([]);
	const selectedModel = ref<ModelInfo | null>(null);
	const importStatus = ref<ImportStatus | null>(null);
	const isImporting = ref(false);
	const isLoading = ref(false);
	const error = ref<string | null>(null);

	// ─── 计算属性 ────────────────────────────
	const modelCount = computed(() => models.value.length);
	const publicModels = computed(() => models.value.filter((m) => m.isPublic));

	// ─── 操作 ────────────────────────────────

	/** 加载所有可用模型 */
	async function loadModels() {
		isLoading.value = true;
		error.value = null;
		try {
			const list = await ModelService.listModels();
			models.value = list;
		} catch (err: unknown) {
			error.value =
				err instanceof Error ? err.message : "加载模型列表失败";
		} finally {
			isLoading.value = false;
		}
	}

	/** 导入模型文件 */
	async function importModel(filePath: string, name: string) {
		// 先进行客户端校验
		const validation = ModelService.validateModelFile(filePath);
		if (!validation.valid) {
			error.value = validation.error ?? "模型文件校验失败";
			return false;
		}

		isImporting.value = true;
		importStatus.value = { stage: "uploading", progress: 0 };
		error.value = null;

		try {
			const result = await ModelService.importModel(filePath, name);
			if (result) {
				importStatus.value = { stage: "completed", progress: 100 };
				// 刷新模型列表
				await loadModels();
				return true;
			}
			error.value = "导入模型失败";
			return false;
		} catch (err: unknown) {
			const msg = err instanceof Error ? err.message : "导入模型时出错";
			error.value = msg;
			importStatus.value = {
				stage: "error",
				progress: 0,
				error: msg,
			};
			return false;
		} finally {
			isImporting.value = false;
		}
	}

	/** 获取模型详情 */
	async function getModelInfo(modelId: string) {
		isLoading.value = true;
		try {
			const info = await ModelService.getModelInfo(modelId);
			if (info) {
				selectedModel.value = info;
			}
		} catch (err: unknown) {
			error.value =
				err instanceof Error ? err.message : "获取模型信息失败";
		} finally {
			isLoading.value = false;
		}
	}

	/** 删除模型 */
	async function deleteModel(modelId: string) {
		try {
			await ModelService.deleteModel(modelId);
			models.value = models.value.filter((m) => m.modelId !== modelId);
			if (selectedModel.value?.modelId === modelId) {
				selectedModel.value = null;
			}
			return true;
		} catch (err: unknown) {
			error.value = err instanceof Error ? err.message : "删除模型失败";
			return false;
		}
	}

	/** 同步局域网模型 */
	async function syncLANModels() {
		isLoading.value = true;
		try {
			await ModelService.syncModelsInLAN();
			await loadModels();
		} catch (err: unknown) {
			error.value =
				err instanceof Error ? err.message : "同步局域网模型失败";
		} finally {
			isLoading.value = false;
		}
	}

	/** 下载局域网模型 */
	async function downloadFromLAN(modelId: string, sourceIP: string) {
		isImporting.value = true;
		importStatus.value = { stage: "downloading", progress: 0 };
		try {
			await ModelService.downloadModelFromLAN(modelId, sourceIP);
			importStatus.value = { stage: "completed", progress: 100 };
			await loadModels();
			return true;
		} catch (err: unknown) {
			error.value = err instanceof Error ? err.message : "下载模型失败";
			importStatus.value = { stage: "error", progress: 0 };
			return false;
		} finally {
			isImporting.value = false;
		}
	}

	function selectModel(model: ModelInfo | null) {
		selectedModel.value = model;
	}

	function clearImportStatus() {
		importStatus.value = null;
	}

	function $reset() {
		models.value = [];
		selectedModel.value = null;
		importStatus.value = null;
		isImporting.value = false;
		isLoading.value = false;
		error.value = null;
	}

	return {
		// State
		models,
		selectedModel,
		importStatus,
		isImporting,
		isLoading,
		error,
		// Computed
		modelCount,
		publicModels,
		// Actions
		loadModels,
		importModel,
		getModelInfo,
		deleteModel,
		syncLANModels,
		downloadFromLAN,
		selectModel,
		clearImportStatus,
		$reset,
	};
});
