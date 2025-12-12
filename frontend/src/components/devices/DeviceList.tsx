import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Network,
  Server,
  Camera,
  Pencil,
  Trash2,
  Search,
  Circle,
  Eye,
  Radio,
  Terminal,
  Loader2,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import { PingDevice, OpenPingCmd } from '../../../wailsjs/go/main/App'
import { useTranslation } from '@/i18n'

interface Device {
  id: number
  name: string
  ip_address: string
  type: string
  manufacturer: string
  model: string
  status: string
  last_check?: string
}

interface DeviceListProps {
  devices: Device[]
  onEdit: (device: Device) => void
  onDelete: (id: number) => void
  onView?: (device: Device) => void
  isLoading?: boolean
}

const deviceTypeIcons: Record<string, React.ReactNode> = {
  switch: <Network className="h-5 w-5" />,
  server: <Server className="h-5 w-5" />,
  camera: <Camera className="h-5 w-5" />,
}

const statusColors: Record<string, string> = {
  online: 'text-green-500',
  offline: 'text-red-500',
  unknown: 'text-gray-500',
}

export function DeviceList({
  devices,
  onEdit,
  onDelete,
  onView,
  isLoading = false,
}: DeviceListProps) {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [statusFilter, setStatusFilter] = useState<string>('all')

  const deviceTypeLabels: Record<string, string> = {
    switch: t('devices.types.switch') as string,
    server: t('devices.types.server') as string,
    camera: t('devices.types.camera') as string,
  }

  const statusLabels: Record<string, string> = {
    online: t('status.online') as string,
    offline: t('status.offline') as string,
    unknown: t('status.unknown') as string,
  }

  const filteredDevices = devices.filter((device) => {
    const matchesSearch =
      device.name.toLowerCase().includes(search.toLowerCase()) ||
      device.ip_address.includes(search) ||
      device.model.toLowerCase().includes(search.toLowerCase())

    const matchesType = typeFilter === 'all' || device.type === typeFilter
    const matchesStatus =
      statusFilter === 'all' || device.status === statusFilter

    return matchesSearch && matchesType && matchesStatus
  })

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t('common.search') as string}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-10"
          />
        </div>
        <Select value={typeFilter} onValueChange={setTypeFilter}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t('devices.form.deviceType')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.all')}</SelectItem>
            <SelectItem value="switch">{t('devices.stats.switches')}</SelectItem>
            <SelectItem value="server">{t('devices.stats.servers')}</SelectItem>
            <SelectItem value="camera">{t('devices.stats.cameras')}</SelectItem>
          </SelectContent>
        </Select>
        <Select value={statusFilter} onValueChange={setStatusFilter}>
          <SelectTrigger className="w-[150px]">
            <SelectValue placeholder={t('common.status')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.all')}</SelectItem>
            <SelectItem value="online">{t('status.online')}</SelectItem>
            <SelectItem value="offline">{t('status.offline')}</SelectItem>
            <SelectItem value="unknown">{t('status.unknown')}</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Device Cards */}
      {isLoading ? (
        <div className="text-center py-8 text-muted-foreground">
          {t('common.loading')}
        </div>
      ) : filteredDevices.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          {devices.length === 0
            ? t('devices.addDevice')
            : t('events.noEvents')}
        </div>
      ) : (
        <div className="grid gap-3">
          {filteredDevices.map((device) => (
            <DeviceCard
              key={device.id}
              device={device}
              deviceTypeLabels={deviceTypeLabels}
              statusLabels={statusLabels}
              onEdit={() => onEdit(device)}
              onDelete={() => onDelete(device.id)}
              onView={onView ? () => onView(device) : undefined}
            />
          ))}
        </div>
      )}
    </div>
  )
}

interface PingResult {
  success: boolean
  packets_sent: number
  packets_recv: number
  packet_loss: number
  avg_latency_ms: number
  min_latency_ms: number
  max_latency_ms: number
  error?: string
}

interface DeviceCardProps {
  device: Device
  deviceTypeLabels: Record<string, string>
  statusLabels: Record<string, string>
  onEdit: () => void
  onDelete: () => void
  onView?: () => void
}

