import { useState, useEffect, useCallback } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  Activity,
  ArrowLeft,
  Clock,
  Network,
  Server,
  Camera,
  Monitor,
  Wifi,
  WifiOff,
  TrendingUp,
  TrendingDown,
  RefreshCw,
  Zap,
  ZapOff,
  Power,
  ArrowDownToLine,
  ArrowUpFromLine,
  AlertCircle,
  CheckCircle2,
  Loader2,
  Link2,
  Play,
  Square,
} from 'lucide-react'
import {
  GetDevice,
  GetDeviceMonitoringStats,
  GetDeviceLatencyHistory,
  GetDeviceEvents,
  GetSwitchPorts,
  GetSwitchSNMPData,
  RestartPoEPort,
  SetPoEEnabled,
  SetPortEnabled,
  RestartPort,
  RefreshCameraStreams,
  GetCameraStreamURL,
  FetchCameraSnapshotBase64,
} from '../../../wailsjs/go/main/App'

interface DeviceStats {
  device_id: number
  total_checks: number
  online_count: number
  offline_count: number
  uptime_percent: number
  avg_latency: number
  min_latency: number
  max_latency: number
  last_online?: string
  last_offline?: string
  current_streak: number
  streak_status: string
}

interface LatencyPoint {
  timestamp: string
  latency: number
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

interface SNMPPortInfo {
  port_number: number
  status: string
  speed: number
  speed_str: string
  rx_bytes: number
  tx_bytes: number
  description: string
}

interface SNMPPoEInfo {
  port_number: number
  enabled: boolean
  active: boolean
  status: string
  power_mw: number
  power_w: number
}

interface SNMPSystemInfo {
  firmware_version: string
  ups?: {
    present: boolean
    status: string
    charge: number
  }
}

interface SwitchSNMPData {
  device_id: number
  system_info?: SNMPSystemInfo
  ports: SNMPPortInfo[]
  poe: SNMPPoEInfo[]
  error?: string
}

interface DeviceDetailsPanelProps {
  deviceId: number
  onBack: () => void
}

export function DeviceDetailsPanel({ deviceId, onBack }: DeviceDetailsPanelProps) {
  const [device, setDevice] = useState<any>(null)
  const [stats, setStats] = useState<DeviceStats | null>(null)
  const [latencyPoints, setLatencyPoints] = useState<LatencyPoint[]>([])
  const [events, setEvents] = useState<any[]>([])
  const [ports, setPorts] = useState<SwitchPort[]>([])
  const [snmpData, setSNMPData] = useState<SwitchSNMPData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isSNMPLoading, setIsSNMPLoading] = useState(false)
  const [restartingPort, setRestartingPort] = useState<number | null>(null)
  const [activeTab, setActiveTab] = useState<'overview' | 'ports' | 'events' | 'preview'>('overview')
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState<string | null>(null)

  const loadData = useCallback(async () => {
    setIsLoading(true)
    try {
      const [deviceData, statsData, latencyData, eventsData] = await Promise.all([
        GetDevice(deviceId),
        GetDeviceMonitoringStats(deviceId),
        GetDeviceLatencyHistory(deviceId, 24),
        GetDeviceEvents(deviceId, 50),
      ])

      setDevice(deviceData)
      setStats(statsData)
      setLatencyPoints(latencyData || [])
      setEvents(eventsData || [])

      // Load ports for switches
      if (deviceData?.type === 'switch') {
        const portsData = await GetSwitchPorts(deviceId)
        setPorts(portsData || [])
      }
    } catch (err) {
      console.error('Failed to load device details:', err)
    } finally {
      setIsLoading(false)
    }
  }, [deviceId])

  const loadSNMPData = useCallback(async () => {
    if (!device || device.type !== 'switch') return

    setIsSNMPLoading(true)
    try {
      const data = await GetSwitchSNMPData(deviceId)
      setSNMPData(data)
    } catch (err) {
      console.error('Failed to load SNMP data:', err)
    } finally {
      setIsSNMPLoading(false)
    }
  }, [deviceId, device])

  useEffect(() => {
    loadData()
  }, [loadData])

