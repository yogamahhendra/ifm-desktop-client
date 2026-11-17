<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import Sidebar from './components/sidebar/Sidebar.vue'
import TopNavbar from './components/chat/TopNavbar.vue'
import MessageItem from './components/chat/MessageItem.vue'
import ChatInput from './components/chat/ChatInput.vue'
import SettingsModal from './components/settings/SettingsModal.vue'
import WorkspaceModal from './components/chat/WorkspaceModal.vue'
import ConfirmDialog from './components/ui/ConfirmDialog.vue'
import { useConversations } from './composables/useConversations'
import { useChat } from './composables/useChat'
import { useSettings } from './composables/useSettings'
import { useWorkspace } from './composables/useWorkspace'
import { AlertCircle, Key, Sparkles, Folder } from 'lucide-vue-next'
import { UpdateConversationPreset } from '../wailsjs/go/main/App'

const {
  config,
  apiKeys,
  activeApiKey,
  models,
  presets,
  appVersion,
  loadConfig,
  switchActiveKey
} = useSettings()

const {
  conversations,
  activeId,
  groupedConversations,
  activeConversation,
  loadConversations,
  newConversation,
  removeConversation,
  renameConversation,
  clearMessages,
  changeModel,
  exportMarkdown,
  setActiveId
} = useConversations()

const {
  messages,
  isLoadingMessages,
  isGenerating,
  streamingContent,
  streamingReasoning,
  streamingMessageId,
  streamingConversationId,
  streamingSearchSource,
  streamingSearchCredits,
  activeGeneratingIds,
  streamError,
  loadMessages,
  send,
  stop,
  regenerate,
  removeMessage
} = useChat()

const {
  currentWorkspace,
  isLoadingWorkspace,
  loadWorkspace,
  chooseAndAttachFolder,
  removeFolder,
  refresh: refreshWorkspace,
  openFolder
} = useWorkspace()

const isSidebarCollapsed = ref(false)
const isSettingsOpen = ref(false)
const isWorkspaceModalOpen = ref(false)
const settingsInitialTab = ref<'api' | 'general' | 'generation' | 'storage' | 'about'>('api')
const chatInputRef = ref<any>(null)
const messagesContainerRef = ref<HTMLDivElement | null>(null)

function openSettings(tab: 'api' | 'general' | 'generation' | 'storage' | 'about' = 'api') {
  settingsInitialTab.value = tab
  isSettingsOpen.value = true
}

function toggleSidebar() {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.ctrlKey && e.key.toLowerCase() === 'b') {
    e.preventDefault()
    toggleSidebar()
  } else if (e.ctrlKey && e.key.toLowerCase() === 'n') {
    e.preventDefault()
    handleNewChat()
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeyDown)
  await loadConfig()
  await loadConversations()
  if (conversations.value.length === 0) {
    const defaultM = config.value?.default_model || 'IFM/K2-Horizon-375B-A23B'
    await newConversation('New Chat', defaultM)
  }
  if (activeId.value) {
    await Promise.all([
      loadMessages(activeId.value),
      loadWorkspace(activeId.value)
    ])
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

// When active conversation changes, load messages and workspace context
watch(activeId, async (newVal) => {
  if (newVal) {
    await Promise.all([
      loadMessages(newVal),
      loadWorkspace(newVal)
    ])
    scrollToBottom()
  }
})

// Auto scroll on new messages or during streaming
watch([() => messages.value.length, streamingContent, streamingReasoning], () => {
  scrollToBottom()
})

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainerRef.value) {
      messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight
    }
  })
}

async function handleNewChat() {
  const defaultM = config.value?.default_model || 'IFM/K2-Horizon-375B-A23B'
  const defaultP = config.value?.default_preset || ''
  let presetSysPrompt = ''
  if (defaultP) {
    const p = presets.value.find(pr => pr.id === defaultP)
    if (p) {
      presetSysPrompt = p.system_prompt
    }
  }
  const created = await newConversation('New Chat', defaultM, defaultP, presetSysPrompt)
  if (created) {
    chatInputRef.value?.focusInput()
  }
  return created
}

