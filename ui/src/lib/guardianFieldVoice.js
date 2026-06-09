/**
 * Phase 67 — push-to-talk STT and optional TTS for Guardian field assistant.
 */
import api from '../api'
import { loadGuardianFieldPrefs } from './guardianFieldPrefs.js'

export function speechRecognitionSupported() {
  if (typeof window === 'undefined') return false
  return !!(window.SpeechRecognition || window.webkitSpeechRecognition)
}

export function speechSynthesisSupported() {
  return typeof window !== 'undefined' && !!window.speechSynthesis
}

/**
 * @param {object} handlers
 * @returns {{ start: () => void, stop: () => void, abort: () => void } | null}
 */
export function createPushToTalkRecognizer({ onPartial, onFinal, onError, onState }) {
  const Ctor = window.SpeechRecognition || window.webkitSpeechRecognition
  if (!Ctor) return null

  const rec = new Ctor()
  rec.continuous = false
  rec.interimResults = true
  rec.lang = navigator.language || 'en-US'

  rec.onstart = () => onState?.('listening')
  rec.onend = () => onState?.('idle')
  rec.onerror = (ev) => {
    onState?.('idle')
    onError?.(ev?.error || 'speech_error')
  }
  rec.onresult = (ev) => {
    let interim = ''
    let final = ''
    for (let i = ev.resultIndex; i < ev.results.length; i++) {
      const t = ev.results[i][0]?.transcript || ''
      if (ev.results[i].isFinal) final += t
      else interim += t
    }
    if (interim) onPartial?.(interim.trim())
    if (final) onFinal?.(final.trim())
  }

  return {
    start() {
      try {
        rec.start()
      } catch (e) {
        onError?.(e?.message || 'start_failed')
      }
    },
    stop() {
      try {
        rec.stop()
      } catch { /* ignore */ }
    },
    abort() {
      try {
        rec.abort()
      } catch { /* ignore */ }
    },
  }
}

let activeUtterance = null

/** Stop any in-progress read-aloud. */
export function stopSpeaking() {
  if (typeof window === 'undefined' || !window.speechSynthesis) return
  window.speechSynthesis.cancel()
  activeUtterance = null
}

/**
 * @param {string} text
 * @returns {boolean}
 */
export function speakText(text) {
  stopSpeaking()
  if (!speechSynthesisSupported()) return false
  const plain = String(text || '').replace(/\s+/g, ' ').trim()
  if (!plain) return false
  const u = new SpeechSynthesisUtterance(plain.slice(0, 4000))
  u.rate = 1
  u.pitch = 1
  activeUtterance = u
  window.speechSynthesis.speak(u)
  return true
}

/**
 * Optional LAN whisper.cpp proxy via POST /v1/chat/stt.
 * @param {Blob} audioBlob
 */
export async function transcribeLocalAudio(audioBlob) {
  const fd = new FormData()
  fd.append('audio', audioBlob, 'field.webm')
  const r = await api.post('/v1/chat/stt', fd, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return String(r.data?.text || '').trim()
}

export function preferredSttProvider(sttLocalEnabled) {
  const prefs = loadGuardianFieldPrefs()
  if (prefs.sttProvider === 'local' && sttLocalEnabled) return 'local'
  return 'browser'
}
