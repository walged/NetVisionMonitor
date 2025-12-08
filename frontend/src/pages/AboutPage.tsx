import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Network,
  FileText,
  Shield,
  Code,
  Folder,
  HardDrive,
  Monitor,
  Server,
  Camera,
  ExternalLink,
  Copy,
  Check,
  ScrollText,
} from "lucide-react"
import { GetAppInfo, GetDeviceStats, GetDataPath, OpenDataFolder, GetLogPath, OpenLogFolder } from '../../wailsjs/go/main/App'

interface AppInfo {
  name: string
  version: string
  isPortable: boolean
}

interface Stats {
  total: number
  switch: number
  server: number
  camera: number
  online: number
  offline: number
}

export function AboutPage() {
  const [appInfo, setAppInfo] = useState<AppInfo | null>(null)
  const [stats, setStats] = useState<Stats | null>(null)
  const [dataPath, setDataPath] = useState('')
  const [logPath, setLogPath] = useState('')
  const [copied, setCopied] = useState(false)
  const [copiedLog, setCopiedLog] = useState(false)

  useEffect(() => {
    const load = async () => {
      try {
        const [info, statsData, path, logFilePath] = await Promise.all([
          GetAppInfo(),
          GetDeviceStats(),
          GetDataPath(),
          GetLogPath(),
        ])
        setAppInfo(info as AppInfo)
        setStats(statsData as unknown as Stats)
        setDataPath(path)
        setLogPath(logFilePath)
      } catch (err) {
        console.error('Failed to load app info:', err)
      }
    }
    load()
  }, [])

  const handleCopyPath = () => {
    navigator.clipboard.writeText(dataPath)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleCopyLogPath = () => {
    navigator.clipboard.writeText(logPath)
    setCopiedLog(true)
    setTimeout(() => setCopiedLog(false), 2000)
  }

  return (
    <div className="space-y-6 max-w-3xl">
      {/* Main Info Card */}
      <Card>
        <CardHeader className="text-center">
          <div className="flex justify-center mb-4">
            <div className="p-4 rounded-full bg-primary/10">
              <Network className="h-12 w-12 text-primary" />
            </div>
          </div>
          <CardTitle className="text-2xl">{appInfo?.name || 'NetVisionMonitor'}</CardTitle>
          <CardDescription>
            Система мониторинга сетевой инфраструктуры
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex justify-center gap-4">
            <div className="text-center">
              <p className="text-3xl font-bold text-primary">{appInfo?.version || '1.0.0'}</p>
              <p className="text-sm text-muted-foreground">Версия</p>
            </div>
            {appInfo?.isPortable && (
              <Badge variant="secondary" className="h-fit mt-2">
                <HardDrive className="h-3 w-3 mr-1" />
                Portable
              </Badge>
            )}
          </div>

          <Separator />

          {/* Stats */}
          {stats && (
            <>
              <div className="grid grid-cols-3 gap-4 text-center">
                <div className="p-4 rounded-lg bg-muted/50">
                  <Monitor className="h-5 w-5 mx-auto mb-2 text-blue-500" />
                  <p className="text-2xl font-bold">{stats.switch}</p>
                  <p className="text-xs text-muted-foreground">Коммутаторы</p>
                </div>
                <div className="p-4 rounded-lg bg-muted/50">
                  <Server className="h-5 w-5 mx-auto mb-2 text-green-500" />
                  <p className="text-2xl font-bold">{stats.server}</p>
                  <p className="text-xs text-muted-foreground">Серверы</p>
                </div>
                <div className="p-4 rounded-lg bg-muted/50">
                  <Camera className="h-5 w-5 mx-auto mb-2 text-purple-500" />
                  <p className="text-2xl font-bold">{stats.camera}</p>
                  <p className="text-xs text-muted-foreground">Камеры</p>
                </div>
              </div>
              <div className="flex justify-center gap-8 text-sm">
                <span className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-green-500" />
                  {stats.online} онлайн
                </span>
                <span className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-red-500" />
                  {stats.offline} оффлайн
                </span>
              </div>
              <Separator />
            </>
          )}

          {/* Features */}
          <div className="space-y-4">
            <InfoItem
              icon={<FileText className="h-5 w-5" />}
              title="Возможности"
              description="Мониторинг коммутаторов (SNMP), серверов (Ping/TCP), IP-камер (RTSP/ONVIF)"
            />
            <InfoItem
              icon={<Shield className="h-5 w-5" />}
              title="Безопасность"
              description="Локальное хранение данных, шифрование паролей AES-256-GCM"
            />
            <InfoItem
              icon={<Code className="h-5 w-5" />}
              title="Технологии"
              description="Go 1.22+, Wails v2, React 18, TypeScript, SQLite, Tailwind CSS"
            />
          </div>

          <Separator />

          {/* Data Path */}
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm font-medium">
              <Folder className="h-4 w-4" />
              Расположение данных
            </div>
            <div className="flex items-center gap-2">
              <code className="flex-1 p-2 rounded bg-muted text-xs font-mono truncate">
                {dataPath || 'Загрузка...'}
              </code>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleCopyPath}
                title="Копировать путь"
              >
                {copied ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => OpenDataFolder()}
                title="Открыть папку"
              >
                <ExternalLink className="h-4 w-4" />
              </Button>
            </div>
          </div>

          {/* Log Path */}
          <div className="space-y-2">
            <div className="flex items-center gap-2 text-sm font-medium">
              <ScrollText className="h-4 w-4" />
              Файл логов
            </div>
            <div className="flex items-center gap-2">
              <code className="flex-1 p-2 rounded bg-muted text-xs font-mono truncate">
                {logPath || 'Загрузка...'}
              </code>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleCopyLogPath}
                title="Копировать путь"
              >
                {copiedLog ? (
                  <Check className="h-4 w-4 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => OpenLogFolder()}
                title="Открыть папку логов"
              >
                <ExternalLink className="h-4 w-4" />
              </Button>
            </div>
          </div>

          <Separator />

          <div className="text-center text-sm text-muted-foreground space-y-1">
            <p>Работает полностью автономно без подключения к интернету</p>
            <p>Поддержка до 300+ устройств в сети</p>
            <p className="pt-2">© 2024 NetVisionMonitor</p>
          </div>
        </CardContent>
      </Card>

      {/* Keyboard Shortcuts Card */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Горячие клавиши</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <ShortcutItem keys={['Ctrl', 'R']} description="Обновить данные" />
            <ShortcutItem keys={['Ctrl', 'M']} description="Запуск мониторинга" />
            <ShortcutItem keys={['Ctrl', 'N']} description="Добавить устройство" />
            <ShortcutItem keys={['Ctrl', ',']} description="Настройки" />
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

interface InfoItemProps {
  icon: React.ReactNode
  title: string
  description: string
}

function InfoItem({ icon, title, description }: InfoItemProps) {
  return (
    <div className="flex gap-4">
      <div className="text-muted-foreground">{icon}</div>
      <div>
        <p className="font-medium">{title}</p>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  )
}

interface ShortcutItemProps {
  keys: string[]
  description: string
}

function ShortcutItem({ keys, description }: ShortcutItemProps) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{description}</span>
      <div className="flex gap-1">
        {keys.map((key, i) => (
          <span key={i}>
            <kbd className="px-2 py-1 text-xs rounded bg-muted border">
              {key}
            </kbd>
            {i < keys.length - 1 && <span className="mx-1">+</span>}
          </span>
        ))}
      </div>
    </div>
  )
}
