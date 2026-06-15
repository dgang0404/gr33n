import { defineStore } from 'pinia'
import { ref, reactive, computed } from 'vue'

export const usePiWizardStore = defineStore('piWizard', () => {
  const currentStep = ref(1)
  
  const formData = reactive({
    // Step 2: Register Pi
    device: {
      name: '',
      uid: '',
      apiKey: '',
      farmId: null,
    },
    // Step 3: Assign Channels
    channelAssignments: {}, // { channel: actuatorId }
    // Step 4: Network
    network: {
      apiBaseUrl: '',
      testInProgress: false,
      testResult: null, // { success, latency_ms, error }
    },
    // Step 5: Config
    configYaml: '',
  })

  const validation = reactive({
    step1: { complete: true, errors: [] },
    step2: { complete: false, errors: [] },
    step3: { complete: false, errors: [] },
    step4: { complete: false, errors: [] },
    step5: { complete: false, errors: [] },
    step6: { complete: false, errors: [] },
  })

  // Computed
  const canAdvance = computed(() => {
    const stepValidation = validation[`step${currentStep.value}`]
    return stepValidation && !stepValidation.errors.length
  })

  const channelCount = computed(() => {
    return Object.keys(formData.channelAssignments).length
  })

  // Actions
  function setStep(step) {
    if (step >= 1 && step <= 6) {
      currentStep.value = step
    }
  }

  function nextStep() {
    if (canAdvance.value && currentStep.value < 6) {
      currentStep.value++
    }
  }

  function prevStep() {
    if (currentStep.value > 1) {
      currentStep.value--
    }
  }

  function updateDevice(data) {
    Object.assign(formData.device, data)
  }

  function setApiKey(key) {
    formData.device.apiKey = key
  }

  function updateChannelAssignment(channel, actuatorId) {
    if (actuatorId === null) {
      delete formData.channelAssignments[channel]
    } else {
      formData.channelAssignments[channel] = actuatorId
    }
  }

  function clearChannelAssignments() {
    formData.channelAssignments = {}
  }

  function updateNetworkConfig(data) {
    Object.assign(formData.network, data)
  }

  function setTestInProgress(inProgress) {
    formData.network.testInProgress = inProgress
  }

  function setTestResult(result) {
    formData.network.testResult = result
  }

  function setConfigYaml(yaml) {
    formData.configYaml = yaml
  }

  function updateValidation(step, errors = []) {
    if (validation[`step${step}`]) {
      validation[`step${step}`].errors = errors
      validation[`step${step}`].complete = errors.length === 0
    }
  }

  function reset() {
    currentStep.value = 1
    formData.device = {
      name: '',
      uid: '',
      apiKey: '',
      farmId: null,
    }
    formData.channelAssignments = {}
    formData.network = {
      apiBaseUrl: '',
      testInProgress: false,
      testResult: null,
    }
    formData.configYaml = ''
    Object.keys(validation).forEach(key => {
      validation[key] = { complete: false, errors: [] }
    })
  }

  return {
    currentStep,
    formData,
    validation,
    canAdvance,
    channelCount,
    setStep,
    nextStep,
    prevStep,
    updateDevice,
    setApiKey,
    updateChannelAssignment,
    clearChannelAssignments,
    updateNetworkConfig,
    setTestInProgress,
    setTestResult,
    setConfigYaml,
    updateValidation,
    reset,
  }
})
