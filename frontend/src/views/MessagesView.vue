<template>
  <div class="messages-container">
    <ConversationList
      :home-data="homeData"
      :selected-cid="selectedCid"
      :loading="userStore.homeLoading"
      @select="onSelect"
      @jump-to-message="onJumpToMessage"
    />
    <MessagePanel
      :messages="messages"
      :has-more="hasMore"
      :has-more-after="hasMoreAfter"
      :loading="loading"
      :conversation="selectedConversation"
      :highlight-keyword="highlightKeyword"
      :target-message-id="targetMessageId"
      @load-more="loadMore"
      @load-more-after="loadMoreAfter"
      @search-and-jump="onSearchAndJump"
      @target-scrolled="onTargetScrolled"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import ConversationList from '@/components/ConversationList.vue'
import MessagePanel from '@/components/MessagePanel.vue'
import { fetchMessages, fetchMessagesAfter, type Message, type Conversation, type SearchAggregatedItem } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const homeData = computed(() => userStore.homeData)
const selectedCid = ref('')
const selectedConversation = ref<Conversation | null>(null)
const messages = ref<Message[]>([])
const hasMore = ref(false)
const hasMoreAfter = ref(false)
const loading = ref(false)
const highlightKeyword = ref('')
const targetMessageId = ref<number | null>(null)

const loadMessages = async (cid: string) => {
  if (!cid) return
  loading.value = true
  messages.value = []
  targetMessageId.value = null
  hasMoreAfter.value = false
  const res = await fetchMessages(cid)
  messages.value = res.items
  hasMore.value = res.has_more
  loading.value = false
}

const onSelect = (item: Conversation) => {
  selectedCid.value = item.cid
  selectedConversation.value = item
  highlightKeyword.value = ''
  loadMessages(item.cid)
}

const onJumpToMessage = async (cid: string, messageId: number, keyword: string, conversation: SearchAggregatedItem, timestamp: number) => {
  selectedCid.value = cid
  selectedConversation.value = {
    id: 0,
    cid: conversation.cid,
    type: conversation.type,
    title: conversation.title,
    is_top: false,
    message_count: 0,
    last_message_at: 0,
    last_message_id: 0,
    last_message_preview: '',
    created_at: 0,
  }
  highlightKeyword.value = keyword

  // 加载目标消息前后的上下文
  loading.value = true
  messages.value = []
  targetMessageId.value = messageId

  const beforeRes = await fetchMessages(cid, timestamp + 1)
  const afterRes = await fetchMessagesAfter(cid, timestamp, 3)

  // 合并去重
  const allMessages = [...beforeRes.items, ...afterRes.items]
  const uniqueMessages = Array.from(new Map(allMessages.map(m => [m.id, m])).values())
  uniqueMessages.sort((a, b) => a.created_at - b.created_at)

  messages.value = uniqueMessages
  hasMore.value = beforeRes.has_more
  hasMoreAfter.value = afterRes.has_more
  loading.value = false
}

const onSearchAndJump = async (messageId: number, keyword: string, timestamp: number) => {
  if (!selectedConversation.value) return

  highlightKeyword.value = keyword
  targetMessageId.value = messageId

  // 加载目标消息前后的上下文
  loading.value = true
  messages.value = []

  const beforeRes = await fetchMessages(selectedCid.value, timestamp + 1)
  const afterRes = await fetchMessagesAfter(selectedCid.value, timestamp, 3)

  // 合并去重
  const allMessages = [...beforeRes.items, ...afterRes.items]
  const uniqueMessages = Array.from(new Map(allMessages.map(m => [m.id, m])).values())
  uniqueMessages.sort((a, b) => a.created_at - b.created_at)

  messages.value = uniqueMessages
  hasMore.value = beforeRes.has_more
  hasMoreAfter.value = afterRes.has_more
  loading.value = false
}

const onTargetScrolled = () => {
  targetMessageId.value = null
}

// 当首页数据加载完毕，默认选中第一条会话，与传统 IM 行为一致
watch(homeData, (data) => {
  if (data && !selectedCid.value) {
    const firstItem = data.top.items[0]
    if (firstItem) onSelect(firstItem)
  }
}, { immediate: true })

const loadMore = async () => {
  const firstMsg = messages.value[0]
  if (!hasMore.value || loading.value || !firstMsg) return
  loading.value = true
  const res = await fetchMessages(selectedCid.value, firstMsg.created_at)
  messages.value = [...res.items, ...messages.value]
  hasMore.value = res.has_more
  loading.value = false
}

const loadMoreAfter = async () => {
  const lastMsg = messages.value[messages.value.length - 1]
  if (!hasMoreAfter.value || loading.value || !lastMsg) return
  loading.value = true
  const res = await fetchMessagesAfter(selectedCid.value, lastMsg.created_at)
  messages.value = [...messages.value, ...res.items]
  hasMoreAfter.value = res.has_more
  loading.value = false
}
</script>

<style scoped>
.messages-container {
  display: flex;
  height: calc(100vh - 48px);
}
</style>
