import { en, type TranslationKey } from './en';
import { vi } from './vi';

export type Locale = 'en' | 'vi';

const dictionaries: Record<Locale, Record<TranslationKey, string>> = { en, vi };

export function getTranslations(locale: Locale) {
  const dict = dictionaries[locale] ?? dictionaries.en;

  return function t(key: TranslationKey, params?: Record<string, string | number>): string {
    let value = dict[key] ?? dictionaries.en[key] ?? key;
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        value = value.replace(`{${k}}`, String(v));
      }
    }
    return value;
  };
}

export type { TranslationKey };
