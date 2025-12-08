import { useEffect, useCallback, useState } from 'react'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import {
  StartMonitoring,
  StopMonitoring,
  GetMonitoringStatus,
  SetMonitoringInterval,
  RunMonitoringOnce,
} from '../../wailsjs/go/main/App'

interface MonitoringStatus {
  running: boolean
  interval: number
  workers: number
}

interface DeviceStatusEvent {
  device_id: number
  old_status: string
  new_status: string
}

interface MonitoringEvent {
  id: number
  device_id: number | null
  type: string
  level: string
  message: string
  created_at: string
}

export function useMonitoring() {
  const [status, setStatus] = useState<MonitoringStatus>({
    running: false,
    interval: 30,
    workers: 10,
  })
  const [isLoading, setIsLoading] = useState(false)

  // Load initial status
  useEffect(() => {
    GetMonitoringStatus().then(setStatus)
  }, [])

  // Listen for monitoring start/stop events
  useEffect(() => {
    const handleStarted = () => {
      setStatus((prev) => ({ ...prev, running: true }))
    }
    const handleStopped = () => {
      setStatus((prev) => ({ ...prev, running: false }))
    }

    EventsOn('monitoring:started', handleStarted)
    EventsOn('monitoring:stopped', handleStopped)

    return () => {
      EventsOff('monitoring:started')
      EventsOff('monitoring:stopped')
    }
  }, [])

  const start = useCallback(async () => {
    setIsLoading(true)
    try {
      await StartMonitoring()
    } finally {
      setIsLoading(false)
    }
  }, [])

  const stop = useCallback(async () => {
    setIsLoading(true)
    try {
      await StopMonitoring()
    } finally {
      setIsLoading(false)
    }
  }, [])

  const setInterval = useCallback(async (seconds: number) => {
    await SetMonitoringInterval(seconds)
    setStatus((prev) => ({ ...prev, interval: seconds }))
  }, [])

  const runOnce = useCallback(async () => {
    setIsLoading(true)
    try {
      await RunMonitoringOnce()
    } finally {
      setIsLoading(false)
    }
  }, [])

  return {
    status,
    isLoading,
    start,
    stop,
    setInterval,
    runOnce,
  }
}

// Hook for subscribing to device status changes
export function useDeviceStatusEvents(
  callback: (event: DeviceStatusEvent) => void
) {
  useEffect(() => {
    EventsOn('device:status', callback)
    return () => {
      EventsOff('device:status')
    }
  }, [callback])
}

// Hook for subscribing to new events
export function useMonitoringEvents(callback: (event: MonitoringEvent) => void) {
  useEffect(() => {
    EventsOn('event:new', callback)
    return () => {
      EventsOff('event:new')
    }
  }, [callback])
}
