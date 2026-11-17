<script setup lang="ts">
import {
  AlertDialogRoot,
  AlertDialogPortal,
  AlertDialogOverlay,
  AlertDialogContent,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogAction,
  AlertDialogCancel,
} from 'radix-vue'
import { AlertTriangle } from 'lucide-vue-next'

const props = withDefaults(
  defineProps<{
    open: boolean
    title?: string
    description?: string
    confirmText?: string
    cancelText?: string
    variant?: 'danger' | 'default'
  }>(),
  {
    title: 'Are you sure?',
    description: 'This action cannot be undone.',
    confirmText: 'Delete',
    cancelText: 'Cancel',
    variant: 'danger',
  }
)

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
  (e: 'confirm'): void
  (e: 'cancel'): void
}>()

function handleConfirm() {
  emit('confirm')
  emit('update:open', false)
}

function handleCancel() {
  emit('cancel')
  emit('update:open', false)
}
</script>

<template>
  <AlertDialogRoot :open="open" @update:open="val => emit('update:open', val)">
    <AlertDialogPortal>
      <!-- Backdrop Overlay -->
      <AlertDialogOverlay
        class="fixed inset-0 z-50 bg-black/70 backdrop-blur-xs transition-opacity duration-150 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
      />

      <!-- Dialog Content Modal -->
      <AlertDialogContent
        class="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-xl border border-zinc-800 bg-[#121214] p-6 shadow-2xl duration-150 focus:outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 select-none"
      >
        <div class="flex items-start gap-4">
          <!-- Danger Icon Badge -->
          <div
            v-if="variant === 'danger'"
            class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-red-500/10 text-red-400 border border-red-500/20"
          >
            <AlertTriangle class="h-5 w-5" />
          </div>

          <div class="flex-1 min-w-0">
            <AlertDialogTitle class="text-sm font-semibold text-zinc-100">
              {{ title }}
            </AlertDialogTitle>
            <AlertDialogDescription class="mt-2 text-xs text-zinc-400 leading-relaxed break-words">
              {{ description }}
            </AlertDialogDescription>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="mt-6 flex items-center justify-end gap-2.5">
          <AlertDialogCancel
            as-child
          >
            <button
              type="button"
              @click="handleCancel"
              class="px-3.5 py-1.5 rounded-lg text-xs font-medium text-zinc-300 bg-zinc-800/80 hover:bg-zinc-800 hover:text-white border border-zinc-700/60 transition-colors focus:outline-none"
            >
              {{ cancelText }}
            </button>
          </AlertDialogCancel>

          <AlertDialogAction
            as-child
          >
            <button
              type="button"
              @click="handleConfirm"
              :class="
                variant === 'danger'
                  ? 'bg-red-600 hover:bg-red-500 text-white shadow-xs'
                  : 'bg-zinc-100 hover:bg-white text-zinc-900 shadow-xs'
              "
              class="px-3.5 py-1.5 rounded-lg text-xs font-medium transition-colors focus:outline-none"
            >
              {{ confirmText }}
            </button>
          </AlertDialogAction>
        </div>
      </AlertDialogContent>
    </AlertDialogPortal>
  </AlertDialogRoot>
</template>
