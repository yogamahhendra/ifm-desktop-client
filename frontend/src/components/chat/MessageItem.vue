<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Copy,
  Check,
  RotateCw,
  Trash2,
  Edit2,
  Sparkles,
  Zap,
  Globe
} from 'lucide-vue-next'
import { renderMarkdown } from '../../utils/markdown'
import ReasoningBlock from './ReasoningBlock.vue'
import type { Message } from '../../types'

const props = defineProps<{
  message: Message
  isStreaming?: boolean
  streamingReasoning?: string
  streamingContent?: string
}>()

const emit = defineEmits<{
  (e: 'regenerate', id: string): void
  (e: 'delete', id: string): void
  (e: 'edit', content: string): void
}>()

const isCopied = ref(false)

const displayContent = computed(() => {
  if (props.isStreaming && props.streamingContent !== undefined) {
    return props.streamingContent
  }
  return props.message.content
})

const displayReasoning = computed(() => {
  if (props.isStreaming && props.streamingReasoning !== undefined) {
    return props.streamingReasoning
  }
  return props.message.reasoning_content || null
})

const renderedMarkdown = computed(() => {
  return renderMarkdown(displayContent.value)
})

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    isCopied.value = true
    setTimeout(() => {
      isCopied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy text:', err)
  }
}

function handleBodyClick(event: MouseEvent) {
  const target = (event.target as HTMLElement).closest('.copy-code-btn')
  if (target) {
    const rawCode = target.getAttribute('data-code')
    if (rawCode) {
      const code = decodeURIComponent(rawCode)
      navigator.clipboard.writeText(code)
      const span = target.querySelector('span')
      if (span) {
        const orig = span.textContent
        span.textContent = 'Copied!'
        setTimeout(() => {
          span.textContent = orig
        }, 2000)
      }
    }
  }
}
</script>

<template>
  <!-- USER MESSAGE (Right-aligned bubble like ChatGPT) -->
  <div v-if="message.role === 'user'" class="flex justify-end w-full group py-1.5">
    <div class="flex flex-col items-end max-w-[85%] sm:max-w-[75%]">
      <!-- Chat Bubble -->
      <div
        class="rounded-2xl rounded-tr-sm bg-[#27272a] text-zinc-100 px-4 py-2.5 shadow-xs text-sm whitespace-pre-wrap break-words leading-relaxed select-text border border-zinc-700/50"
      >
        {{ displayContent }}
      </div>

      <!-- Action toolbar under user bubble (visible on hover) -->
      <div class="flex items-center gap-1 mt-1 opacity-0 group-hover:opacity-100 transition-opacity pr-1">
        <button
          type="button"
          @click="copyText(displayContent)"
          class="p-1 text-zinc-400 hover:text-zinc-100 rounded hover:bg-[#18181b] transition-colors"
          title="Copy message"
        >
          <Check v-if="isCopied" class="w-3.5 h-3.5 text-zinc-200" />
          <Copy v-else class="w-3.5 h-3.5" />
        </button>

        <button
          type="button"
          @click="emit('edit', message.content)"
          class="p-1 text-zinc-400 hover:text-zinc-100 rounded hover:bg-[#18181b] transition-colors"
          title="Edit & Resend"
        >
          <Edit2 class="w-3.5 h-3.5" />
        </button>

        <button
          type="button"
          @click="emit('delete', message.id)"
          class="p-1 text-zinc-400 hover:text-red-400 rounded hover:bg-[#18181b] transition-colors"
          title="Delete message"
        >
          <Trash2 class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>
  </div>

  <!-- ASSISTANT MESSAGE (Option B: Inline Avatar Header, Full-Width Reading Content) -->
  <div v-else-if="message.role === 'assistant'" class="flex flex-col w-full group py-2">
    <!-- Header: Inline Avatar + Label + Tokens + Action Toolbar -->
    <div class="flex items-center justify-between mb-2 select-none">
      <div class="flex items-center gap-2">
        <div class="w-6 h-6 rounded-full bg-zinc-100 flex items-center justify-center text-zinc-900 shadow-xs flex-shrink-0">
          <Sparkles class="w-3.5 h-3.5 fill-current" />
        </div>
        <span class="text-xs font-semibold text-zinc-200">
          IFM Assistant
        </span>
        <span
          v-if="message.completion_tokens > 0"
          class="flex items-center gap-1 text-[10px] text-zinc-400 font-mono"
        >
          <Zap class="w-2.5 h-2.5 text-zinc-300" />
          {{ message.completion_tokens }} tokens
        </span>

        <!-- Web Search Provider & Credit Usage Pill -->
        <span
          v-if="message.search_source"
          class="flex items-center gap-1 px-1.5 py-0.5 rounded bg-emerald-950/50 border border-emerald-800/40 text-[10px] text-emerald-400 font-mono"
          :title="`Real-time web search grounding via ${message.search_source}`"
        >
          <Globe class="w-2.5 h-2.5 text-emerald-400" />
          {{ message.search_source }}
          <span v-if="message.search_credits" class="text-emerald-300/80">
            • {{ message.search_credits }} credit
          </span>
        </span>
      </div>

      <!-- Actions Toolbar (Visible on hover) -->
      <div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
        <button
          type="button"
          @click="copyText(displayContent)"
          class="p-1 text-zinc-400 hover:text-zinc-100 rounded hover:bg-[#27272a] transition-colors"
          title="Copy response"
        >
          <Check v-if="isCopied" class="w-3.5 h-3.5 text-zinc-200" />
          <Copy v-else class="w-3.5 h-3.5" />
        </button>

        <button
          v-if="!isStreaming"
          type="button"
          @click="emit('regenerate', message.id)"
          class="p-1 text-zinc-400 hover:text-zinc-100 rounded hover:bg-[#27272a] transition-colors"
          title="Regenerate response"
        >
          <RotateCw class="w-3.5 h-3.5" />
        </button>

        <button
          type="button"
          @click="emit('delete', message.id)"
          class="p-1 text-zinc-400 hover:text-red-400 rounded hover:bg-[#27272a] transition-colors"
          title="Delete response"
        >
          <Trash2 class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>

    <!-- Collapsible Reasoning Block (K2 Horizon reasoning thought chain) -->
    <ReasoningBlock
      v-if="displayReasoning || isStreaming"
      :reasoning="displayReasoning || ''"
      :is-streaming="isStreaming && !displayContent"
    />

    <!-- Rendered Markdown Body (Full Width, Perfectly Aligned) -->
    <div
      v-if="displayContent"
      class="text-sm text-zinc-200 leading-relaxed selectable-text prose prose-invert max-w-none pt-0.5"
      @click="handleBodyClick"
      v-html="renderedMarkdown"
    ></div>
  </div>

  <!-- SYSTEM / OTHER MESSAGE -->
  <div v-else class="flex justify-center my-2">
    <div class="px-3 py-1 rounded-md bg-[#18181b] border border-[#27272a] text-xs text-zinc-400 font-mono">
      {{ displayContent }}
    </div>
  </div>
</template>
