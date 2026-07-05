import { describe, it, expect } from 'vitest'
import {
  GROUNDED_MIN_CONTEXT_WINDOW,
  effectiveModelSource,
  filterGroundedCapableModels,
  findModelByName,
  groundedChatBlockReason,
  modelBlocksGroundedChat,
  pickPreferredGroundedModel,
  resolveEffectiveChatModelName,
  serverDefaultUsableForGrounded,
  sessionDefaultOptionLabel,
} from '../lib/guardianModelGrounded.js'

describe('guardianModelGrounded', () => {
  it('flags tinyllama as blocking grounded chat', () => {
    const m = { name: 'tinyllama:latest', context_window: 2048 }
    expect(modelBlocksGroundedChat(m)).toBe(true)
    expect(groundedChatBlockReason(m)).toContain(String(GROUNDED_MIN_CONTEXT_WINDOW))
    expect(groundedChatBlockReason(m)).toContain('400')
  })

  it('allows phi3 for grounded chat', () => {
    const m = { name: 'phi3:mini', context_window: 131072 }
    expect(modelBlocksGroundedChat(m)).toBe(false)
    expect(groundedChatBlockReason(m)).toBe('')
  })

  it('resolves effective model precedence', () => {
    expect(
      resolveEffectiveChatModelName({
        sessionModel: 'tinyllama',
        farmModel: 'phi3:mini',
        serverDefault: 'llama3.1:8b',
      }),
    ).toBe('tinyllama')
    expect(
      resolveEffectiveChatModelName({
        sessionModel: '',
        farmModel: 'phi3:mini',
        serverDefault: 'llama3.1:8b',
      }),
    ).toBe('phi3:mini')
  })

  it('findModelByName tolerates :latest', () => {
    const models = [{ name: 'phi3:mini', context_window: 8192 }]
    expect(findModelByName('phi3:mini', models)?.name).toBe('phi3:mini')
  })

  it('labels session default option with resolved model', () => {
    expect(
      sessionDefaultOptionLabel({ farmModel: 'phi3:mini', serverDefault: 'tinyllama' }),
    ).toContain('phi3:mini')
    expect(
      sessionDefaultOptionLabel({ farmModel: '', serverDefault: 'tinyllama', models: [{ name: 'tinyllama', context_window: 2048 }] }),
    ).toContain('not usable')
    expect(effectiveModelSource({ farmModel: 'phi3:mini', serverDefault: 'tinyllama' })).toBe(
      'farm default (saved)',
    )
  })

  it('filters and picks grounded models', () => {
    const models = [
      { name: 'tinyllama', context_window: 2048 },
      { name: 'phi3:mini', context_window: 131072 },
    ]
    expect(filterGroundedCapableModels(models).map((m) => m.name)).toEqual(['phi3:mini'])
    expect(pickPreferredGroundedModel(models)).toBe('phi3:mini')
    expect(serverDefaultUsableForGrounded('tinyllama', models)).toBe(false)
  })
})
