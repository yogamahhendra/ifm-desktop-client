<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Folder,
  FolderOpen,
  X,
  RefreshCw,
  Trash2,
  Copy,
  Check,
  Search,
  FileCode,
  Zap,
  ExternalLink
} from 'lucide-vue-next'
import type { WorkspaceSummary } from '../../types'

const props = defineProps<{
  isOpen: boolean
  workspace: WorkspaceSummary | null
  isLoading?: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'refresh'): void
  (e: 'detach'): void
  (e: 'change-folder'): void
  (e: 'open-explorer', path: string): void
}>()

const searchQuery = ref('')
const isCopied = ref(false)

const filteredFiles = computed(() => {
  if (!props.workspace?.files) return []
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return props.workspace.files
  return props.workspace.files.filter(f =>
    f.relative_path.toLowerCase().includes(q)
  )
})

async function copyPath(path: string) {
  try {
    await navigator.clipboard.writeText(path)
    isCopied.value = true
    setTimeout(() => {
      isCopied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed copying path:', err)
  }
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  const kb = bytes / 1024
  if (kb < 1024) return kb.toFixed(1) + ' KB'
  return (kb / 1024).toFixed(1) + ' MB'
}
</script>

<template>
  <div
    v-if="isOpen && workspace"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-xs select-none"
    @click.self="emit('close')"
  >
    <div
      class="bg-[#121214] border border-[#27272a] rounded-xl shadow-2xl w-full max-w-2xl max-h-[85vh] flex flex-col overflow-hidden text-zinc-100"
    >
      <!-- Header -->
      <div class="px-5 py-3.5 border-b border-[#27272a] flex items-center justify-between bg-[#18181b]">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-blue-950/60 border border-blue-800/40 flex items-center justify-center text-blue-400">
            <Folder class="w-4 h-4" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-zinc-100 flex items-center gap-2">
              <span>{{ workspace.folder_name }}</span>
              <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-950/70 border border-emerald-800/50 text-emerald-400 font-normal">
                Context Active
              </span>
            </h3>
            <p class="text-[11px] text-zinc-400">Attached Workspace Folder Context</p>
          </div>
        </div>
        <button
          type="button"
          @click="emit('close')"
          class="p-1 rounded-md text-zinc-400 hover:text-white hover:bg-[#27272a] transition-colors"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Path Bar & Quick Actions -->
      <div class="px-5 py-2.5 bg-[#18181b]/60 border-b border-[#27272a] flex items-center justify-between gap-3 text-xs">
        <div class="flex items-center gap-2 min-w-0 flex-1 font-mono text-[11px] text-zinc-300 bg-[#09090b] px-3 py-1.5 rounded border border-[#27272a]">
          <span class="truncate">{{ workspace.folder_path }}</span>
          <button
            type="button"
            @click="copyPath(workspace.folder_path)"
            class="ml-auto text-zinc-500 hover:text-zinc-200 transition-colors flex-shrink-0"
            title="Copy Path"
          >
            <Check v-if="isCopied" class="w-3.5 h-3.5 text-emerald-400" />
            <Copy v-else class="w-3.5 h-3.5" />
          </button>
        </div>

        <button
          type="button"
          @click="emit('open-explorer', workspace.folder_path)"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded bg-[#27272a] hover:bg-zinc-700 text-zinc-200 text-xs transition-colors flex-shrink-0"
          title="Open in Windows Explorer"
        >
          <ExternalLink class="w-3.5 h-3.5" />
          <span class="hidden sm:inline">Explorer</span>
        </button>
      </div>

      <!-- Summary Metrics Bar -->
      <div class="px-5 py-3 border-b border-[#27272a] grid grid-cols-3 gap-3 bg-[#09090b]/40">
        <div class="bg-[#18181b] p-2.5 rounded-lg border border-[#27272a]">
          <div class="text-[10px] text-zinc-500 uppercase font-semibold">Total Files</div>
          <div class="text-sm font-semibold text-zinc-200 mt-0.5">{{ workspace.total_files }} files</div>
        </div>
        <div class="bg-[#18181b] p-2.5 rounded-lg border border-[#27272a]">
          <div class="text-[10px] text-zinc-500 uppercase font-semibold">Total Size</div>
          <div class="text-sm font-semibold text-zinc-200 mt-0.5">{{ formatBytes(workspace.total_size) }}</div>
        </div>
        <div class="bg-[#18181b] p-2.5 rounded-lg border border-[#27272a]">
          <div class="text-[10px] text-zinc-500 uppercase font-semibold flex items-center gap-1">
            <Zap class="w-3 h-3 text-amber-400" />
            <span>Est. Context</span>
          </div>
          <div class="text-sm font-semibold text-amber-300 mt-0.5">~{{ workspace.total_tokens.toLocaleString() }} tokens</div>
        </div>
      </div>

      <!-- File List Area -->
      <div class="flex-1 flex flex-col min-h-0 p-5">
        <div class="flex items-center justify-between mb-3 gap-2">
          <div class="text-xs font-medium text-zinc-300">
            Included Files in Context ({{ filteredFiles.length }})
          </div>
          <!-- Search files -->
          <div class="relative w-56">
            <Search class="w-3.5 h-3.5 text-zinc-500 absolute left-2.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              v-model="searchQuery"
              placeholder="Filter files..."
              class="w-full bg-[#18181b] border border-[#27272a] rounded-md pl-8 pr-2.5 py-1 text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-zinc-500"
            />
          </div>
        </div>

        <div class="flex-1 overflow-y-auto border border-[#27272a] rounded-lg bg-[#09090b] divide-y divide-[#27272a]/60">
          <div
            v-for="file in filteredFiles"
            :key="file.path"
            class="px-3 py-2 flex items-center justify-between hover:bg-[#18181b] transition-colors text-xs"
          >
            <div class="flex items-center gap-2 min-w-0 pr-3">
              <FileCode class="w-3.5 h-3.5 text-zinc-400 flex-shrink-0" />
              <span class="font-mono text-zinc-300 truncate">{{ file.relative_path }}</span>
            </div>
            <div class="flex items-center gap-3 text-[11px] text-zinc-500 flex-shrink-0 font-mono">
              <span>{{ formatBytes(file.size) }}</span>
              <span class="text-zinc-400">~{{ file.tokens }} tokens</span>
            </div>
          </div>

          <div v-if="filteredFiles.length === 0" class="p-6 text-center text-xs text-zinc-500">
            No files match your search filter.
          </div>
        </div>
      </div>

      <!-- Footer Buttons -->
      <div class="px-5 py-3 border-t border-[#27272a] bg-[#18181b] flex items-center justify-between flex-wrap gap-2">
        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="emit('refresh')"
            :disabled="isLoading"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[#27272a] hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-colors disabled:opacity-50"
            title="Re-scan directory files"
          >
            <RefreshCw class="w-3.5 h-3.5" :class="{ 'animate-spin': isLoading }" />
            <span>Re-scan</span>
          </button>

          <button
            type="button"
            @click="emit('change-folder')"
            :disabled="isLoading"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[#27272a] hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-colors"
          >
            <FolderOpen class="w-3.5 h-3.5" />
            <span>Change Folder...</span>
          </button>
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="emit('detach')"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-red-400 hover:text-red-300 hover:bg-red-950/40 transition-colors text-xs font-medium border border-red-900/40"
          >
            <Trash2 class="w-3.5 h-3.5" />
            <span>Detach Folder</span>
          </button>

          <button
            type="button"
            @click="emit('close')"
            class="px-4 py-1.5 rounded-lg bg-zinc-100 hover:bg-white text-zinc-900 text-xs font-semibold transition-colors"
          >
            Done
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
