/**
 * CraftFire 3D 模型相关类型定义。
 */

/** 模型格式 */
export type ModelFormat = "gltf" | "glb" | "fbx" | "obj" | "dae";

/** 支持的模型格式列表 */
export const SUPPORTED_FORMATS: ModelFormat[] = [
	"gltf",
	"glb",
	"fbx",
	"obj",
	"dae",
];

/** 模型元数据 */
export interface ModelMetadata {
	vertexCount?: number;
	triangleCount?: number;
	materials?: number;
	textures?: number;
	hasAnimations?: boolean;
}

/** 模型信息 */
export interface ModelInfo {
	modelId: string;
	name: string;
	format: ModelFormat;
	fileSize: number;
	filePath: string;
	md5Hash: string;
	uploadedAt: number;
	uploadedBy: string;
	thumbnailUrl?: string;
	metadata?: ModelMetadata;
	version: number;
	isPublic: boolean;
}

/** 模型导入状态 */
export interface ImportStatus {
	stage: "idle" | "validating" | "uploading" | "processing" | "completed" | "downloading" | "error";
	progress: number;
	error?: string;
}

/** 最大文件大小（10MB） */
export const MAX_MODEL_FILE_SIZE = 10 * 1024 * 1024;
