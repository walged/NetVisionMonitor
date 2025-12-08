import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Monitor, Server, Camera, Network, X, Radio, ExternalLink, Loader2, CheckCircle, XCircle } from 'lucide-react'
import { PingDevice } from '../../../wailsjs/go/main/App'
import type { SchemaItem } from '@/types'

interface PingResult {
  success: boolean
  avg_latency_ms: number
  error?: string
}

interface LinkedCamera {
  id: number
  name: string
  status: string
  port_number: number
}

interface DevicePopupProps {
  item: SchemaItem | null
  onClose: () => void
  onViewDetails?: (deviceId: number) => void
  linkedCameras?: LinkedCamera[]
}

export function DevicePopup({ item, onClose, onViewDetails, linkedCameras = [] }: DevicePopupProps) {
  const [isPinging, setIsPinging] = useState(false)
  const [pingResult, setPingResult] = useState<PingResult | null>(null)

  if (!item) return null

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

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'online':
        return <Badge variant="default" className="bg-green-500">В сети</Badge>
      case 'offline':
        return <Badge variant="destructive">Не в сети</Badge>
      default:
        return <Badge variant="secondary">Неизвестно</Badge>
    }
  }

  const handlePing = async () => {
    if (!item.device_ip) return
    setIsPinging(true)
    setPingResult(null)
    try {
      const result = await PingDevice(item.device_ip)
      setPingResult(result)
    } catch (err) {
      setPingResult({
        success: false,
        avg_latency_ms: 0,
        error: err instanceof Error ? err.message : 'Ошибка пинга',
      })
    } finally {
      setIsPinging(false)
    }
  }

  const handleOpenWebUI = () => {
    if (!item.device_ip) return
    window.open(`http://${item.device_ip}`, '_blank')
  }

  return (
    <Card className="z-50 w-80 shadow-xl animate-in fade-in-0 zoom-in-95 !bg-white dark:!bg-zinc-900 border">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            {getTypeIcon(item.device_type || '')}
            <div>
              <CardTitle className="text-base">{item.device_name || 'Устройство'}</CardTitle>
              <CardDescription>{getTypeName(item.device_type || '')}</CardDescription>
            </div>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={onClose}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">Статус</span>
          {getStatusBadge(item.device_status || 'unknown')}
        </div>
        <div className="flex items-center justify-between">
          <span className="text-sm text-muted-foreground">IP-адрес</span>
          <span className="text-sm font-mono">{item.device_ip}</span>
        </div>

        {/* Ping result */}
        {pingResult && (
          <div className="p-2 bg-muted rounded-md text-sm">
            <div className="flex items-center gap-2">
              {pingResult.success ? (
                <>
                  <CheckCircle className="h-4 w-4 text-green-500" />
                  <span className="text-green-500">
                    Доступен ({pingResult.avg_latency_ms.toFixed(0)} мс)
                  </span>
                </>
              ) : (
                <>
                  <XCircle className="h-4 w-4 text-red-500" />
                  <span className="text-red-500">Недоступен</span>
                </>
              )}
            </div>
          </div>
        )}

        {/* Linked cameras for switches */}
        {item.device_type === 'switch' && linkedCameras.length > 0 && (
          <div className="pt-2 border-t">
            <div className="text-sm font-medium mb-2 flex items-center gap-1">
              <Camera className="h-4 w-4" />
              Камеры ({linkedCameras.length})
            </div>
            <div className="space-y-1 max-h-24 overflow-y-auto">
              {linkedCameras.map((cam) => (
                <div
                  key={cam.id}
                  className="flex items-center justify-between text-xs p-1.5 bg-muted/50 rounded"
                >
                  <span className="truncate flex-1">{cam.name}</span>
                  <div className="flex items-center gap-1.5 ml-2">
                    <span className="text-muted-foreground">Порт {cam.port_number}</span>
                    <div
                      className={`w-2 h-2 rounded-full ${
                        cam.status === 'online' ? 'bg-green-500' : 'bg-red-500'
                      }`}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Action buttons */}
        <div className="flex gap-2 pt-2">
          <Button
            variant="outline"
            size="sm"
            className="flex-1"
            onClick={handlePing}
            disabled={isPinging}
          >
            {isPinging ? (
              <Loader2 className="h-4 w-4 mr-1 animate-spin" />
            ) : (
              <Radio className="h-4 w-4 mr-1" />
            )}
            Ping
          </Button>
          <Button
            variant="outline"
            size="sm"
            className="flex-1"
            onClick={handleOpenWebUI}
          >
            <ExternalLink className="h-4 w-4 mr-1" />
            Открыть
          </Button>
        </div>

        {onViewDetails && (
          <Button
            variant="default"
            className="w-full"
            onClick={() => onViewDetails(item.device_id)}
          >
            Подробнее
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
