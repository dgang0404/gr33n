/**
 * Phase 60 — Config YAML Generator
 * Generate minimal bootstrap config.yaml from form data
 */

export function generateConfigYaml(formData) {
  const { device, network } = formData

  const config = {
    api: {
      base_url: network.apiBaseUrl,
      timeout_seconds: 5,
      api_key: device.apiKey,
    },
    device: {
      uid: device.uid,
    },
    farm: {
      farm_id: device.farmId || 1,
    },
    schedule_poll_interval_seconds: 30,
    offline_queue_path: '/var/lib/gr33n/queue.db',
  }

  return toYamlString(config)
}

function toYamlString(obj, indent = 0) {
  const prefix = ' '.repeat(indent)
  let yaml = ''

  for (const [key, value] of Object.entries(obj)) {
    if (value === null || value === undefined) continue

    if (typeof value === 'object' && !Array.isArray(value)) {
      yaml += `${prefix}${key}:\n`
      yaml += toYamlString(value, indent + 2)
    } else if (Array.isArray(value)) {
      yaml += `${prefix}${key}:\n`
      value.forEach(item => {
        yaml += `${prefix}  - ${item}\n`
      })
    } else {
      yaml += `${prefix}${key}: ${formatYamlValue(value)}\n`
    }
  }

  return yaml
}

function formatYamlValue(value) {
  if (typeof value === 'string') {
    return `"${value.replace(/"/g, '\\"')}"`
  }
  return String(value)
}

export function downloadYaml(yamlContent, filename = 'config.yaml') {
  const blob = new Blob([yamlContent], { type: 'text/plain;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export function copyToClipboard(text) {
  return navigator.clipboard.writeText(text)
}
