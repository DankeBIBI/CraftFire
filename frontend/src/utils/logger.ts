import { LogInfo, LogDebug, LogError, LogWarning } from "@/wailsjs/runtime";
/**
 * 前端统一日志工具。
 * 所有日志以【前端】前缀输出到控制台，方便与后端日志区分。
 */

const PREFIX = "【前端】";

function formatArgs(module: string, ...args: unknown[]): string {
	return JSON.stringify([PREFIX, module, ...args]);
}

export const logger = {
	/** 普通信息 */
	info(module: string, ...args: unknown[]) {
		LogInfo(formatArgs(module, ...args));
	},

	/** 警告 */
	warn(module: string, ...args: unknown[]) {
		LogWarning(formatArgs(module, ...args));
	},

	/** 错误 */
	error(module: string, ...args: unknown[]) {
		LogError(formatArgs(module, ...args));
	},

	/** 调试（仅在开发模式下输出） */
	debug(module: string, ...args: unknown[]) {
		if (import.meta.env.DEV) {
			LogDebug(formatArgs(module, ...args));
		}
	},
};
