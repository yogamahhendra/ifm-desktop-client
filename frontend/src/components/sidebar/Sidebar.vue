<script setup lang="ts">
import { ref } from 'vue'
import {
  Plus,
  Search,
  MessageSquare,
  Edit2,
  Trash2,
  Settings,
  Sparkles,
  Check,
  X,
  PanelLeftClose,
  PanelLeftOpen
} from 'lucide-vue-next'
import type { Conversation } from '../../types'

const props = defineProps<{
  groupedConversations: { title: string; items: Conversation[] }[]
  activeId: string | null
  appVersion: string
  isCollapsed: boolean
  generatingIds?: string[]
}>()

const emit = defineEmits<{
  (e: 'new-chat'): void
  (e: 'select-chat', id: string): void
  (e: 'delete-chat', id: string): void
  (e: 'rename-chat', id: string, title: string): void
  (e: 'search', query: string): void
  (e: 'open-settings'): void
  (e: 'toggle-collapse'): void
}>()

const searchInput = ref('')
const editingId = ref<string | null>(null)
const editTitle = ref('')

function handleSearch() {
  emit('search', searchInput.value)
}

function startRename(conv: Conversation) {
  editingId.value = conv.id
  editTitle.value = conv.title
}

function saveRename(id: string) {
  if (editTitle.value.trim()) {
    emit('rename-chat', id, editTitle.value.trim())
  }
  editingId.value = null
}

function cancelRename() {
  editingId.value = null
}

function handleSearchIconClick() {
  if (props.isCollapsed) {
    emit('toggle-collapse')
  }
}
</script>

