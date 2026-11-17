import { ref, computed } from 'vue'
import {
  GetAppConfig,
  SaveSettings,
  SaveAPIKey,
  DeleteAPIKey,
  TestConnection,
  ListAPIKeys,
  AddAPIKey,
  SetActiveAPIKey,
  DeleteAPIKeyByID,
  ListModels,
  CreateModel,
  UpdateModel,
  DeleteModel,
  SetDefaultModel,
  ListPresets,
  CreatePreset,
  UpdatePreset,
  DeletePreset,
  ExportConversations,
  ImportConversations,
  ClearApplicationData,
  GetAppVersion,
  SaveTavilyKey,
  GetTavilyKeyMask,
  DeleteTavilyKey,
  SelectFolderForDatabase,
  MoveDatabaseLocation
} from '../../wailsjs/go/main/App'
import type { AppConfig, ModelItem, PresetItem, APIKeyItem } from '../types'

const config = ref<AppConfig | null>(null)
const apiKeys = ref<APIKeyItem[]>([])
const models = ref<ModelItem[]>([])
const presets = ref<PresetItem[]>([])
const appVersion = ref<string>('1.0.0')
const isTestingConnection = ref<boolean>(false)
const testConnectionResult = ref<{ success: boolean; message: string } | null>(null)

export function useSettings() {
  async function loadConfig() {
    try {
      const [cfg, keys, mdlList, prsList, ver] = await Promise.all([
        GetAppConfig(),
        ListAPIKeys(),
        ListModels(),
        ListPresets(),
        GetAppVersion()
      ])
      config.value = cfg as AppConfig
      apiKeys.value = (keys || []) as APIKeyItem[]
      models.value = mdlList as ModelItem[]
      presets.value = prsList as PresetItem[]
      appVersion.value = ver
    } catch (err) {
      console.error('Failed to load settings:', err)
    }
  }

  const activeApiKey = computed(() => {
    return apiKeys.value.find(k => k.is_active) || apiKeys.value[0] || null
  })

  async function addNewKey(name: string, key: string) {
    await AddAPIKey(name, key)
    apiKeys.value = (await ListAPIKeys()) as APIKeyItem[]
    if (config.value) {
      config.value.has_api_key = true
    }
  }

  async function switchActiveKey(id: string) {
    await SetActiveAPIKey(id)
    apiKeys.value = (await ListAPIKeys()) as APIKeyItem[]
  }

  async function deleteKey(id: string) {
    await DeleteAPIKeyByID(id)
    apiKeys.value = (await ListAPIKeys()) as APIKeyItem[]
    if (config.value && apiKeys.value.length === 0) {
      config.value.has_api_key = false
    }
  }

  const defaultModel = computed(() => {
    return models.value.find(m => m.is_default) || models.value[0] || null
  })

  async function updateApiKey(key: string) {
    await SaveAPIKey(key)
    if (config.value) {
      config.value.has_api_key = true
    }
  }

  async function removeApiKey() {
    await DeleteAPIKey()
    if (config.value) {
      config.value.has_api_key = false
    }
  }

  async function runConnectionTest() {
    isTestingConnection.value = true
    testConnectionResult.value = null
    try {
      await TestConnection()
      testConnectionResult.value = {
        success: true,
        message: 'Connected to IFM API successfully!'
      }
    } catch (err: any) {
      testConnectionResult.value = {
        success: false,
        message: err || 'Connection failed. Please check your API key and network connection.'
      }
    } finally {
      isTestingConnection.value = false
    }
  }

  async function updateSettings(settingsMap: Record<string, string>) {
    await SaveSettings(settingsMap)
    await loadConfig()
  }

  async function createOrUpdateModel(m: ModelItem) {
    if (m.id) {
      await UpdateModel(m as any)
    } else {
      await CreateModel(m as any)
    }
    models.value = (await ListModels()) as ModelItem[]
  }

  async function removeModel(id: string) {
    await DeleteModel(id)
    models.value = (await ListModels()) as ModelItem[]
  }

  async function makeDefaultModel(id: string) {
    await SetDefaultModel(id)
    models.value = (await ListModels()) as ModelItem[]
  }

  async function createOrUpdatePreset(p: Partial<PresetItem> & { name: string; system_prompt: string }) {
    const payload: any = {
      id: p.id || '',
      name: p.name.trim(),
      description: p.description?.trim() || '',
      system_prompt: p.system_prompt.trim(),
      model_id: p.model_id || null,
      temperature: p.temperature !== undefined && p.temperature !== null ? Number(p.temperature) : 0.7,
      top_p: p.top_p !== undefined && p.top_p !== null ? Number(p.top_p) : 0.95,
      max_tokens: p.max_tokens !== undefined && p.max_tokens !== null ? Number(p.max_tokens) : 4096,
      reasoning_effort: p.reasoning_effort || '',
      created_at: p.created_at || new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    if (p.id) {
      await UpdatePreset(payload)
    } else {
      await CreatePreset(payload)
    }
    presets.value = (await ListPresets()) as PresetItem[]
  }

  async function removePreset(id: string) {
    await DeletePreset(id)
    presets.value = (await ListPresets()) as PresetItem[]
  }

  async function exportData(): Promise<string> {
    return await ExportConversations()
  }

  async function importData(jsonContent: string) {
    await ImportConversations(jsonContent)
  }

  async function resetAllData() {
    await ClearApplicationData()
  }

  const tavilyKeyMask = ref<string>('')

  async function loadTavilyKey() {
    try {
      tavilyKeyMask.value = await GetTavilyKeyMask()
    } catch (err) {
      console.error('Failed loading Tavily key mask:', err)
    }
  }

  async function setTavilyKey(key: string) {
    await SaveTavilyKey(key)
    await loadTavilyKey()
    await loadConfig()
  }

  async function removeTavilyKey() {
    await DeleteTavilyKey()
    await loadTavilyKey()
    await loadConfig()
  }

  async function relocateDatabase(): Promise<string | null> {
    const dir = await SelectFolderForDatabase()
    if (!dir) return null
    const newPath = await MoveDatabaseLocation(dir)
    await loadConfig()
    return newPath
  }

  return {
    config,
    apiKeys,
    activeApiKey,
    models,
    presets,
    defaultModel,
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
    updateApiKey,
    removeApiKey,
    runConnectionTest,
    updateSettings,
    createOrUpdateModel,
    removeModel,
    makeDefaultModel,
    createOrUpdatePreset,
    removePreset,
    exportData,
    importData,
    resetAllData
  }
}
