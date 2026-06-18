import { describe, it, expect, beforeEach, vi } from 'vitest'
import { getAssetUrl } from '../lib/api'

describe('getAssetUrl helper', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
    vi.unstubAllEnvs()
  })

  it('resolves assets correctly in production mode', () => {
    vi.stubEnv('MODE', 'production')
    vi.stubEnv('BASE_URL', '/')

    expect(getAssetUrl('logos/discord.svg')).toBe('/logos/discord.svg')
    expect(getAssetUrl('/icon.svg')).toBe('/icon.svg')
  })

  it('resolves assets correctly in development mode with custom BASE_URL', () => {
    vi.stubEnv('MODE', 'development')
    vi.stubEnv('BASE_URL', '/custom-base/')

    expect(getAssetUrl('logos/discord.svg')).toBe('/custom-base/logos/discord.svg')
    expect(getAssetUrl('/icon.svg')).toBe('/custom-base/icon.svg')
  })

  it('resolves assets correctly in demo mode with trailing slash in URL', () => {
    vi.stubEnv('MODE', 'demo')
    vi.stubGlobal('location', { pathname: '/demo/aetheris/' })

    expect(getAssetUrl('logos/discord.svg')).toBe('/demo/aetheris/logos/discord.svg')
    expect(getAssetUrl('/icon.svg')).toBe('/demo/aetheris/icon.svg')
  })

  it('resolves assets correctly in demo mode without trailing slash in URL', () => {
    vi.stubEnv('MODE', 'demo')
    vi.stubGlobal('location', { pathname: '/demo/aetheris' })

    expect(getAssetUrl('logos/discord.svg')).toBe('/demo/aetheris/logos/discord.svg')
    expect(getAssetUrl('/icon.svg')).toBe('/demo/aetheris/icon.svg')
  })

  it('resolves assets correctly in demo mode when pathname ends with index.html', () => {
    vi.stubEnv('MODE', 'demo')
    vi.stubGlobal('location', { pathname: '/demo/aetheris/index.html' })

    expect(getAssetUrl('logos/discord.svg')).toBe('/demo/aetheris/logos/discord.svg')
    expect(getAssetUrl('/icon.svg')).toBe('/demo/aetheris/icon.svg')
  })
})
