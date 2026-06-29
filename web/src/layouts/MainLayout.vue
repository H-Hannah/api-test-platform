<template>
  <div class="app-shell" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <aside class="sidebar">
      <div class="brand">
        <img src="/assets/icon48.png" class="brand-logo" alt="" />
        <div v-show="!sidebarCollapsed" class="brand-text">
          <span class="brand-title">接口录制助手</span>
          <span class="brand-sub">测试平台</span>
        </div>
      </div>

      <nav class="nav">
        <template v-for="section in navSections" :key="section.key">
          <p v-show="!sidebarCollapsed" class="nav-section">{{ section.label }}</p>
          <router-link
            v-for="item in section.items"
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isNavActive(item.path) }"
            :title="sidebarCollapsed ? item.label : ''"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span v-show="!sidebarCollapsed" class="nav-label">{{ item.label }}</span>
          </router-link>
        </template>
      </nav>

      <div class="sidebar-foot">
        <button
          type="button"
          class="foot-btn"
          :title="sidebarCollapsed ? '退出登录' : ''"
          @click="logout"
        >
          <el-icon><SwitchButton /></el-icon>
          <span v-show="!sidebarCollapsed">退出登录</span>
        </button>
        <button
          type="button"
          class="foot-btn toggle-btn"
          :title="sidebarCollapsed ? '展开侧栏' : '折叠侧栏'"
          @click="toggleSidebar"
        >
          <el-icon><component :is="sidebarCollapsed ? Expand : Fold" /></el-icon>
          <span v-show="!sidebarCollapsed">收起侧栏</span>
        </button>
      </div>
    </aside>

    <div class="main-area">
      <header class="top-header">
        <h2 class="page-title">{{ route.meta.title }}</h2>
        <div class="runtime-bar">
          <div class="runtime-field runtime-field--env">
            <span class="runtime-label">运行环境</span>
            <el-select v-model="envId" class="runtime-select" placeholder="选择环境">
              <el-option v-for="e in environments" :key="e.id" :label="e.name" :value="e.id" />
            </el-select>
            <router-link to="/environments" class="env-manage-link">管理</router-link>
          </div>
        </div>
      </header>
      <main class="page-content">
        <div class="page-view">
          <router-view />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, provide } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { SwitchButton, Document, List, DataLine, Fold, Expand, HomeFilled, Setting, Coin, Aim } from '@element-plus/icons-vue'
import { useAppStore } from '@/composables/useAppStore'
import { clearToken } from '@/api/client'

const SIDEBAR_KEY = 'platform_sidebar_collapsed'

const route = useRoute()
const router = useRouter()
const { environments, envId, loadEnvironments, bootstrap } = useAppStore()

const sidebarCollapsed = ref(localStorage.getItem(SIDEBAR_KEY) === '1')

provide('appStore', { envId, environments, loadEnvironments, bootstrap })

const navSections = [
  {
    key: 'overview',
    label: '总览',
    items: [{ path: '/', label: '首页', icon: HomeFilled }]
  },
  {
    key: 'prepare',
    label: '测试准备',
    items: [{ path: '/testdata', label: '测试数据', icon: Coin }]
  },
  {
    key: 'execute',
    label: '执行与资产',
    items: [
      { path: '/apis', label: '接口定义', icon: Document },
      { path: '/cases', label: '接口用例', icon: List },
      { path: '/reports', label: '测试报告', icon: DataLine },
      { path: '/environments', label: '环境管理', icon: Setting }
    ]
  },
  {
    key: 'impact',
    label: '精准测试',
    items: [{ path: '/impact', label: 'MR 变更分析', icon: Aim }]
  }
]

function isNavActive(path) {
  if (path === '/') return route.path === '/'
  return route.path === path || route.path.startsWith(path + '/')
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem(SIDEBAR_KEY, sidebarCollapsed.value ? '1' : '0')
}

function logout() {
  clearToken()
  router.replace({ name: 'login' })
}

onMounted(bootstrap)
</script>

<style scoped lang="scss">
.app-shell {
  display: flex;
  height: 100vh;
  overflow: hidden;

  --sidebar-current-w: var(--sidebar-w);

  &.sidebar-collapsed {
    --sidebar-current-w: var(--sidebar-w-collapsed);
  }
}

.sidebar {
  width: var(--sidebar-current-w);
  flex-shrink: 0;
  background: linear-gradient(180deg, #1a2332 0%, #141b26 100%);
  color: #c8d0dc;
  display: flex;
  flex-direction: column;
  transition: width 0.2s ease;
  overflow: hidden;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 12px 20px;
  min-height: 72px;
}

.app-shell.sidebar-collapsed .brand {
  padding: 16px 8px;
  justify-content: center;
}

.brand-logo {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  flex-shrink: 0;
  object-fit: contain;
}

.brand-text {
  min-width: 0;
  overflow: hidden;
}

.brand-title {
  display: block;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
}
.brand-sub {
  font-size: 11px;
  color: #7a8494;
  white-space: nowrap;
}

.nav {
  flex: 1;
  padding: 6px 8px;
  overflow-y: auto;
}

.nav-section {
  margin: 10px 10px 6px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: #6b7585;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 12px;
  margin-bottom: 4px;
  border-radius: 8px;
  color: #a8b0bc;
  text-decoration: none;
  font-size: 14px;
  transition: background 0.15s, color 0.15s;

  &:hover {
    background: var(--color-sidebar-hover);
    color: #e8ecf1;
  }
  &.active {
    background: rgba(51, 112, 255, 0.18);
    color: #fff;
    font-weight: 500;
  }
}

.nav-label {
  white-space: nowrap;
}

.sidebar-foot {
  padding: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.foot-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 12px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #a8b0bc;
  font-size: 13px;
  cursor: pointer;
  white-space: nowrap;

  &:hover {
    background: var(--color-sidebar-hover);
    color: #fff;
  }
}

.app-shell.sidebar-collapsed .nav-item {
  justify-content: center;
  padding: 11px;
}

.app-shell.sidebar-collapsed .foot-btn {
  justify-content: center;
  padding: 10px;
}

.toggle-btn {
  color: #7a8494;
  &:hover {
    color: #c8d0dc;
  }
}

.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--color-bg);
}

.top-header {
  height: var(--header-h);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background: linear-gradient(135deg, #eff6ff 0%, #f0fdf4 100%);
  border-bottom: 1px solid #bfdbfe;
  box-shadow: 0 1px 4px rgba(37, 99, 235, 0.06);
}

.page-title {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: #1e3a5f;
}

.runtime-bar {
  display: flex;
  align-items: center;
  gap: 16px;
}

.runtime-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.runtime-field--env .runtime-select {
  min-width: 120px;
}

.runtime-label {
  font-size: 12px;
  font-weight: 700;
  color: #1e40af;
  white-space: nowrap;
}

.runtime-field--env .runtime-label {
  color: #047857;
}

.runtime-select {
  min-width: 140px;

  :deep(.el-input__wrapper) {
    background: #fff;
    box-shadow: 0 0 0 1px #93c5fd inset;
  }
}

.runtime-field--env .runtime-select :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #6ee7b7 inset;
}

.env-manage-link {
  font-size: 12px;
  color: #047857;
  text-decoration: none;
  white-space: nowrap;
  padding: 4px 8px;
  border-radius: 4px;

  &:hover {
    background: rgba(16, 185, 129, 0.12);
    color: #065f46;
  }
}

.page-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.page-view {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
</style>
