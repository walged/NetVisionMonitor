import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { Separator } from '@/components/ui/separator'
import { GetSwitchesWithPorts } from '../../../wailsjs/go/main/App'
import { switchManufacturers, getModelsByManufacturer, getPortCountForModel, getSfpPortCountForModel } from '@/data/switchModels'

interface DeviceFormData {
  id?: number
  name: string
  ip_address: string
  type: string
  manufacturer: string
  model: string
  credential_id?: number
  // Switch
  snmp_community: string
  snmp_write_community: string
  snmp_version: string
  port_count: number
  sfp_port_count: number
  // SNMPv3 settings
  snmpv3_user: string
  snmpv3_security: string
  snmpv3_auth_proto: string
  snmpv3_auth_pass: string
  snmpv3_priv_proto: string
  snmpv3_priv_pass: string
  // Camera
  rtsp_url: string
  onvif_port: number
  snapshot_url: string
  stream_type: string
  switch_port_id?: number // Link to switch port
  // Server
  tcp_ports: string
  use_snmp: boolean
  // Uplink (for switches and servers)
  uplink_switch_id?: number
  uplink_port_id?: number
}

interface Credential {
  id: number
  name: string
  type: string
}

interface SwitchPort {
  id: number
  switch_id: number
  port_number: number
  name: string
  status: string
  port_type: string
  linked_camera_id?: number
  linked_switch_id?: number
}

interface SwitchWithPorts {
  device_id: number
  device_name: string
  ip_address: string
  port_count: number
  ports: SwitchPort[]
}

interface DeviceFormProps {
  open: boolean
  onClose: () => void
  onSubmit: (data: DeviceFormData) => Promise<void>
  initialData?: Partial<DeviceFormData>
  credentials: Credential[]
  isLoading?: boolean
}

const defaultFormData: DeviceFormData = {
  name: '',
  ip_address: '',
  type: 'switch',
  manufacturer: '',
  model: '',
  snmp_community: 'public',
  snmp_write_community: 'private',
  snmp_version: 'v2c',
  port_count: 8,
  sfp_port_count: 2,
  snmpv3_user: '',
  snmpv3_security: 'noAuthNoPriv',
  snmpv3_auth_proto: '',
  snmpv3_auth_pass: '',
  snmpv3_priv_proto: '',
  snmpv3_priv_pass: '',
  rtsp_url: '',
  onvif_port: 80,
  snapshot_url: '',
  stream_type: 'jpeg',
  tcp_ports: '[]',
  use_snmp: false,
}