async function handleSendMessage(content: string, searchMode: 'auto' | 'on' | 'off' = 'auto') {
  if (!activeId.value) {
    const created = await handleNewChat()
    if (!created) return
  }
  if (activeId.value) {
    await send(activeId.value, content, searchMode)
  }
}

async function handleRegenerate(assistantMessageId: string) {
  if (activeId.value) {
    await regenerate(activeId.value, assistantMessageId, 'auto')
  }
}

// Confirm Dialog State (Radix-vue)
const confirmState = ref<{
  open: boolean
  title: string
  description: string
  confirmText: string
  action: () => Promise<void> | void
}>({
  open: false,
  title: '',
  description: '',
  confirmText: 'Delete',
  action: () => {}
})

function triggerConfirm(params: {
  title: string
  description: string
  confirmText?: string
  action: () => Promise<void> | void
}) {
  confirmState.value = {
    open: true,
    title: params.title,
    description: params.description,
    confirmText: params.confirmText || 'Delete',
    action: params.action
  }
}

async function handleConfirmAction() {
  if (confirmState.value.action) {
    await confirmState.value.action()
  }
}

async function handleDeleteConversation(id: string) {
  const conv = conversations.value.find(c => c.id === id)
  const title = conv?.title ? `"${conv.title}"` : 'this conversation'
  triggerConfirm({
    title: 'Delete Conversation',
    description: `Are you sure you want to delete ${title}? All messages within this thread will be permanently deleted.`,
    confirmText: 'Delete',
    action: async () => {
      await removeConversation(id)
    }
  })
}

async function handleDeleteMessage(messageId: string) {
  triggerConfirm({
    title: 'Delete Message',
    description: 'Are you sure you want to delete this message? This action cannot be undone.',
    confirmText: 'Delete',
    action: async () => {
      if (activeId.value) {
        await removeMessage(messageId, activeId.value)
      }
    }
  })
}

function handleEditMessage(content: string) {
  chatInputRef.value?.setText(content)
}

async function handleClearChat() {
  if (!activeId.value) return
  triggerConfirm({
    title: 'Clear Chat Messages',
    description: 'Are you sure you want to clear all messages in this conversation? The conversation thread itself will be preserved.',
    confirmText: 'Clear All',
    action: async () => {
      if (activeId.value) {
        await clearMessages(activeId.value)
        await loadMessages(activeId.value)
      }
    }
  })
}

// Instant on-the-fly model switching during chat
async function handleModelChange(modelId: string) {
  if (activeId.value) {
    await changeModel(activeId.value, modelId)
  }
}

// Instant on-the-fly API key switching during chat
async function handleSwitchApiKey(keyId: string) {
  await switchActiveKey(keyId)
}

// Workspace Folder Context Handlers
async function handleAddWorkspace() {
  if (!activeId.value) {
    await handleNewChat()
  }
  if (activeId.value) {
    const success = await chooseAndAttachFolder(activeId.value)
    if (success) {
      await loadConversations()
    }
  }
}

async function handleDetachWorkspace() {
  if (activeId.value) {
    await removeFolder(activeId.value)
    await loadConversations()
    isWorkspaceModalOpen.value = false
  }
}

async function handleRefreshWorkspace() {
  if (activeId.value) {
    await refreshWorkspace(activeId.value)
  }
}

async function handlePresetChange(presetId: string) {
  if (activeConversation.value) {
    activeConversation.value.preset_id = presetId || null
    if (presetId) {
      const p = presets.value.find(pr => pr.id === presetId)
      if (p) {
        activeConversation.value.system_prompt = p.system_prompt
      }
    } else {
      activeConversation.value.system_prompt = ''
    }
    try {
      await UpdateConversationPreset(activeConversation.value.id, presetId || '')
    } catch (err) {
      console.error('Failed to update conversation preset:', err)
    }
  }
}

