/**
 * 3D 模型管理服务。
 * 封装模型导入、列表、下载和局域网同步的后端接口。
 */

import type { ModelInfo } from "@/types/model";
import { SUPPORTED_FORMATS, MAX_MODEL_FILE_SIZE } from "@/types/model";

/** 文件验证结果 */
export interface ValidateResult {
	valid: boolean;
	error?: string;
}

/**
 * 验证模型文件格式和大小。
 *
 * @param filePath - 待验证的文件路径
 * @param fileSize - 文件大小（字节），可选
 * @returns 验证结果对象
 */
export function validateModelFile(
	filePath: string,
	fileSize?: number,
): ValidateResult {
	const ext = filePath.split(".").pop()?.toLowerCase();
	if (!ext) {
		return { valid: false, error: "文件缺少扩展名" };
	}

	const validFormats = new Set(SUPPORTED_FORMATS);
	if (!validFormats.has(ext as (typeof SUPPORTED_FORMATS)[number])) {
		return {
			valid: false,
			error: `不支持的文件格式: .${ext}`,
		};
	}

	if (fileSize !== undefined) {
		if (fileSize > MAX_MODEL_FILE_SIZE) {
			return {
				valid: false,
				error: `文件大小超限: ${(fileSize / 1024 / 1024).toFixed(1)}MB (最大 10MB)`,
			};
		}

		if (fileSize === 0) {
			return { valid: false, error: "文件为空" };
		}
	}

	return { valid: true };
}

/** 导入 3D 模型 */
export async function importModel(
	filePath: string,
	roomId = "",
): Promise<string> {
	if (window.go?.main?.App) {
		return window.go.main.App.ImportModel(filePath, roomId);
	}
	return "mock-model-id";
}

/** 获取可用模型列表 */
export async function listModels(roomId = ""): Promise<ModelInfo[]> {
	if (window.go?.main?.App) {
		return window.go.main.App.ListAvailableModels(roomId) as Promise<
			ModelInfo[]
		>;
	}
	return [];
}

/** 获取模型详细信息 */
export async function getModelInfo(modelId: string): Promise<ModelInfo> {
	if (window.go?.main?.App) {
		return window.go.main.App.GetModelInfo(modelId) as Promise<ModelInfo>;
	}
	throw new Error("后端不可用");
}

/** 删除模型 */
export async function deleteModel(modelId: string): Promise<void> {
	if (window.go?.main?.App) {
		return window.go.main.App.DeleteModel(modelId);
	}
}

/** 同步局域网内模型 */
export async function syncModelsInLAN(): Promise<ModelInfo[]> {
	if (window.go?.main?.App) {
		return window.go.main.App.SyncModelsInLAN() as Promise<ModelInfo[]>;
	}
	return [];
}

/** 从局域网下载模型 */
export async function downloadModelFromLAN(
	modelId: string,
	sourceIP: string,
): Promise<void> {
	if (window.go?.main?.App) {
		return window.go.main.App.DownloadModelFromLAN(modelId, sourceIP);
	}
}
