/**
 * Phase 180 WS5 — citation doc view helpers.
 */

export function humanDocTitle(docPath) {
  const base = String(docPath || '').split('/').pop() || docPath
  return base
    .replace(/\.md$/i, '')
    .replace(/^crop-/, '')
    .replace(/[-_]/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

export function fieldGuideSlugFromDocPath(docPath) {
  const base = String(docPath || '').split('/').pop() || docPath
  return base.replace(/\.md$/i, '')
}

/** Strip RAG ingest header lines (source type, doc_path) from chunk body text. */
export function chunkDisplayText(contentText) {
  const lines = String(contentText || '').split('\n')
  let i = 0
  if (i < lines.length && /^(field_guide|platform_doc|symptom_guide)$/i.test(lines[i].trim())) {
    i++
  }
  while (i < lines.length) {
    const line = lines[i].trim()
    if (line.startsWith('doc_path:') || line.startsWith('type=')) {
      i++
      continue
    }
    if (line === '') {
      i++
      break
    }
    break
  }
  return lines.slice(i).join('\n').trim()
}

export function guardianDocPrefill(title) {
  const label = String(title || 'this document').trim() || 'this document'
  return `I'm reading "${label}" from our indexed knowledge. What should I know from this doc for my farm right now?`
}

export function citationTypeLabel(citedType) {
  if (citedType === 'platform_doc') return 'Platform doc'
  if (citedType === 'field_guide') return 'Field guide'
  return 'Indexed doc'
}
