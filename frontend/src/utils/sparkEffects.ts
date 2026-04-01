export const SPARK_EFFECT_EVENT = "craftfire:spark";

export type SparkEffectKind = "muzzle" | "impact";

export interface SparkEffectPayload {
	kind: SparkEffectKind;
	position: { x: number; y: number; z: number };
	normal?: { x: number; y: number; z: number };
}

export function emitSparkEffect(payload: SparkEffectPayload): void {
	window.dispatchEvent(
		new CustomEvent<SparkEffectPayload>(SPARK_EFFECT_EVENT, {
			detail: payload,
		}),
	);
}
