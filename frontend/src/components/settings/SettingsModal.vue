<script setup lang="ts">
import { ref, watch } from 'vue'
import {
  X,
  Key,
  Sliders,
  Database,
  Info,
  Check,
  AlertCircle,
  ExternalLink,
  Download,
  Upload,
  Trash2,
  Eye,
  EyeOff,
  Plus,
  Zap,
  Clock,
  Globe,
  Folder,
  Edit2,
  ChevronDown,
  ChevronUp,
  Sparkles
} from 'lucide-vue-next'
import { useSettings } from '../../composables/useSettings'
import ConfirmDialog from '../ui/ConfirmDialog.vue'
import type { PresetItem } from '../../types'

const props = defineProps<{
  isOpen: boolean
  initialTab?: 'api' | 'general' | 'generation' | 'storage' | 'about'
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const {
  config,
  apiKeys,
  models,
  presets,
  appVersion,
  isTestingConnection,
  testConnectionResult,
  tavilyKeyMask,
  loadConfig,
  loadTavilyKey,
  setTavilyKey,
  removeTavilyKey,
  relocateDatabase,
  addNewKey,
  switchActiveKey,
  deleteKey,
  runConnectionTest,
  updateSettings,
  exportData,
  importData,
  createOrUpdatePreset,
  removePreset,
  resetAllData
} = useSettings()

const activeTab = ref<'api' | 'general' | 'generation' | 'storage' | 'about'>('api')

watch(() => props.isOpen, (open) => {
  if (open) {
    if (props.initialTab) {
      activeTab.value = props.initialTab
    }
    loadTavilyKey()
  }
})

// Tavily Key input
const inputTavilyKey = ref('')
const showTavilyKey = ref(false)

// New Key Form
const newKeyName = ref('')
const newKeyValue = ref('')
const showNewKey = ref(false)
const saveMessage = ref<string | null>(null)

// Form fields for general & generation
const ifmBaseUrl = ref('https://api.ifm.ai/v1')
const requestTimeout = ref('60')
const defaultModel = ref('IFM/K2-Horizon-375B-A23B')
const defaultPreset = ref('')
const streamingEnabled = ref('true')
const temperature = ref('1.0')
const topP = ref('0.95')
const maxTokens = ref('4096')
const reasoningEffort = ref('high')

// Preset Management in Settings > General
const isEditingPreset = ref(false)
const expandedPresetPromptId = ref<string | null>(null)
const presetForm = ref<{
  id?: string
  name: string
  description: string
  system_prompt: string
  temperature: number
  top_p: number
  max_tokens: number
  reasoning_effort: string
}>({
  id: undefined,
  name: '',
  description: '',
  system_prompt: '',
  temperature: 0.7,
  top_p: 0.95,
  max_tokens: 4096,
  reasoning_effort: 'high'
})

function startCreatePreset() {
  presetForm.value = {
    id: undefined,
    name: '',
    description: '',
    system_prompt: '',
    temperature: 0.7,
    top_p: 0.95,
    max_tokens: 4096,
    reasoning_effort: 'high'
  }
  isEditingPreset.value = true
}

function startEditPreset(p: PresetItem) {
  presetForm.value = {
    id: p.id,
    name: p.name,
    description: p.description || '',
    system_prompt: p.system_prompt || '',
    temperature: p.temperature !== undefined && p.temperature !== null ? p.temperature : 0.7,
    top_p: p.top_p !== undefined && p.top_p !== null ? p.top_p : 0.95,
    max_tokens: p.max_tokens !== undefined && p.max_tokens !== null ? p.max_tokens : 4096,
    reasoning_effort: p.reasoning_effort || 'high'
  }
  isEditingPreset.value = true
}

function cancelPresetEdit() {
  isEditingPreset.value = false
}

async function handleSavePreset() {
  if (!presetForm.value.name.trim()) {
    alert('Please enter a name for this preset.')
    return
  }
  if (!presetForm.value.system_prompt.trim()) {
    alert('Please enter system instructions/prompt for this preset.')
    return
  }

  try {
    await createOrUpdatePreset({
      id: presetForm.value.id,
      name: presetForm.value.name.trim(),
      description: presetForm.value.description.trim(),
      system_prompt: presetForm.value.system_prompt.trim(),
      temperature: Number(presetForm.value.temperature),
      top_p: Number(presetForm.value.top_p),
      max_tokens: Number(presetForm.value.max_tokens),
      reasoning_effort: presetForm.value.reasoning_effort
    })
    saveMessage.value = `Preset "${presetForm.value.name.trim()}" saved successfully!`
    setTimeout(() => { saveMessage.value = null }, 3500)
    isEditingPreset.value = false
  } catch (err: any) {
    alert('Failed to save preset: ' + err)
  }
}

async function handleDeletePreset(p: PresetItem) {
  triggerConfirm({
    title: 'Delete Preset',
    description: `Are you sure you want to delete preset "${p.name}"? This action cannot be undone.`,
    confirmText: 'Delete Preset',
    action: async () => {
      try {
        await removePreset(p.id)
        if (defaultPreset.value === p.id) {
          defaultPreset.value = ''
          await updateSettings({ default_preset: '' })
        }
        saveMessage.value = `Preset "${p.name}" deleted.`
        setTimeout(() => { saveMessage.value = null }, 2500)
      } catch (err: any) {
        alert('Failed deleting preset: ' + err)
      }
    }
  })
}

async function handleSetDefaultPreset(presetId: string) {
  defaultPreset.value = presetId
  await updateSettings({ default_preset: presetId })
  const found = presets.value.find(pr => pr.id === presetId)
  saveMessage.value = `Default persona preset set to "${found ? found.name : 'Standard Assistant'}"`
  setTimeout(() => { saveMessage.value = null }, 2500)
}

watch(() => config.value, (cfg) => {
  if (cfg) {
    ifmBaseUrl.value = cfg.ifm_base_url || 'https://api.ifm.ai/v1'
    requestTimeout.value = cfg.request_timeout || '60'
    defaultModel.value = cfg.default_model || 'IFM/K2-Horizon-375B-A23B'
    defaultPreset.value = cfg.default_preset || ''
    streamingEnabled.value = cfg.streaming_enabled || 'true'
    temperature.value = cfg.temperature || '1.0'
    topP.value = cfg.top_p || '0.95'
    maxTokens.value = cfg.max_tokens || '4096'
    reasoningEffort.value = cfg.reasoning_effort || 'high'
  }
}, { immediate: true })

async function handleAddKey() {
  if (!newKeyValue.value.trim()) return
  try {
    await addNewKey(newKeyName.value.trim() || 'API Key', newKeyValue.value.trim())
    newKeyName.value = ''
    newKeyValue.value = ''
    saveMessage.value = 'API key added and secured in Windows Credential Manager!'
    setTimeout(() => { saveMessage.value = null }, 3500)
  } catch (err: any) {
    alert('Failed adding API key: ' + err)
  }
}

async function handleSaveTavily() {
  if (!inputTavilyKey.value.trim()) return
  try {
    await setTavilyKey(inputTavilyKey.value.trim())
    inputTavilyKey.value = ''
    saveMessage.value = 'Tavily Search API key saved securely in Windows Credential Manager!'
    setTimeout(() => { saveMessage.value = null }, 3500)
  } catch (err: any) {
    alert('Failed saving Tavily key: ' + err)
  }
}

async function handleSwitchKey(id: string) {
  try {
    await switchActiveKey(id)
    saveMessage.value = 'Active API key updated!'
    setTimeout(() => { saveMessage.value = null }, 2500)
  } catch (err: any) {
    alert('Failed switching key: ' + err)
  }
}

// Confirm Dialog state (Radix-vue)
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

async function handleDeleteKey(id: string, name: string) {
  triggerConfirm({
    title: 'Delete API Key',
    description: `Are you sure you want to remove the API key "${name}" from Windows Credential Manager?`,
    confirmText: 'Delete Key',
    action: async () => {
      try {
        await deleteKey(id)
        saveMessage.value = 'API key removed.'
        setTimeout(() => { saveMessage.value = null }, 2500)
      } catch (err: any) {
        alert('Failed deleting key: ' + err)
      }
    }
  })
}

async function handleSaveAllSettings() {
  await updateSettings({
    ifm_base_url: ifmBaseUrl.value,
    request_timeout: requestTimeout.value,
    default_model: defaultModel.value,
    default_preset: defaultPreset.value,
    streaming_enabled: streamingEnabled.value,
    temperature: temperature.value,
    top_p: topP.value,
    max_tokens: maxTokens.value,
    reasoning_effort: reasoningEffort.value
  })
  saveMessage.value = 'Settings saved successfully.'
  setTimeout(() => { saveMessage.value = null }, 3000)
}

async function handleExport() {
  try {
    const jsonStr = await exportData()
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `ifm-chat-export-${new Date().toISOString().slice(0, 10)}.json`
    a.click()
    URL.revokeObjectURL(url)
  } catch (err) {
    alert('Failed to export data: ' + err)
  }
}

const isRelocating = ref(false)

async function handleRelocateDatabase() {
  triggerConfirm({
    title: 'Move SQLite Database Location',
    description: 'You will be prompted to select a new folder. The application will safely flush WAL journals and move ifm-chat.db to the new location.',
    confirmText: 'Select Folder & Move',
    action: async () => {
      try {
        isRelocating.value = true
        const newPath = await relocateDatabase()
        if (newPath) {
          saveMessage.value = `Database successfully moved to: ${newPath}`
          setTimeout(() => { saveMessage.value = null }, 4000)
        }
      } catch (err: any) {
        alert('Failed moving database: ' + err)
      } finally {
        isRelocating.value = false
      }
    }
  })
}

function handleImportClick() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'application/json'
  input.onchange = async (e: any) => {
    const file = e.target?.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = async (re) => {
        try {
          await importData(re.target?.result as string)
          alert('Conversations imported successfully!')
          window.location.reload()
        } catch (err) {
          alert('Failed to import: ' + err)
        }
      }
      reader.readAsText(file)
    }
  }
  input.click()
}

