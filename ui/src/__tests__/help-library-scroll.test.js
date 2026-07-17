/**
 * Help Library scroll helpers — section jumps inside WorkspaceShell content pane.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { helpLibraryScrollRoot, scrollToHelpLibrarySection } from '../lib/helpLibraryScroll.js'

describe('helpLibraryScroll', () => {
  beforeEach(() => {
    document.body.innerHTML = `
      <div class="workspace-shell__content" style="height:400px;overflow:auto">
        <div id="help-section-guide" style="height:100px;margin-top:200px">Guide</div>
      </div>
    `
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('finds the workspace content scroll root', () => {
    expect(helpLibraryScrollRoot()?.classList.contains('workspace-shell__content')).toBe(true)
  })

  it('scrollToHelpLibrarySection returns true when section and root exist', () => {
    expect(scrollToHelpLibrarySection('guide', { smooth: false })).toBe(true)
    expect(scrollToHelpLibrarySection('missing')).toBe(false)
  })
})
