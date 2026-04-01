<script setup lang="ts">
/**
 * Toast - 全局消息通知组件。
 * 在屏幕右上角显示临时消息提示。
 */
import { useUIStore, type ToastItem } from '@/stores/ui'

const ui = useUIStore()

function getTypeClass(type: ToastItem['type']) {
  switch (type) {
    case 'success':
      return 'border-green-500/50 bg-green-900/30'
    case 'warning':
      return 'border-yellow-500/50 bg-yellow-900/30'
    case 'error':
      return 'border-red-500/50 bg-red-900/30'
    default:
      return 'border-craft-primary/50 bg-craft-primary/10'
  }
}

function getTypeIcon(type: ToastItem['type']) {
  switch (type) {
    case 'success':
      return '✓'
    case 'warning':
      return '⚠'
    case 'error':
      return '✕'
    default:
      return 'ℹ'
  }
}
</script>

<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
      <TransitionGroup name="toast">
        <div
          v-for="toast in ui.toasts"
          :key="toast.id"
          class="pointer-events-auto px-4 py-3 rounded border backdrop-blur-sm min-w-[240px] max-w-[360px] flex items-start gap-2 shadow-lg"
          :class="getTypeClass(toast.type)"
        >
          <span class="text-sm mt-0.5">{{ getTypeIcon(toast.type) }}</span>
          <p class="text-craft-text text-xs font-game flex-1">{{ toast.message }}</p>
          <button
            class="text-craft-text/40 hover:text-craft-text text-xs ml-2"
            @click="ui.removeToast(toast.id)"
          >
            ✕
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active {
  transition: all 0.3s ease-out;
}
.toast-leave-active {
  transition: all 0.2s ease-in;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(30px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(30px);
}
.toast-move {
  transition: transform 0.3s ease;
}
</style>
