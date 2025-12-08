import { useState } from 'react'
import { Layout } from '@/components/layout/Layout'
import { DevicesPage } from '@/pages/DevicesPage'
import { NetworkMapPage } from '@/pages/NetworkMapPage'
import { SchemaPage } from '@/pages/SchemaPage'
import { EventsPage } from '@/pages/EventsPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { AboutPage } from '@/pages/AboutPage'
import { ThemeProvider } from '@/hooks/useTheme'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { useNotificationSound } from '@/hooks/useNotificationSound'

const pageTitles: Record<string, string> = {
  devices: 'Устройства',
  networkmap: 'Карта сети',
  schema: 'Схема',
  events: 'События',
  settings: 'Настройки',
  about: 'О программе',
}

function App() {
  const [currentPage, setCurrentPage] = useState('devices')

  // Initialize notification sounds - listens for device status events globally
  useNotificationSound()

  const renderPage = () => {
    switch (currentPage) {
      case 'devices':
        return <DevicesPage />
      case 'networkmap':
        return <NetworkMapPage />
      case 'schema':
        return <SchemaPage />
      case 'events':
        return <EventsPage />
      case 'settings':
        return <SettingsPage />
      case 'about':
        return <AboutPage />
      default:
        return <DevicesPage />
    }
  }

  return (
    <ThemeProvider>
      <Layout
        currentPage={currentPage}
        pageTitle={pageTitles[currentPage] || 'NetVisionMonitor'}
        onNavigate={setCurrentPage}
      >
        <ErrorBoundary key={currentPage}>
          {renderPage()}
        </ErrorBoundary>
      </Layout>
    </ThemeProvider>
  )
}

export default App