export function DeviceForm({
  open,
  onClose,
  onSubmit,
  initialData,
  credentials,
  isLoading = false,
}: DeviceFormProps) {
  const [formData, setFormData] = useState<DeviceFormData>(defaultFormData)
  const [error, setError] = useState<string | null>(null)
  const [switches, setSwitches] = useState<SwitchWithPorts[]>([])
  const [selectedSwitchId, setSelectedSwitchId] = useState<number | null>(null)
  const [selectedUplinkSwitchId, setSelectedUplinkSwitchId] = useState<number | null>(null)

  const loadSwitches = async () => {
    try {
      const data = await GetSwitchesWithPorts()
      setSwitches(data || [])
      return data || []
    } catch (err) {
      console.error('Failed to load switches:', err)
      return []
    }
  }

  // Load switches when form opens and type is camera, switch, or server
  useEffect(() => {
    if (open && (formData.type === 'camera' || formData.type === 'switch' || formData.type === 'server')) {
      loadSwitches()
    }
  }, [open, formData.type])

  // Initialize form data and find switch for existing camera/uplink
  useEffect(() => {
    const initForm = async () => {
      if (initialData) {
        setFormData({ ...defaultFormData, ...initialData })
        const switchesData = await loadSwitches()

        // If editing a camera with a switch_port_id, find which switch owns that port
        if (initialData.type === 'camera' && initialData.switch_port_id) {
          for (const sw of switchesData) {
            if (sw.ports) {
              const port = sw.ports.find(p => p.id === initialData.switch_port_id)
              if (port) {
                setSelectedSwitchId(sw.device_id)
                break
              }
            }
          }
        }

        // If editing a switch/server with uplink_port_id, find which switch owns that port
        if ((initialData.type === 'switch' || initialData.type === 'server') && initialData.uplink_port_id) {
          for (const sw of switchesData) {
            if (sw.ports) {
              const port = sw.ports.find(p => p.id === initialData.uplink_port_id)
              if (port) {
                setSelectedUplinkSwitchId(sw.device_id)
                break
              }
            }
          }
        }
      } else {
        setFormData(defaultFormData)
        setSelectedSwitchId(null)
        setSelectedUplinkSwitchId(null)
      }
      setError(null)
    }

    if (open) {
      initForm()
    } else {
      setSelectedSwitchId(null)
      setSelectedUplinkSwitchId(null)
    }
  }, [initialData, open])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)

    if (!formData.name.trim()) {
      setError('Введите название устройства')
      return
    }
    if (!formData.ip_address.trim()) {
      setError('Введите IP-адрес')
      return
    }

    // Validate camera requires switch port
    if (formData.type === 'camera' && !formData.switch_port_id) {
      setError('Выберите коммутатор и порт для камеры')
      return
    }

    try {
      await onSubmit(formData)
      onClose()
    } catch (err) {
      console.error('Form submit error:', err)
      setError(err instanceof Error ? err.message : 'Ошибка сохранения')
    }
  }

  const updateField = (field: keyof DeviceFormData, value: unknown) => {
    setFormData((prev) => ({ ...prev, [field]: value }))
  }

  const handleSwitchChange = (switchId: string) => {
    const id = switchId === 'none' ? null : parseInt(switchId)
    setSelectedSwitchId(id)
    updateField('switch_port_id', undefined)
  }

  const handlePortChange = (portId: string) => {
    const portIdNum = portId === 'none' ? undefined : parseInt(portId)
    updateField('switch_port_id', portIdNum)
  }

  const handleManufacturerChange = (manufacturerId: string) => {
    const newManufacturer = manufacturerId === 'none' ? '' : manufacturerId
    setFormData((prev) => ({
      ...prev,
      manufacturer: newManufacturer,
      model: '', // Reset model when manufacturer changes
    }))
  }

  const handleModelChange = (modelName: string) => {
    if (modelName === 'none') {
      setFormData((prev) => ({ ...prev, model: '' }))
      return
    }

    // Auto-set port count and SFP port count based on model
    const portCount = formData.manufacturer
      ? getPortCountForModel(formData.manufacturer, modelName)
      : undefined
    const sfpPortCount = formData.manufacturer
      ? getSfpPortCountForModel(formData.manufacturer, modelName)
      : 0

    // Always update model, port_count, and sfp_port_count together
    setFormData((prev) => ({
      ...prev,
      model: modelName,
      port_count: portCount || prev.port_count,
      sfp_port_count: sfpPortCount,
    }))
  }

  const availableModels = formData.manufacturer
    ? getModelsByManufacturer(formData.manufacturer)
    : []

  const getAvailablePorts = () => {
    if (!selectedSwitchId) return []
    const sw = switches.find(s => s.device_id === selectedSwitchId)
    if (!sw || !sw.ports) return []
    // Filter only copper ports without linked camera (or the current camera when editing)
    return sw.ports.filter(p => p.port_type === 'copper' && (!p.linked_camera_id || p.linked_camera_id === initialData?.id))
  }

  // Get available SFP ports for uplink (exclude current device if editing)
  const getAvailableSfpPorts = () => {
    if (!selectedUplinkSwitchId) return []
    const sw = switches.find(s => s.device_id === selectedUplinkSwitchId)
    if (!sw || !sw.ports) return []
    // Filter only SFP ports without linked switch (or the current device when editing)
    return sw.ports.filter(p =>
      p.port_type === 'sfp' &&
      (!p.linked_switch_id || p.linked_switch_id === initialData?.id)
    )
  }

  // Get switches available for uplink (exclude current switch if editing)
  const getAvailableUplinkSwitches = () => {
    return switches.filter(sw => sw.device_id !== initialData?.id)
  }

  const handleUplinkSwitchChange = (switchId: string) => {
    const id = switchId === 'none' ? null : parseInt(switchId)
    setSelectedUplinkSwitchId(id)
    updateField('uplink_switch_id', id || undefined)
    updateField('uplink_port_id', undefined)
  }

  const handleUplinkPortChange = (portId: string) => {
    const portIdNum = portId === 'none' ? undefined : parseInt(portId)
    updateField('uplink_port_id', portIdNum)
  }

  const isEditing = !!initialData?.id

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[500px] max-h-[90vh] overflow-y-auto">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>
              {isEditing ? 'Редактировать устройство' : 'Добавить устройство'}
            </DialogTitle>
            <DialogDescription>
              {isEditing
                ? 'Измените параметры устройства'
                : 'Заполните информацию о новом устройстве'}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Type selector first for camera */}
            <div className="grid gap-2">
              <Label htmlFor="type">Тип устройства</Label>
              <Select
                value={formData.type}
                onValueChange={(v) => {
                  updateField('type', v)
                  if (v === 'camera') {
                    loadSwitches()
                  }
                }}
                disabled={isEditing}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="switch">Коммутатор</SelectItem>
                  <SelectItem value="server">Сервер</SelectItem>
                  <SelectItem value="camera">Камера</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Camera-specific: Switch and Port selection FIRST */}
            {formData.type === 'camera' && (
              <>
                <Separator />
                <div className="text-sm font-medium text-muted-foreground">
                  Подключение к коммутатору *
                </div>

                {switches.length === 0 ? (
                  <div className="text-sm text-destructive bg-destructive/10 p-3 rounded">
                    Сначала добавьте коммутатор, затем можно добавить камеру
                  </div>
                ) : (
                  <>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="grid gap-2">
                        <Label htmlFor="switch">Коммутатор *</Label>
                        <Select
                          value={selectedSwitchId?.toString() || 'none'}
                          onValueChange={handleSwitchChange}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Выберите коммутатор..." />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">Выберите...</SelectItem>
                            {switches.map((sw) => (
                              <SelectItem key={sw.device_id} value={sw.device_id.toString()}>
                                {sw.device_name} ({sw.ip_address})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="grid gap-2">
                        <Label htmlFor="port">Порт *</Label>
                        <Select
                          value={formData.switch_port_id?.toString() || 'none'}
                          onValueChange={handlePortChange}
                          disabled={!selectedSwitchId}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Выберите порт..." />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">Выберите...</SelectItem>
                            {getAvailablePorts().map((port) => (
                              <SelectItem key={port.id} value={port.id.toString()}>
                                Порт {port.port_number} - {port.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    {selectedSwitchId && getAvailablePorts().length === 0 && (
                      <div className="text-sm text-yellow-600 bg-yellow-500/10 p-2 rounded">
                        Все порты этого коммутатора уже заняты камерами
                      </div>
                    )}
                  </>
                )}
                <Separator />
              </>
            )}

            {/* Basic fields */}
            <div className="grid gap-2">
              <Label htmlFor="name">Название *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => updateField('name', e.target.value)}
                placeholder={formData.type === 'camera' ? 'Камера вход' : formData.type === 'server' ? 'Сервер 1С' : 'Коммутатор 1 этаж'}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="ip_address">IP-адрес *</Label>
              <Input
                id="ip_address"
                value={formData.ip_address}
                onChange={(e) => updateField('ip_address', e.target.value)}
                placeholder="192.168.1.1"
              />
            </div>

            {/* Manufacturer/Model selection for switches */}
            {formData.type === 'switch' && (
              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="manufacturer">Производитель</Label>
                  <Select
                    value={formData.manufacturer || 'none'}
                    onValueChange={handleManufacturerChange}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Выберите производителя..." />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">Выберите...</SelectItem>
                      {switchManufacturers.map((m) => (
                        <SelectItem key={m.id} value={m.id}>
                          {m.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="model">Модель</Label>
                  <Select
                    value={formData.model || 'none'}
                    onValueChange={handleModelChange}
                    disabled={!formData.manufacturer}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Выберите модель...">
                        {formData.model || 'Выберите...'}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">Выберите...</SelectItem>
                      {availableModels.map((model) => (
                        <SelectItem key={model.name} value={model.name}>
                          {model.name} {model.description && `(${model.description})`}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}

            {/* Model input for cameras and servers */}
            {formData.type !== 'switch' && (
              <div className="grid gap-2">
                <Label htmlFor="model">Модель</Label>
                <Input
                  id="model"
                  value={formData.model}
                  onChange={(e) => updateField('model', e.target.value)}
                  placeholder={formData.type === 'camera' ? 'Hikvision DS-2CD' : 'Dell PowerEdge'}
                />
              </div>
            )}

            <div className="grid gap-2">
              <Label htmlFor="credential">Учётные данные</Label>
              <Select
                value={formData.credential_id?.toString() || 'none'}
                onValueChange={(v) =>
                  updateField('credential_id', v === 'none' ? undefined : parseInt(v))
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Выберите..." />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">Без учётных данных</SelectItem>
                  {credentials.map((c) => (
                    <SelectItem key={c.id} value={c.id.toString()}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Switch-specific fields */}
            {formData.type === 'switch' && (
              <>
                <Separator />
                <div className="text-sm font-medium text-muted-foreground">
                  Параметры SNMP
                </div>
                {/* Show port configuration info based on selected model */}
                {formData.model && (
                  <div className="text-sm bg-muted/50 p-3 rounded-md">
                    <div className="font-medium mb-1">Конфигурация портов:</div>
                    <div className="text-muted-foreground">
                      Всего портов: <span className="text-foreground font-medium">{formData.port_count}</span>
                      {formData.sfp_port_count > 0 && (
                        <>
                          {' '}({formData.port_count - formData.sfp_port_count} copper + {formData.sfp_port_count} SFP)
                        </>
                      )}
                    </div>
                  </div>
                )}

                <div className="grid gap-2">
                  <Label htmlFor="snmp_version">Версия SNMP</Label>
                  <Select
                    value={formData.snmp_version}
                    onValueChange={(v) => updateField('snmp_version', v)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="v1">SNMPv1</SelectItem>
                      <SelectItem value="v2c">SNMPv2c</SelectItem>
                      <SelectItem value="v3">SNMPv3</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* SNMPv1/v2c settings */}
                {(formData.snmp_version === 'v1' || formData.snmp_version === 'v2c') && (
                  <div className="grid grid-cols-2 gap-4">
                    <div className="grid gap-2">
                      <Label htmlFor="snmp_community">Community (чтение)</Label>
                      <Input
                        id="snmp_community"
                        value={formData.snmp_community}
                        onChange={(e) => updateField('snmp_community', e.target.value)}
                        placeholder="public"
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="snmp_write_community">Community (запись)</Label>
                      <Input
                        id="snmp_write_community"
                        value={formData.snmp_write_community}
                        onChange={(e) => updateField('snmp_write_community', e.target.value)}
                        placeholder="private"
                      />
                    </div>
                  </div>
                )}

                {/* SNMPv3 settings */}
                {formData.snmp_version === 'v3' && (
                  <>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="grid gap-2">
                        <Label htmlFor="snmpv3_user">Пользователь (User)</Label>
                        <Input
                          id="snmpv3_user"
                          value={formData.snmpv3_user}
                          onChange={(e) => updateField('snmpv3_user', e.target.value)}
                          placeholder="admin"
                        />
                      </div>

                      <div className="grid gap-2">
                        <Label htmlFor="snmpv3_security">Уровень безопасности</Label>
                        <Select
                          value={formData.snmpv3_security}
                          onValueChange={(v) => updateField('snmpv3_security', v)}
                        >
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="noAuthNoPriv">NoAuth, NoPriv</SelectItem>
                            <SelectItem value="authNoPriv">Auth, NoPriv</SelectItem>
                            <SelectItem value="authPriv">Auth + Priv</SelectItem>
                          </SelectContent>
                        </Select>
                      </div>
                    </div>

                    {/* Auth settings - show for authNoPriv and authPriv */}
                    {(formData.snmpv3_security === 'authNoPriv' || formData.snmpv3_security === 'authPriv') && (
                      <div className="grid grid-cols-2 gap-4">
                        <div className="grid gap-2">
                          <Label htmlFor="snmpv3_auth_proto">Протокол аутентификации</Label>
                          <Select
                            value={formData.snmpv3_auth_proto || 'MD5'}
                            onValueChange={(v) => updateField('snmpv3_auth_proto', v)}
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="MD5">MD5</SelectItem>
                              <SelectItem value="SHA">SHA</SelectItem>
                              <SelectItem value="SHA224">SHA-224</SelectItem>
                              <SelectItem value="SHA256">SHA-256</SelectItem>
                              <SelectItem value="SHA384">SHA-384</SelectItem>
                              <SelectItem value="SHA512">SHA-512</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>

                        <div className="grid gap-2">
                          <Label htmlFor="snmpv3_auth_pass">Пароль аутентификации</Label>
                          <Input
                            id="snmpv3_auth_pass"
                            type="password"
                            value={formData.snmpv3_auth_pass}
                            onChange={(e) => updateField('snmpv3_auth_pass', e.target.value)}
                            placeholder="Auth password"
                          />
                        </div>
                      </div>
                    )}

                    {/* Priv settings - show only for authPriv */}
                    {formData.snmpv3_security === 'authPriv' && (
                      <div className="grid grid-cols-2 gap-4">
                        <div className="grid gap-2">
                          <Label htmlFor="snmpv3_priv_proto">Протокол шифрования</Label>
                          <Select
                            value={formData.snmpv3_priv_proto || 'DES'}
                            onValueChange={(v) => updateField('snmpv3_priv_proto', v)}
                          >
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="DES">DES</SelectItem>
                              <SelectItem value="AES">AES-128</SelectItem>
                              <SelectItem value="AES192">AES-192</SelectItem>
                              <SelectItem value="AES256">AES-256</SelectItem>
                            </SelectContent>
                          </Select>
                        </div>

                        <div className="grid gap-2">
                          <Label htmlFor="snmpv3_priv_pass">Пароль шифрования</Label>
                          <Input
                            id="snmpv3_priv_pass"
                            type="password"
                            value={formData.snmpv3_priv_pass}
                            onChange={(e) => updateField('snmpv3_priv_pass', e.target.value)}
                            placeholder="Priv password"
                          />
                        </div>
                      </div>
                    )}
                  </>
                )}

                {/* Uplink settings for switches */}
                {getAvailableUplinkSwitches().length > 0 && (
                  <>
                    <Separator />
                    <div className="text-sm font-medium text-muted-foreground">
                      Оптический Uplink (SFP)
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="grid gap-2">
                        <Label htmlFor="uplink_switch">Родительский коммутатор</Label>
                        <Select
                          value={selectedUplinkSwitchId?.toString() || 'none'}
                          onValueChange={handleUplinkSwitchChange}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Выберите коммутатор..." />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">Не выбран</SelectItem>
                            {getAvailableUplinkSwitches().map((sw) => (
                              <SelectItem key={sw.device_id} value={sw.device_id.toString()}>
                                {sw.device_name} ({sw.ip_address})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="grid gap-2">
                        <Label htmlFor="uplink_port">SFP порт</Label>
                        <Select
                          value={formData.uplink_port_id?.toString() || 'none'}
                          onValueChange={handleUplinkPortChange}
                          disabled={!selectedUplinkSwitchId}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Выберите порт..." />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">Не выбран</SelectItem>
                            {getAvailableSfpPorts().map((port) => (
                              <SelectItem key={port.id} value={port.id.toString()}>
                                SFP {port.port_number} - {port.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    {selectedUplinkSwitchId && getAvailableSfpPorts().length === 0 && (
                      <div className="text-sm text-yellow-600 bg-yellow-500/10 p-2 rounded">
                        Нет свободных SFP портов на выбранном коммутаторе
                      </div>
                    )}
                  </>
                )}
              </>
            )}

            {/* Camera-specific fields */}
            {formData.type === 'camera' && switches.length > 0 && (
              <>
                <div className="text-sm font-medium text-muted-foreground">
                  Параметры камеры
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="onvif_port">ONVIF порт</Label>
                  <Input
                    id="onvif_port"
                    type="number"
                    value={formData.onvif_port}
                    onChange={(e) =>
                      updateField('onvif_port', parseInt(e.target.value) || 80)
                    }
                    placeholder="80"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="snapshot_url">Snapshot URL (опционально)</Label>
                  <Input
                    id="snapshot_url"
                    value={formData.snapshot_url || ''}
                    onChange={(e) => updateField('snapshot_url', e.target.value)}
                    placeholder="/ISAPI/Streaming/channels/101/picture или http://ip/snapshot.jpg"
                  />
                  <p className="text-xs text-muted-foreground">
                    Если ONVIF не работает, укажите URL вручную. Можно путь или полный URL.
                  </p>
                </div>
              </>
            )}

            {/* Server-specific fields */}
            {formData.type === 'server' && (
              <>
                <Separator />
                <div className="text-sm font-medium text-muted-foreground">
                  Параметры сервера
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="tcp_ports">TCP порты (JSON)</Label>
                  <Input
                    id="tcp_ports"
                    value={formData.tcp_ports}
                    onChange={(e) => updateField('tcp_ports', e.target.value)}
                    placeholder='[22, 80, 443]'
                  />
                </div>

                {/* Uplink settings for servers */}
                {switches.length > 0 && (
                  <>
                    <Separator />
                    <div className="text-sm font-medium text-muted-foreground">
                      Оптический Uplink (SFP)
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="grid gap-2">
                        <Label htmlFor="uplink_switch">Коммутатор</Label>
                        <Select
                          value={selectedUplinkSwitchId?.toString() || 'none'}
                          onValueChange={handleUplinkSwitchChange}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Выберите коммутатор..." />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">Не выбран</SelectItem>
                            {switches.map((sw) => (
                              <SelectItem key={sw.device_id} value={sw.device_id.toString()}>
                                {sw.device_name} ({sw.ip_address})
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>

                      <div className="grid gap-2">
                        <Label htmlFor="uplink_port">SFP порт</Label>
                        <Select
                          value={formData.uplink_port_id?.toString() || 'none'}
                          onValueChange={handleUplinkPortChange}
                          disabled={!selectedUplinkSwitchId}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder="Выберите порт..." />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="none">Не выбран</SelectItem>
                            {getAvailableSfpPorts().map((port) => (
                              <SelectItem key={port.id} value={port.id.toString()}>
                                SFP {port.port_number} - {port.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </div>
                    {selectedUplinkSwitchId && getAvailableSfpPorts().length === 0 && (
                      <div className="text-sm text-yellow-600 bg-yellow-500/10 p-2 rounded">
                        Нет свободных SFP портов на выбранном коммутаторе
                      </div>
                    )}
                  </>
                )}
              </>
            )}

            {error && (
              <div className="text-sm text-destructive">{error}</div>
            )}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Отмена
            </Button>
            <Button
              type="submit"
              disabled={isLoading || (formData.type === 'camera' && switches.length === 0)}
            >
              {isLoading ? 'Сохранение...' : isEditing ? 'Сохранить' : 'Добавить'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
