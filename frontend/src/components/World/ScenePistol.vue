<script setup lang="ts">
import { onMounted, onUnmounted, watch, ref } from "vue";
import { useTresContext, useRenderLoop } from "@tresjs/core";
import {
	BoxGeometry,
	Group,
	Mesh,
	MeshStandardMaterial,
	// Vector3 not currently used
	AnimationMixer,
	LoopOnce,
} from "three";
import type { Material, Object3D } from "three";
import { useSettingsStore } from "@/stores/settings";
import { usePlayerStore } from "@/stores/player";
import { useGameStateStore } from "@/stores/gameState";
import { useUIStore } from "@/stores/ui";
import { loadGLTF } from "@/loader/GLTFLoader";
import pistol, {
	type WeaponAnimationConfig,
	type WeaponAnimationState,
} from "@/loader/weapon/pistol";
import type { WeaponSlotConfig } from '@/composables/useWeapon';
import { tryCatch } from "@/utils/promise";
import { weaponManager } from "@/composables/weaponManager";
import { logger } from "@/utils/logger";
import { emitSparkEffect } from "@/utils/sparkEffects";

const { camera } = useTresContext();
const settings = useSettingsStore();
const playerStore = usePlayerStore();
const gameState = useGameStateStore();
const uiStore = useUIStore();

const MAG_SIZE = 30;
const SHOOT_COST = 1;
const FIRE_INTERVAL_MS = 80;

let weaponConfig = pistol.lowpoly_generic_pistol_43;
let animConfig: WeaponAnimationConfig | null = null;
let loadedAnimations: ReturnType<typeof loadGLTF> extends Promise<infer R> ? R extends { animations: infer A } ? A : never : never = [];

const firstPersonRig = new Group();
firstPersonRig.position.set(0, 0, 0);
firstPersonRig.name = "FirstPersonRig";

// 武器挂点（相机局部空间）
const weaponAnchor = new Group();
weaponAnchor.rotation.set(-0.1, -0.35, 0);
firstPersonRig.add(weaponAnchor);

// 简化人物模型：第一人称双前臂（跟随相机，避免"只有枪没有人"）
const leftArm = new Mesh(
	new BoxGeometry(0.05, 0.05, 0.2),
	new MeshStandardMaterial({
		color: "#d7a882",
		roughness: 0.85,
		metalness: 0.05,
	}),
);
leftArm.position.set(-0.22, -0.26, -0.4);
leftArm.rotation.set(0.25, 0.12, -0.12);
firstPersonRig.add(leftArm);

const rightArm = new Mesh(
	new BoxGeometry(0.05, 0.05, 0.2),
	new MeshStandardMaterial({
		color: "#d7a882",
		roughness: 0.85,
		metalness: 0.05,
	}),
);
rightArm.position.set(0.12, -0.16, -0.3);
rightArm.rotation.set(0.22, -0.18, 0.08);
firstPersonRig.add(rightArm);

let pistolModelRoot: Object3D | null = null;
let fallbackWeaponMesh: Mesh | null = null;
let attachedCamera: Object3D | null = null;
let unwatchCamera: (() => void) | null = null;
let unwatchWeaponView: (() => void) | null = null;

const animState: WeaponAnimationState = {
	mixer: null,
	shootAction: null,
	reloadAction: null,
};

const isAnimating = ref(false);
const isReloading = ref(false);
let lastShootAt = 0;

function applyMeshSettings(root: Object3D) {
	root.traverse((node) => {
		const mesh = node as Mesh & {
			isMesh?: boolean;
			material?: Material | Material[];
		};
		if (!mesh.isMesh) return;

		mesh.castShadow = false;
		mesh.receiveShadow = false;
		mesh.frustumCulled = false;

		const material = mesh.material as Material | Material[];
		if (Array.isArray(material)) {
			material.forEach((item) => {
				item.needsUpdate = true;
			});
		} else if (material) {
			material.needsUpdate = true;
		}
	});
}

function createFallbackWeapon() {
	if (fallbackWeaponMesh) return;

	fallbackWeaponMesh = new Mesh(
		new BoxGeometry(0.32, 0.18, 0.58),
		new MeshStandardMaterial({
			color: "#4b4b4b",
			roughness: 0.75,
			metalness: 0.2,
		}),
	);
	fallbackWeaponMesh.position.set(0, 0, 0.04);
	fallbackWeaponMesh.rotation.set(0.08, 0.05, -0.06);
	fallbackWeaponMesh.frustumCulled = false;
	weaponAnchor.add(fallbackWeaponMesh);
}

