
import type { AnimationMixer } from "three";

// weapon animation definitions are now embedded in WeaponConfig so they stay grouped
export type WeaponAnimationState = {
  mixer: AnimationMixer | null;
  shootAction: ReturnType<AnimationMixer["clipAction"]> | null;
  reloadAction: ReturnType<AnimationMixer["clipAction"]> | null;
};

export type WeaponConfig = {
  url: string;
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
  scale: number;
  animationMap: {
    shoot: {
      name: string | string[];
      play?: () => void;
    };
    reload: {
      name: string | string[];
      play?: () => void;
    };
  };
};

// alias for convenience/compatibility with existing code
export type WeaponAnimationConfig = WeaponConfig['animationMap'];

const weaponConfigs = {
  lowpoly_generic_pistol_43: {
    url: "/src/assets/models/lowpoly_generic_pistol_43/scene.gltf",
    position: { x: 0.2, y: 0.14, z: 0.15 },
    rotation: { x: 0, y: 0.8, z: 0 },
    scale: 1.2,
    animationMap: {
      shoot: { name: ["shoot", "reload_magazine_G43"] },
      reload: { name: "shoot_combined" },
    },
  },
} as Record<string, WeaponConfig>;

export default weaponConfigs;