<template>
  <aside
    class="flex-shrink-0 h-full border-r border-[#27272a] bg-[#09090b] flex flex-col justify-between select-none transition-all duration-200 ease-in-out relative z-20"
    :class="isCollapsed ? 'w-14 items-center' : 'w-72'"
  >
    <!-- ==================== COLLAPSED ICON RAIL MODE (w-14) ==================== -->
    <template v-if="isCollapsed">
      <!-- Top Icon Actions -->
      <div class="py-3 flex flex-col items-center gap-2 w-full px-2">
        <!-- Logo -->
        <div class="w-8 h-8 rounded-lg bg-zinc-100 flex items-center justify-center text-zinc-900 shadow-sm mb-1" title="IFM Chat">
          <Sparkles class="w-4 h-4 fill-current" />
        </div>

        <!-- Expand Toggle -->
        <button
          type="button"
          @click="emit('toggle-collapse')"
          class="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-zinc-100 hover:bg-[#18181b] rounded-md transition-colors"
          title="Expand Sidebar (Ctrl + B)"
        >
          <PanelLeftOpen class="w-4 h-4" />
        </button>

        <!-- New Chat Icon Button -->
        <button
          type="button"
          @click="emit('new-chat')"
          class="w-8 h-8 flex items-center justify-center rounded-md bg-zinc-100 hover:bg-white text-zinc-900 transition-colors shadow-xs"
          title="New Chat (Ctrl + N)"
        >
          <Plus class="w-4 h-4" />
        </button>

        <!-- Search Icon (Expands sidebar) -->
        <button
          type="button"
          @click="handleSearchIconClick"
          class="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-zinc-100 hover:bg-[#18181b] rounded-md transition-colors"
          title="Search conversations"
        >
          <Search class="w-4 h-4" />
        </button>
      </div>

      <!-- Middle: Chat Bubble Icons Rail -->
      <div class="flex-1 overflow-y-auto overflow-x-hidden w-full flex flex-col items-center gap-1.5 py-2 px-2">
        <template v-for="group in groupedConversations" :key="group.title">
          <button
            v-for="c in group.items"
            :key="c.id"
            @click="emit('select-chat', c.id)"
            class="w-8 h-8 flex items-center justify-center rounded-md transition-colors relative group"
            :class="activeId === c.id ? 'bg-[#18181b] text-white border border-[#27272a]' : 'text-zinc-500 hover:text-zinc-300 hover:bg-[#18181b]/50'"
            :title="c.title || 'Conversation'"
          >
            <MessageSquare class="w-3.5 h-3.5 flex-shrink-0" />
            <!-- Floating Tooltip on Hover -->
            <span class="absolute left-full ml-2 px-2 py-1 bg-[#18181b] border border-[#27272a] text-zinc-200 text-xs rounded-md shadow-lg whitespace-nowrap opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50">
              {{ c.title || 'Untitled' }}
            </span>
          </button>
        </template>
      </div>

      <!-- Bottom: Settings Icon -->
      <div class="py-3 flex flex-col items-center w-full border-t border-[#27272a] bg-[#09090b]">
        <button
          type="button"
          @click="emit('open-settings')"
          class="w-8 h-8 flex items-center justify-center text-zinc-400 hover:text-zinc-100 hover:bg-[#18181b] rounded-md transition-colors"
          title="Settings"
        >
          <Settings class="w-4 h-4" />
        </button>
      </div>
    </template>

    <!-- ==================== EXPANDED FULL SIDEBAR MODE (w-64) ==================== -->
    <template v-else>
      <!-- Top Area: Brand, Collapse button & New Chat -->
      <div class="p-3 flex flex-col gap-2.5">
        <!-- App Brand & Collapse Toggle -->
        <div class="flex items-center justify-between px-1.5 pt-0.5">
          <div class="flex items-center gap-2">
            <div class="w-6 h-6 rounded-md bg-zinc-100 flex items-center justify-center text-zinc-900 shadow-sm">
              <Sparkles class="w-3.5 h-3.5 fill-current" />
            </div>
            <span class="font-semibold text-xs tracking-tight text-zinc-100">IFM Chat</span>
            <span class="text-[10px] font-mono px-1 py-0.2 rounded bg-zinc-800/80 text-zinc-400 border border-zinc-700/50">
              v{{ appVersion }}
            </span>
          </div>

          <button
            type="button"
            @click="emit('toggle-collapse')"
            class="p-1 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 rounded transition-colors"
            title="Collapse to icons (Ctrl + B)"
          >
            <PanelLeftClose class="w-4 h-4" />
          </button>
        </div>

        <!-- New Chat Button -->
        <button
          type="button"
          @click="emit('new-chat')"
          class="w-full flex items-center justify-between py-2 px-3 rounded-lg bg-zinc-100 hover:bg-white text-zinc-900 text-xs font-semibold shadow-sm transition-all active:scale-[0.99]"
        >
          <span class="flex items-center gap-2">
            <Plus class="w-3.5 h-3.5" />
            <span>New Chat</span>
          </span>
          <kbd class="text-[10px] bg-zinc-200/80 text-zinc-600 px-1.5 py-0.5 rounded font-mono font-normal">Ctrl+N</kbd>
        </button>

        <!-- Search Input -->
        <div class="relative">
          <Search class="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-zinc-500" />
          <input
            v-model="searchInput"
            @input="handleSearch"
            type="text"
            placeholder="Search conversations..."
            class="w-full pl-8 pr-7 py-1.5 rounded-lg bg-[#18181b] border border-[#27272a] text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-600"
          />
          <button
            v-if="searchInput"
            @click="searchInput = ''; handleSearch()"
            class="absolute right-2 top-2 text-zinc-500 hover:text-zinc-300"
          >
            <X class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      <!-- Middle: Conversations List Grouped by Date -->
      <div class="flex-1 overflow-y-auto px-2 space-y-4">
        <div v-if="groupedConversations.length === 0" class="text-center py-10 px-4 text-xs text-zinc-500">
          No conversations found.
        </div>

        <div v-for="group in groupedConversations" :key="group.title" class="space-y-0.5">
          <div class="px-2 py-1 text-[10px] font-semibold text-zinc-500 uppercase tracking-wider">
            {{ group.title }}
          </div>

          <div v-for="c in group.items" :key="c.id" class="group relative">
            <!-- Normal Item -->
            <div
              v-if="editingId !== c.id"
              @click="emit('select-chat', c.id)"
              class="flex items-center justify-between px-2.5 py-2 rounded-lg cursor-pointer text-xs transition-colors"
              :class="activeId === c.id ? 'bg-[#18181b] text-zinc-100 font-medium border border-[#27272a]' : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#18181b]/60'"
            >
              <div class="flex items-center gap-2 min-w-0 flex-1 pr-1">
                <MessageSquare class="w-3.5 h-3.5 flex-shrink-0" :class="activeId === c.id ? 'text-zinc-200' : 'text-zinc-500'" />
                <span class="truncate">{{ c.title || 'Untitled' }}</span>
                <span
                  v-if="generatingIds?.includes(c.id)"
                  class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse flex-shrink-0"
                  title="Generating in background..."
                ></span>
              </div>

              <!-- Action buttons (hover) -->
              <div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                <button
                  type="button"
                  @click.stop="startRename(c)"
                  class="p-1 text-zinc-400 hover:text-zinc-100 rounded hover:bg-zinc-800"
                  title="Rename"
                >
                  <Edit2 class="w-3 h-3" />
                </button>
                <button
                  type="button"
                  @click.stop="emit('delete-chat', c.id)"
                  class="p-1 text-zinc-400 hover:text-red-400 rounded hover:bg-zinc-800"
                  title="Delete"
                >
                  <Trash2 class="w-3 h-3" />
                </button>
              </div>
            </div>

            <!-- Rename Input -->
            <div v-else class="flex items-center gap-1 p-1 bg-[#18181b] rounded-lg border border-zinc-500">
              <input
                v-model="editTitle"
                @keydown.enter="saveRename(c.id)"
                @keydown.esc="cancelRename"
                type="text"
                class="w-full bg-transparent px-2 py-1 text-xs text-white focus:outline-none"
                autofocus
              />
              <button @click="saveRename(c.id)" class="p-1 text-zinc-200 hover:bg-zinc-800 rounded">
                <Check class="w-3 h-3" />
              </button>
              <button @click="cancelRename" class="p-1 text-zinc-400 hover:bg-zinc-800 rounded">
                <X class="w-3 h-3" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom: Settings -->
      <div class="p-3 border-t border-[#27272a] bg-[#09090b]">
        <button
          type="button"
          @click="emit('open-settings')"
          class="w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-xs text-zinc-400 hover:text-zinc-100 hover:bg-[#18181b] transition-colors"
        >
          <div class="flex items-center gap-2">
            <Settings class="w-4 h-4 text-zinc-500" />
            <span>Settings</span>
          </div>
          <span class="text-[10px] text-zinc-500 font-mono">Config</span>
        </button>
      </div>
    </template>
  </aside>
</template>