  useEffect(() => {
    if (device?.type === 'switch' && activeTab === 'ports') {
      loadSNMPData()
    }
  }, [device, activeTab, loadSNMPData])

  const handleRestartPoE = async (portNumber: number) => {
    setRestartingPort(portNumber)
    try {
      await RestartPoEPort(deviceId, portNumber)
      // Reload SNMP data after restart
      setTimeout(() => loadSNMPData(), 4000)
    } catch (err) {
      console.error('Failed to restart PoE:', err)
    } finally {
      setRestartingPort(null)
    }
  }

  const [togglingPoE, setTogglingPoE] = useState<number | null>(null)
  const [togglingPort, setTogglingPort] = useState<number | null>(null)

  const handleTogglePoE = async (portNumber: number, enabled: boolean) => {
    setTogglingPoE(portNumber)
    try {
      console.log(`Setting PoE port ${portNumber} to ${enabled}`)
      await SetPoEEnabled(deviceId, portNumber, enabled)
      // Wait for the switch to process the command
      await new Promise(resolve => setTimeout(resolve, 1500))
      // Force refresh SNMP data to get new state
      const data = await GetSwitchSNMPData(deviceId)
      setSNMPData(data)
    } catch (err) {
      console.error('Failed to toggle PoE:', err)
    } finally {
      setTogglingPoE(null)
    }
  }

  const handleTogglePort = async (portNumber: number, enabled: boolean) => {
    setTogglingPort(portNumber)
    try {
      console.log(`Setting port ${portNumber} to ${enabled}`)
      await SetPortEnabled(deviceId, portNumber, enabled)
      // Wait for the switch to process the command
      await new Promise(resolve => setTimeout(resolve, 1500))
      // Force refresh SNMP data to get new state
      const data = await GetSwitchSNMPData(deviceId)
      setSNMPData(data)
    } catch (err) {
      console.error('Failed to toggle port:', err)
    } finally {
      setTogglingPort(null)
    }
  }

