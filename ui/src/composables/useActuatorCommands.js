import { useFarmStore } from '../stores/farm.js'
import { ref } from 'vue'

/**
 * Phase 167 — shared actuator ON/OFF and pulse commands (ActuatorCard + quick actions).
 */
export function useActuatorCommands() {
  const store = useFarmStore()
  const busyId = ref(null)
  const feedback = ref('')

  async function sendCommand(actuator, command, reason = '') {
    if (!actuator?.id) return
    busyId.value = actuator.id
    feedback.value = ''
    try {
      await store.enqueueActuatorCommand(
        actuator.id,
        command,
        reason || `Quick action: ${command}`,
      )
      feedback.value = `Queued ${String(command).toUpperCase()} for ${actuator.name}`
      return true
    } catch (e) {
      feedback.value = e?.response?.data?.error || e.message || 'Command failed'
      return false
    } finally {
      busyId.value = null
    }
  }

  async function runPulse(actuator, seconds, reason = '') {
    if (!actuator?.id || !seconds) return false
    busyId.value = actuator.id
    feedback.value = ''
    try {
      await store.enqueueActuatorCommand(
        actuator.id,
        'on',
        reason || `Quick action: ${seconds}s pulse`,
        Math.round(seconds),
      )
      feedback.value = `Queued ${seconds}s pulse on ${actuator.name}`
      return true
    } catch (e) {
      feedback.value = e?.response?.data?.error || e.message || 'Pulse failed'
      return false
    } finally {
      busyId.value = null
    }
  }

  return {
    busyId,
    feedback,
    sendCommand,
    runPulse,
  }
}
