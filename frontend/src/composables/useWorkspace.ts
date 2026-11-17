import { ref } from 'vue'
import {
  SelectFolderDialog,
  AttachFolder,
  DetachFolder,
  GetWorkspaceSummary,
  OpenFolderInExplorer
} from '../../wailsjs/go/main/App'
import type { WorkspaceSummary } from '../types'

const currentWorkspace = ref<WorkspaceSummary | null>(null)
const isLoadingWorkspace = ref(false)
const workspaceError = ref<string | null>(null)

export function useWorkspace() {
  async function loadWorkspace(conversationId: string) {
    if (!conversationId) {
      currentWorkspace.value = null
      return
    }
    isLoadingWorkspace.value = true
    workspaceError.value = null
    try {
      const summary = await GetWorkspaceSummary(conversationId)
      currentWorkspace.value = (summary as unknown as WorkspaceSummary) || null
    } catch (err: any) {
      console.error('Failed to load workspace:', err)
      currentWorkspace.value = null
      workspaceError.value = String(err)
    } finally {
      isLoadingWorkspace.value = false
    }
  }

  async function chooseAndAttachFolder(conversationId: string): Promise<boolean> {
    if (!conversationId) return false
    isLoadingWorkspace.value = true
    workspaceError.value = null
    try {
      const selectedDir = await SelectFolderDialog()
      if (!selectedDir) {
        isLoadingWorkspace.value = false
        return false // User cancelled
      }
      const summary = await AttachFolder(conversationId, selectedDir)
      currentWorkspace.value = (summary as unknown as WorkspaceSummary) || null
      return true
    } catch (err: any) {
      console.error('Failed to attach folder:', err)
      workspaceError.value = String(err)
      return false
    } finally {
      isLoadingWorkspace.value = false
    }
  }

  async function removeFolder(conversationId: string) {
    if (!conversationId) return
    isLoadingWorkspace.value = true
    try {
      await DetachFolder(conversationId)
      currentWorkspace.value = null
    } catch (err: any) {
      console.error('Failed to detach folder:', err)
      workspaceError.value = String(err)
    } finally {
      isLoadingWorkspace.value = false
    }
  }

  async function refresh(conversationId: string) {
    if (!conversationId || !currentWorkspace.value?.folder_path) return
    isLoadingWorkspace.value = true
    try {
      const summary = await AttachFolder(conversationId, currentWorkspace.value.folder_path)
      currentWorkspace.value = (summary as unknown as WorkspaceSummary) || null
    } catch (err: any) {
      console.error('Failed to refresh workspace:', err)
      workspaceError.value = String(err)
    } finally {
      isLoadingWorkspace.value = false
    }
  }

  async function openFolder(folderPath: string) {
    try {
      await OpenFolderInExplorer(folderPath)
    } catch (err: any) {
      console.error('Failed to open explorer:', err)
    }
  }

  return {
    currentWorkspace,
    isLoadingWorkspace,
    workspaceError,
    loadWorkspace,
    chooseAndAttachFolder,
    removeFolder,
    refresh,
    openFolder
  }
}