async function handleExportMarkdown() {
  if (!activeId.value) return
  const mdContent = await exportMarkdown(activeId.value)
  if (!mdContent) return

  const blob = new Blob([mdContent], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  const safeTitle = (activeConversation.value?.title || 'conversation')
    .replace(/[^a-zA-Z0-9_-]/g, '_')
    .slice(0, 30)
  a.download = `${safeTitle}-${new Date().toISOString().slice(0, 10)}.md`
  a.click()
  URL.revokeObjectURL(url)
}

const activeModelName = () => {
  const mid = activeConversation.value?.model_id
  const m = models.value.find(mod => mod.model_identifier === mid)
  return m ? m.name : (mid || 'IFM Foundation Model')
}

const activePresetName = () => {
  const pid = activeConversation.value?.preset_id
  if (!pid) return undefined
  const p = presets.value.find(prs => prs.id === pid)
  return p?.name
}
</script>

<template>
  <div class="h-screen w-screen flex bg-[#09090b] text-[#fafafa] overflow-hidden font-sans">
    <!-- Collapsible Sidebar -->
    <Sidebar
      :grouped-conversations="groupedConversations"
      :active-id="activeId"
      :app-version="appVersion"
      :is-collapsed="isSidebarCollapsed"
      :generating-ids="activeGeneratingIds"
      @new-chat="handleNewChat"
      @select-chat="setActiveId"
      @delete-chat="handleDeleteConversation"
      @rename-chat="renameConversation"
      @search="q => loadConversations(q)"
      @open-settings="openSettings('general')"
      @toggle-collapse="toggleSidebar"
    />

    <!-- Main Chat Workspace (Flat Dark) -->
    <main class="flex-1 flex flex-col h-full min-w-0 bg-[#09090b] relative">
      <!-- Top Navigation Bar -->
      <TopNavbar
        :conversation="activeConversation"
        :models="models"
        :presets="presets"
        :api-keys="apiKeys"
        :active-api-key="activeApiKey"
        :workspace="currentWorkspace"
        :is-sidebar-collapsed="isSidebarCollapsed"
        @change-model="handleModelChange"
        @change-preset="handlePresetChange"
        @switch-api-key="handleSwitchApiKey"
        @open-workspace="isWorkspaceModalOpen = true"
        @add-workspace="handleAddWorkspace"
        @clear-chat="handleClearChat"
        @export-markdown="handleExportMarkdown"
        @open-settings="openSettings"
        @toggle-sidebar="toggleSidebar"
      />

      <!-- API Key Notice Banner (Flat Dark) -->
      <div
        v-if="!config?.has_api_key"
        class="bg-[#18181b] border-b border-amber-900/50 px-4 py-2 flex items-center justify-between text-xs text-amber-300"
      >
        <div class="flex items-center gap-2">
          <AlertCircle class="w-4 h-4 flex-shrink-0 text-amber-400" />
          <span>IFM API key is not configured. Please add your API key to start chatting.</span>
        </div>
        <button
          type="button"
          @click="openSettings('api')"
          class="flex items-center gap-1 px-2.5 py-1 rounded bg-amber-950/60 hover:bg-amber-900/80 font-medium transition-colors border border-amber-800/60"
        >
          <Key class="w-3.5 h-3.5" />
          <span>Setup Key</span>
        </button>
      </div>

      <!-- Messages Stream Area (ChatGPT layout) -->
      <div
        ref="messagesContainerRef"
        class="flex-1 overflow-y-auto overflow-x-hidden relative scroll-smooth"
      >
        <!-- Empty State -->
        <div
          v-if="messages.length === 0 && (!isGenerating || streamingConversationId !== activeId)"
          class="h-full flex flex-col items-center justify-center p-8 text-center"
        >
          <div class="w-12 h-12 rounded-xl bg-zinc-100 flex items-center justify-center text-zinc-900 mb-4 shadow-sm">
            <Sparkles class="w-6 h-6 fill-current" />
          </div>
          <h2 class="text-lg font-semibold text-zinc-100 mb-1.5">Welcome to IFM Chat</h2>
          <p class="text-xs text-zinc-400 max-w-md leading-relaxed mb-6">
            Directly interact with high-capability Institute of Foundation Models with deep reasoning, local codebase context, and streaming responses.
          </p>

          <!-- Prompt Suggestions & Workspace Button -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2 max-w-lg w-full text-left">
            <button
              @click="handleSendMessage('Explain Go concurrency patterns and goroutine best practices with code examples.')"
              class="p-3 rounded-lg border border-[#27272a] bg-[#121214] hover:bg-[#18181b] text-xs text-zinc-300 hover:text-white transition-all text-left"
            >
              <div class="font-medium text-zinc-100 mb-0.5">Explain Go Concurrency</div>
              <div class="text-[11px] text-zinc-500">Goroutines, channels, and synchronization</div>
            </button>
            <button
              @click="handleAddWorkspace"
              class="p-3 rounded-lg border border-blue-900/40 bg-blue-950/20 hover:bg-blue-900/30 text-xs text-blue-300 hover:text-white transition-all text-left"
            >
              <div class="font-medium text-blue-200 mb-0.5 flex items-center gap-1.5">
                <Folder class="w-3.5 h-3.5 text-blue-400" />
                <span>Attach Folder Context</span>
              </div>
              <div class="text-[11px] text-zinc-400">Add local code or documents as chat context</div>
            </button>
          </div>
        </div>

        <!-- Render Messages (ChatGPT styled reading stream: user bubbles right, assistant left) -->
        <div v-else class="max-w-3xl mx-auto px-4 py-6 space-y-4">
          <MessageItem
            v-for="msg in messages"
            :key="msg.id"
            :message="msg"
            @regenerate="handleRegenerate"
            @delete="handleDeleteMessage"
            @edit="handleEditMessage"
          />

          <!-- Live Streaming Message (Only rendered if current chat is the one generating) -->
          <MessageItem
            v-if="isGenerating && streamingConversationId === activeId"
            :message="{
              id: streamingMessageId || 'streaming',
              conversation_id: activeId || '',
              role: 'assistant',
              content: streamingContent,
              reasoning_content: streamingReasoning || null,
              prompt_tokens: 0,
              completion_tokens: 0,
              search_source: streamingSearchSource || undefined,
              search_credits: streamingSearchCredits || undefined,
              created_at: new Date().toISOString()
            }"
            :is-streaming="true"
            :streaming-content="streamingContent"
            :streaming-reasoning="streamingReasoning"
          />
        </div>

        <!-- Error Notification Banner -->
        <div
          v-if="streamError"
          class="max-w-3xl mx-auto my-4 p-3.5 rounded-lg bg-red-950/40 border border-red-900/60 text-red-300 text-xs flex items-center justify-between"
        >
          <div class="flex items-center gap-2">
            <AlertCircle class="w-4 h-4 flex-shrink-0 text-red-400" />
            <span>{{ streamError }}</span>
          </div>
          <button
            type="button"
            @click="streamError = null"
            class="text-xs font-medium underline hover:text-white"
          >
            Dismiss
          </button>
        </div>
      </div>

      <!-- Message Composer (with on-the-fly model, API key switcher, & folder context) -->
      <ChatInput
        ref="chatInputRef"
        :is-generating="isGenerating && streamingConversationId === activeId"
        :model-name="activeModelName()"
        :models="models"
        :current-model-id="activeConversation?.model_id"
        :preset-name="activePresetName()"
        :api-keys="apiKeys"
        :active-api-key="activeApiKey"
        :workspace="currentWorkspace"
        :has-tavily-search="config?.has_tavily_key"
        @send="handleSendMessage"
        @stop="stop"
        @change-model="handleModelChange"
        @switch-api-key="handleSwitchApiKey"
        @open-workspace="isWorkspaceModalOpen = true"
        @add-workspace="handleAddWorkspace"
        @detach-workspace="handleDetachWorkspace"
        @open-settings="openSettings"
      />
    </main>

    <!-- Settings Modal -->
    <SettingsModal
      :is-open="isSettingsOpen"
      :initial-tab="settingsInitialTab"
      @close="isSettingsOpen = false"
    />

    <!-- Workspace Folder Context Details Modal -->
    <WorkspaceModal
      :is-open="isWorkspaceModalOpen"
      :workspace="currentWorkspace"
      :is-loading="isLoadingWorkspace"
      @close="isWorkspaceModalOpen = false"
      @refresh="handleRefreshWorkspace"
      @detach="handleDetachWorkspace"
      @change-folder="handleAddWorkspace"
      @open-explorer="openFolder"
    />

    <!-- Radix Vue Confirmation Dialog -->
    <ConfirmDialog
      v-model:open="confirmState.open"
      :title="confirmState.title"
      :description="confirmState.description"
      :confirm-text="confirmState.confirmText"
      variant="danger"
      @confirm="handleConfirmAction"
    />
  </div>
</template>
