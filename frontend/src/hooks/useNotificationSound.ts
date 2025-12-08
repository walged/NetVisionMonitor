import { useEffect, useRef, useCallback } from 'react'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { GetAppSettings } from '../../wailsjs/go/main/App'

interface DeviceStatusEvent {
  device_id: number
  old_status: string
  new_status: string
}

interface SoundSettings {
  enabled: boolean
  volume: number
  notifyOnOnline: boolean
  notifyOnOffline: boolean
}

export function useNotificationSound() {
  const okSoundRef = useRef<HTMLAudioElement | null>(null)
  const errorSoundRef = useRef<HTMLAudioElement | null>(null)
  const settingsRef = useRef<SoundSettings>({
    enabled: true,
    volume: 0.5,
    notifyOnOnline: true,
    notifyOnOffline: true,
  })

  // Load settings
  const loadSettings = useCallback(async () => {
    try {
      const appSettings = await GetAppSettings()
      settingsRef.current = {
        enabled: appSettings.sound_enabled,
        volume: appSettings.sound_volume,
        notifyOnOnline: appSettings.notify_on_online,
        notifyOnOffline: appSettings.notify_on_offline,
      }
    } catch (err) {
      console.error('Failed to load sound settings:', err)
    }
  }, [])

  // Initialize audio elements
  useEffect(() => {
    okSoundRef.current = new Audio('/sounds/ok.mp3')
    errorSoundRef.current = new Audio('/sounds/error.mp3')

    // Preload
    okSoundRef.current.load()
    errorSoundRef.current.load()

    // Load initial settings
    loadSettings()

    return () => {
      okSoundRef.current = null
      errorSoundRef.current = null
    }
  }, [loadSettings])

  // Listen for settings changes
  useEffect(() => {
    const handleSettingsChanged = (settings: {
      sound_enabled: boolean
      sound_volume: number
      notify_on_online: boolean
      notify_on_offline: boolean
    }) => {
      settingsRef.current = {
        enabled: settings.sound_enabled,
        volume: settings.sound_volume,
        notifyOnOnline: settings.notify_on_online,
        notifyOnOffline: settings.notify_on_offline,
      }
    }

    EventsOn('settings:changed', handleSettingsChanged)
    return () => {
      EventsOff('settings:changed')
    }
  }, [])

  // Play sound based on type
  const playSound = useCallback((type: 'online' | 'offline') => {
    const settings = settingsRef.current
    if (!settings.enabled) return

    if (type === 'online' && !settings.notifyOnOnline) return
    if (type === 'offline' && !settings.notifyOnOffline) return

    const audio = type === 'online' ? okSoundRef.current : errorSoundRef.current
    if (audio) {
      audio.volume = settings.volume
      audio.currentTime = 0
      audio.play().catch(() => {
        // Ignore autoplay errors
      })
    }
  }, [])

  // Listen for device status changes
  useEffect(() => {
    const handleStatusChange = (event: DeviceStatusEvent) => {
      if (event.new_status === 'online' && event.old_status !== 'online') {
        playSound('online')
      } else if (event.new_status === 'offline' && event.old_status !== 'offline') {
        playSound('offline')
      }
    }

    EventsOn('device:status', handleStatusChange)
    return () => {
      EventsOff('device:status')
    }
  }, [playSound])

  return { playSound, loadSettings }
}
