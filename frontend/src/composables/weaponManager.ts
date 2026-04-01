import { ref } from 'vue';
import type { WeaponSlot, WeaponSlotConfig } from './useWeapon';
import { WEAPON_SLOTS, WEAPON_SLOT_ORDER } from './useWeapon';
import { wsService } from '@/services/WebSocketService';

export type WeaponSwitchCallback = (config: WeaponSlotConfig) => void;

/** 武器类型标识符映射 */
const WEAPON_TYPE_MAP: Record<WeaponSlot, string> = {
	primary: 'pistol',
	secondary: 'pistol',
};

class WeaponManager {
	private currentSlot = ref<WeaponSlot>('primary');
	private onSwitchCallback: WeaponSwitchCallback | null = null;
	private initialized = false;

	readonly slotOrder = WEAPON_SLOT_ORDER;

	get currentWeaponSlot(): WeaponSlot {
		return this.currentSlot.value;
	}

	get currentWeaponConfig(): WeaponSlotConfig {
		return WEAPON_SLOTS[this.currentSlot.value];
	}

	get currentWeaponType(): string {
		return WEAPON_TYPE_MAP[this.currentSlot.value] || 'pistol';
	}

	getWeaponConfig(slot: WeaponSlot): WeaponSlotConfig {
		return WEAPON_SLOTS[slot];
	}

	get isPrimary(): boolean {
		return this.currentSlot.value === 'primary';
	}

	get isSecondary(): boolean {
		return this.currentSlot.value === 'secondary';
	}

	onSwitch(callback: WeaponSwitchCallback): void {
		this.onSwitchCallback = callback;
	}

	private notifySwitch(): void {
		if (this.onSwitchCallback) {
			this.onSwitchCallback(this.currentWeaponConfig);
		}
	}

	/** 内部切换武器逻辑 */
	private doSwitchWeapon(slot: WeaponSlot): void {
		if (slot === this.currentSlot.value) return;
		if (!WEAPON_SLOTS[slot]) return;

		this.currentSlot.value = slot;
		this.notifySwitch();

		// 同步装备切换到服务器
		this.syncEquipment();
	}

	switchWeapon(slot: WeaponSlot): void {
		this.doSwitchWeapon(slot);
	}

	switchToPrimary(): void {
		this.doSwitchWeapon('primary');
	}

	switchToSecondary(): void {
		this.doSwitchWeapon('secondary');
	}

	toggleWeapon(): void {
		const currentIndex = WEAPON_SLOT_ORDER.indexOf(this.currentSlot.value);
		const nextIndex = (currentIndex + 1) % WEAPON_SLOT_ORDER.length;
		this.doSwitchWeapon(WEAPON_SLOT_ORDER[nextIndex]);
	}

	/** 同步当前装备到服务器 */
	syncEquipment(): void {
		if (wsService.state === 'connected') {
			wsService.sendPlayerEquip(this.currentWeaponType);
		}
	}

	getSlotDisplayInfo(slot: WeaponSlot): { slot: WeaponSlot; name: string; isActive: boolean } {
		return {
			slot,
			name: WEAPON_SLOTS[slot].name,
			isActive: this.currentSlot.value === slot,
		};
	}

	getAllSlotsDisplayInfo(): Array<{ slot: WeaponSlot; name: string; isActive: boolean }> {
		return WEAPON_SLOT_ORDER.map((slot) => this.getSlotDisplayInfo(slot));
	}

	init(callback?: WeaponSwitchCallback): void {
		if (this.initialized) return;
		this.initialized = true;

		if (callback) {
			this.onSwitchCallback = callback;
		}

		this.notifySwitch();
	}

	reset(): void {
		this.currentSlot.value = 'primary';
		this.initialized = false;
	}
}

export const weaponManager = new WeaponManager();