  const handleRestartPort = async (portNumber: number) => {
    setTogglingPort(portNumber)
    try {
      await RestartPort(deviceId, portNumber)
      // Reload SNMP data after restart
      setTimeout(() => loadSNMPData(), 4000)
    } catch (err) {
      console.error('Failed to restart port:', err)
    } finally {
      setTogglingPort(null)
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'online':
        return <Badge variant="success">В сети</Badge>
      case 'offline':
        return <Badge variant="destructive">Не в сети</Badge>
      default:
        return <Badge variant="secondary">Неизвестно</Badge>
    }
  }

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'switch':
        return <Network className="h-5 w-5" />
      case 'server':
        return <Server className="h-5 w-5" />
      case 'camera':
        return <Camera className="h-5 w-5" />
      default:
        return <Monitor className="h-5 w-5" />
    }
  }

  const getTypeName = (type: string) => {
    switch (type) {
      case 'switch':
        return 'Коммутатор'
      case 'server':
        return 'Сервер'
      case 'camera':
        return 'Камера'
      default:
        return 'Устройство'
    }
  }

  const formatLatency = (ms: number) => {
    if (ms < 1) return '<1 мс'
    if (ms < 1000) return `${Math.round(ms)} мс`
    return `${(ms / 1000).toFixed(2)} с`
  }

  const formatUptime = (percent: number) => {
    return `${percent.toFixed(2)}%`
  }

  const formatDate = (dateStr: string | undefined) => {
    if (!dateStr) return 'Н/Д'
    try {
      return new Date(dateStr).toLocaleString('ru-RU')
    } catch {
      return dateStr
    }
  }

  const formatBytes = (bytes: number) => {
    if (bytes < 1024) return `${bytes} Б`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} МБ`
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} ГБ`
  }

  const getSNMPPortInfo = (portNumber: number): SNMPPortInfo | undefined => {
    return snmpData?.ports?.find(p => p.port_number === portNumber)
  }

  const getSNMPPoEInfo = (portNumber: number): SNMPPoEInfo | undefined => {
    return snmpData?.poe?.find(p => p.port_number === portNumber)
  }

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    )
  }

  if (!device) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <p className="text-muted-foreground">Устройство не найдено</p>
          <Button variant="outline" className="mt-4" onClick={onBack}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Назад
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Button variant="ghost" size="icon" onClick={onBack}>
                <ArrowLeft className="h-5 w-5" />
              </Button>
              <div className="flex items-center gap-3">
                {getTypeIcon(device.type)}
                <div>
                  <CardTitle className="text-xl">{device.name}</CardTitle>
                  <CardDescription>
                    {getTypeName(device.type)} • {device.ip_address}
                    {device.model && ` • ${device.model}`}
                    {snmpData?.system_info?.firmware_version && (
                      <> • FW: {snmpData.system_info.firmware_version}</>
                    )}
                  </CardDescription>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-4">
              {snmpData?.system_info?.ups && (
                <Badge variant={snmpData.system_info.ups.status === 'full' ? 'success' : 'warning'}>
                  <Zap className="h-3 w-3 mr-1" />
                  UPS: {snmpData.system_info.ups.charge}%
                </Badge>
              )}
              {getStatusBadge(device.status)}
              <Button variant="outline" size="sm" onClick={loadData}>
                <RefreshCw className="h-4 w-4 mr-2" />
                Обновить
              </Button>
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Tabs */}
      <div className="flex gap-2">
        <Button
          variant={activeTab === 'overview' ? 'default' : 'outline'}
          onClick={() => setActiveTab('overview')}
        >
          <Activity className="h-4 w-4 mr-2" />
          Обзор
        </Button>
        {device.type === 'switch' && (
          <Button
            variant={activeTab === 'ports' ? 'default' : 'outline'}
            onClick={() => setActiveTab('ports')}
          >
            <Network className="h-4 w-4 mr-2" />
            Порты ({ports.length})
          </Button>
        )}
        {device.type === 'camera' && (
          <Button
            variant={activeTab === 'preview' ? 'default' : 'outline'}
            onClick={() => setActiveTab('preview')}
          >
            <Monitor className="h-4 w-4 mr-2" />
            Превью
          </Button>
        )}
        <Button
          variant={activeTab === 'events' ? 'default' : 'outline'}
          onClick={() => setActiveTab('events')}
        >
          <Clock className="h-4 w-4 mr-2" />
          События ({events.length})
        </Button>
      </div>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {/* Uptime Card */}
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Uptime</CardDescription>
              <CardTitle className="text-2xl">
                {stats ? formatUptime(stats.uptime_percent) : 'Н/Д'}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-xs text-muted-foreground">
                {stats && (
                  <>
                    {stats.online_count} онлайн / {stats.offline_count} оффлайн
                    <br />
                    из {stats.total_checks} проверок
                  </>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Latency Card */}
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Средняя задержка</CardDescription>
              <CardTitle className="text-2xl">
                {stats ? formatLatency(stats.avg_latency) : 'Н/Д'}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-xs text-muted-foreground">
                {stats && (
                  <>
                    мин: {formatLatency(stats.min_latency)}
                    <br />
                    макс: {formatLatency(stats.max_latency)}
                  </>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Current Streak Card */}
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Текущая серия</CardDescription>
              <CardTitle className="text-2xl flex items-center gap-2">
                {stats?.streak_status === 'online' ? (
                  <TrendingUp className="h-5 w-5 text-green-500" />
                ) : (
                  <TrendingDown className="h-5 w-5 text-red-500" />
                )}
                {stats?.current_streak || 0} проверок
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-xs text-muted-foreground">
                Статус: {stats?.streak_status === 'online' ? 'онлайн' : 'оффлайн'}
              </div>
            </CardContent>
          </Card>

          {/* Last Status Change Card */}
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Последние изменения</CardDescription>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="flex items-center gap-2 text-xs">
                <Wifi className="h-3 w-3 text-green-500" />
                <span className="text-muted-foreground">Онлайн:</span>
                <span>{formatDate(stats?.last_online)}</span>
              </div>
              <div className="flex items-center gap-2 text-xs">
                <WifiOff className="h-3 w-3 text-red-500" />
                <span className="text-muted-foreground">Оффлайн:</span>
                <span>{formatDate(stats?.last_offline)}</span>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Latency Graph (simplified) */}
      {activeTab === 'overview' && latencyPoints.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>История задержки (последние 24ч)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-32 flex items-end gap-px">
              {latencyPoints.slice(-100).map((point, i) => {
                const maxLatency = Math.max(...latencyPoints.map(p => p.latency), 1)
                const height = (point.latency / maxLatency) * 100
                return (
                  <div
                    key={i}
                    className={`flex-1 min-w-[2px] rounded-t ${
                      point.status === 'online' ? 'bg-green-500' : 'bg-red-500'
                    }`}
                    style={{ height: `${Math.max(height, 2)}%` }}
                    title={`${formatLatency(point.latency)} - ${point.status}`}
                  />
                )
              })}
            </div>
            <div className="flex justify-between text-xs text-muted-foreground mt-2">
              <span>24ч назад</span>
              <span>Сейчас</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Ports Tab */}
      {activeTab === 'ports' && device.type === 'switch' && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Порты коммутатора</CardTitle>
                <CardDescription>
                  Всего {ports.length} портов • {ports.filter(p => p.linked_camera_id).length} камер подключено
                  {snmpData?.error && (
                    <span className="text-destructive ml-2">• SNMP: {snmpData.error}</span>
                  )}
                </CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={loadSNMPData}
                disabled={isSNMPLoading}
              >
                {isSNMPLoading ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4 mr-2" />
                )}
                Обновить SNMP
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            {/* System Info Section */}
            {snmpData && !snmpData.error && snmpData.system_info && (
              <div className="mb-6 p-4 bg-muted/50 rounded-lg">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div className="flex items-center gap-2">
                    {isSNMPLoading ? (
                      <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                    ) : (
                      <CheckCircle2 className="h-4 w-4 text-green-500" />
                    )}
                    <span className="text-muted-foreground">SNMP:</span>
                    <span className="font-medium">OK</span>
                  </div>
                  {snmpData.system_info.firmware_version && (
                    <div>
                      <span className="text-muted-foreground">Прошивка: </span>
                      <span className="font-medium">{snmpData.system_info.firmware_version}</span>
                    </div>
                  )}
                  {snmpData.system_info.ups && (
                    <>
                      <div className="flex items-center gap-2">
                        <Zap className="h-4 w-4 text-yellow-500" />
                        <span className="text-muted-foreground">UPS:</span>
                        <span className="font-medium">
                          {snmpData.system_info.ups.status === 'charging' ? 'Заряжается' :
                           snmpData.system_info.ups.status === 'discharging' ? 'Разряжается' :
                           snmpData.system_info.ups.status === 'full' ? 'Заряжен' : 'N/A'}
                        </span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Заряд: </span>
                        <span className="font-medium">{snmpData.system_info.ups.charge}%</span>
                      </div>
                    </>
                  )}
                  <div>
                    <span className="text-muted-foreground">PoE всего: </span>
                    <span className="font-medium">
                      {snmpData.poe ? snmpData.poe.reduce((sum, p) => sum + p.power_w, 0).toFixed(1) : 0} Вт
                    </span>
                  </div>
                </div>
              </div>
            )}

            {ports.length === 0 ? (
              <p className="text-muted-foreground text-center py-8">
                Порты не настроены
              </p>
            ) : (
              <>
                {/* Port Grid View */}
                <TooltipProvider>
                  <div className="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-12 gap-2 mb-6">
                    {ports.map((port) => {
                      const snmpPort = getSNMPPortInfo(port.port_number)
                      const snmpPoe = getSNMPPoEInfo(port.port_number)
                      const actualStatus = snmpPort?.status || port.status
                      const isSFP = port.port_type === 'sfp'
                      const hasUplink = !!port.linked_switch_id && isSFP

                      return (
                        <Tooltip key={port.id}>
                          <TooltipTrigger asChild>
                            <div
                              className={`p-2 border text-center cursor-pointer transition-colors relative ${
                                isSFP ? 'rounded-md' : 'rounded-lg'
                              } ${
                                hasUplink
                                  ? 'bg-blue-500/10 border-blue-500/50 hover:bg-blue-500/20'
                                  : isSFP
                                  ? 'bg-amber-500/10 border-amber-500/50 hover:bg-amber-500/20'
                                  : port.linked_camera_id
                                  ? 'bg-purple-500/10 border-purple-500/50 hover:bg-purple-500/20'
                                  : actualStatus === 'up'
                                  ? 'bg-green-500/10 border-green-500/50 hover:bg-green-500/20'
                                  : actualStatus === 'down'
                                  ? 'bg-red-500/10 border-red-500/50 hover:bg-red-500/20'
                                  : 'bg-muted hover:bg-muted/80'
                              }`}
                            >
                              <div className={`text-xs font-medium ${isSFP ? 'text-amber-600' : ''}`}>{port.port_number}</div>
                              {isSFP ? (
                                <>
                                  <span className="text-[8px] text-amber-500 block">SFP</span>
                                  {hasUplink && (
                                    <Link2 className="w-3 h-3 mx-auto text-blue-500" />
                                  )}
                                  {!hasUplink && (
                                    <div className={`w-2 h-2 rounded-full mx-auto mt-0.5 ${
                                      actualStatus === 'up' ? 'bg-green-500' :
                                      actualStatus === 'down' ? 'bg-red-500' : 'bg-gray-400'
                                    }`} />
                                  )}
                                </>
                              ) : port.linked_camera_id ? (
                                <Camera className="w-3 h-3 mx-auto mt-1 text-purple-500" />
                              ) : snmpPoe?.active ? (
                                <Zap className="w-3 h-3 mx-auto mt-1 text-yellow-500" />
                              ) : (
                                <div className={`w-2 h-2 rounded-full mx-auto mt-1 ${
                                  actualStatus === 'up' ? 'bg-green-500' :
                                  actualStatus === 'down' ? 'bg-red-500' : 'bg-gray-400'
                                }`} />
                              )}
                            </div>
                          </TooltipTrigger>
                          <TooltipContent>
                            <div className="text-xs space-y-1">
                              <div className="font-medium">{port.name} {isSFP && '(SFP)'}</div>
                              <div>Статус: {actualStatus === 'up' ? 'Активен' : actualStatus === 'down' ? 'Неактивен' : 'Неизвестно'}</div>
                              {snmpPort && (
                                <>
                                  <div>Скорость: {snmpPort.speed_str || 'N/A'}</div>
                                  <div>RX: {formatBytes(snmpPort.rx_bytes)}</div>
                                  <div>TX: {formatBytes(snmpPort.tx_bytes)}</div>
                                </>
                              )}
                              {!isSFP && snmpPoe && (
                                <div>
                                  PoE: {snmpPoe.active ? `${snmpPoe.power_w.toFixed(1)} Вт` : 'Выкл'}
                                </div>
                              )}
                              {port.linked_camera_id && !isSFP && <div>Камера подключена</div>}
                              {hasUplink && <div className="text-blue-500">Uplink подключен</div>}
                              {isSFP && !hasUplink && <div className="text-amber-500">SFP порт (оптика)</div>}
                            </div>
                          </TooltipContent>
                        </Tooltip>
                      )
                    })}
                  </div>
                </TooltipProvider>

                {/* Detailed Port Table */}
                {snmpData && !snmpData.error && (
                  <>
                    <Separator className="my-4" />
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-16">Порт</TableHead>
                          <TableHead>Статус</TableHead>
                          <TableHead>Скорость</TableHead>
                          <TableHead>
                            <div className="flex items-center gap-1">
                              <ArrowDownToLine className="h-3 w-3" />
                              RX
                            </div>
                          </TableHead>
                          <TableHead>
                            <div className="flex items-center gap-1">
                              <ArrowUpFromLine className="h-3 w-3" />
                              TX
                            </div>
                          </TableHead>
                          <TableHead>PoE</TableHead>
                          <TableHead>Мощность</TableHead>
                          <TableHead>Устройство</TableHead>
                          <TableHead>Действия</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {ports.map((port) => {
                          const snmpPort = getSNMPPortInfo(port.port_number)
                          const snmpPoe = getSNMPPoEInfo(port.port_number)
                          const isRestarting = restartingPort === port.port_number

                          return (
                            <TableRow key={port.id}>
                              <TableCell className="font-medium">{port.port_number}</TableCell>
                              <TableCell>
                                {snmpPort?.status === 'up' ? (
                                  <Badge variant="success">Активен</Badge>
                                ) : snmpPort?.status === 'down' ? (
                                  <Badge variant="destructive">Неактивен</Badge>
                                ) : (
                                  <Badge variant="secondary">N/A</Badge>
                                )}
                              </TableCell>
                              <TableCell>{snmpPort?.speed_str || '-'}</TableCell>
                              <TableCell>{snmpPort ? formatBytes(snmpPort.rx_bytes) : '-'}</TableCell>
                              <TableCell>{snmpPort ? formatBytes(snmpPort.tx_bytes) : '-'}</TableCell>
                              <TableCell>
                                {port.port_type === 'sfp' ? (
                                  <span className="text-muted-foreground">—</span>
                                ) : snmpPoe?.active ? (
                                  <Zap className="h-4 w-4 text-yellow-500" />
                                ) : snmpPoe?.enabled ? (
                                  <ZapOff className="h-4 w-4 text-gray-400" />
                                ) : (
                                  <span className="text-muted-foreground">-</span>
                                )}
                              </TableCell>
                              <TableCell>
                                {port.port_type === 'sfp' ? '—' : snmpPoe?.active ? `${snmpPoe.power_w.toFixed(1)} Вт` : '-'}
                              </TableCell>
                              <TableCell>
                                {port.port_type === 'sfp' ? (
                                  port.linked_switch_id ? (
                                    <Badge variant="outline" className="border-blue-500 text-blue-500">
                                      <Link2 className="h-3 w-3 mr-1" />
                                      Uplink
                                    </Badge>
                                  ) : (
                                    <Badge variant="outline" className="border-amber-500 text-amber-500">
                                      SFP
                                    </Badge>
                                  )
                                ) : port.linked_camera_id ? (
                                  <Badge variant="outline">
                                    <Camera className="h-3 w-3 mr-1" />
                                    Камера
                                  </Badge>
                                ) : (
                                  '-'
                                )}
                              </TableCell>
                              <TableCell>
                                <div className="flex gap-2">
                                  {/* 1. PoE restart - only for copper ports with PoE */}
                                  {snmpPoe && port.port_type !== 'sfp' && (
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      onClick={() => handleRestartPoE(port.port_number)}
                                      disabled={restartingPort === port.port_number}
                                      className="text-orange-600 hover:text-orange-700 hover:bg-orange-50"
                                    >
                                      {restartingPort === port.port_number ? (
                                        <Loader2 className="h-3 w-3 mr-1 animate-spin" />
                                      ) : (
                                        <Power className="h-3 w-3 mr-1" />
                                      )}
                                      {restartingPort === port.port_number ? '...' : 'PoE'}
                                    </Button>
                                  )}
                                  {/* 2. Port on/off - for ALL ports */}
                                  {snmpPort && (
                                    <>
                                      {togglingPort === port.port_number ? (
                                        <Button variant="outline" size="sm" disabled>
                                          <Loader2 className="h-3 w-3 mr-1 animate-spin" />
                                          ...
                                        </Button>
                                      ) : snmpPort?.status === 'up' ? (
                                        <Button
                                          variant="outline"
                                          size="sm"
                                          onClick={() => handleTogglePort(port.port_number, false)}
                                          className="text-red-600 hover:text-red-700 hover:bg-red-50"
                                        >
                                          <ZapOff className="h-3 w-3 mr-1" />
                                          Откл
                                        </Button>
                                      ) : (
                                        <Button
                                          variant="outline"
                                          size="sm"
                                          onClick={() => handleTogglePort(port.port_number, true)}
                                          className="text-green-600 hover:text-green-700 hover:bg-green-50"
                                        >
                                          <Zap className="h-3 w-3 mr-1" />
                                          Вкл
                                        </Button>
                                      )}
                                    </>
                                  )}
                                </div>
                              </TableCell>
                            </TableRow>
                          )
                        })}
                      </TableBody>
                    </Table>
                  </>
                )}

                {/* Legend */}
                <Separator className="my-4" />
                <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded bg-green-500" />
                    <span>Порт активен</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded bg-red-500" />
                    <span>Порт неактивен</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Zap className="w-3 h-3 text-yellow-500" />
                    <span>PoE активен</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Camera className="w-3 h-3 text-purple-500" />
                    <span>Камера подключена</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-sm border-2 border-amber-500" />
                    <span>SFP порт</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Link2 className="w-3 h-3 text-blue-500" />
                    <span>Uplink</span>
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      )}

      {/* Events Tab */}
      {activeTab === 'events' && (
        <Card>
          <CardHeader>
            <CardTitle>События устройства</CardTitle>
          </CardHeader>
          <CardContent>
            {events.length === 0 ? (
              <p className="text-muted-foreground text-center py-8">
                Нет событий
              </p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Время</TableHead>
                    <TableHead>Тип</TableHead>
                    <TableHead>Уровень</TableHead>
                    <TableHead>Сообщение</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {events.slice(0, 20).map((event) => (
                    <TableRow key={event.id}>
                      <TableCell className="text-xs">
                        {formatDate(event.created_at)}
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">{event.type}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={
                            event.level === 'error'
                              ? 'destructive'
                              : event.level === 'warn'
                              ? 'warning'
                              : 'secondary'
                          }
                        >
                          {event.level}
                        </Badge>
                      </TableCell>
                      <TableCell className="max-w-md truncate">
                        {event.message}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}

      {/* Preview Tab for Cameras */}
      {activeTab === 'preview' && device.type === 'camera' && (
        <CameraPreview
          device={device}
          deviceId={deviceId}
          onDeviceUpdate={loadData}
        />
      )}
    </div>
  )
}

// Camera Preview component with ONVIF auto-discovery
interface CameraPreviewProps {
  device: any
  deviceId: number
  onDeviceUpdate: () => void
}

function CameraPreview({ device, deviceId, onDeviceUpdate }: CameraPreviewProps) {
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [rtspUrl, setRtspUrl] = useState<string | null>(null)
  const [previewLoading, setPreviewLoading] = useState(false)
  const [previewError, setPreviewError] = useState<string | null>(null)
  const [isDiscovering, setIsDiscovering] = useState(false)
  const [autoRefresh, setAutoRefresh] = useState(false)
  const [refreshInterval, setRefreshIntervalState] = useState<ReturnType<typeof setInterval> | null>(null)

  // Set RTSP URL from device data
  useEffect(() => {
    if (device.camera?.rtsp_url) {
      setRtspUrl(device.camera.rtsp_url)
    }
  }, [deviceId, device.camera?.rtsp_url])

  // Auto-refresh effect
  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(async () => {
        try {
          const base64Data = await FetchCameraSnapshotBase64(deviceId)
          if (base64Data) {
            setPreviewUrl(base64Data)
            setPreviewError(null)
          }
        } catch (err) {
          // Silent fail during auto-refresh
        }
      }, 83) // ~12 fps
      setRefreshIntervalState(interval)
      return () => clearInterval(interval)
    } else {
      if (refreshInterval) {
        clearInterval(refreshInterval)
        setRefreshIntervalState(null)
      }
    }
  }, [autoRefresh, deviceId])

  const handleDiscover = async () => {
    setIsDiscovering(true)
    setPreviewError(null)
    setPreviewUrl(null)
    try {
      // Refresh camera streams via ONVIF
      await RefreshCameraStreams(deviceId)
      // Reload device data
      onDeviceUpdate()
      // Fetch snapshot via proxy (Go backend fetches and returns base64)
      const base64Data = await FetchCameraSnapshotBase64(deviceId)
      if (base64Data) {
        setPreviewUrl(base64Data)
      }
      // Get RTSP URL
      const streamUrl = await GetCameraStreamURL(deviceId)
      if (streamUrl) {
        setRtspUrl(streamUrl)
      }
    } catch (err: any) {
      setPreviewError(err?.message || 'Ошибка ONVIF discovery')
    } finally {
      setIsDiscovering(false)
    }
  }

  const loadSnapshot = async () => {
    setPreviewLoading(true)
    setPreviewError(null)
    try {
      // Fetch snapshot via Go proxy - returns base64 data URI
      const base64Data = await FetchCameraSnapshotBase64(deviceId)
      if (base64Data) {
        setPreviewUrl(base64Data)
      } else {
        setPreviewError('Не удалось получить снимок')
      }
    } catch (err: any) {
      setPreviewError(err?.message || 'Ошибка получения снимка')
    } finally {
      setPreviewLoading(false)
    }
  }

  const handleRefresh = async () => {
    await loadSnapshot()
  }

  const toggleAutoRefresh = () => {
    if (!autoRefresh && !previewUrl) {
      // Start with one snapshot first
      loadSnapshot()
    }
    setAutoRefresh(!autoRefresh)
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>Превью камеры</CardTitle>
            <CardDescription>
              ONVIF порт: {device.camera?.onvif_port || 80}
              {device.camera?.snapshot_url && ' • Snapshot настроен'}
              {autoRefresh && ' • Live режим'}
            </CardDescription>
          </div>
          <div className="flex gap-2 flex-wrap">
            <Button
              variant={autoRefresh ? 'default' : 'outline'}
              onClick={toggleAutoRefresh}
              disabled={isDiscovering}
            >
              {autoRefresh ? (
                <Square className="h-4 w-4 mr-2" />
              ) : (
                <Play className="h-4 w-4 mr-2" />
              )}
              {autoRefresh ? 'Стоп' : 'Live'}
            </Button>
            <Button
              variant="outline"
              onClick={handleRefresh}
              disabled={previewLoading || isDiscovering || autoRefresh}
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${previewLoading ? 'animate-spin' : ''}`} />
              Снимок
            </Button>
            <Button
              variant="outline"
              onClick={handleDiscover}
              disabled={isDiscovering || previewLoading}
            >
              {isDiscovering ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4 mr-2" />
              )}
              ONVIF
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* Preview area */}
          <div className="relative bg-black rounded-lg overflow-hidden aspect-video flex items-center justify-center">
            {(previewLoading || isDiscovering) && (
              <div className="absolute inset-0 flex items-center justify-center bg-black/50 z-10">
                <div className="text-center text-white">
                  <Loader2 className="h-8 w-8 animate-spin mx-auto mb-2" />
                  <p className="text-sm">{isDiscovering ? 'Поиск камеры через ONVIF...' : 'Загрузка снимка...'}</p>
                </div>
              </div>
            )}
            {previewError ? (
              <div className="text-center text-red-400 p-4">
                <AlertCircle className="h-12 w-12 mx-auto mb-2" />
                <p>Ошибка</p>
                <p className="text-xs mt-1 max-w-md whitespace-pre-wrap">{previewError}</p>
                <p className="text-xs mt-3 text-muted-foreground">
                  Отредактируйте камеру и укажите Snapshot URL, например:
                </p>
                <code className="text-xs text-muted-foreground">/ISAPI/Streaming/channels/101/picture</code>
              </div>
            ) : previewUrl ? (
              <img
                src={previewUrl}
                alt="Camera preview"
                className="max-w-full max-h-full object-contain"
              />
            ) : !previewLoading && !isDiscovering ? (
              <div className="text-center text-muted-foreground p-8">
                <Camera className="h-16 w-16 mx-auto mb-4 opacity-50" />
                <p>Нажмите "Снимок" для получения изображения</p>
                <p className="text-xs mt-2 text-muted-foreground/60">
                  Или "Live" для автообновления
                </p>
              </div>
            ) : null}
          </div>

          {/* URLs info */}
          {(device.camera?.snapshot_url || device.camera?.rtsp_url || rtspUrl) && (
            <div className="text-sm text-muted-foreground space-y-2 p-3 bg-muted/50 rounded-lg">
              {device.camera?.snapshot_url && (
                <p>
                  <span className="text-foreground font-medium">Snapshot:</span>{' '}
                  <code className="bg-muted px-1 rounded text-xs">{device.camera.snapshot_url}</code>
                </p>
              )}
              {(device.camera?.rtsp_url || rtspUrl) && (
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-foreground font-medium">RTSP:</span>
                  <code className="bg-muted px-1 rounded text-xs">{device.camera?.rtsp_url || rtspUrl}</code>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      navigator.clipboard.writeText(device.camera?.rtsp_url || rtspUrl || '')
                    }}
                  >
                    Копировать
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
