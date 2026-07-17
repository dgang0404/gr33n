/**
 * Help workspace internal scroll + farm-wide alerts inbox.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Help workspace scroll + alerts inbox', () => {
  it('WorkspaceShell pins chrome and scrolls content internally', () => {
    const shell = readFileSync(join(repoRoot, 'ui/src/components/WorkspaceShell.vue'), 'utf8')
    expect(shell).toContain('workspace-shell flex-1 min-h-0 flex flex-col')
    expect(shell).toContain('workspace-shell__subnav shrink-0')
    expect(shell).toContain('workspace-shell__content flex-1 min-h-0 overflow-y-auto overscroll-y-contain')
    expect(shell).not.toContain('sticky top-0')
  })

  it('App.vue locks body scroll and uses a single workspace scroll container', () => {
    const app = readFileSync(join(repoRoot, 'ui/src/App.vue'), 'utf8')
    expect(app).toContain('workspaceByRoute')
    expect(app).toContain('overflow-hidden flex flex-col')
    expect(app).toContain('routeShellClass')
    const styles = readFileSync(join(repoRoot, 'ui/src/style.css'), 'utf8')
    expect(styles).toContain('#app')
    expect(styles).toContain('overflow: hidden')
  })

  it('HelpLibrarySectionNav scrolls within workspace content pane', () => {
    const nav = readFileSync(join(repoRoot, 'ui/src/components/HelpLibrarySectionNav.vue'), 'utf8')
    expect(nav).toContain('scrollToHelpLibrarySection')
    expect(nav).not.toContain("if (section !== 'guide')")
  })

  it('Help workspace uses unified single-row chrome', () => {
    const help = readFileSync(join(repoRoot, 'ui/src/views/workspaces/HelpWorkspace.vue'), 'utf8')
    expect(help).toContain('unified-header')

    const shell = readFileSync(join(repoRoot, 'ui/src/components/WorkspaceShell.vue'), 'utf8')
    expect(shell).toContain('workspace-shell-unified-chrome')
    expect(shell).toContain('unifiedHeader')
  })

  it('HelpLibraryHub shows What lives where map without duplicate Guide card', () => {
    const hub = readFileSync(join(repoRoot, 'ui/src/views/HelpLibraryHub.vue'), 'utf8')
    expect(hub).toContain('HelpKnowledgeSurfacesMap')

    const map = readFileSync(join(repoRoot, 'ui/src/components/HelpKnowledgeSurfacesMap.vue'), 'utf8')
    expect(map).not.toContain("id: 'guide'")
    expect(map).toContain("id: 'knowledge'")
    expect(map).toContain("id: 'catalog'")
    expect(map).toContain("id: 'symptoms'")
  })

  it('App shell opens Guardian from TopBar only, not edge tab', () => {
    const app = readFileSync(join(repoRoot, 'ui/src/App.vue'), 'utf8')
    expect(app).not.toContain('GuardianEdgeTab')
    expect(app).toContain('GuardianDrawer')
  })

  it('/alerts is a farm-wide inbox route (zone_id still deep-links to zone Ops)', () => {
    const router = readFileSync(join(repoRoot, 'ui/src/router/index.js'), 'utf8')
    expect(router).toContain("path: '/alerts'")
    expect(router).toContain('component: Alerts')
    const workspaces = readFileSync(join(repoRoot, 'ui/src/lib/workspaces.js'), 'utf8')
    expect(workspaces).not.toContain("path: '/alerts'")
  })
})
