<template>
  <div class="contacts-container">
    <div class="search-box">
      <a-input
        v-model:value="searchKeyword"
        placeholder="搜索联系人"
        allow-clear
        @compositionstart="isComposing = true"
        @compositionend="onCompositionEnd"
        @input="onInputChange"
        @clear="exitSearch"
      >
        <template #prefix>
          <SearchOutlined style="color: rgba(128, 128, 128, 0.6)" />
        </template>
      </a-input>
    </div>

    <div class="user-list" @scroll="onScroll">
      <a-spin :spinning="loading" class="spin-container">
        <div v-if="users.length === 0 && !loading" class="empty-tip">
          未找到相关用户
        </div>
        <div
          v-for="user in users"
          :key="user.id"
          class="user-item"
        >
          <a-avatar :size="40" :style="{ backgroundColor: '#1677ff' }">
            {{ user.nickname?.charAt(0) }}
          </a-avatar>
          <div class="user-info">
            <div class="user-name">{{ user.nickname }}</div>
            <div class="user-email">{{ user.email }}</div>
          </div>
        </div>
        <div v-if="hasMore && !searchMode" class="load-more-spinner">
          <a-spin size="small" />
        </div>
      </a-spin>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { SearchOutlined } from '@ant-design/icons-vue'
import { fetchUsers, searchUsers, type User } from '@/api'

const users = ref<User[]>([])
const loading = ref(false)
const searchKeyword = ref('')
const searchMode = ref(false)
const isComposing = ref(false)
const page = ref(1)
const total = ref(0)
let searchTimer: ReturnType<typeof setTimeout> | null = null

const hasMore = computed(() => users.value.length < total.value)

const loadUsers = async () => {
  if (loading.value) return
  loading.value = true
  const res = await fetchUsers(page.value)
  users.value = [...users.value, ...res.items]
  total.value = res.total
  loading.value = false
}

const loadMore = async () => {
  if (loading.value || !hasMore.value || searchMode.value) return
  page.value++
  await loadUsers()
}

const doSearch = async (keyword: string) => {
  if (!keyword) {
    exitSearch()
    return
  }
  loading.value = true
  searchMode.value = true
  const results = await searchUsers(keyword)
  users.value = results
  loading.value = false
}

const debouncedSearch = () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    doSearch(searchKeyword.value.trim())
  }, 300)
}

const onInputChange = () => {
  if (!isComposing.value) {
    debouncedSearch()
  }
}

const onCompositionEnd = () => {
  isComposing.value = false
  debouncedSearch()
}

const exitSearch = () => {
  searchMode.value = false
  searchKeyword.value = ''
  users.value = []
  page.value = 1
  total.value = 0
  loadUsers()
}

const onScroll = (e: Event) => {
  const el = e.target as HTMLElement
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 50) {
    loadMore()
  }
}

loadUsers()
</script>

<style scoped>
.contacts-container {
  height: calc(100vh - 48px);
  display: flex;
  flex-direction: column;
}
.search-box {
  padding: 12px;
  border-bottom: 1px solid rgba(128, 128, 128, 0.2);
}
.user-list {
  flex: 1;
  overflow-y: auto;
}
.user-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  cursor: pointer;
  gap: 10px;
}
.user-item:hover {
  background: rgba(128, 128, 128, 0.1);
}
.user-info {
  flex: 1;
  min-width: 0;
}
.user-name {
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.user-email {
  font-size: 12px;
  color: rgba(128, 128, 128, 0.8);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.load-more-spinner {
  text-align: center;
  padding: 12px;
}
.empty-tip {
  text-align: center;
  padding: 24px;
  color: rgba(128, 128, 128, 0.6);
}
.spin-container {
  min-height: 100px;
}
</style>
