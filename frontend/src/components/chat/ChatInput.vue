<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import {
  Send,
  Square,
  Sparkles,
  ChevronDown,
  Key,
  Check,
  Plus,
  Folder,
  FolderPlus,
  X,
  Globe
} from 'lucide-vue-next'
import type { ModelItem, APIKeyItem, WorkspaceSummary } from '../../types'

const props = defineProps<{
  isGenerating: boolean
  modelName: string
  models: ModelItem[]
  currentModelId?: string
  presetName?: string
  initialContent?: string
  apiKeys?: APIKeyItem[]
  activeApiKey?: APIKeyItem | null
  workspace?: WorkspaceSummary | null
  hasTavilySearch?: boolean
}>()

const emit = defineEmits<{
  (e: 'send', content: string, searchMode: 'auto' | 'on' | 'off'): void
  (e: 'stop'): void
  (e: 'change-model', modelId: string): void
  (e: 'switch-api-key', keyId: string): void
  (e: 'open-workspace'): void
  (e: 'add-workspace'): void
  (e: 'detach-workspace'): void
  (e: 'open-settings', tab?: 'api' | 'general' | 'generation' | 'storage' | 'about'): void
}>()

const text = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const isModelPickerOpen = ref(false)
const isKeyPickerOpen = ref(false)

const charCount = computed(() => text.value.length)
const estTokens = computed(() => Math.ceil(text.value.length / 4))

watch(() => props.initialContent, (newVal) => {
  if (newVal !== undefined) {
    text.value = newVal
    focusInput()
  }
})

function closeAllPopups() {
  isModelPickerOpen.value = false
  isKeyPickerOpen.value = false
}

function adjustHeight() {
  if (!textareaRef.value) return
  textareaRef.value.style.height = 'auto'
  const maxHeight = 180
  const newHeight = Math.min(textareaRef.value.scrollHeight, maxHeight)
  textareaRef.value.style.height = `${newHeight}px`
}

function handleInput() {
  adjustHeight()
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    submit()
  }
}

// Search Mode: 'auto' (smart keyword trigger), 'on' (forced for next/all queries), 'off' (never)
const searchMode = ref<'auto' | 'on' | 'off'>('auto')

function cycleSearchMode() {
  if (searchMode.value === 'auto') {
    searchMode.value = 'on'
  } else if (searchMode.value === 'on') {
    searchMode.value = 'off'
  } else {
    searchMode.value = 'auto'
  }
}

function submit() {
  const content = text.value.trim()
  if (!content || props.isGenerating) return
  emit('send', content, searchMode.value)
  text.value = ''
  nextTick(() => {
    adjustHeight()
  })
}

function focusInput() {
  nextTick(() => {
    textareaRef.value?.focus()
  })
}

defineExpose({
  focusInput,
  setText: (val: string) => {
    text.value = val
    nextTick(adjustHeight)
  }
})
</script>

