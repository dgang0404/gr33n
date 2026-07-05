import { ref } from 'vue'
import api from '../api'

/**
 * Shared Ollama model list for Guardian chat UI (selector + grounded gate warnings).
 */
export function useGuardianModels() {
  const models = ref([])
  const serverDefault = ref('')
  const loading = ref(false)
  const loadError = ref(null)

  async function loadModels() {
    loading.value = true
    loadError.value = null
    try {
      const r = await api.get('/guardian/models')
      models.value = Array.isArray(r.data?.available_models) ? r.data.available_models : []
      serverDefault.value = r.data?.server_default || ''
    } catch (e) {
      loadError.value = e.response?.data?.error || 'Could not load models'
      models.value = []
    } finally {
      loading.value = false
    }
  }

  return { models, serverDefault, loading, loadError, loadModels }
}
