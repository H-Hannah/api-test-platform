<template>
  <div class="home-page">
    <section class="hero panel-card">
      <img src="/assets/icon128.png" class="hero-logo" alt="" />
      <h1>欢迎使用接口录制助手 · 测试平台</h1>
      <p class="hero-desc">
        测试数据 → 接口与场景 → 精准测试分析 MR 变更并执行回归。
      </p>
      <p class="hero-links">
        <router-link to="/testdata">测试数据</router-link>
        <span class="sep">·</span>
        <router-link to="/impact">精准测试</router-link>
        <span class="sep">·</span>
        <router-link to="/apis">接口定义</router-link>
        <span class="sep">·</span>
        <router-link to="/scenarios">测试场景</router-link>
        <span class="sep">·</span>
        <router-link to="/reports">测试报告</router-link>
        <span class="sep">·</span>
        <router-link to="/environments">环境管理</router-link>
      </p>
    </section>
  </div>
</template>

<style scoped lang="scss">
.home-page {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  overflow: auto;
}

.hero {
  text-align: center;
  padding: 48px 40px;
  max-width: 520px;
  width: 100%;
}

.hero-logo {
  width: 72px;
  height: 72px;
  border-radius: 16px;
  margin-bottom: 20px;
}

.hero h1 {
  margin: 0 0 12px;
  font-size: 22px;
  font-weight: 700;
  line-height: 1.35;
}

.hero-desc {
  margin: 0;
  font-size: 14px;
  line-height: 1.65;
  color: var(--color-muted);
}

.hero-links {
  margin: 16px 0 0;
  font-size: 14px;

  a {
    color: var(--color-primary);
    text-decoration: none;
    font-weight: 500;
  }
  .sep {
    margin: 0 8px;
    color: var(--color-muted);
  }
}
</style>
