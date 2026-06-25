import { ref, watch } from 'vue'
import { api } from '@/api/client'

const STORAGE_ENV = 'platform_env_id'

const environments = ref([])
const envId = ref(null)
let bootstrapPromise = null

export function useAppStore() {
  async function loadEnvironments() {
    const list = await api.listEnvironments()
    const order = { BETA: 0, PRE: 1, PROD: 2 }
    environments.value = [...list].sort(
      (a, b) => (order[a.name] ?? 9) - (order[b.name] ?? 9)
    )
    const savedEnv = Number(localStorage.getItem(STORAGE_ENV) || '')
    if (savedEnv && environments.value.some((e) => e.id === savedEnv)) {
      envId.value = savedEnv
      return
    }
    const def =
      environments.value.find((e) => e.name === 'PROD' && e.is_default) ||
      environments.value.find((e) => e.is_default) ||
      environments.value[0]
    if (def && !envId.value) envId.value = def.id
    if (envId.value && !environments.value.some((e) => e.id === envId.value)) {
      envId.value = def?.id ?? null
    }
  }

  async function bootstrap() {
    if (!bootstrapPromise) {
      bootstrapPromise = loadEnvironments()
    }
    await bootstrapPromise
  }

  watch(envId, (id) => {
    if (id) localStorage.setItem(STORAGE_ENV, String(id))
  })

  return {
    environments,
    envId,
    loadEnvironments,
    bootstrap
  }
}
