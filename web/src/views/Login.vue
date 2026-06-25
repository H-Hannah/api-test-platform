<template>
  <div class="login-page">
    <div class="login-card panel-card">
      <div class="login-brand">
        <img src="/assets/icon128.png" class="login-logo" alt="接口录制助手" />
        <h1>接口录制助手</h1>
        <p class="login-sub">测试平台 · 请使用访问令牌登录</p>
      </div>

      <el-form class="login-form" @submit.prevent="handleLogin">
        <el-form-item label="API Token">
          <el-input
            v-model="token"
            type="password"
            show-password
            placeholder="与服务器 .env 中 API_TOKEN 一致"
            size="large"
            :disabled="loading"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <p class="login-hint">
          内网共享令牌，用于调用平台 API。插件入库与 Web 管理使用同一套鉴权。
        </p>
        <el-button type="primary" class="login-btn" size="large" :loading="loading" @click="handleLogin">
          登 录
        </el-button>
      </el-form>
    </div>

    <p class="login-foot">默认令牌 <code>TEST123</code>（与后端 <code>.env</code> 中 <code>API_TOKEN</code> 一致）</p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getToken, setToken, verifyToken } from '@/api/client'

const route = useRoute()
const router = useRouter()
const token = ref('')
const loading = ref(false)

onMounted(() => {
  const saved = getToken()
  if (saved) token.value = saved
})

async function handleLogin() {
  const value = token.value.trim()
  if (!value) {
    ElMessage.warning('请输入 API Token')
    return
  }
  loading.value = true
  try {
    await verifyToken(value)
    setToken(value)
    ElMessage.success('登录成功')
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/apis'
    router.replace(redirect)
  } catch (e) {
    ElMessage.error(e.message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background:
    radial-gradient(ellipse 80% 60% at 50% 0%, rgba(51, 112, 255, 0.12), transparent),
    var(--color-bg);
}

.login-card {
  width: 100%;
  max-width: 400px;
  padding: 40px 36px 32px;
}

.login-brand {
  text-align: center;
  margin-bottom: 32px;
}

.login-logo {
  width: 72px;
  height: 72px;
  border-radius: 16px;
  margin-bottom: 16px;
  box-shadow: var(--shadow-md);
}

.login-brand h1 {
  margin: 0 0 8px;
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text);
}

.login-sub {
  margin: 0;
  font-size: 13px;
  color: var(--color-muted);
}

.login-form :deep(.el-form-item__label) {
  font-weight: 600;
  color: #374151;
}

.login-hint {
  margin: -8px 0 20px;
  font-size: 12px;
  line-height: 1.55;
  color: var(--color-muted);
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 8px;
}

.login-foot {
  margin-top: 20px;
  font-size: 12px;
  color: var(--color-muted);
  text-align: center;

  code {
    background: #fff;
    padding: 2px 6px;
    border-radius: 4px;
    border: 1px solid var(--color-border);
    font-size: 11px;
  }
}
</style>
