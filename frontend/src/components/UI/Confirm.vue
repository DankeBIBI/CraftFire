<script setup lang="ts">
/**
 * Confirm - 全局确认弹窗组件。
 * 在屏幕中央显示确认对话框。
 * 通过 window 事件触发：show-confirm
 */
import { ref, onMounted, onUnmounted } from 'vue'
import { useUIStore } from '@/stores/ui'

const ui = useUIStore()

const showConfirmModal = ref(false)
const confirmMessage = ref('')

/** 处理确认 */
function handleConfirm() {
  showConfirmModal.value = false
  ui.confirmResult(true)
}

/** 处理取消 */
function handleCancel() {
  showConfirmModal.value = false
  ui.confirmResult(false)
}

/** 监听显示确认弹窗事件 */
function onShowConfirm(event: Event) {
  const customEvent = event as CustomEvent<{ message: string }>
  confirmMessage.value = customEvent.detail.message
  showConfirmModal.value = true
}

onMounted(() => {
  window.addEventListener('show-confirm', onShowConfirm)
})

onUnmounted(() => {
  window.removeEventListener('show-confirm', onShowConfirm)
})
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="showConfirmModal"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-[200] pointer-events-auto"
        @click.self="handleCancel"
      >
        <div class="panel-lowpoly w-full max-w-sm p-6 pointer-events-auto">
          <h3 class="text-craft-light text-sm font-game mb-4">⚠️ 确认操作</h3>
          <p class="text-craft-text text-xs font-game mb-6">{{ confirmMessage }}</p>
          <div class="flex gap-2">
            <button class="btn-lowpoly flex-1" @click="handleCancel">取消</button>
            <button class="btn-danger flex-1" @click="handleConfirm">确认</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-enter-active {
  transition: all 0.2s ease-out;
}
.modal-leave-active {
  transition: all 0.15s ease-in;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
.modal-enter-from .panel-lowpoly,
.modal-leave-to .panel-lowpoly {
  transform: scale(0.95);
}
</style>
