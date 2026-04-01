/**
 * UI 状态管理。
 * 控制 Toast 通知、弹窗、调试面板等 UI 全局状态。
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";

export interface ToastItem {
  id: string;
  message: string;
  type: "info" | "success" | "warning" | "error";
  duration: number;
  createdAt: number;
}

export const useUIStore = defineStore("ui", () => {
  // ─── 状态 ────────────────────────────────
  const toasts = ref<ToastItem[]>([]);
  const showPauseMenu = ref(false);
  const showSettings = ref(false);
  const showDebugPanel = ref(false);
  const showAdminPanel = ref(false);
  const showPlayerProfile = ref(false);
  const showModelImporter = ref(false);
  const crosshairVisible = ref(true);
  const chatOpen = ref(false);
  const uiScale = ref(1.0);

  // ─── 计算属性 ────────────────────────────
  const hasActiveOverlay = computed(
    () =>
      showPauseMenu.value ||
      showSettings.value ||
      showAdminPanel.value ||
      showPlayerProfile.value ||
      showModelImporter.value,
  );

  // ─── Toast 操作 ──────────────────────────
  let _toastCounter = 0;

  function showToast(
    message: string,
    type: ToastItem["type"] = "info",
    duration = 3000,
  ) {
    const id = `toast-${++_toastCounter}`;
    const toast: ToastItem = {
      id,
      message,
      type,
      duration,
      createdAt: Date.now(),
    };
    toasts.value.push(toast);

    // 自动移除
    if (duration > 0) {
      setTimeout(() => removeToast(id), duration);
    }
    return id;
  }

  function removeToast(id: string) {
    const idx = toasts.value.findIndex((t) => t.id === id);
    if (idx >= 0) toasts.value.splice(idx, 1);
  }

  function clearAllToasts() {
    toasts.value = [];
  }

  // ─── Confirm 操作 ──────────────────────────
  let _confirmResolve: ((value: boolean) => void) | null = null;

  /**
   * 显示确认弹窗，返回 Promise
   */
  function confirm(message: string): Promise<boolean> {
    return new Promise((resolve) => {
      _confirmResolve = resolve;
      // 触发全局确认弹窗显示（需配合 Confirm 组件）
      window.dispatchEvent(new CustomEvent('show-confirm', { detail: { message } }));
    });
  }

  /**
   * 确认弹窗结果处理
   */
  function confirmResult(confirmed: boolean) {
    if (_confirmResolve) {
      _confirmResolve(confirmed);
      _confirmResolve = null;
    }
  }

  // ─── 面板切换 ────────────────────────────
  function togglePauseMenu() {
    showPauseMenu.value = !showPauseMenu.value;
  }

  function toggleSettings() {
    showSettings.value = !showSettings.value;
  }

  function toggleDebugPanel() {
    showDebugPanel.value = !showDebugPanel.value;
  }

  function toggleAdminPanel() {
    showAdminPanel.value = !showAdminPanel.value;
  }

  function togglePlayerProfile() {
    showPlayerProfile.value = !showPlayerProfile.value;
  }

  function toggleModelImporter() {
    showModelImporter.value = !showModelImporter.value;
  }

  function closeAllOverlays() {
    showPauseMenu.value = false;
    showSettings.value = false;
    showAdminPanel.value = false;
    showPlayerProfile.value = false;
    showModelImporter.value = false;
  }

  function setUIScale(scale: number) {
    uiScale.value = Math.max(0.5, Math.min(2.0, scale));
  }

  function $reset() {
    toasts.value = [];
    showPauseMenu.value = false;
    showSettings.value = false;
    showDebugPanel.value = false;
    showAdminPanel.value = false;
    showPlayerProfile.value = false;
    showModelImporter.value = false;
    crosshairVisible.value = true;
    chatOpen.value = false;
    uiScale.value = 1.0;
  }

  return {
    // State
    toasts,
    showPauseMenu,
    showSettings,
    showDebugPanel,
    showAdminPanel,
    showPlayerProfile,
    showModelImporter,
    crosshairVisible,
    chatOpen,
    uiScale,
    // Computed
    hasActiveOverlay,
    // Actions
    showToast,
    removeToast,
    clearAllToasts,
    confirm,
    confirmResult,
    togglePauseMenu,
    toggleSettings,
    toggleDebugPanel,
    toggleAdminPanel,
    togglePlayerProfile,
    toggleModelImporter,
    closeAllOverlays,
    setUIScale,
    $reset,
  };
});
