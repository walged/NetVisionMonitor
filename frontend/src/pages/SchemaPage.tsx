import { useState, useEffect, useCallback, useMemo } from 'react'
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
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Upload,
  Plus,
  LayoutGrid,
  Edit,
  Trash2,
  Save,
  Monitor,
  Server,
  Camera,
  Network,
  Link2,
  Eye,
  EyeOff,
  GripVertical,
} from 'lucide-react'
import { SchemaCanvas } from '@/components/schema/SchemaCanvas'
import { DevicePopup } from '@/components/schema/DevicePopup'
import {
  GetSchemas,
  GetSchemaItems,
  CreateSchema,
  UpdateSchema,
  DeleteSchema,
  AddDeviceToSchema,
  UpdateSchemaItemPosition,
  RemoveDeviceFromSchema,
  SelectBackgroundImage,
  GetBackgroundImage,
  GetDevices,
  GetSwitchesWithPorts,
} from '../../wailsjs/go/main/App'
import { useDeviceStatusEvents } from '@/hooks/useMonitoring'
import type { Schema, SchemaItem, Device } from '@/types'

interface SwitchPort {
  id: number
  switch_id: number
  port_number: number
  port_type: string
  linked_camera_id?: number
  linked_switch_id?: number
}

interface SwitchWithPorts {
  device_id: number
  device_name: string
  ports: SwitchPort[]
}

interface LinkedCamera {
  id: number
  name: string
  status: string
  port_number: number
}

interface ConnectionLine {
  fromId: number
  toId: number
  fromX: number
  fromY: number
  toX: number
  toY: number
  type: 'camera' | 'uplink'  // Type of connection
}

