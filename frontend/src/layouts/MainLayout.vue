<template>
  <a-config-provider :theme="{ algorithm: themeAlgorithm }">
    <a-layout style="min-height: 100vh">
      <a-layout-sider v-model:collapsed="appStore.collapsed" :theme="siderTheme" :trigger="null">
        <div class="user-info">
          <a-avatar :style="{ backgroundColor: '#1677ff' }">
            {{ userStore.user?.nickname }}
          </a-avatar>
        </div>
        <a-menu v-model:selectedKeys="selectedKeys" :theme="siderTheme" mode="inline" @click="onMenuClick">
          <a-menu-item key="messages">
            <MessageOutlined />
            <span>消息</span>
          </a-menu-item>
          <a-menu-item key="contacts">
            <TeamOutlined />
            <span>联系人</span>
          </a-menu-item>
        </a-menu>
        <div class="sider-footer">
          <div class="footer-btn" @click="appStore.isDark = !appStore.isDark">
            <BulbFilled v-if="appStore.isDark" />
            <BulbOutlined v-else />
          </div>
          <div class="footer-btn" @click="appStore.collapsed = !appStore.collapsed">
            <MenuUnfoldOutlined v-if="appStore.collapsed" />
            <MenuFoldOutlined v-else />
          </div>
        </div>
      </a-layout-sider>
      <a-layout-content style="padding: 24px">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-config-provider>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { theme } from 'ant-design-vue'
import {
  MessageOutlined,
  TeamOutlined,
  BulbOutlined,
  BulbFilled,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons-vue'
import { useAppStore } from '@/stores/app'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const appStore = useAppStore()
const userStore = useUserStore()
const selectedKeys = ref(['messages'])

const themeAlgorithm = computed(() =>
  appStore.isDark ? theme.darkAlgorithm : theme.defaultAlgorithm
)
const siderTheme = computed(() => (appStore.isDark ? 'dark' : 'light'))

const onMenuClick = ({ key }: { key: string }) => {
  router.push(`/${key}`)
}
</script>

<style scoped>
.user-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px;
}
.user-name {
  color: inherit;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 100%;
}
.sider-footer {
  display: flex;
  border-top: 1px solid rgba(128, 128, 128, 0.2);
}
.footer-btn {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  height: 48px;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.2s;
}
.footer-btn:hover {
  background-color: rgba(128, 128, 128, 0.1);
}
:deep(.ant-layout-sider-children) {
  display: flex;
  flex-direction: column;
}
:deep(.ant-menu) {
  flex: 1;
}

</style>
