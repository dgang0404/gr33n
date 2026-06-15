/**
 * Phase 60 — Pi Setup Wizard Validation
 * Real-time validation for each step
 */

export function validateStep2(formData) {
  const errors = []
  
  if (!formData.device.name?.trim()) {
    errors.push('Device name is required')
  }
  if (!formData.device.uid?.trim()) {
    errors.push('Device UID is required')
  }
  if (!formData.device.apiKey) {
    errors.push('API key must be generated')
  }
  
  return errors
}

export function validateStep3(formData) {
  const errors = []
  
  const assignedCount = Object.keys(formData.channelAssignments).length
  if (assignedCount === 0) {
    errors.push('At least one actuator must be assigned to a channel')
  }
  
  return errors
}

export function validateStep4(formData) {
  const errors = []
  
  if (!formData.network.apiBaseUrl?.trim()) {
    errors.push('API base URL is required')
  } else if (!isValidUrl(formData.network.apiBaseUrl)) {
    errors.push('API base URL must be a valid URL (e.g., http://192.168.1.50:8080)')
  }
  
  return errors
}

export function validateStep5(formData) {
  const errors = []
  
  if (!formData.configYaml?.trim()) {
    errors.push('Config YAML not generated yet')
  }
  
  return errors
}

export function validateStep6(formData) {
  // Step 6 is confirmation; no hard blockers
  return []
}

function isValidUrl(string) {
  try {
    new URL(string)
    return true
  } catch {
    return false
  }
}