function DeviceCard({ device, deviceTypeLabels, statusLabels, onEdit, onDelete, onView }: DeviceCardProps) {
  const { t } = useTranslation()
  const [isPinging, setIsPinging] = useState(false)
  const [pingResult, setPingResult] = useState<PingResult | null>(null)
  const [pingOpen, setPingOpen] = useState(false)

  const handleInternalPing = async () => {
    setIsPinging(true)
    setPingResult(null)
    try {
      const result = await PingDevice(device.ip_address)
      setPingResult(result)
    } catch (err) {
      setPingResult({
        success: false,
        packets_sent: 0,
        packets_recv: 0,
        packet_loss: 100,
        avg_latency_ms: 0,
        min_latency_ms: 0,
        max_latency_ms: 0,
        error: err instanceof Error ? err.message : t('common.error') as string,
      })
    } finally {
      setIsPinging(false)
    }
  }

  const handleCmdPing = async () => {
    try {
      await OpenPingCmd(device.ip_address)
    } catch (err) {
      console.error('Failed to open ping cmd:', err)
    }
  }

  return (
    <TooltipProvider>
      <Card className="hover:bg-accent/50 transition-colors">
        <CardContent className="flex items-center gap-4 p-4">
          {/* Status indicator */}
          <Tooltip>
            <TooltipTrigger>
              <Circle
                className={`h-3 w-3 fill-current ${statusColors[device.status]}`}
              />
            </TooltipTrigger>
            <TooltipContent>
              <p>{statusLabels[device.status]}</p>
            </TooltipContent>
          </Tooltip>

          {/* Icon */}
          <div className="text-muted-foreground">
            {deviceTypeIcons[device.type]}
          </div>

          {/* Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              <span className="font-medium truncate">{device.name}</span>
              <Badge variant="secondary" className="text-xs">
                {deviceTypeLabels[device.type]}
              </Badge>
            </div>
            <div className="flex items-center gap-4 text-sm text-muted-foreground">
              <span className="font-mono">{device.ip_address}</span>
              {device.model && <span>{device.model}</span>}
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-1">
            {/* Ping button with popover */}
            <Popover open={pingOpen} onOpenChange={setPingOpen}>
              <Tooltip>
                <TooltipTrigger asChild>
                  <PopoverTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <Radio className="h-4 w-4" />
                    </Button>
                  </PopoverTrigger>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Ping</p>
                </TooltipContent>
              </Tooltip>
              <PopoverContent className="w-72" align="end">
                <div className="space-y-3">
                  <div className="font-medium">Ping {device.ip_address}</div>

                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={handleInternalPing}
                      disabled={isPinging}
                      className="flex-1"
                    >
                      {isPinging ? (
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      ) : (
                        <Radio className="h-4 w-4 mr-2" />
                      )}
                      Internal
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={handleCmdPing}
                      className="flex-1"
                    >
                      <Terminal className="h-4 w-4 mr-2" />
                      CMD
                    </Button>
                  </div>

                  {pingResult && (
                    <div className="p-2 bg-muted rounded-md text-sm space-y-1">
                      <div className="flex items-center gap-2">
                        {pingResult.success ? (
                          <CheckCircle className="h-4 w-4 text-green-500" />
                        ) : (
                          <XCircle className="h-4 w-4 text-red-500" />
                        )}
                        <span className={pingResult.success ? 'text-green-500' : 'text-red-500'}>
                          {pingResult.success ? t('status.online') : t('status.offline')}
                        </span>
                      </div>
                      {pingResult.success ? (
                        <>
                          <div>Sent: {pingResult.packets_sent}, Received: {pingResult.packets_recv}</div>
                          <div>Latency: {pingResult.avg_latency_ms.toFixed(0)} ms</div>
                          <div className="text-xs text-muted-foreground">
                            (min: {pingResult.min_latency_ms.toFixed(0)}, max: {pingResult.max_latency_ms.toFixed(0)} ms)
                          </div>
                        </>
                      ) : (
                        <div className="text-muted-foreground">
                          {pingResult.error || t('status.offline')}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </PopoverContent>
            </Popover>

            {onView && (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" onClick={onView}>
                    <Eye className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{t('common.details')}</p>
                </TooltipContent>
              </Tooltip>
            )}

            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" onClick={onEdit}>
                  <Pencil className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{t('common.edit')}</p>
              </TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={onDelete}
                  className="text-destructive hover:text-destructive"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>{t('common.delete')}</p>
              </TooltipContent>
            </Tooltip>
          </div>
        </CardContent>
      </Card>
    </TooltipProvider>
  )
}
