import { ref, computed } from 'vue'
import {
  ListConversations,
  CreateConversation,
  DeleteConversation,
  UpdateConversationTitle,
  UpdateConversationModel,
  ClearConversationMessages,
  ExportConversationMarkdown
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { Conversation } from '../types'

const conversations = ref<Conversation[]>([])
const activeId = ref<string | null>(null)
const searchQuery = ref<string>('')
const isLoadingConversations = ref<boolean>(false)

export function useConversations() {
  async function loadConversations(search = '') {
    isLoadingConversations.value = true
    try {
      const list = await ListConversations(search)
      conversations.value = (list || []) as Conversation[]
      if (!activeId.value && conversations.value.length > 0) {
        activeId.value = conversations.value[0].id
      }
    } catch (err) {
      console.error('Failed loading conversations:', err)
    } finally {
      isLoadingConversations.value = false
    }
  }

  // Grouping logic for conversations
  const groupedConversations = computed(() => {
    const today: Conversation[] = []
    const yesterday: Conversation[] = []
    const prev7Days: Conversation[] = []
    const older: Conversation[] = []

    const now = new Date()
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
    const oneDayMs = 24 * 60 * 60 * 1000
    const yesterdayStart = todayStart - oneDayMs
    const sevenDaysStart = todayStart - 7 * oneDayMs

    for (const c of conversations.value) {
      const updatedTime = new Date(c.updated_at).getTime()
      if (updatedTime >= todayStart) {
        today.push(c)
      } else if (updatedTime >= yesterdayStart) {
        yesterday.push(c)
      } else if (updatedTime >= sevenDaysStart) {
        prev7Days.push(c)
      } else {
        older.push(c)
      }
    }

    return [
      { title: 'Today', items: today },
      { title: 'Yesterday', items: yesterday },
      { title: 'Previous 7 days', items: prev7Days },
      { title: 'Older', items: older }
    ].filter(g => g.items.length > 0)
  })

  const activeConversation = computed(() => {
    return conversations.value.find(c => c.id === activeId.value) || null
  })

  async function newConversation(title = 'New Chat', modelId = 'IFM/K2-Horizon-375B-A23B', presetId = '', systemPrompt = '') {
    try {
      const created = await CreateConversation(title, modelId, presetId, systemPrompt)
      if (created) {
        conversations.value.unshift(created as Conversation)
        activeId.value = created.id
        return created
      }
    } catch (err) {
      console.error('Failed creating conversation:', err)
    }
    return null
  }

  async function removeConversation(id: string) {
    try {
      await DeleteConversation(id)
      conversations.value = conversations.value.filter(c => c.id !== id)
      if (activeId.value === id) {
        activeId.value = conversations.value.length > 0 ? conversations.value[0].id : null
      }
    } catch (err) {
      console.error('Failed deleting conversation:', err)
    }
  }

  async function renameConversation(id: string, newTitle: string) {
    try {
      await UpdateConversationTitle(id, newTitle)
      const conv = conversations.value.find(c => c.id === id)
      if (conv) {
        conv.title = newTitle
      }
    } catch (err) {
      console.error('Failed renaming conversation:', err)
    }
  }

  async function clearMessages(id: string) {
    try {
      await ClearConversationMessages(id)
    } catch (err) {
      console.error('Failed clearing messages:', err)
    }
  }

  async function changeModel(id: string, modelId: string) {
    try {
      await UpdateConversationModel(id, modelId)
      const conv = conversations.value.find(c => c.id === id)
      if (conv) {
        conv.model_id = modelId
      }
    } catch (err) {
      console.error('Failed updating model:', err)
    }
  }

  async function exportMarkdown(id: string): Promise<string> {
    try {
      return await ExportConversationMarkdown(id)
    } catch (err) {
      console.error('Failed exporting markdown:', err)
      return ''
    }
  }

  function setActiveId(id: string) {
    activeId.value = id
  }

  // Subscribe to backend auto title update event
  EventsOn('conversation_title_updated', (data: { id: string; title: string }) => {
    const conv = conversations.value.find(c => c.id === data.id)
    if (conv) {
      conv.title = data.title
    }
  })

  return {
    conversations,
    activeId,
    searchQuery,
    isLoadingConversations,
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
  }
}
