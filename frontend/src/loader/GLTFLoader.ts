import { logger } from "@/utils/logger";
import { tryCatch } from "@/utils/promise";
import { GLTFLoader } from "three/examples/jsm/loaders/GLTFLoader";
import type { AnimationClip } from "three";

type xyz = {
	x: number;
	y: number;
	z: number;
};

export type GLTFType = {
	url: string;
	position: xyz;
	rotation: xyz;
	scale: number;
};

export type GLTFResult = {
	scene: import("three").Group;
	animations: AnimationClip[];
};

/** 加载GLFT */
export async function loadGLTF(model: GLTFType) {
	const { url, position, rotation, scale } = model;
	const loader = new GLTFLoader();
	const [err, gltf] = await tryCatch(loader.loadAsync(url));
	if (err) {
		logger.error("GLTFLoader", "加载模型失败:", url);
		return;
	}

	const { scene, animations } = gltf;
	scene.scale.setScalar(scale);
	scene.position.set(position.x, position.y, position.z);
	scene.rotation.set(rotation.x, rotation.y, rotation.z);

	return { scene, animations };
}
