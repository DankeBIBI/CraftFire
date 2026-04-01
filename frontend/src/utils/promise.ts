import { logger } from "./logger";

export async function tryCatch<T>(
	promise: Promise<T>,
): Promise<[null, T] | [Error, never]> {
	try {
		const res = await promise;
		return [null, res];
	} catch (e: unknown) {
		logger.error("PromiseUtils", "Promise rejected:", e);
		const error = e instanceof Error ? e : new Error(String(e));
		return [error, undefined as never];
	}
}