export function SchemaPage() {
  const [schemas, setSchemas] = useState<Schema[]>([])
  const [activeSchema, setActiveSchema] = useState<Schema | null>(null)
  const [items, setItems] = useState<SchemaItem[]>([])
  const [backgroundImage, setBackgroundImage] = useState<string>('')
  const [devices, setDevices] = useState<Device[]>([])
  const [switchesWithPorts, setSwitchesWithPorts] = useState<SwitchWithPorts[]>([])
  const [editMode, setEditMode] = useState(false)
  const [selectedItem, setSelectedItem] = useState<SchemaItem | null>(null)
  const [popupItem, setPopupItem] = useState<SchemaItem | null>(null)

  // Filters
  const [showSwitches, setShowSwitches] = useState(true)
  const [showServers, setShowServers] = useState(true)
  const [showCameras, setShowCameras] = useState(true)
  const [showConnections, setShowConnections] = useState(true)

  // Dialog states
  const [schemaDialogOpen, setSchemaDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [newSchemaName, setNewSchemaName] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  // Load schemas
  const loadSchemas = useCallback(async () => {
    try {
      const data = await GetSchemas()
      setSchemas(data || [])
      if (data && data.length > 0 && !activeSchema) {
        setActiveSchema(data[0])
      }
    } catch (err) {
      console.error('Failed to load schemas:', err)
    }
  }, [activeSchema])

  // Load schema items
  const loadItems = useCallback(async () => {
    if (!activeSchema) {
      setItems([])
      setBackgroundImage('')
      return
    }

    try {
      const itemsData = await GetSchemaItems(activeSchema.id)
      setItems(itemsData || [])

      if (activeSchema.background_image) {
        const bg = await GetBackgroundImage(activeSchema.background_image)
        setBackgroundImage(bg)
      } else {
        setBackgroundImage('')
      }
    } catch (err) {
      console.error('Failed to load schema items:', err)
    }
  }, [activeSchema])

  // Load devices for adding to schema
  const loadDevices = useCallback(async () => {
    try {
      const [devicesData, switchesData] = await Promise.all([
        GetDevices(),
        GetSwitchesWithPorts(),
      ])
      setDevices(devicesData || [])
      setSwitchesWithPorts(switchesData || [])
    } catch (err) {
      console.error('Failed to load devices:', err)
    }
  }, [])

  // Get linked cameras for a switch
  const getLinkedCameras = useCallback((switchDeviceId: number): LinkedCamera[] => {
    const sw = switchesWithPorts.find((s) => s.device_id === switchDeviceId)
    if (!sw || !sw.ports) return []

    const linkedCameras: LinkedCamera[] = []
    for (const port of sw.ports) {
      if (port.linked_camera_id) {
        const camera = devices.find((d) => d.id === port.linked_camera_id)
        if (camera) {
          linkedCameras.push({
            id: camera.id,
            name: camera.name,
            status: camera.status,
            port_number: port.port_number,
          })
        }
      }
    }
    return linkedCameras
  }, [switchesWithPorts, devices])

  // Calculate connection lines between switches and their cameras/uplinks
  const connections = useMemo((): ConnectionLine[] => {
    const lines: ConnectionLine[] = []
    const addedUplinks = new Set<string>() // Track added uplinks to avoid duplicates

    for (const sw of switchesWithPorts) {
      const switchItem = items.find(i => i.device_id === sw.device_id)
      if (!switchItem || !sw.ports) continue

      for (const port of sw.ports) {
        // Camera connections (copper ports)
        if (port.linked_camera_id) {
          const cameraItem = items.find(i => i.device_id === port.linked_camera_id)
          if (cameraItem) {
            lines.push({
              fromId: sw.device_id,
              toId: port.linked_camera_id,
              fromX: switchItem.x + (switchItem.width || 70) / 2,
              fromY: switchItem.y + (switchItem.height || 70) / 2,
              toX: cameraItem.x + (cameraItem.width || 70) / 2,
              toY: cameraItem.y + (cameraItem.height || 70) / 2,
              type: 'camera',
            })
          }
        }

        // Uplink connections (SFP ports)
        if (port.linked_switch_id && port.port_type === 'sfp') {
          const linkedSwitchItem = items.find(i => i.device_id === port.linked_switch_id)
          if (linkedSwitchItem) {
            // Create a unique key for this connection (sorted to avoid duplicates)
            const key = [sw.device_id, port.linked_switch_id].sort().join('-')
            if (!addedUplinks.has(key)) {
              addedUplinks.add(key)
              lines.push({
                fromId: sw.device_id,
                toId: port.linked_switch_id,
                fromX: switchItem.x + (switchItem.width || 70) / 2,
                fromY: switchItem.y + (switchItem.height || 70) / 2,
                toX: linkedSwitchItem.x + (linkedSwitchItem.width || 70) / 2,
                toY: linkedSwitchItem.y + (linkedSwitchItem.height || 70) / 2,
                type: 'uplink',
              })
            }
          }
        }
      }
    }

    return lines
  }, [items, switchesWithPorts])

  useEffect(() => {
    loadSchemas()
    loadDevices()
  }, [loadSchemas, loadDevices])

  useEffect(() => {
    loadItems()
    // Reload devices/switches data when items change to keep connections updated
    loadDevices()
  }, [loadItems, loadDevices])

  // Listen for device status changes
  useDeviceStatusEvents(
    useCallback((event) => {
      setItems((prev) =>
        prev.map((item) =>
          item.device_id === event.device_id
            ? { ...item, device_status: event.new_status }
            : item
        )
      )
      setDevices((prev) =>
        prev.map((d) =>
          d.id === event.device_id
            ? { ...d, status: event.new_status }
            : d
        )
      )
    }, [])
  )

  // Create new schema
  const handleCreateSchema = async () => {
    if (!newSchemaName.trim()) return

    setIsLoading(true)
    try {
      const schema = await CreateSchema({
        name: newSchemaName,
        background_image: backgroundImage,
      })
      if (schema) {
        setSchemas((prev) => [...prev, schema])
        setActiveSchema(schema)
        setSchemaDialogOpen(false)
        setNewSchemaName('')
        setBackgroundImage('')
      }
    } catch (err) {
      console.error('Failed to create schema:', err)
    } finally {
      setIsLoading(false)
    }
  }

  // Delete schema
  const handleDeleteSchema = async () => {
    if (!activeSchema) return

    setIsLoading(true)
    try {
      await DeleteSchema(activeSchema.id)
      setSchemas((prev) => prev.filter((s) => s.id !== activeSchema.id))
      setActiveSchema(schemas.length > 1 ? schemas[0] : null)
      setDeleteDialogOpen(false)
    } catch (err) {
      console.error('Failed to delete schema:', err)
    } finally {
      setIsLoading(false)
    }
  }

  // Select background image
  const handleSelectBackground = async () => {
    try {
      const base64 = await SelectBackgroundImage()
      if (base64) {
        setBackgroundImage(base64)

        // Update existing schema
        if (activeSchema) {
          await UpdateSchema({
            id: activeSchema.id,
            name: activeSchema.name,
            background_image: base64,
          })
          setActiveSchema({ ...activeSchema, background_image: base64 })
        }
      }
    } catch (err) {
      console.error('Failed to select background:', err)
    }
  }

  // Add device to schema (with linked cameras for switches)
  const handleAddDevice = async (deviceId: number) => {
    if (!activeSchema) return

    setIsLoading(true)
    try {
      // Find the device to check if it's a switch
      const device = devices.find(d => d.id === deviceId)
      const baseX = 100 + Math.random() * 200
      const baseY = 100 + Math.random() * 200

      // Add the main device (switch/server)
      const item = await AddDeviceToSchema({
        device_id: deviceId,
        schema_id: activeSchema.id,
        x: baseX,
        y: baseY,
        width: 70,
        height: 70,
      })

      // If it's a switch, also add all linked cameras
      if (device?.type === 'switch') {
        const linkedCameras = getLinkedCameras(deviceId)
        const cameraCount = linkedCameras.length

        for (let i = 0; i < linkedCameras.length; i++) {
          const cam = linkedCameras[i]
          // Check if camera is not already on schema
          if (!items.some(itm => itm.device_id === cam.id)) {
            // Position cameras in a circle around the switch
            const angle = (2 * Math.PI * i) / cameraCount - Math.PI / 2
            const radius = 150
            const camX = baseX + Math.cos(angle) * radius
            const camY = baseY + Math.sin(angle) * radius

            await AddDeviceToSchema({
              device_id: cam.id,
              schema_id: activeSchema.id,
              x: Math.max(0, camX),
              y: Math.max(0, camY),
              width: 70,
              height: 70,
            })
          }
        }
      }

      if (item) {
        await loadItems()
      }
    } catch (err) {
      console.error('Failed to add device:', err)
    } finally {
      setIsLoading(false)
    }
  }

  // Move item
  const handleItemMove = async (itemId: number, x: number, y: number) => {
    setItems((prev) =>
      prev.map((item) => (item.id === itemId ? { ...item, x, y } : item))
    )

    try {
      await UpdateSchemaItemPosition(itemId, x, y)
    } catch (err) {
      console.error('Failed to update item position:', err)
    }
  }

  // Remove item from schema
  const handleItemRemove = async (itemId: number) => {
    try {
      await RemoveDeviceFromSchema(itemId)
      setItems((prev) => prev.filter((item) => item.id !== itemId))
      setSelectedItem(null)
    } catch (err) {
      console.error('Failed to remove item:', err)
    }
  }

  // Handle item click
  const handleItemClick = (item: SchemaItem) => {
    if (editMode) {
      setSelectedItem(item)
      setPopupItem(null)
    } else {
      setPopupItem(item)
      setSelectedItem(null)
    }
  }

  // Get available devices (not already on schema)
  const availableDevices = devices.filter(
    (d) => !items.some((item) => item.device_id === d.id)
  )

  // Filter items based on visibility settings
  const filteredItems = items.filter((item) => {
    if (item.device_type === 'switch' && !showSwitches) return false
    if (item.device_type === 'server' && !showServers) return false
    if (item.device_type === 'camera' && !showCameras) return false
    return true
  })

  const getDeviceIcon = (type: string) => {
    switch (type) {
      case 'switch':
        return <Network className="h-4 w-4 text-blue-500" />
      case 'server':
        return <Server className="h-4 w-4 text-green-500" />
      case 'camera':
        return <Camera className="h-4 w-4 text-purple-500" />
      default:
        return <Monitor className="h-4 w-4" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return 'bg-green-500'
      case 'offline':
        return 'bg-red-500'
      default:
        return 'bg-gray-400'
    }
  }

  // Group available devices by type
  const groupedDevices = useMemo(() => {
    const switches = availableDevices.filter(d => d.type === 'switch')
    const servers = availableDevices.filter(d => d.type === 'server')
    const cameras = availableDevices.filter(d => d.type === 'camera')
    return { switches, servers, cameras }
  }, [availableDevices])

  return (
    <div className="flex gap-4 h-[calc(100vh-180px)]">
      {/* Left sidebar - Devices panel */}
      {editMode && (
        <Card className="w-64 flex-shrink-0 flex flex-col">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Устройства</CardTitle>
            <CardDescription className="text-xs">
              Кликните чтобы добавить
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 p-2 overflow-hidden">
            <ScrollArea className="h-full pr-2">
              {/* Switches */}
              {groupedDevices.switches.length > 0 && (
                <div className="mb-4">
                  <div className="flex items-center gap-2 mb-2 px-1">
                    <Network className="h-4 w-4 text-blue-500" />
                    <span className="text-xs font-medium">Коммутаторы</span>
                    <Badge variant="secondary" className="text-xs ml-auto">
                      {groupedDevices.switches.length}
                    </Badge>
                  </div>
                  <div className="space-y-1">
                    {groupedDevices.switches.map((d) => (
                      <button
                        key={d.id}
                        onClick={() => handleAddDevice(d.id)}
                        className="w-full flex items-center gap-2 p-2 rounded-lg hover:bg-accent text-left text-sm transition-colors"
                        disabled={isLoading}
                      >
                        <div className={`w-2 h-2 rounded-full ${getStatusColor(d.status)}`} />
                        <span className="truncate flex-1">{d.name}</span>
                        <Plus className="h-3 w-3 opacity-50" />
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Servers */}
              {groupedDevices.servers.length > 0 && (
                <div className="mb-4">
                  <div className="flex items-center gap-2 mb-2 px-1">
                    <Server className="h-4 w-4 text-green-500" />
                    <span className="text-xs font-medium">Серверы</span>
                    <Badge variant="secondary" className="text-xs ml-auto">
                      {groupedDevices.servers.length}
                    </Badge>
                  </div>
                  <div className="space-y-1">
                    {groupedDevices.servers.map((d) => (
                      <button
                        key={d.id}
                        onClick={() => handleAddDevice(d.id)}
                        className="w-full flex items-center gap-2 p-2 rounded-lg hover:bg-accent text-left text-sm transition-colors"
                        disabled={isLoading}
                      >
                        <div className={`w-2 h-2 rounded-full ${getStatusColor(d.status)}`} />
                        <span className="truncate flex-1">{d.name}</span>
                        <Plus className="h-3 w-3 opacity-50" />
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Info about cameras */}
              <div className="text-xs text-muted-foreground px-1 py-2 border-t mt-2">
                <Camera className="h-3 w-3 inline mr-1" />
                Камеры добавляются автоматически вместе с коммутатором
              </div>

              {groupedDevices.switches.length === 0 && groupedDevices.servers.length === 0 && (
                <div className="text-center text-muted-foreground text-sm py-8">
                  Все устройства добавлены
                </div>
              )}
            </ScrollArea>
          </CardContent>
        </Card>
      )}

      {/* Main content */}
      <Card className="flex-1 flex flex-col min-w-0">
        <CardHeader className="flex flex-row items-center justify-between py-3">
          <div>
            <CardTitle className="text-lg">Схема размещения</CardTitle>
            <CardDescription className="text-xs">
              {activeSchema ? activeSchema.name : 'Выберите или создайте схему'}
            </CardDescription>
          </div>
          <div className="flex gap-2">
            {schemas.length > 0 && (
              <Select
                value={activeSchema?.id.toString() || ''}
                onValueChange={(v) => {
                  const schema = schemas.find((s) => s.id === parseInt(v))
                  if (schema) setActiveSchema(schema)
                }}
              >
                <SelectTrigger className="w-[160px] h-9">
                  <SelectValue placeholder="Выберите схему" />
                </SelectTrigger>
                <SelectContent>
                  {schemas.map((s) => (
                    <SelectItem key={s.id} value={s.id.toString()}>
                      {s.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}

            {activeSchema && (
              <>
                <Button variant="outline" size="sm" onClick={handleSelectBackground}>
                  <Upload className="h-4 w-4 mr-1" />
                  Фон
                </Button>
                <Button
                  variant={editMode ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => {
                    setEditMode(!editMode)
                    setSelectedItem(null)
                    setPopupItem(null)
                  }}
                >
                  {editMode ? (
                    <>
                      <Save className="h-4 w-4 mr-1" />
                      Готово
                    </>
                  ) : (
                    <>
                      <Edit className="h-4 w-4 mr-1" />
                      Редактировать
                    </>
                  )}
                </Button>
                {editMode && (
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => setDeleteDialogOpen(true)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </>
            )}

            <Button size="sm" onClick={() => setSchemaDialogOpen(true)}>
              <Plus className="h-4 w-4 mr-1" />
              Новая
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex-1 relative p-2 overflow-hidden">
          {activeSchema ? (
            <div className="h-full relative">
              <SchemaCanvas
                backgroundImage={backgroundImage}
                items={filteredItems}
                allItems={items}
                connections={showConnections ? connections : []}
                onItemMove={handleItemMove}
                onItemClick={handleItemClick}
                onItemRemove={handleItemRemove}
                selectedItemId={selectedItem?.id}
                editMode={editMode}
                showConnections={showConnections}
              />

              {/* Device popup */}
              {popupItem && (
                <div className="absolute top-4 left-4 z-20">
                  <DevicePopup
                    item={popupItem}
                    onClose={() => setPopupItem(null)}
                    linkedCameras={
                      popupItem.device_type === 'switch'
                        ? getLinkedCameras(popupItem.device_id)
                        : undefined
                    }
                  />
                </div>
              )}

              {/* Legend & Filters */}
              <div className="absolute bottom-2 right-2 bg-white dark:bg-zinc-900 border rounded-lg p-3 shadow-lg">
                <div className="text-xs font-medium mb-2">Фильтры и легенда</div>
                <div className="space-y-2">
                  <label className="flex items-center gap-2 text-xs cursor-pointer">
                    <Checkbox
                      checked={showSwitches}
                      onCheckedChange={(c: boolean | 'indeterminate') => setShowSwitches(!!c)}
                    />
                    <Network className="h-3.5 w-3.5 text-blue-500" />
                    <span>Коммутаторы</span>
                  </label>
                  <label className="flex items-center gap-2 text-xs cursor-pointer">
                    <Checkbox
                      checked={showServers}
                      onCheckedChange={(c: boolean | 'indeterminate') => setShowServers(!!c)}
                    />
                    <Server className="h-3.5 w-3.5 text-green-500" />
                    <span>Серверы</span>
                  </label>
                  <label className="flex items-center gap-2 text-xs cursor-pointer">
                    <Checkbox
                      checked={showCameras}
                      onCheckedChange={(c: boolean | 'indeterminate') => setShowCameras(!!c)}
                    />
                    <Camera className="h-3.5 w-3.5 text-purple-500" />
                    <span>Камеры</span>
                  </label>
                  <div className="border-t pt-2 mt-2">
                    <label className="flex items-center gap-2 text-xs cursor-pointer">
                      <Checkbox
                        checked={showConnections}
                        onCheckedChange={(c: boolean | 'indeterminate') => setShowConnections(!!c)}
                      />
                      <Link2 className="h-3.5 w-3.5" />
                      <span>Связи</span>
                    </label>
                  </div>
                  <div className="border-t pt-2 mt-2 space-y-1">
                    <div className="flex items-center gap-2 text-xs">
                      <div className="w-2.5 h-2.5 rounded-full bg-green-500" />
                      <span>В сети</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                      <div className="w-2.5 h-2.5 rounded-full bg-red-500" />
                      <span>Не в сети</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                      <div className="w-2.5 h-2.5 rounded-full bg-gray-400" />
                      <span>Неизвестно</span>
                    </div>
                  </div>
                  <div className="border-t pt-2 mt-2 space-y-1">
                    <div className="flex items-center gap-2 text-xs">
                      <div className="w-6 h-0.5 bg-blue-500 opacity-70" style={{ borderStyle: 'dashed' }} />
                      <span>Камера</span>
                    </div>
                    <div className="flex items-center gap-2 text-xs">
                      <div className="w-6 h-1 bg-amber-500 opacity-90 rounded-full" />
                      <span>SFP Uplink</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground border-2 border-dashed rounded-lg">
              <LayoutGrid className="h-16 w-16 mb-4 opacity-50" />
              <p className="text-lg font-medium">Нет активной схемы</p>
              <p className="text-sm mb-4">
                Создайте новую схему или выберите существующую
              </p>
              <Button onClick={() => setSchemaDialogOpen(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Создать схему
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {/* New Schema Dialog */}
      <Dialog open={schemaDialogOpen} onOpenChange={setSchemaDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Новая схема</DialogTitle>
            <DialogDescription>
              Создайте новую схему размещения устройств
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="schema-name">Название</Label>
              <Input
                id="schema-name"
                value={newSchemaName}
                onChange={(e) => setNewSchemaName(e.target.value)}
                placeholder="План этажа 1"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setSchemaDialogOpen(false)}>
              Отмена
            </Button>
            <Button onClick={handleCreateSchema} disabled={isLoading}>
              {isLoading ? 'Создание...' : 'Создать'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Schema Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Удалить схему?</DialogTitle>
            <DialogDescription>
              Схема "{activeSchema?.name}" и все размещённые на ней устройства
              будут удалены. Это действие нельзя отменить.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Отмена
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteSchema}
              disabled={isLoading}
            >
              {isLoading ? 'Удаление...' : 'Удалить'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
