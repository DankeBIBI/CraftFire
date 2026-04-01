/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_API_BASE_URL: string;
	readonly VITE_DEBUG_MODE: string;
	readonly VITE_LOG_LEVEL: string;
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}

declare module "*.vue" {
	import type { DefineComponent } from "vue";
	const component: DefineComponent<{}, {}, any>;
	export default component;
}

declare module "*.gltf?url" {
	const src: string;
	export default src;
}

declare module "*.gltf" {
	const src: string;
	export default src;
}

declare module "*?url" {
	const src: string;
	export default src;
}
