import { ref } from 'vue'
import { Capacitor } from '@capacitor/core'
import api from '../api'

const registered = ref(false)
const token = ref(null)
const error = ref(null)

let initialized = false

/**
 * Registers for push notifications on native Capacitor platforms.
 * On web this is a no-op for now (Firebase web push can be added later).
 * Call once from App.vue onMounted when the user is authenticated.
 */
export function usePush() {
  async function init() {
    if (initialized) return
    initialized = true

    if (!Capacitor.isNativePlatform()) return

    try {
      const { PushNotifications } = await import('@capacitor/push-notifications')

      const permResult = await PushNotifications.requestPermissions()
      if (permResult.receive !== 'granted') {
        error.value = 'Push permission denied'
        return
      }

      PushNotifications.addListener('registration', async (tokenData) => {
        token.value = tokenData.value
        const platform = Capacitor.getPlatform() // 'android' | 'ios'
        try {
          await api.post('/profile/push-tokens', {
            fcm_token: tokenData.value,
            platform,
            device_label: `${platform} device`,
          })
          registered.value = true
        } catch (e) {
          error.value = e.message || 'Failed to register token'
        }
      })

      PushNotifications.addListener('registrationError', (err) => {
        error.value = err.error || 'Push registration error'
      })

      PushNotifications.addListener('pushNotificationReceived', (notification) => {
        // Foreground notification — could show an in-app toast
        console.info('[push] foreground:', notification)
      })

      PushNotifications.addListener('pushNotificationActionPerformed', (action) => {
        const data = action.notification?.data
        if (data?.route) {
          window.location.hash = ''
          window.location.pathname = data.route
        }
      })

      await PushNotifications.register()
    } catch (e) {
      error.value = e.message || 'Push setup failed'
    }
  }

  return { init, registered, token, error }
}
