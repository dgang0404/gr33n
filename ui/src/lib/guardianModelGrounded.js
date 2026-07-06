/** Matches server GuardianMinContextWindow (internal/farmguardian/model_cache.go). */
export const GROUNDED_MIN_CONTEXT_WINDOW = 8192

/**
 * True when Ollama-reported context is too small for grounded (farm context) chat.
 * @param {{ context_window?: number } | null | undefined} model
 */
export function modelBlocksGroundedChat(model) {
  if (!model) return false
  const w = Number(model.context_window) || 0
  return w > 0 && w < GROUNDED_MIN_CONTEXT_WINDOW
}

/** True when the model can run grounded (farm context) chat. */
export function isGroundedCapableModel(model) {
  return !!model && !modelBlocksGroundedChat(model)
}

/** Chat models that meet the grounded context minimum. */
export function filterGroundedCapableModels(models) {
  return (Array.isArray(models) ? models : []).filter(isGroundedCapableModel)
}

const GROUNDED_MODEL_PREFERENCES = ['phi3:mini', 'llama3.1:8b', 'llama3.1:8b-instruct-q4_0']

/** Best grounded model to auto-select when the current choice would fail. */
export function pickPreferredGroundedModel(models) {
  const grounded = filterGroundedCapableModels(models)
  for (const prefer of GROUNDED_MODEL_PREFERENCES) {
    const m = findModelByName(prefer, grounded)
    if (m) return m.name
  }
  return grounded[0]?.name || ''
}

export function serverDefaultUsableForGrounded(serverDefault, models) {
  const info = findModelByName(serverDefault, models)
  return isGroundedCapableModel(info)
}

/** True when session-empty resolution (farm → server) would fail grounded chat. */
export function resolvedDefaultBlocksGrounded({ farmModel = '', serverDefault = '', models = [] } = {}) {
  const name = resolveEffectiveChatModelName({ sessionModel: '', farmModel, serverDefault })
  return modelBlocksGroundedChat(findModelByName(name, models))
}

/** Resolve a model name against the /guardian/models list (tolerates :latest). */
export function findModelByName(name, models) {
  const n = String(name || '').trim()
  if (!n || !Array.isArray(models)) return null
  return (
    models.find((m) => m.name === n) ||
    models.find((m) => m.name === `${n}:latest`) ||
    models.find((m) => n === String(m.name).replace(/:latest$/, '')) ||
    null
  )
}

/**
 * Effective chat model for this session: session override → farm policy → server default.
 * @param {{ sessionModel?: string, farmModel?: string, farmCounselModel?: string, farmQuickModel?: string, serverDefault?: string, grounded?: boolean }} opts
 */
export function resolveEffectiveChatModelName({
  sessionModel = '',
  farmModel = '',
  farmCounselModel = '',
  farmQuickModel = '',
  serverDefault = '',
  grounded = true,
} = {}) {
  const s = String(sessionModel || '').trim()
  if (s) return s
  const policy = grounded
    ? (String(farmCounselModel || farmModel || '').trim())
    : String(farmQuickModel || '').trim()
  if (policy) return policy
  return String(serverDefault || '').trim()
}

/** Human-readable source for the effective model (for UI labels). */
export function effectiveModelSource({ sessionModel = '', farmModel = '', serverDefault = '' } = {}) {
  if (String(sessionModel || '').trim()) return 'this chat'
  if (String(farmModel || '').trim()) return 'farm default (saved)'
  if (String(serverDefault || '').trim()) return 'server .env (LLM_MODEL)'
  return 'unconfigured'
}

/** Label for the empty session option in the model dropdown. */
export function sessionDefaultOptionLabel({ farmModel = '', serverDefault = '', models = [] } = {}) {
  const name = resolveEffectiveChatModelName({ sessionModel: '', farmModel, serverDefault })
  if (!name) return 'Farm / server default'
  if (resolvedDefaultBlocksGrounded({ farmModel, serverDefault, models })) {
    return `Farm / server default — not usable (${name}, ctx too small)`
  }
  const src = String(farmModel || '').trim() ? 'farm default' : 'server .env'
  return `Farm / server default → ${name} (${src})`
}

export function farmServerDefaultOptionLabel({ serverDefault = '', models = [] } = {}) {
  if (!serverDefaultUsableForGrounded(serverDefault, models)) {
    return `Server default — not usable for farm context (${serverDefault || 'env'})`
  }
  return `Server default (${serverDefault || 'env'})`
}

export function groundedChatBlockReason(model) {
  if (!model || !modelBlocksGroundedChat(model)) return ''
  const name = model.name || 'This model'
  const w = model.context_window
  return (
    `${name} cannot use farm context: context window is ${w}, below the ${GROUNDED_MIN_CONTEXT_WINDOW} ` +
    'minimum for grounded chat — the server would reject the request (400). ' +
    'Turn off Use farm context for quick chat, or switch to phi3:mini / llama3.1:8b.'
  )
}
