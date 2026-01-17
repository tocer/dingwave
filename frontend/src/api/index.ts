export interface User {
  id: number
  nickname: string
  email: string
}

export interface Conversation {
  id: number
  cid: string
  type: number
  title: string
  is_top: boolean
  message_count: number
  last_message_at: number
  last_message_id: number
  last_message_preview: string
  created_at: number
}

export interface ConversationList {
  total: number
  items: Conversation[]
}

export interface HomeResponse {
  current_user: User
  top: ConversationList
  single: ConversationList
  group: ConversationList
}

export interface Message {
  id: number
  cid: string
  sender_id: number
  sender_name: string
  content_type: number
  content_text: string
  content_json: string
  created_at: number
  is_recall: boolean
}

export interface MessagesResponse {
  has_more: boolean
  items: Message[]
}

export interface ConversationPageResponse {
  total: number
  page: number
  size: number
  items: Conversation[]
}

export interface SearchAggregatedItem {
  cid: string
  title: string
  type: number
  match_count: number
}

export interface SearchAggregatedResponse {
  total: number
  page: number
  size: number
  items: SearchAggregatedItem[]
}

export interface SearchMessageItem {
  id: number
  cid: string
  sender_id: number
  sender_name: string
  content_type: number
  content_text: string
  created_at: number
  is_recall: boolean
}

export interface SearchConversationResponse {
  total: number
  page: number
  size: number
  items: SearchMessageItem[]
}

export interface UserListResponse {
  total: number
  page: number
  size: number
  items: User[]
}

export const fetchHome = (limit = 5): Promise<HomeResponse> =>
  fetch(`/api/conversations/home?limit=${limit}`).then((r) => r.json())

export const fetchConversations = (
  type: number,
  page = 1,
  size = 20
): Promise<ConversationPageResponse> =>
  fetch(`/api/conversations?type=${type}&page=${page}&size=${size}`).then((r) =>
    r.json()
  )

export const fetchMessages = (cid: string, before?: number): Promise<MessagesResponse> =>
  fetch(`/api/conversations/${cid}/messages${before ? `?before=${before}` : ''}`).then(
    (r) => r.json()
  )

export const fetchMessagesAfter = (cid: string, after: number, size = 50): Promise<MessagesResponse> =>
  fetch(`/api/conversations/${cid}/messages?after=${after}&size=${size}`).then((r) => r.json())

export const searchMessages = (keyword: string, page = 1, size = 20): Promise<SearchAggregatedResponse> =>
  fetch(`/api/messages/search?q=${encodeURIComponent(keyword)}&page=${page}&size=${size}`).then(
    (r) => r.json()
  )

export const searchConversationMessages = (
  cid: string,
  keyword: string,
  page = 1,
  size = 20
): Promise<SearchConversationResponse> =>
  fetch(
    `/api/conversations/${cid}/messages/search?q=${encodeURIComponent(keyword)}&page=${page}&size=${size}`
  ).then((r) => r.json())

export const fetchUsers = (page = 1, size = 50): Promise<UserListResponse> =>
  fetch(`/api/users?page=${page}&size=${size}`).then((r) => r.json())

export const searchUsers = (keyword: string): Promise<User[]> =>
  fetch(`/api/users/search?q=${encodeURIComponent(keyword)}`).then((r) => r.json())