async function handleClearAll() {
  triggerConfirm({
    title: 'Clear Application Storage',
    description: 'Warning: This will permanently delete ALL conversations, messages, and settings from local SQLite storage. This action cannot be reversed.',
    confirmText: 'Reset & Reload',
    action: async () => {
      await resetAllData()
      window.location.reload()
    }
  })
}
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4 select-none">
    <div class="w-full max-w-2xl bg-[#09090b] border border-[#27272a] rounded-xl shadow-2xl overflow-hidden flex flex-col max-h-[88vh]">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-[#27272a] bg-[#18181b]/50">
        <h2 class="text-sm font-semibold text-zinc-100">Application Settings</h2>
        <button @click="emit('close')" class="p-1 text-zinc-400 hover:text-zinc-100 rounded-md hover:bg-zinc-800">
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Tabs Navigation -->
      <div class="flex border-b border-[#27272a] bg-[#121214] px-6 gap-2 text-xs font-medium">
        <button
          @click="activeTab = 'api'"
          class="py-2.5 px-3 border-b-2 transition-colors flex items-center gap-1.5"
          :class="activeTab === 'api' ? 'border-zinc-100 text-zinc-100' : 'border-transparent text-zinc-500 hover:text-zinc-300'"
        >
          <Key class="w-3.5 h-3.5" />
          <span>API Keys & Config</span>
        </button>
        <button
          @click="activeTab = 'general'"
          class="py-2.5 px-3 border-b-2 transition-colors flex items-center gap-1.5"
          :class="activeTab === 'general' ? 'border-zinc-100 text-zinc-100' : 'border-transparent text-zinc-500 hover:text-zinc-300'"
        >
          <Sliders class="w-3.5 h-3.5" />
          <span>General</span>
        </button>
        <button
          @click="activeTab = 'generation'"
          class="py-2.5 px-3 border-b-2 transition-colors flex items-center gap-1.5"
          :class="activeTab === 'generation' ? 'border-zinc-100 text-zinc-100' : 'border-transparent text-zinc-500 hover:text-zinc-300'"
        >
          <Sliders class="w-3.5 h-3.5" />
          <span>Generation</span>
        </button>
        <button
          @click="activeTab = 'storage'"
          class="py-2.5 px-3 border-b-2 transition-colors flex items-center gap-1.5"
          :class="activeTab === 'storage' ? 'border-zinc-100 text-zinc-100' : 'border-transparent text-zinc-500 hover:text-zinc-300'"
        >
          <Database class="w-3.5 h-3.5" />
          <span>Storage</span>
        </button>
        <button
          @click="activeTab = 'about'"
          class="py-2.5 px-3 border-b-2 transition-colors flex items-center gap-1.5"
          :class="activeTab === 'about' ? 'border-zinc-100 text-zinc-100' : 'border-transparent text-zinc-500 hover:text-zinc-300'"
        >
          <Info class="w-3.5 h-3.5" />
          <span>About</span>
        </button>
      </div>

      <!-- Tab Content -->
      <div class="flex-1 overflow-y-auto p-6 space-y-5 text-xs text-zinc-300">
        <!-- Notification Message -->
        <div v-if="saveMessage" class="p-3 bg-zinc-900 border border-zinc-700 rounded-lg text-zinc-200 flex items-center gap-2">
          <Check class="w-4 h-4 flex-shrink-0 text-zinc-100" />
          <span>{{ saveMessage }}</span>
        </div>

        <!-- TAB: API Configuration & Multi-Keys -->
        <div v-if="activeTab === 'api'" class="space-y-5">
          <!-- Endpoint URL -->
          <div>
            <label class="block font-medium text-zinc-200 mb-1">IFM API Base URL</label>
            <input
              v-model="ifmBaseUrl"
              type="text"
              class="w-full px-3 py-2 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 font-mono text-xs focus:outline-none focus:border-zinc-500"
            />
            <p class="mt-1 text-[11px] text-zinc-500">Default: https://api.ifm.ai/v1</p>
          </div>

          <!-- Multi-API Keys Section -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="block font-medium text-zinc-200">Configured IFM API Keys</label>
              <a
                href="https://platform.ifm.ai/api-keys"
                target="_blank"
                class="inline-flex items-center gap-1 text-[11px] text-zinc-400 hover:text-white underline"
              >
                <span>Check Official Platform Quota / Balance</span>
                <ExternalLink class="w-3 h-3" />
              </a>
            </div>

            <!-- Keys List -->
            <div v-if="apiKeys.length === 0" class="p-4 rounded-lg border border-dashed border-[#27272a] text-center text-zinc-500 text-xs">
              No API keys configured yet. Add an API key below to begin.
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="k in apiKeys"
                :key="k.id"
                class="flex items-center justify-between p-3 rounded-lg border transition-colors"
                :class="k.is_active ? 'bg-[#18181b] border-zinc-600' : 'bg-[#121214] border-[#27272a]'"
              >
                <!-- Key Details -->
                <div class="flex flex-col gap-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="font-medium text-zinc-100 text-xs truncate">{{ k.name }}</span>
                    <span
                      v-if="k.is_active"
                      class="px-1.5 py-0.2 rounded text-[10px] bg-zinc-100 text-zinc-900 font-semibold"
                    >
                      ACTIVE
                    </span>
                  </div>
                  <div class="flex items-center gap-3 text-[11px] text-zinc-400 font-mono">
                    <span>{{ k.key_mask }}</span>
                    <span class="text-zinc-600">|</span>
                    <span class="flex items-center gap-1 text-zinc-400">
                      <Zap class="w-2.5 h-2.5 text-zinc-500" />
                      {{ (k.total_prompt_tokens + k.total_completion_tokens).toLocaleString() }} tokens used
                    </span>
                    <span class="text-zinc-600">|</span>
                    <span>{{ k.total_requests }} requests</span>
                  </div>
                </div>

                <!-- Actions: Activate, Delete -->
                <div class="flex items-center gap-2 flex-shrink-0">
                  <button
                    v-if="!k.is_active"
                    type="button"
                    @click="handleSwitchKey(k.id)"
                    class="px-2.5 py-1 rounded bg-[#27272a] hover:bg-zinc-700 text-zinc-200 text-xs transition-colors"
                  >
                    Activate
                  </button>
                  <button
                    type="button"
                    @click="handleDeleteKey(k.id, k.name)"
                    class="p-1.5 text-zinc-500 hover:text-red-400 rounded hover:bg-zinc-800 transition-colors"
                    title="Delete Key"
                  >
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Add Key Form -->
            <div class="mt-3 p-3 rounded-lg border border-[#27272a] bg-[#121214] space-y-3">
              <span class="font-medium text-xs text-zinc-200 flex items-center gap-1.5">
                <Plus class="w-3.5 h-3.5" />
                <span>Add New API Key</span>
              </span>
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                <input
                  v-model="newKeyName"
                  type="text"
                  placeholder="Label (e.g. Work, Backup)"
                  class="px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                />
                <div class="relative sm:col-span-2">
                  <input
                    v-model="newKeyValue"
                    :type="showNewKey ? 'text' : 'password'"
                    placeholder="IFM API Key (sk-...)"
                    class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 font-mono text-xs focus:outline-none focus:border-zinc-500 pr-8"
                  />
                  <button
                    type="button"
                    @click="showNewKey = !showNewKey"
                    class="absolute right-2 top-2 text-zinc-500 hover:text-zinc-300"
                  >
                    <EyeOff v-if="showNewKey" class="w-3.5 h-3.5" />
                    <Eye v-else class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
              <div class="flex justify-end">
                <button
                  type="button"
                  @click="handleAddKey"
                  :disabled="!newKeyValue.trim()"
                  class="px-3 py-1.5 bg-zinc-100 hover:bg-white disabled:opacity-40 text-zinc-900 rounded-md font-medium text-xs transition-colors"
                >
                  Save API Key
                </button>
              </div>
            </div>
          </div>

          <!-- Connection Test -->
          <div class="pt-3 border-t border-[#27272a]">
            <div class="flex items-center justify-between">
              <div>
                <span class="font-medium text-zinc-200">Test API Connection</span>
                <p class="text-[11px] text-zinc-500">Sends a quick ping to the IFM API endpoint with 60s timeout</p>
              </div>
              <button
                type="button"
                @click="runConnectionTest"
                :disabled="isTestingConnection || apiKeys.length === 0"
                class="px-3 py-1.5 rounded-md border border-[#27272a] bg-[#18181b] hover:bg-[#27272a] text-zinc-200 disabled:opacity-40 font-medium"
              >
                {{ isTestingConnection ? 'Testing (up to 60s)...' : 'Test Connection' }}
              </button>
            </div>

            <!-- Test Result -->
            <div
              v-if="testConnectionResult"
              class="mt-3 p-3 rounded-md border flex items-start gap-2"
              :class="testConnectionResult.success ? 'bg-zinc-900 border-zinc-700 text-zinc-200' : 'bg-red-950/40 border-red-900/50 text-red-300'"
            >
              <Check v-if="testConnectionResult.success" class="w-4 h-4 flex-shrink-0 mt-0.5 text-zinc-100" />
              <AlertCircle v-else class="w-4 h-4 flex-shrink-0 mt-0.5 text-red-400" />
              <span class="text-xs">{{ testConnectionResult.message }}</span>
            </div>
          </div>

          <!-- Tavily Real-time Web Search Integration -->
          <div class="pt-4 border-t border-[#27272a] space-y-3">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div class="w-7 h-7 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-400 border border-emerald-500/20">
                  <Globe class="w-4 h-4" />
                </div>
                <div>
                  <div class="flex items-center gap-1.5">
                    <span class="font-medium text-zinc-200 text-xs">Tavily AI Search (Real-time Web Grounding)</span>
                    <span v-if="tavilyKeyMask" class="px-1.5 py-0.2 bg-emerald-950/60 border border-emerald-800/60 text-emerald-300 rounded text-[10px] font-mono">
                      Active
                    </span>
                  </div>
                  <p class="text-[11px] text-zinc-500">Enables real-time internet search for current news, fuel prices, research & live facts</p>
                </div>
              </div>
            </div>

            <!-- Active Tavily Key Badge -->
            <div v-if="tavilyKeyMask" class="flex items-center justify-between p-2.5 rounded-lg bg-[#18181b] border border-[#27272a]">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono text-zinc-300">{{ tavilyKeyMask }}</span>
                <span class="text-[10px] text-zinc-500">Secured in Windows Vault</span>
              </div>
              <button
                type="button"
                @click="removeTavilyKey"
                class="px-2 py-1 text-xs text-red-400 hover:text-red-300 hover:bg-red-950/40 rounded transition-colors"
              >
                Remove
              </button>
            </div>

            <!-- New Tavily Key Input Form -->
            <div v-else class="space-y-2 p-3 bg-[#18181b]/60 rounded-lg border border-[#27272a]">
              <label class="block text-xs font-medium text-zinc-300">Enter Tavily API Key</label>
              <div class="flex items-center gap-2">
                <div class="relative flex-1">
                  <input
                    v-model="inputTavilyKey"
                    :type="showTavilyKey ? 'text' : 'password'"
                    placeholder="tvly-..."
                    class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 font-mono text-xs focus:outline-none focus:border-zinc-500 pr-8"
                  />
                  <button
                    type="button"
                    @click="showTavilyKey = !showTavilyKey"
                    class="absolute right-2 top-2 text-zinc-500 hover:text-zinc-300"
                  >
                    <EyeOff v-if="showTavilyKey" class="w-3.5 h-3.5" />
                    <Eye v-else class="w-3.5 h-3.5" />
                  </button>
                </div>
                <button
                  type="button"
                  @click="handleSaveTavily"
                  :disabled="!inputTavilyKey.trim()"
                  class="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white font-medium text-xs rounded-md disabled:opacity-40 transition-colors"
                >
                  Save Key
                </button>
              </div>
              <p class="text-[10px] text-zinc-500">
                Get a free API key at <a href="https://tavily.com" target="_blank" class="text-blue-400 hover:underline">tavily.com</a> (1,000 free searches/month).
              </p>
            </div>
          </div>
        </div>

        <!-- TAB: General -->
        <div v-if="activeTab === 'general'" class="space-y-5">
          <!-- Application Defaults Section -->
          <div class="space-y-3">
            <h3 class="font-medium text-xs text-zinc-200">Application Defaults</h3>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block font-medium text-zinc-300 text-[11px] mb-1">Default Model</label>
                <select
                  v-model="defaultModel"
                  class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                >
                  <option v-for="m in models" :key="m.id" :value="m.model_identifier">
                    {{ m.name }} ({{ m.model_identifier }})
                  </option>
                </select>
              </div>

              <div>
                <label class="block font-medium text-zinc-300 text-[11px] mb-1">Default Persona Preset</label>
                <select
                  v-model="defaultPreset"
                  class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                >
                  <option value="">None (Standard AI Assistant)</option>
                  <option v-for="p in presets" :key="p.id" :value="p.id">
                    {{ p.name }}
                  </option>
                </select>
              </div>
            </div>

            <div class="w-48">
              <label class="block font-medium text-zinc-300 text-[11px] mb-1">HTTP Request Timeout (seconds)</label>
              <input
                v-model="requestTimeout"
                type="number"
                min="10"
                max="300"
                class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
              />
            </div>
          </div>

          <!-- Persona Presets Section -->
          <div class="pt-4 border-t border-[#27272a] space-y-3">
            <div class="flex items-center justify-between">
              <div>
                <div class="flex items-center gap-2">
                  <span class="font-medium text-zinc-200 text-xs">Persona Presets</span>
                  <span class="px-1.5 py-0.2 rounded-full text-[10px] bg-zinc-800 text-zinc-400 font-mono">
                    {{ presets.length }}
                  </span>
                </div>
                <p class="text-[11px] text-zinc-500">
                  Custom personas, system instructions, and behavior parameters saved to local database.
                </p>
              </div>
              <button
                type="button"
                @click="startCreatePreset"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-zinc-100 hover:bg-white text-zinc-900 rounded-md font-medium text-xs transition-colors shadow-sm"
              >
                <Plus class="w-3.5 h-3.5" />
                <span>New Preset</span>
              </button>
            </div>

            <!-- Preset Create / Edit Card Form -->
            <div v-if="isEditingPreset" class="p-4 rounded-xl border border-zinc-700 bg-[#121214] space-y-3.5 shadow-xl">
              <div class="flex items-center justify-between border-b border-[#27272a] pb-2.5">
                <div class="flex items-center gap-2 text-zinc-100 font-medium text-xs">
                  <Sparkles class="w-4 h-4 text-zinc-400" />
                  <span>{{ presetForm.id ? `Edit Preset: ${presetForm.name || 'Unnamed'}` : 'Create New Custom Preset' }}</span>
                </div>
                <button
                  type="button"
                  @click="cancelPresetEdit"
                  class="p-1 text-zinc-500 hover:text-zinc-300 rounded hover:bg-zinc-800"
                >
                  <X class="w-3.5 h-3.5" />
                </button>
              </div>

              <!-- Preset Name & Description -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label class="block text-zinc-300 font-medium text-[11px] mb-1">
                    Preset Name <span class="text-red-400">*</span>
                  </label>
                  <input
                    v-model="presetForm.name"
                    type="text"
                    placeholder="e.g. Senior Golang Architect"
                    class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                  />
                </div>
                <div>
                  <label class="block text-zinc-300 font-medium text-[11px] mb-1">
                    Description
                  </label>
                  <input
                    v-model="presetForm.description"
                    type="text"
                    placeholder="e.g. Clean architecture & high-concurrency specialist"
                    class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                  />
                </div>
              </div>

              <!-- System Prompt / Instructions -->
              <div>
                <label class="block text-zinc-300 font-medium text-[11px] mb-1">
                  System Instructions (Prompt) <span class="text-red-400">*</span>
                </label>
                <textarea
                  v-model="presetForm.system_prompt"
                  rows="4"
                  placeholder="You are an expert principal software engineer. Provide clean, secure, idiomatic code..."
                  class="w-full px-3 py-2 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs font-mono leading-relaxed focus:outline-none focus:border-zinc-500"
                ></textarea>
              </div>

              <!-- Parameters: Temperature, Reasoning Effort, Max Tokens -->
              <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-1">
                <div>
                  <label class="block text-zinc-400 text-[11px] mb-1">Temperature ({{ presetForm.temperature }})</label>
                  <input
                    v-model.number="presetForm.temperature"
                    type="number"
                    min="0.0"
                    max="2.0"
                    step="0.05"
                    class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                  />
                </div>
                <div>
                  <label class="block text-zinc-400 text-[11px] mb-1">Reasoning Effort</label>
                  <select
                    v-model="presetForm.reasoning_effort"
                    class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                  >
                    <option value="">Default (None)</option>
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                  </select>
                </div>
                <div>
                  <label class="block text-zinc-400 text-[11px] mb-1">Max Tokens</label>
                  <input
                    v-model.number="presetForm.max_tokens"
                    type="number"
                    min="256"
                    max="32768"
                    step="256"
                    class="w-full px-3 py-1.5 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 text-xs focus:outline-none focus:border-zinc-500"
                  />
                </div>
              </div>

              <!-- Form Buttons -->
              <div class="flex items-center justify-end gap-2 pt-2 border-t border-[#27272a]">
                <button
                  type="button"
                  @click="cancelPresetEdit"
                  class="px-3 py-1.5 rounded-md bg-[#18181b] hover:bg-[#27272a] text-zinc-300 text-xs transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  @click="handleSavePreset"
                  :disabled="!presetForm.name.trim() || !presetForm.system_prompt.trim()"
                  class="px-3.5 py-1.5 bg-zinc-100 hover:bg-white text-zinc-900 font-medium rounded-md text-xs disabled:opacity-40 transition-colors shadow-sm"
                >
                  Save Preset
                </button>
              </div>
            </div>

            <!-- Presets List -->
            <div v-if="presets.length === 0" class="p-6 text-center border border-dashed border-[#27272a] rounded-lg text-zinc-500 text-xs">
              No presets configured. Click "+ New Preset" above to create a custom persona.
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="p in presets"
                :key="p.id"
                class="p-3 rounded-lg border bg-[#121214] transition-colors"
                :class="defaultPreset === p.id ? 'border-emerald-800/60 bg-[#141816]/30' : 'border-[#27272a] hover:border-zinc-700'"
              >
                <!-- Card Header -->
                <div class="flex items-start justify-between gap-3">
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 flex-wrap">
                      <span class="font-semibold text-zinc-100 text-xs">{{ p.name }}</span>
                      <span
                        v-if="defaultPreset === p.id"
                        class="px-1.5 py-0.2 rounded text-[10px] bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 font-semibold"
                      >
                        DEFAULT
                      </span>
                      <span
                        v-if="p.temperature !== undefined && p.temperature !== null"
                        class="px-1.5 py-0.2 rounded text-[10px] bg-zinc-800 text-zinc-400 font-mono"
                      >
                        temp: {{ p.temperature }}
                      </span>
                      <span
                        v-if="p.reasoning_effort"
                        class="px-1.5 py-0.2 rounded text-[10px] bg-zinc-800 text-zinc-400 font-mono"
                      >
                        effort: {{ p.reasoning_effort }}
                      </span>
                    </div>
                    <p v-if="p.description" class="text-[11px] text-zinc-400 mt-1 line-clamp-2">
                      {{ p.description }}
                    </p>
                  </div>

                  <!-- Actions -->
                  <div class="flex items-center gap-1.5 flex-shrink-0">
                    <button
                      v-if="defaultPreset !== p.id"
                      type="button"
                      @click="handleSetDefaultPreset(p.id)"
                      class="px-2 py-1 text-[11px] text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 rounded transition-colors"
                      title="Set as Default Preset"
                    >
                      Make Default
                    </button>
                    <button
                      type="button"
                      @click="startEditPreset(p)"
                      class="p-1.5 text-zinc-400 hover:text-zinc-100 rounded hover:bg-zinc-800 transition-colors"
                      title="Edit Preset"
                    >
                      <Edit2 class="w-3.5 h-3.5" />
                    </button>
                    <button
                      type="button"
                      @click="handleDeletePreset(p)"
                      class="p-1.5 text-zinc-500 hover:text-red-400 rounded hover:bg-zinc-800 transition-colors"
                      title="Delete Preset"
                    >
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>

                <!-- Collapsible System Prompt Preview -->
                <div class="mt-2 pt-2 border-t border-[#1e1e22]">
                  <button
                    type="button"
                    @click="expandedPresetPromptId = expandedPresetPromptId === p.id ? null : p.id"
                    class="inline-flex items-center gap-1 text-[10px] text-zinc-500 hover:text-zinc-300 font-medium transition-colors"
                  >
                    <span>{{ expandedPresetPromptId === p.id ? 'Hide System Prompt' : 'View System Prompt' }}</span>
                    <ChevronUp v-if="expandedPresetPromptId === p.id" class="w-3 h-3" />
                    <ChevronDown v-else class="w-3 h-3" />
                  </button>
                  <div
                    v-if="expandedPresetPromptId === p.id"
                    class="mt-1.5 p-2 bg-[#09090b] rounded border border-[#27272a] text-[11px] font-mono text-zinc-300 whitespace-pre-wrap leading-relaxed max-h-40 overflow-y-auto"
                  >
                    {{ p.system_prompt }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- TAB: Generation -->
        <div v-if="activeTab === 'generation'" class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <span class="font-medium text-zinc-200">Temperature ({{ temperature }})</span>
              <p class="text-[11px] text-zinc-500">Randomness (1.0 default for K2 Horizon reasoning)</p>
            </div>
            <input
              v-model="temperature"
              type="range"
              min="0.0"
              max="2.0"
              step="0.05"
              class="w-48 accent-zinc-200"
            />
          </div>

          <div class="flex items-center justify-between">
            <div>
              <span class="font-medium text-zinc-200">Top P ({{ topP }})</span>
              <p class="text-[11px] text-zinc-500">Nucleus sampling threshold (0.95 recommended)</p>
            </div>
            <input
              v-model="topP"
              type="range"
              min="0.1"
              max="1.0"
              step="0.05"
              class="w-48 accent-zinc-200"
            />
          </div>

          <div>
            <label class="block font-medium text-zinc-200 mb-1">Max Output Tokens</label>
            <input
              v-model="maxTokens"
              type="number"
              min="256"
              max="32768"
              step="256"
              class="w-40 px-3 py-2 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 focus:outline-none focus:border-zinc-500"
            />
          </div>

          <div>
            <label class="block font-medium text-zinc-200 mb-1">Default Reasoning Effort</label>
            <select
              v-model="reasoningEffort"
              class="w-40 px-3 py-2 bg-[#18181b] border border-[#27272a] rounded-md text-zinc-100 focus:outline-none focus:border-zinc-500"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
            </select>
          </div>
        </div>

        <!-- TAB: Storage -->
        <div v-if="activeTab === 'storage'" class="space-y-4">
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="block font-medium text-zinc-200">SQLite Database Location</label>
              <button
                type="button"
                @click="handleRelocateDatabase"
                :disabled="isRelocating"
                class="inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium text-zinc-200 bg-[#18181b] hover:bg-[#27272a] border border-[#27272a] rounded-md transition-colors"
              >
                <Folder class="w-3.5 h-3.5 text-zinc-400" />
                <span>{{ isRelocating ? 'Moving...' : 'Move Location...' }}</span>
              </button>
            </div>
            <div class="p-2.5 bg-[#121214] border border-[#27272a] rounded-md text-xs font-mono text-zinc-400 break-all select-all">
              {{ config?.database_path || '%APPDATA%/IFMChat/database/ifm-chat.db' }}
            </div>
            <p class="text-[11px] text-zinc-500 mt-1">
              Click "Move Location" to safely relocate `ifm-chat.db` and its WAL logs to any other SSD, drive, or folder.
            </p>
          </div>

          <div class="pt-2 flex flex-col gap-2">
            <span class="font-medium text-zinc-200">Backup & Restore</span>
            <div class="flex items-center gap-2">
              <button
                type="button"
                @click="handleExport"
                class="flex items-center gap-1.5 px-3 py-2 bg-[#18181b] hover:bg-[#27272a] border border-[#27272a] rounded-md text-zinc-200 font-medium"
              >
                <Download class="w-3.5 h-3.5" />
                <span>Export JSON</span>
              </button>
              <button
                type="button"
                @click="handleImportClick"
                class="flex items-center gap-1.5 px-3 py-2 bg-[#18181b] hover:bg-[#27272a] border border-[#27272a] rounded-md text-zinc-200 font-medium"
              >
                <Upload class="w-3.5 h-3.5" />
                <span>Import JSON</span>
              </button>
            </div>
          </div>

          <div class="pt-4 border-t border-zinc-800">
            <span class="font-medium text-red-400">Clear All Data</span>
            <p class="text-[11px] text-zinc-500 mb-2">Erase all conversations and chat history from local database</p>
            <button
              type="button"
              @click="handleClearAll"
              class="flex items-center gap-1.5 px-3 py-2 bg-red-950/40 hover:bg-red-900/60 border border-red-900/50 text-red-300 rounded-md font-medium transition-colors"
            >
              <Trash2 class="w-3.5 h-3.5" />
              <span>Clear All Data</span>
            </button>
          </div>
        </div>

        <!-- TAB: About -->
        <div v-if="activeTab === 'about'" class="space-y-4">
          <div class="p-4 bg-[#121214] border border-[#27272a] rounded-lg flex items-center gap-3.5">
            <div class="w-10 h-10 rounded-lg bg-zinc-100 flex items-center justify-center text-zinc-900 text-base font-bold shadow-sm">
              IFM
            </div>
            <div>
              <h3 class="font-semibold text-xs text-zinc-100">IFM Chat Desktop Client</h3>
              <p class="text-[11px] text-zinc-500">Native Windows client for Institute of Foundation Models</p>
            </div>
          </div>

          <div class="space-y-2 text-xs">
            <div class="flex justify-between py-1 border-b border-[#27272a]">
              <span class="text-zinc-500">Version</span>
              <span class="font-mono text-zinc-200">v{{ appVersion }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-[#27272a]">
              <span class="text-zinc-500">Go Runtime</span>
              <span class="font-mono text-zinc-200">go1.27.4 windows/amd64</span>
            </div>
            <div class="flex justify-between py-1 border-b border-[#27272a]">
              <span class="text-zinc-500">Framework</span>
              <span class="font-mono text-zinc-200">Wails v2.17.1</span>
            </div>
            <div class="flex justify-between py-1 border-b border-[#27272a]">
              <span class="text-zinc-500">UI Stack</span>
              <span class="text-zinc-200">Vue 3 + TypeScript + shadcn-vue (Flat Dark)</span>
            </div>
          </div>

          <div class="pt-2">
            <a
              href="https://docs.ifm.ai/"
              target="_blank"
              class="inline-flex items-center gap-1.5 text-zinc-300 hover:text-white underline text-xs"
            >
              <span>Visit Official IFM Documentation</span>
              <ExternalLink class="w-3.5 h-3.5" />
            </a>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end px-6 py-3 border-t border-[#27272a] bg-[#121214] gap-2.5">
        <button
          type="button"
          @click="emit('close')"
          class="px-3.5 py-1.5 text-xs text-zinc-400 hover:text-zinc-200"
        >
          Cancel
        </button>
        <button
          type="button"
          @click="handleSaveAllSettings"
          class="px-3.5 py-1.5 bg-zinc-100 hover:bg-white text-zinc-900 text-xs font-semibold rounded-md shadow-xs transition-colors"
        >
          Save Changes
        </button>
      </div>
    </div>

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
