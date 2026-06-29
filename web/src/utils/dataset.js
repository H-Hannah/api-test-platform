export function parseTags(raw) {
  try {
    const list = JSON.parse(raw || '[]')
    return Array.isArray(list) ? list.filter(Boolean) : []
  } catch {
    return []
  }
}

export function hasTag(raw, tag) {
  return parseTags(raw).includes(tag)
}

export function removeTag(raw, tag) {
  return parseTags(raw).filter((t) => t !== tag)
}

export function tagsLabel(raw) {
  const tags = parseTags(raw)
  if (!tags.length) return ''
  return tags.join(', ')
}
