import DOMPurify from 'dompurify'
import { Marked } from 'marked'

const renderer = new Marked({
  gfm: true,
  breaks: true,
})

export function renderMarkdown(text: string): string {
  if (!text) return ''

  const rendered = renderer.parse(text)
  if (typeof rendered !== 'string') return ''

  return DOMPurify.sanitize(rendered)
}
