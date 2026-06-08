/**
 * Phase 63 — Guardian session memory UI helpers.
 */

export const SESSION_TOPIC_LABELS = {
  alerts: 'Alerts',
  feeding: 'Feeding',
  comfort: 'Comfort',
  grow: 'Grow',
  stock: 'Stock',
  money: 'Money',
  setup: 'Setup',
}

export function topicChipLabel(topic) {
  return SESSION_TOPIC_LABELS[topic] || topic
}

/** Build chat payload for the "Pick up where I left off" chip. */
export function buildContinueTopicPayload(recent) {
  if (!recent?.summary_text) return null
  const topic = recent.topics?.[0] ? topicChipLabel(recent.topics[0]) : 'this topic'
  return {
    message: `I'd like to continue our earlier discussion about ${topic}.`,
    contextRef: {
      type: 'route',
      path: '/chat',
    },
  }
}
