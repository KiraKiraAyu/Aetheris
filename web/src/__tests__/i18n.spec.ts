import { describe, it, expect, beforeEach, vi } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useI18nStore } from '../stores/i18n';

describe('i18n Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('translates static keys correctly in both languages', () => {
    const store = useI18nStore();
    
    // Test English
    store.setLanguage('en');
    expect(store.t('nav.overview')).toBe('Overview');
    expect(store.t('settings.status_connected')).toBe('Connected');

    // Test Chinese
    store.setLanguage('zh');
    expect(store.t('nav.overview')).toBe('仪表盘');
    expect(store.t('settings.status_connected')).toBe('已连接');
  });

  it('handles variable interpolation', () => {
    const store = useI18nStore();
    
    // Test English interpolation
    store.setLanguage('en');
    expect(store.t('overview.kpi_delivered_desc', { pct: '95' })).toBe('95% delivery success rate');

    // Test Chinese interpolation
    store.setLanguage('zh');
    expect(store.t('overview.kpi_delivered_desc', { pct: '95' })).toBe('投递成功率 95%');
  });

  it('detects language from localStorage', () => {
    localStorage.setItem('aetheris.locale', 'zh');
    const store = useI18nStore();
    store.detectLanguage();
    expect(store.locale).toBe('zh');
  });

  it('detects language from navigator.language when localStorage is empty', () => {
    // Mock navigator.language
    const langMock = vi.spyOn(navigator, 'language', 'get').mockReturnValue('zh-CN');
    
    const store = useI18nStore();
    store.detectLanguage();
    expect(store.locale).toBe('zh');

    langMock.mockReturnValue('en-US');
    store.detectLanguage();
    expect(store.locale).toBe('en');
  });
});