async function loadWeapon(config: WeaponSlotConfig) {
	logger.info("[Pistol] loadWeapon 被调用, config:", config);

	if (pistolModelRoot) {
		weaponAnchor.remove(pistolModelRoot);
		pistolModelRoot = null;
	}

	if (fallbackWeaponMesh) {
		weaponAnchor.remove(fallbackWeaponMesh);
		fallbackWeaponMesh.geometry.dispose();
		const mat = fallbackWeaponMesh.material as Material;
		if (Array.isArray(mat)) {
			mat.forEach((m) => m.dispose());
		} else {
			mat.dispose();
		}
		fallbackWeaponMesh = null;
	}

	if (animState.mixer) {
		animState.mixer.stopAllAction();
		animState.mixer = null;
		animState.shootAction = null;
		animState.reloadAction = null;
	}

	weaponConfig = config.config;
animConfig = weaponConfig.animationMap ?? null;

if (!animConfig) {
	logger.warn("[Pistol] weapon config missing animationMap, animations will not play", weaponConfig);
}
logger.info("[Pistol] 开始加载武器:", weaponConfig);

	const [err, gltfResult] = await tryCatch(loadGLTF(weaponConfig));

	if (err || !gltfResult) {
		logger.error("[Pistol] 加载GLTF失败:", err);
		createFallbackWeapon();
		return;
	}

	const { scene, animations } = gltfResult;
	logger.info("[Pistol] GLTF加载成功, 动画数量:", animations.length);

	// 保存动画列表供 playShootAnimation 使用
	loadedAnimations = animations;

	applyMeshSettings(scene);
	weaponAnchor.add(scene);

	pistolModelRoot = scene;

	// 设置 AnimationMixer - 直接使用动画名称查找
	if (animations.length > 0) {
		logger.info(
			"[Pistol] 加载到的动画:",
			animations.map((a) => a.name),
		);
		logger.info("[Pistol] 动画配置:", animConfig);

		animState.mixer = new AnimationMixer(scene);

			// 获取动画名称（支持对象配置）
			const shootName =
				typeof animConfig?.shoot === "string"
					? animConfig.shoot
					: animConfig?.shoot?.name;
			const reloadName =
				typeof animConfig?.reload === "string"
					? animConfig.reload
					: animConfig?.reload?.name;

			// 查找射击动画 - 使用配置中的名称
			const shootClip = animations.find((clip) => clip.name === shootName);
		logger.info(
			"[Pistol] 射击动画:",
			shootClip ? `找到: ${shootClip.name}` : "未找到",
		);
		if (shootClip) {
			animState.shootAction = animState.mixer.clipAction(shootClip);
			// loop once (second arg makes TypeScript happy)
			animState.shootAction.setLoop(LoopOnce, 1);
			animState.shootAction.clampWhenFinished = true;
		}

		// 查找换弹动画 - 使用配置中的名称
		const reloadClip = animations.find(
			(clip) => clip.name === reloadName,
		);
		logger.info(
			"[Pistol] 换弹动画:",
			reloadClip ? `找到: ${reloadClip.name}` : "未找到",
		);
		if (reloadClip) {
			animState.reloadAction = animState.mixer.clipAction(reloadClip);
			// loop once
			animState.reloadAction.setLoop(LoopOnce, 1);
			animState.reloadAction.clampWhenFinished = true;				if (animConfig && typeof animConfig.reload !== "string") {
					animConfig.reload.play = () => {
						animState.reloadAction?.reset();
						animState.reloadAction?.play();
					};
				}		}
	} else {
		logger.warn("[Pistol] 没有找到动画");
	}
}

function attachRigToCamera(cameraObject: Object3D | null) {
	if (!cameraObject) return;
	if (attachedCamera === cameraObject) return;

	if (attachedCamera) {
		attachedCamera.remove(firstPersonRig);
	}

	cameraObject.add(firstPersonRig);
	attachedCamera = cameraObject;
}

/**
 * 播放射击动画
 * 使用 Blender 导出的 GLTF 原生动画
 */
function playShootAnimation() {
	if (isReloading.value) return;

	// 获取动画名称（支持数组或字符串）
	const getAnimNames = (config: typeof animConfig) => {
		if (!config) return [];
		const shoot = config.shoot;
		if (!shoot) return [];
		if (typeof shoot === "string") return [shoot];
		const name = shoot.name;
		return Array.isArray(name) ? name : [name];
	};

	// 播放所有射击动画
	const names = getAnimNames(animConfig);
	for (const name of names) {
		const clip = animState.mixer?.clipAction(
			loadedAnimations.find((a) => a.name === name),
		);
		if (clip) {
			clip.setLoop(LoopOnce, 1);
			clip.clampWhenFinished = true;
			clip.reset();
			clip.play();
		}
	}

	if (names.length > 0) {
		isAnimating.value = true;
	}
}

/**
 * 播放换弹动画
 * 使用 Blender 导出的 GLTF 原生动画 - "reload_magazine"
 */
