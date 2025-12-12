import { useState, useEffect, useCallback, useRef } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Moon,
  Sun,
  Monitor,
  Volume2,
  VolumeX,
  Clock,
  Key,
  Download,
  Upload,
  FolderOpen,
  Trash2,
  Save,
  AlertCircle,
  CheckCircle,
  Plus,
  Edit,
  Trash,
  Power,
  Play,
  Languages,
} from 'lucide-react'
import {
  GetAppSettings,
  SaveAppSettings,
  ExportConfiguration,
  ImportConfiguration,
  GetDataPath,
  OpenDataFolder,
  ClearOldData,
  GetCredentials,
  CreateCredential,
  UpdateCredential,
  DeleteCredential,
  GetAutostartEnabled,
  SetAutostartEnabled,
  GetMinimizeToTray,
  SetMinimizeToTray,
} from '../../wailsjs/go/main/App'
import { main } from '../../wailsjs/go/models'
import { useTheme } from '@/hooks/useTheme'
import { useTranslation } from '@/i18n'
import { changeLanguage, languages } from '@/i18n'

type AppSettings = main.AppSettings

interface Credential {
  id?: number
  name: string
  type: string
  username: string
  password?: string
  note: string
}

export function SettingsPage() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme } = useTheme()
  const [settings, setSettings] = useState<AppSettings>(
    main.AppSettings.createFrom({
      theme: 'light',
      monitoring_interval: 30,
      ping_timeout: 3,
      snmp_timeout: 5,
      monitoring_workers: 10,
      auto_start_monitor: true,
      sound_enabled: true,
      sound_volume: 0.5,
      notify_on_offline: true,
      notify_on_online: true,
      notify_on_port_change: false,
      event_retention_days: 30,
      camera_snapshot_interval: 60,
      camera_stream_type: 'jpeg',
    })
  )

  // Audio refs for test playback
  const okSoundRef = useRef<HTMLAudioElement | null>(null)
  const errorSoundRef = useRef<HTMLAudioElement | null>(null)

  // Play test sound
  const playTestSound = (type: 'ok' | 'error') => {
    const audio = type === 'ok' ? okSoundRef.current : errorSoundRef.current
    if (audio) {
      audio.volume = settings.sound_volume
      audio.currentTime = 0
      audio.play().catch(() => {})
    }
  }
  const [originalSettings, setOriginalSettings] = useState<AppSettings | null>(null)
  const [dataPath, setDataPath] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Credentials
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [credentialDialogOpen, setCredentialDialogOpen] = useState(false)
  const [editingCredential, setEditingCredential] = useState<Partial<Credential> | null>(null)
  const [deleteCredentialId, setDeleteCredentialId] = useState<number | null>(null)

  // Clear data dialog
  const [clearDataDialogOpen, setClearDataDialogOpen] = useState(false)
  const [daysToKeep, setDaysToKeep] = useState(30)

  // Autostart
  const [autostartEnabled, setAutostartEnabled] = useState(false)

  // Minimize to tray
  const [minimizeToTray, setMinimizeToTrayState] = useState(true)

  const loadSettings = useCallback(async () => {
    try {
      const data = await GetAppSettings()
      setSettings(data)
      setOriginalSettings(data)
    } catch (err) {
      console.error('Failed to load settings:', err)
    }
  }, [])

  const loadCredentials = useCallback(async () => {
    try {
      const data = await GetCredentials()
      // Map to local Credential type (password is not returned from backend)
      const creds: Credential[] = (data || []).map((c) => ({
        id: c.id,
        name: c.name,
        type: c.type,
        username: c.username,
        note: c.note,
      }))
      setCredentials(creds)
    } catch (err) {
      console.error('Failed to load credentials:', err)
    }
  }, [])

  const loadDataPath = useCallback(async () => {
    try {
      const path = await GetDataPath()
      setDataPath(path)
    } catch (err) {
      console.error('Failed to get data path:', err)
    }
  }, [])

  const loadAutostart = useCallback(async () => {
    try {
      const enabled = await GetAutostartEnabled()
      setAutostartEnabled(enabled)
    } catch (err) {
      console.error('Failed to get autostart state:', err)
    }
  }, [])

  const loadMinimizeToTray = useCallback(async () => {
    try {
      const enabled = await GetMinimizeToTray()
      setMinimizeToTrayState(enabled)
    } catch (err) {
      console.error('Failed to get minimize to tray state:', err)
    }
  }, [])

  useEffect(() => {
    loadSettings()
    loadCredentials()
    loadDataPath()
    loadAutostart()
    loadMinimizeToTray()
  }, [loadSettings, loadCredentials, loadDataPath, loadAutostart, loadMinimizeToTray])

  // Sync theme from hook to settings state
  useEffect(() => {
    setSettings((prev) => ({ ...prev, theme }))
  }, [theme])

  const hasChanges = JSON.stringify(settings) !== JSON.stringify(originalSettings)

  const handleSave = async () => {
    setIsSaving(true)
    setError(null)
    setSaveSuccess(false)

    try {
      await SaveAppSettings(settings)
      setOriginalSettings(settings)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    } finally {
      setIsSaving(false)
    }
  }

  const handleExport = async () => {
    try {
      const path = await ExportConfiguration()
      if (path) {
        setSaveSuccess(true)
        setTimeout(() => setSaveSuccess(false), 3000)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка экспорта')
    }
  }

  const handleImport = async () => {
    try {
      const success = await ImportConfiguration()
      if (success) {
        loadSettings()
        loadCredentials()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка импорта')
    }
  }

  const handleClearData = async () => {
    try {
      const deleted = await ClearOldData(daysToKeep)
      setClearDataDialogOpen(false)
      setError(null)
      // Show success message
      console.log(`Deleted ${deleted} old events`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка очистки')
    }
  }

  const handleSaveCredential = async () => {
    if (!editingCredential?.name || !editingCredential?.type) return

    try {
      if (editingCredential.id) {
        await UpdateCredential(editingCredential as never)
      } else {
        await CreateCredential(editingCredential as never)
      }
      loadCredentials()
      setCredentialDialogOpen(false)
      setEditingCredential(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    }
  }

  const handleDeleteCredential = async () => {
    if (!deleteCredentialId) return

    try {
      await DeleteCredential(deleteCredentialId)
      loadCredentials()
      setDeleteCredentialId(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка удаления')
    }
  }

  const handleAutostartToggle = async (enabled: boolean) => {
    try {
      await SetAutostartEnabled(enabled)
      setAutostartEnabled(enabled)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка настройки автозапуска')
    }
  }

  const handleMinimizeToTrayToggle = async (enabled: boolean) => {
    try {
      await SetMinimizeToTray(enabled)
      setMinimizeToTrayState(enabled)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка настройки сворачивания в трей')
    }
  }

  const updateSetting = <K extends keyof AppSettings>(key: K, value: AppSettings[K]) => {
    setSettings((prev) => ({ ...prev, [key]: value }))
  }

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Status messages */}
      {error && (
        <div className="flex items-center gap-2 p-4 text-destructive bg-destructive/10 rounded-lg">
          <AlertCircle className="h-5 w-5" />
          <span>{error}</span>
        </div>
      )}
      {saveSuccess && (
        <div className="flex items-center gap-2 p-4 text-green-500 bg-green-500/10 rounded-lg">
          <CheckCircle className="h-5 w-5" />
          <span>{t('settings.saved')}</span>
        </div>
      )}

      {/* Interface Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sun className="h-5 w-5" />
            {t('settings.interface.title')}
          </CardTitle>
          <CardDescription>{t('settings.interface.subtitle')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.interface.theme')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.interface.themeHint')}
              </p>
            </div>
            <div className="flex gap-2">
              <Button
                variant={theme === 'light' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setTheme('light')}
              >
                <Sun className="h-4 w-4 mr-2" />
                {t('settings.interface.light')}
              </Button>
              <Button
                variant={theme === 'dark' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setTheme('dark')}
              >
                <Moon className="h-4 w-4 mr-2" />
                {t('settings.interface.dark')}
              </Button>
              <Button
                variant={theme === 'system' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setTheme('system')}
              >
                <Monitor className="h-4 w-4 mr-2" />
                {t('settings.interface.system')}
              </Button>
            </div>
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.interface.language')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.interface.languageHint')}
              </p>
            </div>
            <div className="flex gap-2">
              {languages.map((lang) => (
                <Button
                  key={lang.code}
                  variant={i18n.language === lang.code ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => changeLanguage(lang.code)}
                >
                  <Languages className="h-4 w-4 mr-2" />
                  {lang.name}
                </Button>
              ))}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Notification Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {settings.sound_enabled ? <Volume2 className="h-5 w-5" /> : <VolumeX className="h-5 w-5" />}
            {t('settings.notifications.title')}
          </CardTitle>
          <CardDescription>{t('settings.notifications.subtitle')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.notifications.soundEnabled')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.notifications.soundEnabledHint')}
              </p>
            </div>
            <Switch
              checked={settings.sound_enabled}
              onCheckedChange={(v) => updateSetting('sound_enabled', v)}
            />
          </div>

          {settings.sound_enabled && (
            <>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label>{t('settings.notifications.volume')}: {Math.round(settings.sound_volume * 100)}%</Label>
                </div>
                <input
                  type="range"
                  min="0"
                  max="100"
                  value={Math.round(settings.sound_volume * 100)}
                  onChange={(e) => updateSetting('sound_volume', parseInt(e.target.value) / 100)}
                  className="w-full h-2 bg-muted rounded-lg appearance-none cursor-pointer accent-primary"
                />
              </div>

              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => playTestSound('ok')}
                  className="flex-1"
                >
                  <Play className="h-4 w-4 mr-2" />
                  {t('settings.notifications.testOnline')}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => playTestSound('error')}
                  className="flex-1"
                >
                  <Play className="h-4 w-4 mr-2" />
                  {t('settings.notifications.testOffline')}
                </Button>
              </div>
            </>
          )}

          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.notifications.notifyOffline')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.notifications.notifyOfflineHint')}
              </p>
            </div>
            <Switch
              checked={settings.notify_on_offline}
              onCheckedChange={(v) => updateSetting('notify_on_offline', v)}
            />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.notifications.notifyOnline')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.notifications.notifyOnlineHint')}
              </p>
            </div>
            <Switch
              checked={settings.notify_on_online}
              onCheckedChange={(v) => updateSetting('notify_on_online', v)}
            />
          </div>
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.notifications.notifyPortChange')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.notifications.notifyPortChangeHint')}
              </p>
            </div>
            <Switch
              checked={settings.notify_on_port_change}
              onCheckedChange={(v) => updateSetting('notify_on_port_change', v)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Hidden audio elements for test playback */}
      <audio ref={okSoundRef} src="/sounds/ok.mp3" preload="auto" />
      <audio ref={errorSoundRef} src="/sounds/error.mp3" preload="auto" />

      {/* Monitoring Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5" />
            {t('settings.monitoring.title')}
          </CardTitle>
          <CardDescription>{t('settings.monitoring.subtitle')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.monitoring.autoStart')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.monitoring.autoStartHint')}
              </p>
            </div>
            <Switch
              checked={settings.auto_start_monitor}
              onCheckedChange={(v) => updateSetting('auto_start_monitor', v)}
            />
          </div>
          <Separator />
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>{t('settings.monitoring.interval')}</Label>
              <Input
                type="number"
                min={5}
                max={300}
                value={settings.monitoring_interval}
                onChange={(e) =>
                  updateSetting('monitoring_interval', parseInt(e.target.value) || 30)
                }
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.monitoring.workers')}</Label>
              <Input
                type="number"
                min={1}
                max={50}
                value={settings.monitoring_workers}
                onChange={(e) =>
                  updateSetting('monitoring_workers', parseInt(e.target.value) || 10)
                }
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.monitoring.pingTimeout')}</Label>
              <Input
                type="number"
                min={1}
                max={30}
                value={settings.ping_timeout}
                onChange={(e) =>
                  updateSetting('ping_timeout', parseInt(e.target.value) || 3)
                }
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.monitoring.snmpTimeout')}</Label>
              <Input
                type="number"
                min={1}
                max={30}
                value={settings.snmp_timeout}
                onChange={(e) =>
                  updateSetting('snmp_timeout', parseInt(e.target.value) || 5)
                }
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Credentials */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-5 w-5" />
            {t('settings.credentials.title')}
          </CardTitle>
          <CardDescription>
            {t('settings.credentials.subtitle')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {credentials.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {t('settings.credentials.noCredentials')}
            </p>
          ) : (
            <div className="space-y-2">
              {credentials.map((cred) => (
                <div
                  key={cred.id}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <div>
                    <p className="font-medium">{cred.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {cred.type.toUpperCase()}
                      {cred.username && ` • ${cred.username}`}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => {
                        setEditingCredential(cred)
                        setCredentialDialogOpen(true)
                      }}
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setDeleteCredentialId(cred.id ?? null)}
                    >
                      <Trash className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
          <Button
            onClick={() => {
              setEditingCredential({
                name: '',
                type: 'snmp',
                username: '',
                password: '',
                note: '',
              })
              setCredentialDialogOpen(true)
            }}
          >
            <Plus className="h-4 w-4 mr-2" />
            {t('settings.credentials.addCredential')}
          </Button>
        </CardContent>
      </Card>

      {/* System Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Power className="h-5 w-5" />
            {t('settings.system.title')}
          </CardTitle>
          <CardDescription>{t('settings.system.subtitle')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.system.autostart')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.system.autostartHint')}
              </p>
            </div>
            <Switch
              checked={autostartEnabled}
              onCheckedChange={handleAutostartToggle}
            />
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.system.minimizeToTray')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.system.minimizeToTrayHint')}
              </p>
            </div>
            <Switch
              checked={minimizeToTray}
              onCheckedChange={handleMinimizeToTrayToggle}
            />
          </div>
        </CardContent>
      </Card>

      {/* Data Management */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Download className="h-5 w-5" />
            {t('settings.data.title')}
          </CardTitle>
          <CardDescription>
            {t('settings.data.subtitle')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.data.dataPath')}</p>
              <p className="text-sm text-muted-foreground font-mono">
                {dataPath || t('common.loading')}
              </p>
            </div>
            <Button variant="outline" size="sm" onClick={() => OpenDataFolder()}>
              <FolderOpen className="h-4 w-4 mr-2" />
              {t('settings.data.openFolder')}
            </Button>
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="font-medium">{t('settings.data.eventRetention')}</p>
              <p className="text-sm text-muted-foreground">
                {t('settings.data.eventRetentionHint')}
              </p>
            </div>
            <Select
              value={settings.event_retention_days.toString()}
              onValueChange={(v) => updateSetting('event_retention_days', parseInt(v))}
            >
              <SelectTrigger className="w-[120px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="7">7 {t('settings.data.days')}</SelectItem>
                <SelectItem value="14">14 {t('settings.data.days')}</SelectItem>
                <SelectItem value="30">30 {t('settings.data.days')}</SelectItem>
                <SelectItem value="60">60 {t('settings.data.days')}</SelectItem>
                <SelectItem value="90">90 {t('settings.data.days')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Separator />
          <div className="flex gap-4">
            <Button variant="outline" onClick={handleExport}>
              <Download className="h-4 w-4 mr-2" />
              {t('settings.data.export')}
            </Button>
            <Button variant="outline" onClick={handleImport}>
              <Upload className="h-4 w-4 mr-2" />
              {t('settings.data.import')}
            </Button>
            <Button
              variant="outline"
              onClick={() => setClearDataDialogOpen(true)}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              {t('settings.data.clearEvents')}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Save Button */}
      {hasChanges && (
        <div className="sticky bottom-4 flex justify-end">
          <Button onClick={handleSave} disabled={isSaving} size="lg">
            <Save className="h-4 w-4 mr-2" />
            {isSaving ? t('settings.saving') : t('settings.saveSettings')}
          </Button>
        </div>
      )}

      {/* Credential Dialog */}
      <Dialog open={credentialDialogOpen} onOpenChange={setCredentialDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingCredential?.id ? t('settings.credentials.editCredential') : t('settings.credentials.newCredential')}
            </DialogTitle>
            <DialogDescription>
              {t('settings.credentials.credentialHint')}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>{t('settings.credentials.credentialName')}</Label>
              <Input
                value={editingCredential?.name || ''}
                onChange={(e) =>
                  setEditingCredential((prev) => ({ ...prev, name: e.target.value }))
                }
                placeholder="SNMP Public"
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.credentials.credentialType')}</Label>
              <Select
                value={editingCredential?.type || 'snmp'}
                onValueChange={(v) =>
                  setEditingCredential((prev) => ({ ...prev, type: v }))
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="snmp">SNMP</SelectItem>
                  <SelectItem value="rtsp">RTSP</SelectItem>
                  <SelectItem value="onvif">ONVIF</SelectItem>
                  <SelectItem value="ssh">SSH</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>{t('settings.credentials.username')}</Label>
              <Input
                value={editingCredential?.username || ''}
                onChange={(e) =>
                  setEditingCredential((prev) => ({ ...prev, username: e.target.value }))
                }
                placeholder="public"
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.credentials.password')}</Label>
              <Input
                type="password"
                value={editingCredential?.password || ''}
                onChange={(e) =>
                  setEditingCredential((prev) => ({ ...prev, password: e.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label>{t('settings.credentials.note')}</Label>
              <Input
                value={editingCredential?.note || ''}
                onChange={(e) =>
                  setEditingCredential((prev) => ({ ...prev, note: e.target.value }))
                }
                placeholder={t('common.optional') as string}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCredentialDialogOpen(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button onClick={handleSaveCredential}>{t('common.save')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Credential Dialog */}
      <Dialog
        open={deleteCredentialId !== null}
        onOpenChange={() => setDeleteCredentialId(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('settings.credentials.deleteCredential')}</DialogTitle>
            <DialogDescription>
              {t('settings.credentials.deleteCredentialWarning')}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteCredentialId(null)}>
              {t('common.cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDeleteCredential}>
              {t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Clear Data Dialog */}
      <Dialog open={clearDataDialogOpen} onOpenChange={setClearDataDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('settings.data.clearOldData')}</DialogTitle>
            <DialogDescription>
              {t('settings.data.eventRetentionHint')}
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <div className="space-y-2">
              <Label>{t('settings.data.keepEventsFor')}</Label>
              <Select
                value={daysToKeep.toString()}
                onValueChange={(v) => setDaysToKeep(parseInt(v))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="7">7 {t('settings.data.days')}</SelectItem>
                  <SelectItem value="14">14 {t('settings.data.days')}</SelectItem>
                  <SelectItem value="30">30 {t('settings.data.days')}</SelectItem>
                  <SelectItem value="60">60 {t('settings.data.days')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setClearDataDialogOpen(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button onClick={handleClearData}>{t('settings.data.clearEvents')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