<template>
  <div class="p-4 bg-gradient-to-t from-[#09090b] via-[#09090b] to-transparent relative">
    <!-- Click-away backdrop overlay for picker menus -->
    <div
      v-if="isModelPickerOpen || isKeyPickerOpen"
      class="fixed inset-0 z-30"
      @click="closeAllPopups"
    ></div>

    <div class="max-w-3xl mx-auto relative z-40">
      <!-- Composer Card (Flat Dark) -->
      <div
        class="relative flex flex-col rounded-xl border border-[#27272a] bg-[#18181b] shadow-sm transition-all focus-within:border-zinc-500 focus-within:ring-1 focus-within:ring-zinc-600"
      >
        <!-- Attached Folder Context Banner -->
        <div
          v-if="workspace"
          class="px-3.5 py-1.5 bg-[#121214] border-b border-[#27272a] rounded-t-xl flex items-center justify-between text-xs text-zinc-300"
        >
          <div class="flex items-center gap-2 min-w-0">
            <Folder class="w-3.5 h-3.5 text-blue-400 flex-shrink-0" />
            <span class="font-medium text-blue-300 truncate">{{ workspace.folder_name }}</span>
            <span class="text-[11px] text-zinc-500 font-mono hidden sm:inline">
              ({{ workspace.total_files }} files • ~{{ workspace.total_tokens.toLocaleString() }} tokens context)
            </span>
          </div>
          <div class="flex items-center gap-1.5">
            <button
              type="button"
              @click="emit('open-workspace')"
              class="text-[11px] text-blue-400 hover:text-blue-300 px-2 py-0.5 rounded hover:bg-[#27272a] transition-colors"
            >
              View Files
            </button>
            <button
              type="button"
              @click="emit('detach-workspace')"
              class="text-[11px] text-zinc-500 hover:text-red-400 p-1 rounded hover:bg-[#27272a] transition-colors"
              title="Detach folder context"
            >
              <X class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- Textarea -->
        <textarea
          ref="textareaRef"
          v-model="text"
          @input="handleInput"
          @keydown="handleKeyDown"
          rows="1"
          placeholder="Message IFM Chat... (Shift + Enter for new line)"
          class="w-full resize-none bg-transparent px-4 py-3 text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none max-h-48 leading-relaxed"
        ></textarea>

        <!-- Bottom Controls & Quick Switchers -->
        <div class="flex items-center justify-between px-3 py-2 border-t border-[#27272a] bg-[#121214] rounded-b-xl flex-wrap gap-2">
          <!-- Left: Quick Model Switch Pill & Token Switch Pill & Folder Context Pill -->
          <div class="flex items-center gap-2 flex-wrap">
            <!-- Model Picker Pill -->
            <div class="relative">
              <button
                type="button"
                @click="isModelPickerOpen = !isModelPickerOpen; isKeyPickerOpen = false"
                class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-[#27272a] hover:bg-zinc-700/60 text-zinc-300 text-xs transition-colors"
                title="Click to switch model"
              >
                <Sparkles class="w-3 h-3 text-zinc-400" />
                <span class="font-medium text-[11px] max-w-[120px] truncate">{{ modelName }}</span>
                <ChevronDown class="w-2.5 h-2.5 text-zinc-500" />
              </button>

              <!-- Quick Model Switch Menu (pops upward) -->
              <div
                v-if="isModelPickerOpen"
                class="absolute bottom-full left-0 mb-1.5 w-64 rounded-lg border border-[#27272a] bg-[#18181b] p-1 shadow-2xl z-50 text-xs"
              >
                <div class="px-2 py-1 text-[10px] text-zinc-500 uppercase tracking-wider font-semibold border-b border-[#27272a]">
                  Quick Switch Model
                </div>
                <div class="max-h-56 overflow-y-auto py-1">
                  <button
                    v-for="m in models"
                    :key="m.id"
                    @click="emit('change-model', m.model_identifier); isModelPickerOpen = false"
                    class="w-full text-left px-2.5 py-1.5 rounded hover:bg-[#27272a] flex items-center justify-between transition-colors"
                    :class="currentModelId === m.model_identifier ? 'bg-[#27272a] text-white font-medium' : 'text-zinc-400'"
                  >
                    <span class="truncate">{{ m.name }}</span>
                    <span v-if="m.reasoning_effort" class="text-[9px] text-zinc-500 font-mono ml-1">reasoning</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- API Key Picker Pill (Quick Switch on-the-fly) -->
            <div v-if="apiKeys && apiKeys.length > 0" class="relative">
              <button
                type="button"
                @click="isKeyPickerOpen = !isKeyPickerOpen; isModelPickerOpen = false"
                class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-[#27272a] hover:bg-zinc-700/60 text-zinc-300 text-xs transition-colors"
                :title="activeApiKey ? `Current API Token: ${activeApiKey.name} (${activeApiKey.key_mask}). Click to switch.` : 'Switch API Key'"
              >
                <Key class="w-3 h-3 text-amber-400" />
                <span class="font-medium text-[11px] max-w-[90px] truncate">
                  {{ activeApiKey ? activeApiKey.name : 'Key' }}
                </span>
                <ChevronDown class="w-2.5 h-2.5 text-zinc-500" />
              </button>

              <!-- Quick Key Switch Menu (pops upward) -->
              <div
                v-if="isKeyPickerOpen"
                class="absolute bottom-full left-0 mb-1.5 w-68 rounded-lg border border-[#27272a] bg-[#18181b] p-1 shadow-2xl z-50 text-xs"
              >
                <div class="px-2 py-1 text-[10px] text-zinc-500 uppercase tracking-wider font-semibold border-b border-[#27272a] flex items-center justify-between">
                  <span>Switch API Key</span>
                  <span>{{ apiKeys.length }} keys</span>
                </div>
                <div class="max-h-52 overflow-y-auto py-1">
                  <button
                    v-for="k in apiKeys"
                    :key="k.id"
                    @click="emit('switch-api-key', k.id); isKeyPickerOpen = false"
                    class="w-full text-left px-2.5 py-1.5 rounded hover:bg-[#27272a] flex items-center justify-between transition-colors"
                    :class="activeApiKey?.id === k.id ? 'bg-[#27272a] text-white font-medium' : 'text-zinc-300'"
                  >
                    <div class="flex flex-col min-w-0 pr-2">
                      <span class="truncate">{{ k.name }}</span>
                      <span class="text-[10px] text-zinc-500 font-mono">{{ k.key_mask }}</span>
                    </div>
                    <Check v-if="activeApiKey?.id === k.id" class="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" />
                  </button>
                </div>
                <div class="border-t border-[#27272a] pt-1 mt-1">
                  <button
                    type="button"
                    @click="emit('open-settings', 'api'); isKeyPickerOpen = false"
                    class="w-full text-left px-2.5 py-1.5 rounded hover:bg-[#27272a] text-zinc-400 hover:text-white flex items-center gap-1.5 text-[11px]"
                  >
                    <Plus class="w-3 h-3" />
                    <span>Manage / Add API Keys...</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- Quick Add Folder Context Button (when none attached) -->
            <button
              v-if="!workspace"
              type="button"
              @click="emit('add-workspace')"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-[#27272a] hover:bg-zinc-700/60 text-zinc-400 hover:text-zinc-200 text-xs transition-colors"
              title="Attach a local folder as context for this chat"
            >
              <FolderPlus class="w-3 h-3 text-zinc-400" />
              <span class="font-medium text-[11px] hidden sm:inline">Folder Context</span>
            </button>

            <!-- Preset badge if active -->
            <span
              v-if="presetName"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-[#27272a]/60 text-zinc-400 text-[11px]"
            >
              {{ presetName }}
            </span>

            <!-- Web Search Mode Toggle Pill (Auto / Force ON / OFF) -->
            <button
              v-if="hasTavilySearch"
              type="button"
              @click="cycleSearchMode"
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-[11px] font-medium transition-colors border select-none cursor-pointer"
              :class="{
                'bg-emerald-950/40 border-emerald-800/40 text-emerald-400 hover:bg-emerald-900/50': searchMode === 'auto',
                'bg-emerald-600 text-white border-emerald-500 hover:bg-emerald-500 shadow-xs': searchMode === 'on',
                'bg-zinc-800/50 border-zinc-700/40 text-zinc-500 hover:text-zinc-400': searchMode === 'off'
              }"
              :title="`Web Search Mode: ${searchMode.toUpperCase()} (Click to toggle: Auto -> Force ON -> OFF)`"
            >
              <Globe class="w-3 h-3" :class="searchMode === 'off' ? 'opacity-40' : 'text-current'" />
              <span>Web:</span>
              <span class="font-mono uppercase text-[10px]">{{ searchMode }}</span>
            </button>
          </div>

          <!-- Right: Token & Char Counter & Action Buttons -->
          <div class="flex items-center gap-3">
            <span v-if="charCount > 0" class="text-[10px] text-zinc-500 font-mono">
              ~{{ estTokens }} tokens ({{ charCount }} chars)
            </span>

            <!-- Stop Generation Button -->
            <button
              v-if="isGenerating"
              type="button"
              @click="emit('stop')"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-red-600/90 hover:bg-red-600 text-white text-xs font-medium shadow-sm transition-all animate-pulse"
              title="Stop Generation"
            >
              <Square class="w-3 h-3 fill-current" />
              <span>Stop</span>
            </button>

            <!-- Send Button -->
            <button
              v-else
              type="button"
              @click="submit"
              :disabled="!text.trim()"
              class="p-1.5 rounded-lg bg-zinc-100 hover:bg-white text-zinc-900 disabled:opacity-30 disabled:hover:bg-zinc-100 transition-all shadow-xs"
              title="Send message (Enter)"
            >
              <Send class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>

      <div class="mt-1.5 text-center text-[11px] text-zinc-500">
        IFM Chat can make mistakes. Verify important information.
      </div>
    </div>
  </div>
</template>
