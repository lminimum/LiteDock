import { describe, expect, it } from 'vitest'
import { renderMarkdown } from '../markdown'

describe('renderMarkdown', () => {
  it('renders markdown formatting and removes unsafe html', () => {
    const html = renderMarkdown('Hello **world**\n\n<script>alert(1)</script>')

    expect(html).toContain('<strong>world</strong>')
    expect(html).not.toContain('<script>')
  })

  it('renders lists and inline code', () => {
    const html = renderMarkdown('- one\n- two\n\n`ls -la`')

    expect(html).toContain('<ul>')
    expect(html).toContain('<code>ls -la</code>')
  })
})
