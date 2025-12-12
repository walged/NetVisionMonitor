import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { useState, useEffect } from 'react'

import ru from './locales/ru.json'
import en from './locales/en.json'

const resources = {
  ru: { translation: ru },
  en: { translation: en },
}

// Get saved language from localStorage or default to 'ru'
const savedLanguage = localStorage.getItem('language') || 'ru'

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    lng: savedLanguage,
    fallbackLng: 'ru',
    interpolation: {
      escapeValue: false,
    },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
      lookupLocalStorage: 'language',
    },
  })

export default i18n

// Simple translate function
export const t = (key: string): string => {
  return i18n.t(key) as string
}

// Custom hook that doesn't use react-i18next's useTranslation
export const useTranslation = () => {
  const [, forceUpdate] = useState(0)

  useEffect(() => {
    const handleLanguageChanged = () => {
      forceUpdate(x => x + 1)
    }
    i18n.on('languageChanged', handleLanguageChanged)
    return () => {
      i18n.off('languageChanged', handleLanguageChanged)
    }
  }, [])

  return {
    t: (key: string): string => i18n.t(key) as string,
    i18n
  }
}

// Helper to change language and persist
export const changeLanguage = (lang: string) => {
  localStorage.setItem('language', lang)
  i18n.changeLanguage(lang)
}

// Get current language
export const getCurrentLanguage = () => i18n.language

// Available languages
export const languages = [
  { code: 'ru', name: 'Русский', flag: '🇷🇺' },
  { code: 'en', name: 'English', flag: '🇺🇸' },
]
