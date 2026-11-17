<script setup lang="ts">
import { ref, watch, onUnmounted, nextTick } from 'vue'
import { ChevronDown, ChevronRight, Brain, Sparkles } from 'lucide-vue-next'

const props = defineProps<{
  reasoning: string
  isStreaming?: boolean
}>()

const isExpanded = ref(true)
const elapsedSeconds = ref(0)
const contentRef = ref<HTMLDivElement | null>(null)
let timer: any = null

function startTimer() {
  if (timer) return
  elapsedSeconds.value = 0
  timer = setInterval(() => {
    elapsedSeconds.value++
  }, 1000)
}

function stopTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

watch(
  () => props.isStreaming,
  (streaming) => {
    if (streaming) {
      isExpanded.value = true
      startTimer()
    } else {
      stopTimer()
    }
  },
  { immediate: true }
)

// Auto-scroll inside thinking box as thoughts stream in
watch(
  () => props.reasoning,
  () => {
    if (props.isStreaming && isExpanded.value) {
      nextTick(() => {
        if (contentRef.value) {
          contentRef.value.scrollTop = contentRef.value.scrollHeight
        }
      })
    }
  }
)

onUnmounted(() => {
  stopTimer()
})
</script>

<template>
  <div class="my-2 rounded-lg border border-[#27272a] bg-[#121214] overflow-hidden text-xs transition-all shadow-xs">
    <!-- Header / Toggle Button -->
    <button
      type="button"
      @click="isExpanded = !isExpanded"
      class="w-full flex items-center justify-between px-3.5 py-2 bg-[#18181b] hover:bg-[#27272a] transition-colors text-left text-xs font-medium select-none"
      :class="isStreaming ? 'text-amber-300' : 'text-zinc-400'"
    >
      <div class="flex items-center gap-2">
        <div v-if="isStreaming" class="flex items-center gap-2">
          <Brain class="w-3.5 h-3.5 text-amber-400 animate-pulse flex-shrink-0" />
          <span class="font-medium text-amber-300">
            Thinking{{ elapsedSeconds > 0 ? ` (${elapsedSeconds}s)` : '...' }}
          </span>
        </div>
        <div v-else class="flex items-center gap-2">
          <Brain class="w-3.5 h-3.5 text-zinc-400 flex-shrink-0" />
          <span class="text-zinc-300">
            Thought process {{ elapsedSeconds > 0 ? `(${elapsedSeconds}s)` : '' }}
          </span>
        </div>
      </div>
      <div class="flex items-center gap-1 text-zinc-500 font-mono text-[10px]">
        <span>{{ isExpanded ? 'Hide' : 'Show' }}</span>
        <ChevronDown v-if="isExpanded" class="w-3.5 h-3.5" />
        <ChevronRight v-else class="w-3.5 h-3.5" />
      </div>
    </button>

    <!-- Collapsible Live Thought Process Content -->
    <div
      v-show="isExpanded"
      ref="contentRef"
      class="p-3.5 text-xs leading-relaxed text-zinc-300 font-mono whitespace-pre-wrap border-t border-[#27272a] max-h-80 overflow-y-auto selectable-text bg-[#09090b]/90 scroll-smooth"
    >
      <template v-if="reasoning && reasoning.trim()">
        <span>{{ reasoning }}</span>
        <!-- Blinking cursor during streaming thought generation -->
        <span
          v-if="isStreaming"
          class="inline-block w-1.5 h-3.5 bg-amber-400/90 ml-1 align-middle animate-pulse"
        ></span>
      </template>

      <!-- Placeholder before first thought token arrives -->
      <div v-else-if="isStreaming" class="flex items-center gap-2 text-zinc-500 italic py-1">
        <Sparkles class="w-3.5 h-3.5 text-amber-400 animate-spin" />
        <span>Analyzing query and forming reasoning chain...</span>
      </div>

      <span v-else class="text-zinc-500 italic">(No thought process recorded)</span>
    </div>
  </div>
</template>
