import { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Plus, Monitor, Server, Camera, Network, AlertCircle, Play, Square, RefreshCw, Search } from 'lucide-react'
import { DeviceForm } from '@/components/devices/DeviceForm'
import { DeviceList } from '@/components/devices/DeviceList'
import { DeviceDetailsPanel } from '@/components/device/DeviceDetailsPanel'
import { Pagination } from '@/components/ui/pagination'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  GetDevicesPaginated,
  GetDevice,
  GetDeviceStats,
  CreateDevice,
  UpdateDevice,
  DeleteDevice,
  GetCredentials,
  GetCameraPort,
} from '../../wailsjs/go/main/App'
import { useMonitoring, useDeviceStatusEvents } from '@/hooks/useMonitoring'
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

interface Credential {
  id: number
  name: string
  type: string
}

interface DeviceStats {
  total: number
  switch: number
  server: number
  camera: number
  online: number
  offline: number
}

interface DevicesPageProps {
  filterType?: 'switch' | 'server' | 'camera'
}

export function DevicesPage({ filterType }: DevicesPageProps) {
  const { t } = useTranslation()
  const [devices, setDevices] = useState<Device[]>([])
  const [credentials, setCredentials] = useState<Credential[]>([])
  const [stats, setStats] = useState<DeviceStats>({
    total: 0,
    switch: 0,
    server: 0,
    camera: 0,
    online: 0,
    offline: 0,
  })
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [totalItems, setTotalItems] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  // Search state
  const [searchQuery, setSearchQuery] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')

  // Form state
  const [formOpen, setFormOpen] = useState(false)
  const [editingDevice, setEditingDevice] = useState<Partial<Device> | null>(null)
  const [isSaving, setIsSaving] = useState(false)

  // Delete confirmation
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)

  // Device details view
  const [viewingDeviceId, setViewingDeviceId] = useState<number | null>(null)

  // Monitoring
  const { status: monitoringStatus, start, stop, runOnce, isLoading: isMonitoringLoading } = useMonitoring()

  // Listen for device status changes
  useDeviceStatusEvents(
    useCallback((event) => {
      setDevices((prev) =>
        prev.map((d) =>
          d.id === event.device_id ? { ...d, status: event.new_status } : d
        )
      )
      // Update stats
      setStats((prev) => {
        const wasOnline = event.old_status === 'online'
        const isOnline = event.new_status === 'online'
        if (wasOnline && !isOnline) {
          return { ...prev, online: prev.online - 1, offline: prev.offline + 1 }
        } else if (!wasOnline && isOnline) {
          return { ...prev, online: prev.online + 1, offline: prev.offline - 1 }
        }
        return prev
      })
    }, [])
  )

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchQuery !== debouncedSearch) {
        setDebouncedSearch(searchQuery)
        setCurrentPage(1) // Reset to first page on search
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [searchQuery, debouncedSearch])

  const loadData = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)

      const [devicesResult, statsData, credsData] = await Promise.all([
        GetDevicesPaginated({
          type: filterType || '',
          status: '',
          search: debouncedSearch,
          page: currentPage,
          page_size: pageSize,
          sort_by: '',
          sort_order: '',
        }),
        GetDeviceStats(),
        GetCredentials(),
      ])

      setDevices(devicesResult?.devices || [])
      setTotalItems(devicesResult?.total || 0)
      setTotalPages(devicesResult?.total_pages || 1)

      setStats({
        total: statsData?.total || 0,
        switch: statsData?.switch || 0,
        server: statsData?.server || 0,
        camera: statsData?.camera || 0,
        online: statsData?.online || 0,
        offline: statsData?.offline || 0,
      })
      setCredentials(credsData || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : t('errors.loadingData'))
    } finally {
      setIsLoading(false)
    }
  }, [filterType, currentPage, pageSize, debouncedSearch, t])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handlePageChange = (page: number) => {
    setCurrentPage(page)
  }

  const handlePageSizeChange = (size: number) => {
    setPageSize(size)
    setCurrentPage(1)
  }

  const handleCreate = () => {
    setEditingDevice(null)
    setFormOpen(true)
  }

  const handleEdit = async (device: Device) => {
    try {
      // Load full device details including type-specific data
      const fullDevice = await GetDevice(device.id)
      if (fullDevice) {
        const formData: Partial<DeviceFormData> = {
          id: fullDevice.id,
          name: fullDevice.name,
          ip_address: fullDevice.ip_address,
          type: fullDevice.type,
          manufacturer: fullDevice.manufacturer || '',
          model: fullDevice.model,
          credential_id: fullDevice.credential_id,
        }

        // Add type-specific fields
        if (fullDevice.switch) {
          formData.snmp_community = fullDevice.switch.snmp_community
          formData.snmp_version = fullDevice.switch.snmp_version
          formData.port_count = fullDevice.switch.port_count
          formData.sfp_port_count = fullDevice.switch.sfp_port_count || 0
          // SNMPv3 fields
          formData.snmpv3_user = fullDevice.switch.snmpv3_user || ''
          formData.snmpv3_security = fullDevice.switch.snmpv3_security || 'noAuthNoPriv'
          formData.snmpv3_auth_proto = fullDevice.switch.snmpv3_auth_proto || ''
          formData.snmpv3_auth_pass = fullDevice.switch.snmpv3_auth_pass || ''
          formData.snmpv3_priv_proto = fullDevice.switch.snmpv3_priv_proto || ''
          formData.snmpv3_priv_pass = fullDevice.switch.snmpv3_priv_pass || ''
        }
        if (fullDevice.camera) {
          formData.rtsp_url = fullDevice.camera.rtsp_url
          formData.onvif_port = fullDevice.camera.onvif_port
          formData.snapshot_url = fullDevice.camera.snapshot_url
          formData.stream_type = fullDevice.camera.stream_type
          // Find which port this camera is linked to
          const portId = await GetCameraPort(fullDevice.id)
          if (portId) {
            formData.switch_port_id = portId
          }
        }
        if (fullDevice.server) {
          formData.tcp_ports = fullDevice.server.tcp_ports
          formData.use_snmp = fullDevice.server.use_snmp
        }

        setEditingDevice(formData as Partial<Device>)
      } else {
        setEditingDevice(device)
      }
    } catch (err) {
      console.error('Failed to load device details:', err)
      setEditingDevice(device)
    }
    setFormOpen(true)
  }

  interface DeviceFormData {
    id?: number
    name: string
    ip_address: string
    type: string
    manufacturer: string
    model: string
    credential_id?: number
    snmp_community: string
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
    rtsp_url: string
    onvif_port: number
    snapshot_url: string
    stream_type: string
    switch_port_id?: number
    tcp_ports: string
    use_snmp: boolean
  }

  const handleViewDetails = (device: Device) => {
    setViewingDeviceId(device.id)
  }

  const handleFormSubmit = async (data: DeviceFormData) => {
    setIsSaving(true)
    try {
      if (editingDevice?.id) {
        await UpdateDevice({ ...data, id: editingDevice.id } as never)
      } else {
        await CreateDevice(data as never)
      }
      setFormOpen(false)
      await loadData()
    } finally {
      setIsSaving(false)
    }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteId) return

    setIsDeleting(true)
    try {
      await DeleteDevice(deleteId)
      setDeleteId(null)
      await loadData()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('errors.deletingData'))
    } finally {
      setIsDeleting(false)
    }
  }

  // Show device details panel if viewing a device
  if (viewingDeviceId !== null) {
    return (
      <DeviceDetailsPanel
        deviceId={viewingDeviceId}
        onBack={() => setViewingDeviceId(null)}
      />
    )
  }

  return (
    <div className="space-y-6">
      {/* Error message */}
      {error && (
        <div className="flex items-center gap-2 p-4 text-destructive bg-destructive/10 rounded-lg">
          <AlertCircle className="h-5 w-5" />
          <span>{error}</span>
        </div>
      )}

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title={t('devices.totalDevices')}
          value={stats.total.toString()}
          icon={<Monitor className="h-5 w-5" />}
          description={`${stats.online} ${t('devices.inNetwork')}`}
        />
        <StatsCard
          title={t('devices.stats.switches')}
          value={stats.switch.toString()}
          icon={<Network className="h-5 w-5" />}
          description={t('devices.stats.switchesDesc')}
          color="text-blue-500"
        />
        <StatsCard
          title={t('devices.stats.servers')}
          value={stats.server.toString()}
          icon={<Server className="h-5 w-5" />}
          description={t('devices.stats.serversDesc')}
          color="text-green-500"
        />
        <StatsCard
          title={t('devices.stats.cameras')}
          value={stats.camera.toString()}
          icon={<Camera className="h-5 w-5" />}
          description={t('devices.stats.camerasDesc')}
          color="text-purple-500"
        />
      </div>

      {/* Devices List */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>{t('devices.title')}</CardTitle>
            <CardDescription>
              {t('devices.subtitle')}
              {monitoringStatus.running && (
                <span className="ml-2 text-green-500">● {t('settings.monitoring.running')}</span>
              )}
            </CardDescription>
          </div>
          <div className="flex gap-2 items-center">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                type="search"
                placeholder={t('common.search') as string}
                className="pl-8 w-[200px]"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>
            {monitoringStatus.running ? (
              <Button
                variant="outline"
                onClick={stop}
                disabled={isMonitoringLoading}
              >
                <Square className="h-4 w-4 mr-2" />
                {t('settings.monitoring.stop')}
              </Button>
            ) : (
              <Button
                variant="outline"
                onClick={start}
                disabled={isMonitoringLoading}
              >
                <Play className="h-4 w-4 mr-2" />
                {t('settings.monitoring.start')}
              </Button>
            )}
            <Button
              variant="outline"
              onClick={runOnce}
              disabled={isMonitoringLoading}
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              {t('settings.monitoring.check')}
            </Button>
            <Button onClick={handleCreate}>
              <Plus className="h-4 w-4 mr-2" />
              {t('devices.addDevice')}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <DeviceList
            devices={devices}
            onEdit={handleEdit}
            onDelete={(id) => setDeleteId(id)}
            onView={handleViewDetails}
            isLoading={isLoading}
          />
          {/* Pagination */}
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            pageSize={pageSize}
            totalItems={totalItems}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
          />
        </CardContent>
      </Card>

      {/* Add/Edit Form */}
      <DeviceForm
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSubmit={handleFormSubmit}
        initialData={editingDevice || undefined}
        credentials={credentials}
        isLoading={isSaving}
      />

      {/* Delete Confirmation */}
      <Dialog open={deleteId !== null} onOpenChange={() => setDeleteId(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('devices.deleteConfirm')}</DialogTitle>
            <DialogDescription>
              {t('devices.deleteWarning')}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteId(null)}>
              {t('common.cancel')}
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteConfirm}
              disabled={isDeleting}
            >
              {isDeleting ? t('common.loading') : t('common.delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

interface StatsCardProps {
  title: string
  value: string
  icon: React.ReactNode
  description: string
  color?: string
}

function StatsCard({
  title,
  value,
  icon,
  description,
  color = 'text-primary',
}: StatsCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <div className={color}>{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <p className="text-xs text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  )
}
