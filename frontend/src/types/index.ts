export interface Conversation {
  id: string
  title: string
  model_id: string
  preset_id: string | null
  session_id: string
  system_prompt: string
  folder_path?: string | null
  created_at: string
  updated_at: string
}

export interface WorkspaceFile {
  path: string
  relative_path: string
  size: number
  extension: string
  tokens: number
  included: boolean
}

export interface WorkspaceSummary {
  folder_path: string
  folder_name: string
  total_files: number
  total_size: number
  total_tokens: number
  files: WorkspaceFile[]
}

export interface APIKeyItem {
  id: string
  name: string
  key_mask: string
  is_active: boolean
  total_prompt_tokens: number
  total_completion_tokens: number
  total_requests: number
  last_used_at?: string | null
  created_at: string
}

export interface Message {
  id: string
  conversation_id: string
  parent_message_id?: string | null
  role: 'user' | 'assistant' | 'system'
  content: string
  reasoning_content?: string | null
  tool_calls?: string | null
  tool_call_id?: string | null
  finish_reason?: string | null
  prompt_tokens: number
  completion_tokens: number
  search_source?: string
  search_credits?: number
  created_at: string
}

export interface ModelItem {
  id: string
  name: string
  model_identifier: string
  description: string
  temperature: number
  top_p: number
  top_k?: number | null
  max_tokens: number
  reasoning_effort: string
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface PresetItem {
  id: string
  name: string
  description: string
  system_prompt: string
  model_id?: string | null
  temperature?: number | null
  top_p?: number | null
  max_tokens?: number | null
  reasoning_effort?: string | null
  created_at: string
  updated_at: string
}

export interface AppConfig {
  theme: string
  ifm_base_url: string
  default_model: string
  default_preset: string
  request_timeout: string
  streaming_enabled: string
  temperature: string
  top_p: string
  max_tokens: string
  reasoning_effort: string
  has_api_key: boolean
  has_tavily_key?: boolean
  database_path: string
  all_settings: Record<string, string>
}

export interface StreamEvent {
  conversation_id: string
  message_id: string
  content_delta?: string
  reasoning_delta?: string
  full_content: string
  full_reasoning: string
  finish_reason?: string
  is_done: boolean
  error?: string
  prompt_tokens?: number
  completion_tokens?: number
  search_source?: string
  search_credits?: number
}
