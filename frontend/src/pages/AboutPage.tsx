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
import { useTranslation } from '@/i18n'

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
  const { t } = useTranslation()
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
    <div className="flex justify-center">
      <Card className="w-full max-w-lg">
        <CardHeader className="text-center pb-2">
          <div className="flex justify-center mb-3">
            <div className="p-3 rounded-full bg-primary/10">
              <Network className="h-10 w-10 text-primary" />
            </div>
          </div>
          <CardTitle className="text-xl">{appInfo?.name || 'NetVisionMonitor'}</CardTitle>
          <CardDescription className="text-xs">
            {t('about.description')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex justify-center gap-3">
            <div className="text-center">
              <p className="text-2xl font-bold text-primary">{appInfo?.version || '1.0.0'}</p>
              <p className="text-xs text-muted-foreground">{t('about.version')}</p>
            </div>
            {appInfo?.isPortable && (
              <Badge variant="secondary" className="h-fit mt-1">
                <HardDrive className="h-3 w-3 mr-1" />
                Portable
              </Badge>
            )}
          </div>

          <Separator />

          {/* Stats */}
          {stats && (
            <>
              <div className="grid grid-cols-3 gap-2 text-center">
                <div className="p-2 rounded-lg bg-muted/50">
                  <Monitor className="h-4 w-4 mx-auto mb-1 text-blue-500" />
                  <p className="text-lg font-bold">{stats.switch}</p>
                  <p className="text-[10px] text-muted-foreground">{t('devices.stats.switches')}</p>
                </div>
                <div className="p-2 rounded-lg bg-muted/50">
                  <Server className="h-4 w-4 mx-auto mb-1 text-green-500" />
                  <p className="text-lg font-bold">{stats.server}</p>
                  <p className="text-[10px] text-muted-foreground">{t('devices.stats.servers')}</p>
                </div>
                <div className="p-2 rounded-lg bg-muted/50">
                  <Camera className="h-4 w-4 mx-auto mb-1 text-purple-500" />
                  <p className="text-lg font-bold">{stats.camera}</p>
                  <p className="text-[10px] text-muted-foreground">{t('devices.stats.cameras')}</p>
                </div>
              </div>
              <div className="flex justify-center gap-6 text-xs">
                <span className="flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                  {stats.online} {t('status.online').toLowerCase()}
                </span>
                <span className="flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-red-500" />
                  {stats.offline} {t('status.offline').toLowerCase()}
                </span>
              </div>
              <Separator />
            </>
          )}

          {/* Features */}
          <div className="space-y-2">
            <InfoItem
              icon={<FileText className="h-4 w-4" />}
              title={t('about.features')}
              description="SNMP, Ping/TCP, RTSP/ONVIF"
            />
            <InfoItem
              icon={<Shield className="h-4 w-4" />}
              title={t('about.featuresList.monitoring')}
              description="AES-256-GCM"
            />
            <InfoItem
              icon={<Code className="h-4 w-4" />}
              title={t('about.technologies')}
              description="Go, Wails v2, React, TypeScript, SQLite"
            />
          </div>

          <Separator />

          {/* Data Path */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 text-xs font-medium">
              <Folder className="h-3 w-3" />
              {t('settings.data.title')}
            </div>
            <div className="flex items-center gap-1">
              <code className="flex-1 p-1.5 rounded bg-muted text-[10px] font-mono truncate">
                {dataPath || t('common.loading')}
              </code>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleCopyPath}>
                {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
              </Button>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => OpenDataFolder()}>
                <ExternalLink className="h-3 w-3" />
              </Button>
            </div>
          </div>

          {/* Log Path */}
          <div className="space-y-1">
            <div className="flex items-center gap-1.5 text-xs font-medium">
              <ScrollText className="h-3 w-3" />
              {t('events.title')}
            </div>
            <div className="flex items-center gap-1">
              <code className="flex-1 p-1.5 rounded bg-muted text-[10px] font-mono truncate">
                {logPath || t('common.loading')}
              </code>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleCopyLogPath}>
                {copiedLog ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
              </Button>
              <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => OpenLogFolder()}>
                <ExternalLink className="h-3 w-3" />
              </Button>
            </div>
          </div>

          <Separator />

          <div className="text-center text-xs text-muted-foreground">
            <p>
              {t('about.author')}: <span className="font-medium">walged</span> with Claude •{' '}
              <a href="https://arthurdev.ru" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                arthurdev.ru
              </a>
            </p>
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
    <div className="flex gap-3">
      <div className="text-muted-foreground mt-0.5">{icon}</div>
      <div>
        <p className="text-sm font-medium">{title}</p>
        <p className="text-xs text-muted-foreground">{description}</p>
      </div>
    </div>
  )
}
