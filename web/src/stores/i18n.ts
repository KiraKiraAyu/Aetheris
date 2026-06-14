import { defineStore } from 'pinia';
import { ref } from 'vue';
import en, { type TranslationSchema } from '../locales/en';
import zh from '../locales/zh';

type KeysOfUnion<T> = T extends unknown ? keyof T : never;

export type TranslationKey = {
  [K in keyof TranslationSchema]: `${K}.${Extract<KeysOfUnion<TranslationSchema[K]>, string>}`
}[keyof TranslationSchema];

const dictionaries: Record<'en' | 'zh', TranslationSchema> = { en, zh };

export const useI18nStore = defineStore('i18n', () => {
  const locale = ref<'en' | 'zh'>('en');

  function detectLanguage() {
    const saved = localStorage.getItem('aetheris.locale');
    if (saved === 'en' || saved === 'zh') {
      locale.value = saved;
      return;
    }
    const navLang = navigator.language || '';
    if (navLang.toLowerCase().startsWith('zh')) {
      locale.value = 'zh';
    } else {
      locale.value = 'en';
    }
  }

  function setLanguage(lang: 'en' | 'zh') {
    locale.value = lang;
    localStorage.setItem('aetheris.locale', lang);
  }

  function t(key: TranslationKey, variables?: Record<string, unknown>): string {
    const parts = key.split('.');
    if (parts.length !== 2) {
      return key;
    }
    const section = parts[0] as keyof TranslationSchema;
    const name = parts[1] as string;

    const dict = dictionaries[locale.value];
    const sectionDict = dict[section] as Record<string, string>;
    const text = sectionDict?.[name];

    if (text === undefined) {
      return key;
    }

    if (!variables) {
      return text;
    }

    return text.replace(/\{(\w+)\}/g, (match: string, p1: string) => {
      return variables[p1] !== undefined ? String(variables[p1]) : match;
    });
  }

  return {
    locale,
    detectLanguage,
    setLanguage,
    t
  };
});
