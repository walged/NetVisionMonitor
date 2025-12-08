import { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Network, Server, Camera, RefreshCw, Zap, ZapOff, Power, Loader2, Link2 } from 'lucide-react'
import {
  GetDevices,
  GetSwitchesWithPorts,
  GetSwitchSNMPData,
  SetPoEEnabled,
  RestartPoEPort,
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

interface Device {
  id: number
  name: string
  ip_address: string
  type: string
  model: string
  status: string
}

interface SwitchPort {
  id: number
  switch_id: number
  port_number: number
  name: string
  status: string
  speed: string
  port_type: string  // "copper" or "sfp"
  linked_camera_id?: number
  linked_switch_id?: number
}

interface SwitchWithPorts {
  device_id: number
  device_name: string
  ip_address: string
  port_count: number
  sfp_port_count: number
  ports: SwitchPort[]
}

interface SNMPPoEInfo {
  port_number: number
  enabled: boolean
  active: boolean
  status: string
  power_mw: number
  power_w: number
}

interface SwitchSNMPData {
  device_id: number
  poe: SNMPPoEInfo[]
  error?: string
}

const statusColors: Record<string, string> = {
  online: 'bg-green-500',
  offline: 'bg-red-500',
  unknown: 'bg-gray-500',
  up: 'bg-green-500',
  down: 'bg-red-500',
}

export function NetworkMapPage() {
  const [switches, setSwitches] = useState<SwitchWithPorts[]>([])
  const [devices, setDevices] = useState<Device[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [snmpDataMap, setSNMPDataMap] = useState<Record<number, SwitchSNMPData>>({})
  const [togglingPoE, setTogglingPoE] = useState<string | null>(null) // "deviceId-portNumber"
  const [restartingPoE, setRestartingPoE] = useState<string | null>(null)

  const loadData = useCallback(async () => {
    setIsLoading(true)
    try {
      const [switchesData, devicesData] = await Promise.all([
        GetSwitchesWithPorts(),
        GetDevices(),
      ])
      setSwitches(switchesData || [])
      setDevices(devicesData || [])

      // Load SNMP data for all switches
      const snmpMap: Record<number, SwitchSNMPData> = {}
      for (const sw of (switchesData || [])) {
        try {
          const snmpData = await GetSwitchSNMPData(sw.device_id)
          if (snmpData) {
            snmpMap[sw.device_id] = snmpData
          }
        } catch (err) {
          console.error(`Failed to load SNMP for switch ${sw.device_id}:`, err)
        }
      }
      setSNMPDataMap(snmpMap)
    } catch (err) {
      console.error('Failed to load network map data:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const handleTogglePoE = async (deviceId: number, portNumber: number, enabled: boolean) => {
    const key = `${deviceId}-${portNumber}`
    setTogglingPoE(key)
    try {
      await SetPoEEnabled(deviceId, portNumber, enabled)
      await new Promise(resolve => setTimeout(resolve, 1500))
      // Refresh SNMP data for this switch
      const snmpData = await GetSwitchSNMPData(deviceId)
      if (snmpData) {
        setSNMPDataMap(prev => ({ ...prev, [deviceId]: snmpData }))
      }
    } catch (err) {
      console.error('Failed to toggle PoE:', err)
    } finally {
      setTogglingPoE(null)
    }
  }

  const handleRestartPoE = async (deviceId: number, portNumber: number) => {
    const key = `${deviceId}-${portNumber}`
    setRestartingPoE(key)
    try {
      await RestartPoEPort(deviceId, portNumber)
      await new Promise(resolve => setTimeout(resolve, 4000))
      // Refresh SNMP data
      const snmpData = await GetSwitchSNMPData(deviceId)
      if (snmpData) {
        setSNMPDataMap(prev => ({ ...prev, [deviceId]: snmpData }))
      }
    } catch (err) {
      console.error('Failed to restart PoE:', err)
    } finally {
      setRestartingPoE(null)
    }
  }

  const getPoEInfo = (deviceId: number, portNumber: number): SNMPPoEInfo | undefined => {
    return snmpDataMap[deviceId]?.poe?.find(p => p.port_number === portNumber)
  }

  useEffect(() => {
    loadData()
  }, [loadData])

  // Subscribe to device events
  useEffect(() => {
    // Reload data when device is deleted
    const unsubDelete = EventsOn('device:deleted', () => {
      loadData()
    })

    // Reload data when device status changes
    const unsubStatus = EventsOn('device:status', (data: any) => {
      setDevices(prev => prev.map(d =>
        d.id === data.device_id ? { ...d, status: data.new_status } : d
      ))
    })

    return () => {
      unsubDelete()
      unsubStatus()
    }
  }, [loadData])

  // Get device info by ID
  const getDeviceById = (id: number): Device | undefined => {
    return devices.find(d => d.id === id)
  }

  // Get camera linked to port
  const getCameraForPort = (linkedCameraId?: number): Device | undefined => {
    if (!linkedCameraId) return undefined
    return devices.find(d => d.id === linkedCameraId && d.type === 'camera')
  }

  // Get switch linked to SFP port
  const getLinkedSwitch = (linkedSwitchId?: number): Device | undefined => {
    if (!linkedSwitchId) return undefined
    return devices.find(d => d.id === linkedSwitchId && d.type === 'switch')
  }

  // Get servers (devices without switch ports)
  const servers = devices.filter(d => d.type === 'server')

  return (
    <TooltipProvider>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">Карта сети</h2>
            <p className="text-muted-foreground">
              Визуализация сетевой инфраструктуры
            </p>
          </div>
          <Button variant="outline" onClick={loadData} disabled={isLoading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            Обновить
          </Button>
        </div>

        {isLoading ? (
          <div className="text-center py-12 text-muted-foreground">
            Загрузка карты сети...
          </div>
        ) : (
          <div className="space-y-8">
            {/* Switches */}
            {switches.map((sw) => {
              const switchDevice = getDeviceById(sw.device_id)
              const switchStatus = switchDevice?.status || 'unknown'

              return (
                <Card key={sw.device_id}>
                  <CardHeader className="pb-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="p-2 rounded-lg bg-blue-500/10">
                          <Network className="h-6 w-6 text-blue-500" />
                        </div>
                        <div>
                          <CardTitle className="flex items-center gap-2">
                            {sw.device_name}
                            <span
                              className="inline-block h-3 w-3 rounded-full"
                              style={{ backgroundColor: switchStatus === 'online' ? '#22c55e' : switchStatus === 'offline' ? '#ef4444' : '#6b7280' }}
                            />
                          </CardTitle>
                          <p className="text-sm text-muted-foreground">
                            {sw.ip_address} • {sw.port_count} портов
                          </p>
                        </div>
                      </div>
                      <Badge variant={switchStatus === 'online' ? 'default' : 'destructive'}>
                        {switchStatus === 'online' ? 'В сети' : 'Недоступен'}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent>
                    {/* Port Grid */}
                    <div className="flex flex-wrap gap-2">
                      {sw.ports.map((port) => {
                        const isSFP = port.port_type === 'sfp'
                        const linkedCamera = getCameraForPort(port.linked_camera_id)
                        const linkedSwitch = getLinkedSwitch(port.linked_switch_id)
                        const portStatus = port.status || 'unknown'
                        const hasCamera = !!linkedCamera && !isSFP
                        const hasUplink = !!linkedSwitch && isSFP
                        const poeInfo = !isSFP ? getPoEInfo(sw.device_id, port.port_number) : undefined
                        const poeKey = `${sw.device_id}-${port.port_number}`
                        const isTogglingPoe = togglingPoE === poeKey
                        const isRestartingPoe = restartingPoE === poeKey

                        // Определяем цвета
                        const isUp = portStatus === 'up'
                        const isDown = portStatus === 'down'
                        const bgColor = isUp ? '#22c55e' : isDown ? '#ef4444' : '#6b7280'
                        const bgOpacity = isUp ? '0.15' : isDown ? '0.15' : '0.1'

                        // SFP port styling
                        const sfpBorderColor = isSFP ? '#f59e0b' : (hasCamera ? '#a855f7' : (hasUplink ? '#3b82f6' : undefined))

                        return (
                          <Tooltip key={port.id} delayDuration={100}>
                            <TooltipTrigger asChild>
                              <div
                                className={`relative w-12 h-12 cursor-pointer transition-all hover:scale-105 hover:shadow-lg flex flex-col items-center justify-center ${isSFP ? 'rounded-md' : 'rounded-lg'}`}
                                style={{
                                  backgroundColor: `rgba(${isUp ? '34,197,94' : isDown ? '239,68,68' : '107,114,128'}, ${bgOpacity})`,
                                  border: sfpBorderColor ? `2px solid ${sfpBorderColor}` : '2px solid var(--color-border)',
                                }}
                              >
                                {/* Port number */}
                                <span className={`text-sm font-bold ${isSFP ? 'text-amber-600' : ''}`}>{port.port_number}</span>
                                {/* SFP indicator */}
                                {isSFP && <span className="text-[8px] text-amber-500 font-medium">SFP</span>}
                                {/* Status indicator */}
                                {!isSFP && (
                                  <div
                                    className="w-2 h-2 rounded-full mt-0.5"
                                    style={{ backgroundColor: bgColor }}
                                  />
                                )}
                                {/* Camera indicator */}
                                {hasCamera && (
                                  <Camera className="absolute -top-1 -right-1 h-4 w-4 text-purple-500" style={{ backgroundColor: 'var(--color-card)', borderRadius: '50%', padding: '1px' }} />
                                )}
                                {/* Uplink indicator for SFP ports */}
                                {hasUplink && (
                                  <Link2 className="absolute -top-1 -right-1 h-4 w-4 text-blue-500" style={{ backgroundColor: 'var(--color-card)', borderRadius: '50%', padding: '1px' }} />
                                )}
                                {/* PoE indicator */}
                                {poeInfo?.active && (
                                  <Zap className="absolute -bottom-1 -right-1 h-3 w-3 text-yellow-500" style={{ backgroundColor: 'var(--color-card)', borderRadius: '50%', padding: '1px' }} />
                                )}
                              </div>
                            </TooltipTrigger>
                            <TooltipContent side="top" className="p-3 max-w-xs">
                              <div className="space-y-2">
                                <div className="flex items-center justify-between gap-4">
                                  <span className="font-semibold text-base">Порт {port.port_number}</span>
                                  <div className="flex items-center gap-1.5">
                                    <div
                                      className="w-2.5 h-2.5 rounded-full"
                                      style={{ backgroundColor: bgColor }}
                                    />
                                    <span className="text-sm font-medium" style={{ color: bgColor }}>
                                      {isUp ? 'Активен' : isDown ? 'Неактивен' : 'Нет данных'}
                                    </span>
                                  </div>
                                </div>
                                {port.speed && (
                                  <div className="text-sm">
                                    Скорость: <span className="font-medium">{port.speed}</span>
                                  </div>
                                )}
                                {/* PoE Info and Controls */}
                                {poeInfo && (
                                  <div className="pt-2 border-t">
                                    <div className="flex items-center gap-2 mb-2">
                                      {poeInfo.active ? (
                                        <Zap className="h-4 w-4 text-yellow-500" />
                                      ) : (
                                        <ZapOff className="h-4 w-4 text-gray-400" />
                                      )}
                                      <span className="text-sm">
                                        PoE: {poeInfo.active ? `${poeInfo.power_w.toFixed(1)} Вт` : poeInfo.enabled ? 'Вкл (нет нагрузки)' : 'Выкл'}
                                      </span>
                                    </div>
                                    <div className="flex gap-1">
                                      {isTogglingPoe || isRestartingPoe ? (
                                        <Button size="sm" variant="outline" disabled className="h-7 text-xs">
                                          <Loader2 className="h-3 w-3 animate-spin mr-1" />
                                          Применение...
                                        </Button>
                                      ) : (
                                        <>
                                          {poeInfo.enabled ? (
                                            <Button
                                              size="sm"
                                              variant="destructive"
                                              className="h-7 text-xs"
                                              onClick={(e) => {
                                                e.stopPropagation()
                                                handleTogglePoE(sw.device_id, port.port_number, false)
                                              }}
                                            >
                                              <ZapOff className="h-3 w-3 mr-1" />
                                              Выкл
                                            </Button>
                                          ) : (
                                            <Button
                                              size="sm"
                                              variant="default"
                                              className="h-7 text-xs"
                                              onClick={(e) => {
                                                e.stopPropagation()
                                                handleTogglePoE(sw.device_id, port.port_number, true)
                                              }}
                                            >
                                              <Zap className="h-3 w-3 mr-1" />
                                              Вкл
                                            </Button>
                                          )}
                                          <Button
                                            size="sm"
                                            variant="outline"
                                            className="h-7 text-xs"
                                            disabled={!poeInfo.enabled}
                                            onClick={(e) => {
                                              e.stopPropagation()
                                              handleRestartPoE(sw.device_id, port.port_number)
                                            }}
                                          >
                                            <Power className="h-3 w-3 mr-1" />
                                            Restart
                                          </Button>
                                        </>
                                      )}
                                    </div>
                                  </div>
                                )}
                                {linkedCamera && (
                                  <div className="pt-2 border-t">
                                    <div className="flex items-center gap-2 text-purple-500 mb-1">
                                      <Camera className="h-4 w-4" />
                                      <span className="font-medium">Подключена камера</span>
                                    </div>
                                    <div className="pl-6 space-y-0.5">
                                      <div className="font-medium">{linkedCamera.name}</div>
                                      <div className="text-sm text-muted-foreground">{linkedCamera.ip_address}</div>
                                      <div className="flex items-center gap-1.5">
                                        <div
                                          className="w-2 h-2 rounded-full"
                                          style={{ backgroundColor: linkedCamera.status === 'online' ? '#22c55e' : '#ef4444' }}
                                        />
                                        <span className="text-sm" style={{ color: linkedCamera.status === 'online' ? '#22c55e' : '#ef4444' }}>
                                          {linkedCamera.status === 'online' ? 'В сети' : 'Недоступна'}
                                        </span>
                                      </div>
                                    </div>
                                  </div>
                                )}
                                {/* Linked switch info for SFP ports */}
                                {linkedSwitch && (
                                  <div className="pt-2 border-t">
                                    <div className="flex items-center gap-2 text-blue-500 mb-1">
                                      <Link2 className="h-4 w-4" />
                                      <span className="font-medium">Uplink на коммутатор</span>
                                    </div>
                                    <div className="pl-6 space-y-0.5">
                                      <div className="font-medium">{linkedSwitch.name}</div>
                                      <div className="text-sm text-muted-foreground">{linkedSwitch.ip_address}</div>
                                      <div className="flex items-center gap-1.5">
                                        <div
                                          className="w-2 h-2 rounded-full"
                                          style={{ backgroundColor: linkedSwitch.status === 'online' ? '#22c55e' : '#ef4444' }}
                                        />
                                        <span className="text-sm" style={{ color: linkedSwitch.status === 'online' ? '#22c55e' : '#ef4444' }}>
                                          {linkedSwitch.status === 'online' ? 'В сети' : 'Недоступен'}
                                        </span>
                                      </div>
                                    </div>
                                  </div>
                                )}
                                {/* SFP port type indicator */}
                                {isSFP && !linkedSwitch && (
                                  <div className="pt-2 border-t">
                                    <div className="flex items-center gap-2 text-amber-500">
                                      <Network className="h-4 w-4" />
                                      <span className="text-sm">SFP порт (оптика)</span>
                                    </div>
                                    <div className="text-xs text-muted-foreground mt-1">
                                      Можно привязать коммутатор для uplink
                                    </div>
                                  </div>
                                )}
                                {!linkedCamera && !linkedSwitch && !isSFP && !port.speed && !isUp && !isDown && !poeInfo && (
                                  <div className="text-sm text-muted-foreground">
                                    Порт не опрошен
                                  </div>
                                )}
                              </div>
                            </TooltipContent>
                          </Tooltip>
                        )
                      })}
                    </div>

                    {/* Legend */}
                    <div className="flex flex-wrap items-center gap-x-6 gap-y-2 mt-4 pt-4 border-t text-sm">
                      <div className="flex items-center gap-2">
                        <div className="w-4 h-4 rounded" style={{ backgroundColor: 'rgba(34,197,94,0.15)', border: '2px solid var(--color-border)' }}>
                          <div className="w-2 h-2 rounded-full mx-auto mt-0.5" style={{ backgroundColor: '#22c55e' }} />
                        </div>
                        <span>Активен</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-4 h-4 rounded" style={{ backgroundColor: 'rgba(239,68,68,0.15)', border: '2px solid var(--color-border)' }}>
                          <div className="w-2 h-2 rounded-full mx-auto mt-0.5" style={{ backgroundColor: '#ef4444' }} />
                        </div>
                        <span>Неактивен</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-4 h-4 rounded" style={{ backgroundColor: 'rgba(107,114,128,0.1)', border: '2px solid var(--color-border)' }}>
                          <div className="w-2 h-2 rounded-full mx-auto mt-0.5" style={{ backgroundColor: '#6b7280' }} />
                        </div>
                        <span>Нет данных</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-4 h-4 rounded" style={{ border: '2px solid #a855f7' }} />
                        <span>Камера</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-4 h-4 rounded-md" style={{ border: '2px solid #f59e0b' }}>
                          <span className="text-[6px] text-amber-500 block text-center leading-3">SFP</span>
                        </div>
                        <span>SFP порт</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="w-4 h-4 rounded-md" style={{ border: '2px solid #3b82f6' }} />
                        <span>Uplink</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })}

            {/* Servers */}
            {servers.length > 0 && (
              <Card>
                <CardHeader className="pb-3">
                  <div className="flex items-center gap-3">
                    <div className="p-2 rounded-lg bg-green-500/10">
                      <Server className="h-6 w-6 text-green-500" />
                    </div>
                    <div>
                      <CardTitle>Серверы</CardTitle>
                      <p className="text-sm text-muted-foreground">
                        {servers.length} серверов в сети
                      </p>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3">
                    {servers.map((server) => (
                      <div
                        key={server.id}
                        className={`
                          p-3 rounded-lg border
                          ${server.status === 'online' ? 'bg-green-500/5 border-green-500/30' : 'bg-red-500/5 border-red-500/30'}
                        `}
                      >
                        <div className="flex items-center gap-2">
                          <Server className={`h-4 w-4 ${server.status === 'online' ? 'text-green-500' : 'text-red-500'}`} />
                          <span className="font-medium text-sm truncate">{server.name}</span>
                        </div>
                        <p className="text-xs text-muted-foreground mt-1 font-mono">
                          {server.ip_address}
                        </p>
                        {server.model && (
                          <p className="text-xs text-muted-foreground truncate">
                            {server.model}
                          </p>
                        )}
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}

            {switches.length === 0 && servers.length === 0 && (
              <div className="text-center py-12 text-muted-foreground">
                <Network className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>Нет устройств для отображения</p>
                <p className="text-sm">Добавьте коммутаторы или серверы на странице "Устройства"</p>
              </div>
            )}
          </div>
        )}
      </div>
    </TooltipProvider>
  )
}
