<script setup lang="ts">
import { ref } from 'vue'
import {
  Sparkles,
  Layers,
  Settings,
  Trash2,
  ChevronDown,
  Download,
  PanelLeftOpen,
  Check,
  Key,
  Plus,
  Folder,
  FolderPlus
} from 'lucide-vue-next'
import type { ModelItem, PresetItem, Conversation, APIKeyItem, WorkspaceSummary } from '../../types'

const props = defineProps<{
  conversation: Conversation | null
  models: ModelItem[]
  presets: PresetItem[]
  apiKeys: APIKeyItem[]
  activeApiKey: APIKeyItem | null
  workspace: WorkspaceSummary | null
  isSidebarCollapsed: boolean
}>()

const emit = defineEmits<{
  (e: 'change-model', modelId: string): void
  (e: 'change-preset', presetId: string): void
  (e: 'switch-api-key', keyId: string): void
  (e: 'open-workspace'): void
  (e: 'add-workspace'): void
  (e: 'clear-chat'): void
  (e: 'export-markdown'): void
  (e: 'open-settings', tab?: 'api' | 'general' | 'generation' | 'storage' | 'about'): void
  (e: 'toggle-sidebar'): void
}>()

const isModelDropdownOpen = ref(false)
const isPresetDropdownOpen = ref(false)
const isKeyDropdownOpen = ref(false)

function closeAllDropdowns() {
  isModelDropdownOpen.value = false
  isPresetDropdownOpen.value = false
  isKeyDropdownOpen.value = false
}

function selectModel(id: string) {
  emit('change-model', id)
  isModelDropdownOpen.value = false
}

function selectPreset(id: string) {
  emit('change-preset', id)
  isPresetDropdownOpen.value = false
}

function selectKey(id: string) {
  emit('switch-api-key', id)
  isKeyDropdownOpen.value = false
}

const currentModelName = () => {
  const m = props.models.find(mod => mod.model_identifier === props.conversation?.model_id)
  return m ? m.name : (props.conversation?.model_id || 'Select Model')
}

const currentPresetName = () => {
  if (!props.conversation?.preset_id) return 'No Preset'
  const p = props.presets.find(prs => prs.id === props.conversation?.preset_id)
  return p ? p.name : 'Custom Preset'
}
</script>

