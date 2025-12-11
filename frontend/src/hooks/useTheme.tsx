import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { GetAppSettings, SaveAppSettings } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { main } from '../../wailsjs/go/models'

type Theme = 'dark' | 'light' | 'system'

interface ThemeContextType {
  theme: Theme
  setTheme: (theme: Theme) => void
  effectiveTheme: 'dark' | 'light'
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

function getSystemTheme(): 'dark' | 'light' {
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return 'dark'
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>('light')
  const [effectiveTheme, setEffectiveTheme] = useState<'dark' | 'light'>('light')

  // Load theme from settings
  useEffect(() => {
    GetAppSettings()
      .then((settings) => {
        const savedTheme = (settings.theme || 'light') as Theme
        setThemeState(savedTheme)
      })
      .catch((err) => {
        console.error('Failed to load theme:', err)
      })

    // Listen for settings changes
    const unsubscribe = EventsOn('settings:changed', (settings: main.AppSettings) => {
      if (settings.theme) {
        setThemeState(settings.theme as Theme)
      }
    })

    return () => {
      unsubscribe()
    }
  }, [])

  // Update effective theme and apply to document
  useEffect(() => {
    const newEffective = theme === 'system' ? getSystemTheme() : theme
    setEffectiveTheme(newEffective)

    const root = document.documentElement
    root.classList.remove('light', 'dark')
    root.classList.add(newEffective)
  }, [theme])

  // Listen for system theme changes
  useEffect(() => {
    if (theme !== 'system') return

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handler = (e: MediaQueryListEvent) => {
      setEffectiveTheme(e.matches ? 'dark' : 'light')
      const root = document.documentElement
      root.classList.remove('light', 'dark')
      root.classList.add(e.matches ? 'dark' : 'light')
    }

    mediaQuery.addEventListener('change', handler)
    return () => mediaQuery.removeEventListener('change', handler)
  }, [theme])

  const setTheme = async (newTheme: Theme) => {
    setThemeState(newTheme)
    try {
      const settings = await GetAppSettings()
      settings.theme = newTheme
      await SaveAppSettings(settings)
    } catch (err) {
      console.error('Failed to save theme:', err)
    }
  }

  return (
    <ThemeContext.Provider value={{ theme, setTheme, effectiveTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}
