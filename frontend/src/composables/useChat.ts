import { ref, computed } from 'vue'
import {
  GetMessages,
  SendMessage,
  StopGeneration,
  RegenerateMessage,
  DeleteMessage
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { Message, StreamEvent } from '../types'

export interface ActiveStream {
  conversationId: string
  messageId: string
  content: string
  reasoning: string
  isGenerating: boolean
  searchSource?: string
  searchCredits?: number
  error?: string | null
}

const messages = ref<Message[]>([])
const isLoadingMessages = ref<boolean>(false)
const currentViewingConversationId = ref<string>('')
const activeStreams = ref<Record<string, ActiveStream>>({})
const streamError = ref<string | null>(null)

let isEventsBound = false

export function useChat() {
  function bindStreamEvents() {
    if (isEventsBound) return
    isEventsBound = true

    EventsOn('chat_stream', (event: StreamEvent) => {
      const convId = event.conversation_id
      if (!convId) return

      if (event.error) {
        if (activeStreams.value[convId]) {
          activeStreams.value[convId].isGenerating = false
          activeStreams.value[convId].error = event.error
        }
        if (convId === currentViewingConversationId.value) {
          streamError.value = event.error
        }
        delete activeStreams.value[convId]
        return
      }

      if (!activeStreams.value[convId]) {
        activeStreams.value[convId] = {
          conversationId: convId,
          messageId: event.message_id || '',
          content: '',
          reasoning: '',
          isGenerating: true,
          error: null
        }
      }

      const stream = activeStreams.value[convId]
      stream.messageId = event.message_id
      stream.content = event.full_content || ''
      stream.reasoning = event.full_reasoning || ''

      if (event.search_source) {
        stream.searchSource = event.search_source
      }
      if (event.search_credits) {
        stream.searchCredits = event.search_credits
      }

      if (event.is_done) {
        stream.isGenerating = false
        delete activeStreams.value[convId]
        // If user is currently looking at this conversation, load messages from DB to display final message!
        if (convId === currentViewingConversationId.value) {
          loadMessages(convId)
        }
      }
    })
  }

  const currentStream = computed(() => {
    const id = currentViewingConversationId.value
    return (id && activeStreams.value[id]) || null
  })

  const isGenerating = computed(() => {
    return !!currentStream.value?.isGenerating
  })

  const streamingContent = computed(() => {
    return currentStream.value?.content || ''
  })

  const streamingReasoning = computed(() => {
    return currentStream.value?.reasoning || ''
  })

  const streamingMessageId = computed(() => {
    return currentStream.value?.messageId || ''
  })

  const streamingConversationId = computed(() => {
    return currentStream.value?.conversationId || ''
  })

  const streamingSearchSource = computed(() => {
    return currentStream.value?.searchSource || ''
  })

  const streamingSearchCredits = computed(() => {
    return currentStream.value?.searchCredits || 0
  })

  const activeGeneratingIds = computed(() => {
    return Object.keys(activeStreams.value).filter(id => activeStreams.value[id]?.isGenerating)
  })

  async function loadMessages(conversationId: string) {
    currentViewingConversationId.value = conversationId
    if (!conversationId) {
      messages.value = []
      return
    }
    isLoadingMessages.value = true
    streamError.value = null
    try {
      const list = await GetMessages(conversationId)
      messages.value = (list || []) as Message[]
    } catch (err) {
      console.error('Failed loading messages:', err)
    } finally {
      isLoadingMessages.value = false
    }
  }

  async function send(conversationId: string, content: string, webSearchMode: 'auto' | 'on' | 'off' = 'auto') {
    if (!content.trim() || activeStreams.value[conversationId]?.isGenerating) return
    streamError.value = null

    activeStreams.value[conversationId] = {
      conversationId,
      messageId: '',
      content: '',
      reasoning: '',
      isGenerating: true
    }

    // Optimistically push user message to UI only if viewing this conversation
    if (conversationId === currentViewingConversationId.value) {
      messages.value.push({
        id: 'temp-' + Date.now(),
        conversation_id: conversationId,
        role: 'user',
        content: content,
        prompt_tokens: 0,
        completion_tokens: 0,
        created_at: new Date().toISOString()
      })
    }

    try {
      await SendMessage(conversationId, content, webSearchMode)
    } catch (err: any) {
      delete activeStreams.value[conversationId]
      if (conversationId === currentViewingConversationId.value) {
        streamError.value = err?.toString() || 'Failed to send message'
      }
    }
  }

  async function stop(conversationId?: string) {
    const targetId = conversationId || currentViewingConversationId.value
    if (!targetId) return
    try {
      await StopGeneration(targetId)
    } catch (err) {
      console.error('Failed to stop generation:', err)
    } finally {
      delete activeStreams.value[targetId]
    }
  }

  async function regenerate(conversationId: string, assistantMessageId: string, webSearchMode: 'auto' | 'on' | 'off' = 'auto') {
    if (activeStreams.value[conversationId]?.isGenerating) return
    streamError.value = null

    activeStreams.value[conversationId] = {
      conversationId,
      messageId: '',
      content: '',
      reasoning: '',
      isGenerating: true
    }

    try {
      await RegenerateMessage(conversationId, assistantMessageId, webSearchMode)
      if (conversationId === currentViewingConversationId.value) {
        await loadMessages(conversationId)
      }
    } catch (err: any) {
      delete activeStreams.value[conversationId]
      if (conversationId === currentViewingConversationId.value) {
        streamError.value = err?.toString() || 'Failed to regenerate'
      }
    }
  }

  async function removeMessage(messageId: string, conversationId: string) {
    try {
      await DeleteMessage(messageId)
      messages.value = messages.value.filter(m => m.id !== messageId)
    } catch (err) {
      console.error('Failed deleting message:', err)
    }
  }

  bindStreamEvents()

  return {
    messages,
    isLoadingMessages,
    isGenerating,
    currentStream,
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
  }
}