function playReloadAnimation() {
	if (isReloading.value) return;

	isReloading.value = true;
	isAnimating.value = true;

	// 获取动画名称（支持数组或字符串）
	const getAnimNames = (config: typeof animConfig) => {
		if (!config) return [];
		const reload = config.reload;
		if (!reload) return [];
		if (typeof reload === "string") return [reload];
		const name = reload.name;
		return Array.isArray(name) ? name : [name];
	};

	// 播放所有换弹动画
	const names = getAnimNames(animConfig);
	for (const name of names) {
		const clip = animState.mixer?.clipAction(
			loadedAnimations.find((a) => a.name === name),
		);
		if (clip) {
			clip.setLoop(LoopOnce, 1);
			clip.clampWhenFinished = true;
			clip.reset();
			clip.play();
		}
	}

	if (names.length > 0) {
		// 800ms 后自动填满弹药
		window.setTimeout(() => {
			playerStore.refillLocalAmmo(MAG_SIZE);
			isReloading.value = false;
			isAnimating.value = false;
		}, 800);
	} else {
		playerStore.refillLocalAmmo(MAG_SIZE);
		isReloading.value = false;
		isAnimating.value = false;
	}
}

function tryShoot(): void {
	if (!document.pointerLockElement) return;
	if (gameState.isPaused || uiStore.hasActiveOverlay) return;
	if (!gameState.isRunning) return;
	if (!playerStore.localPlayer?.isAlive) return;

	const now = performance.now();
	if (now - lastShootAt < FIRE_INTERVAL_MS) return;

	if (!playerStore.canLocalShoot(SHOOT_COST)) {
		playReloadAnimation();
		return;
	}

	const consumed = playerStore.consumeLocalAmmo(SHOOT_COST);
	if (!consumed) {
		playReloadAnimation();
		return;
	}

	lastShootAt = now;

	const pos = playerStore.position;
	const rot = playerStore.rotation;
	const dir = {
		x: -Math.sin(rot.yaw) * Math.cos(rot.pitch),
		y: Math.sin(rot.pitch),
		z: -Math.cos(rot.yaw) * Math.cos(rot.pitch),
	};
	emitSparkEffect({
		kind: "muzzle",
		position: {
			x: pos.x + dir.x * 0.55,
			y: pos.y + 1.45 + dir.y * 0.55,
			z: pos.z + dir.z * 0.55,
		},
		normal: dir,
	});

	playShootAnimation();

	if (!playerStore.canLocalShoot(SHOOT_COST)) {
		playReloadAnimation();
	}
}

function tryReload(): void {
	if (!playerStore.localPlayer?.isAlive) return;
	if (playerStore.getLocalAmmo() >= MAG_SIZE) return;
	playReloadAnimation();
}

/**
 * 每帧更新动画状态
 * @param delta - 距离上一帧的时间增量（秒）
 */
function updateAnimation(delta: number) {
	if (animState.mixer) {
		animState.mixer.update(delta);
	}

	if (animState.shootAction && !animState.shootAction.isRunning()) {
		isAnimating.value = false;
	}

	if (isReloading.value && animState.reloadAction && !animState.reloadAction.isRunning()) {
		playerStore.refillLocalAmmo(MAG_SIZE);
		isReloading.value = false;
		isAnimating.value = false;
	}
}

function onMouseDown(event: MouseEvent) {
	if (event.button === 0) {
		tryShoot();
	}
}

function onKeyDown(event: KeyboardEvent) {
	if (event.code === "KeyR") {
		tryReload();
	}
}

onMounted(async () => {

	document.addEventListener("mousedown", onMouseDown);
	document.addEventListener("keydown", onKeyDown);

	weaponManager.onSwitch(loadWeapon);

	weaponManager.init(loadWeapon);
	logger.info("[Pistol] weaponManager.init 完成");

	unwatchWeaponView = watch(
		() => settings.weaponView,
		(weaponView) => {
			weaponAnchor.position.set(
				weaponView.offsetX,
				weaponView.offsetY,
				weaponView.offsetZ,
			);
		},
		{ immediate: true, deep: true },
	);

	unwatchCamera = watch(
		() => camera.value,
		(cam) => {
			attachRigToCamera((cam as unknown as Object3D | null) ?? null);
		},
		{ immediate: true },
	);

	const { onLoop } = useRenderLoop();
	onLoop(({ delta }) => {
		updateAnimation(delta);
	});
});

onUnmounted(() => {
	document.removeEventListener("mousedown", onMouseDown);
	document.removeEventListener("keydown", onKeyDown);

	if (unwatchWeaponView) {
		unwatchWeaponView();
		unwatchWeaponView = null;
	}

	if (unwatchCamera) {
		unwatchCamera();
		unwatchCamera = null;
	}

	if (pistolModelRoot) {
		weaponAnchor.remove(pistolModelRoot);
		pistolModelRoot = null;
	}

	if (fallbackWeaponMesh) {
		weaponAnchor.remove(fallbackWeaponMesh);
		fallbackWeaponMesh.geometry.dispose();
		const mat = fallbackWeaponMesh.material;
		if (Array.isArray(mat)) {
			mat.forEach((m) => m.dispose());
		} else {
			mat.dispose();
		}
		fallbackWeaponMesh = null;
	}

	if (attachedCamera) {
		attachedCamera.remove(firstPersonRig);
		attachedCamera = null;
	}
});
</script>

<template />
