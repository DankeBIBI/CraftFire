import type { WeaponConfig } from '@/loader/weapon/pistol';

export type WeaponSlot = 'primary' | 'secondary';

export interface WeaponSlotConfig {
  slot: WeaponSlot;
  name: string;
  config: WeaponConfig;
}

import pistolConfigs from '@/loader/weapon/pistol';

export const WEAPON_SLOTS: Record<WeaponSlot, WeaponSlotConfig> = {
  primary: {
    slot: 'primary',
    name: 'Primary',
    // reuse the config defined in the pistol loader (includes animationMap)
    config: pistolConfigs.lowpoly_generic_pistol_43,
  },
  secondary: {
    slot: 'secondary',
    name: 'Secondary',
    config: pistolConfigs.lowpoly_generic_pistol_43,
  },
};

export const WEAPON_SLOT_ORDER: WeaponSlot[] = ['primary', 'secondary'];