<template>
  <header class="h-13 border-b border-[#27272a] bg-[#09090b]/80 backdrop-blur-sm px-4 flex items-center justify-between z-20 select-none relative">
    <!-- Click-away backdrop overlay for dropdowns -->
    <div
      v-if="isModelDropdownOpen || isPresetDropdownOpen || isKeyDropdownOpen"
      class="fixed inset-0 z-30"
      @click="closeAllDropdowns"
    ></div>

    <!-- Left: Sidebar toggle, Model, Preset, API Key & Workspace Switchers -->
    <div class="flex items-center gap-2 relative z-40 flex-wrap">
      <!-- Sidebar Expand Toggle (when collapsed) -->
      <button
        v-if="isSidebarCollapsed"
        type="button"
        @click="emit('toggle-sidebar')"
        class="p-1.5 text-zinc-400 hover:text-zinc-100 hover:bg-[#18181b] rounded-md transition-colors"
        title="Open Sidebar (Ctrl + B)"
      >
        <PanelLeftOpen class="w-4 h-4" />
      </button>

      <!-- Instant Model Switcher Dropdown (Flat Dark Shadcn) -->
      <div class="relative">
        <button
          type="button"
          @click="isModelDropdownOpen = !isModelDropdownOpen; isPresetDropdownOpen = false; isKeyDropdownOpen = false"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-[#27272a] bg-[#18181b] hover:bg-[#27272a] text-xs font-medium text-zinc-200 transition-colors shadow-xs"
          title="Switch Active Model"
        >
          <Sparkles class="w-3.5 h-3.5 text-zinc-400" />
          <span class="max-w-[130px] truncate">{{ currentModelName() }}</span>
          <ChevronDown class="w-3 h-3 text-zinc-500" />
        </button>

        <!-- Model Dropdown Popover -->
        <div
          v-if="isModelDropdownOpen"
          class="absolute left-0 mt-1.5 w-72 rounded-lg border border-[#27272a] bg-[#18181b] p-1 shadow-2xl z-50 text-xs"
        >
          <div class="px-2.5 py-1.5 text-[10px] font-semibold text-zinc-500 uppercase tracking-wider border-b border-[#27272a]">
            Switch Active Model
          </div>
          <div class="max-h-64 overflow-y-auto py-1">
            <button
              v-for="m in models"
              :key="m.id"
              @click="selectModel(m.model_identifier)"
              class="w-full text-left px-2.5 py-2 rounded-md hover:bg-[#27272a] flex flex-col gap-0.5 transition-colors"
              :class="conversation?.model_id === m.model_identifier ? 'bg-[#27272a] text-white' : 'text-zinc-300'"
            >
              <div class="flex items-center justify-between">
                <span class="font-medium">{{ m.name }}</span>
                <Check v-if="conversation?.model_id === m.model_identifier" class="w-3.5 h-3.5 text-zinc-100" />
                <span
                  v-else-if="m.reasoning_effort"
                  class="text-[9px] px-1 py-0.2 rounded bg-zinc-800 text-zinc-400 font-mono"
                >
                  reasoning
                </span>
              </div>
              <span class="text-[11px] text-zinc-500 truncate">{{ m.description || m.model_identifier }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Quick API Token Switcher Dropdown -->
      <div class="relative">
        <button
          type="button"
          @click="isKeyDropdownOpen = !isKeyDropdownOpen; isModelDropdownOpen = false; isPresetDropdownOpen = false"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-[#27272a] bg-[#18181b] hover:bg-[#27272a] text-xs font-medium text-zinc-200 transition-colors shadow-xs"
          :title="activeApiKey ? `Active Token: ${activeApiKey.name} (${activeApiKey.key_mask})` : 'No API Key configured'"
        >
          <Key class="w-3.5 h-3.5 text-amber-400" />
          <span class="max-w-[110px] truncate">{{ activeApiKey ? activeApiKey.name : 'No Key' }}</span>
          <span v-if="activeApiKey" class="text-[10px] text-zinc-500 font-mono hidden md:inline">{{ activeApiKey.key_mask }}</span>
          <ChevronDown class="w-3 h-3 text-zinc-500" />
        </button>

        <!-- API Key Dropdown Popover -->
        <div
          v-if="isKeyDropdownOpen"
          class="absolute left-0 mt-1.5 w-76 rounded-lg border border-[#27272a] bg-[#18181b] p-1 shadow-2xl z-50 text-xs"
        >
          <div class="px-2.5 py-1.5 text-[10px] font-semibold text-zinc-500 uppercase tracking-wider border-b border-[#27272a] flex items-center justify-between">
            <span>Switch API Token</span>
            <span class="text-[10px] text-zinc-400 font-mono">{{ apiKeys.length }} keys</span>
          </div>

          <div v-if="apiKeys.length === 0" class="p-3 text-center text-zinc-500">
            <p class="text-xs mb-2">No API keys configured yet.</p>
            <button
              type="button"
              @click="emit('open-settings', 'api'); isKeyDropdownOpen = false"
              class="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-[#27272a] hover:bg-zinc-700 text-zinc-200 text-xs font-medium"
            >
              <Plus class="w-3.5 h-3.5" />
              <span>Add Key</span>
            </button>
          </div>

          <div v-else class="max-h-64 overflow-y-auto py-1 space-y-0.5">
            <button
              v-for="k in apiKeys"
              :key="k.id"
              @click="selectKey(k.id)"
              class="w-full text-left px-2.5 py-2 rounded-md hover:bg-[#27272a] flex items-center justify-between transition-colors"
              :class="activeApiKey?.id === k.id ? 'bg-[#27272a] text-white' : 'text-zinc-300'"
            >
              <div class="flex flex-col min-w-0 pr-2">
                <div class="flex items-center gap-1.5">
                  <span class="font-medium truncate">{{ k.name }}</span>
                  <span class="text-[10px] text-zinc-500 font-mono">{{ k.key_mask }}</span>
                </div>
                <span class="text-[10px] text-zinc-500 mt-0.5">
                  {{ (k.total_prompt_tokens + k.total_completion_tokens).toLocaleString() }} tokens • {{ k.total_requests }} reqs
                </span>
              </div>
              <Check v-if="activeApiKey?.id === k.id" class="w-4 h-4 text-emerald-400 flex-shrink-0" />
            </button>
          </div>

          <div class="border-t border-[#27272a] pt-1 mt-1">
            <button
              type="button"
              @click="emit('open-settings', 'api'); isKeyDropdownOpen = false"
              class="w-full text-left px-2.5 py-1.5 rounded hover:bg-[#27272a] text-zinc-400 hover:text-white flex items-center gap-1.5 text-[11px] font-medium transition-colors"
            >
              <Plus class="w-3.5 h-3.5 text-zinc-400" />
              <span>Manage / Add API Keys...</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Workspace Folder Context Pill / Button -->
      <div>
        <button
          v-if="workspace"
          type="button"
          @click="emit('open-workspace')"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-blue-900/60 bg-blue-950/40 hover:bg-blue-900/50 text-xs font-medium text-blue-300 transition-colors shadow-xs"
          :title="`Workspace Context: ${workspace.folder_path} (${workspace.total_files} files, ~${workspace.total_tokens.toLocaleString()} tokens)`"
        >
          <Folder class="w-3.5 h-3.5 text-blue-400" />
          <span class="max-w-[120px] truncate">{{ workspace.folder_name }}</span>
          <span class="text-[10px] text-blue-400/80 font-mono hidden md:inline">({{ workspace.total_files }})</span>
        </button>

        <button
          v-else
          type="button"
          @click="emit('add-workspace')"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-[#27272a] bg-[#18181b] hover:bg-[#27272a] text-xs font-medium text-zinc-400 hover:text-zinc-200 transition-colors shadow-xs"
          title="Attach a local folder as context for this chat"
        >
          <FolderPlus class="w-3.5 h-3.5 text-zinc-400" />
          <span class="hidden sm:inline">Add Folder</span>
        </button>
      </div>

      <!-- Preset Switcher Dropdown -->
      <div class="relative">
        <button
          type="button"
          @click="isPresetDropdownOpen = !isPresetDropdownOpen; isModelDropdownOpen = false; isKeyDropdownOpen = false"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border border-[#27272a] bg-[#18181b] hover:bg-[#27272a] text-xs font-medium text-zinc-400 hover:text-zinc-200 transition-colors shadow-xs"
        >
          <Layers class="w-3.5 h-3.5 text-zinc-500" />
          <span class="max-w-[110px] truncate">{{ currentPresetName() }}</span>
          <ChevronDown class="w-3 h-3 text-zinc-500" />
        </button>

        <div
          v-if="isPresetDropdownOpen"
          class="absolute left-0 mt-1.5 w-64 rounded-lg border border-[#27272a] bg-[#18181b] p-1 shadow-2xl z-50 text-xs"
        >
          <div class="px-2.5 py-1.5 text-[10px] font-semibold text-zinc-500 uppercase tracking-wider border-b border-[#27272a]">
            Select Persona Preset
          </div>
          <div class="max-h-60 overflow-y-auto py-1">
            <button
              @click="selectPreset('')"
              class="w-full text-left px-2.5 py-1.5 rounded-md hover:bg-[#27272a] text-zinc-400 hover:text-zinc-200"
            >
              None (Default Assistant)
            </button>
            <button
              v-for="p in presets"
              :key="p.id"
              @click="selectPreset(p.id)"
              class="w-full text-left px-2.5 py-2 rounded-md hover:bg-[#27272a] flex flex-col gap-0.5 transition-colors"
              :class="conversation?.preset_id === p.id ? 'bg-[#27272a] text-white' : 'text-zinc-300'"
            >
              <span class="font-medium">{{ p.name }}</span>
              <span class="text-[11px] text-zinc-500 truncate">{{ p.description }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Right Actions: Export Markdown, Clear, Settings -->
    <div class="flex items-center gap-1.5 relative z-40">
      <!-- Export Markdown Button -->
      <button
        v-if="conversation"
        type="button"
        @click="emit('export-markdown')"
        class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs text-zinc-400 hover:text-zinc-100 hover:bg-[#18181b] transition-colors"
        title="Export this chat as Markdown (.md)"
      >
        <Download class="w-3.5 h-3.5" />
        <span class="hidden sm:inline">Export .md</span>
      </button>

      <!-- Clear Chat Button -->
      <button
        v-if="conversation"
        type="button"
        @click="emit('clear-chat')"
        class="p-1.5 rounded-md text-zinc-400 hover:text-red-400 hover:bg-[#18181b] transition-colors"
        title="Clear conversation messages"
      >
        <Trash2 class="w-3.5 h-3.5" />
      </button>

      <!-- Settings Button -->
      <button
        type="button"
        @click="emit('open-settings')"
        class="p-1.5 rounded-md text-zinc-400 hover:text-zinc-100 hover:bg-[#18181b] transition-colors"
        title="Settings"
      >
        <Settings class="w-4 h-4" />
      </button>
    </div>
  </header>
</template>
