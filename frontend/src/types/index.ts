// Device types
export interface Device {
  id: number
  name: string
  ip_address: string
  type: string
  manufacturer: string
  model: string
  status: string
  credential_id?: number
  last_check?: string
  created_at?: string
  updated_at?: string
}

// Schema types
export interface Schema {
  id: number
  name: string
  background_image?: string
  created_at?: string
}

export interface SchemaItem {
  id: number
  device_id: number
  schema_id: number
  x: number
  y: number
  width: number
  height: number
  // Device info (joined from devices table)
  device_name?: string
  device_type?: string
  device_status?: string
  device_ip?: string
}

// Event types
export interface Event {
  id: number
  device_id?: number
  type: string
  level: string
  message: string
  created_at: string
}

// Credential types
export interface Credential {
  id: number
  name: string
  type: string
  username?: string
  created_at?: string
}

// Switch types
export interface Switch {
  id: number
  device_id: number
  snmp_community: string
  snmp_version: string
  port_count: number
  sfp_port_count: number
}

export interface SwitchPort {
  id: number
  switch_id: number
  port_number: number
  name: string
  status: string
  speed?: string
  port_type: string  // "copper" or "sfp"
  linked_camera_id?: number  // Only for copper ports
  linked_switch_id?: number  // Only for SFP ports (uplink)
}

// Camera types
export interface Camera {
  id: number
  device_id: number
  rtsp_url: string
  onvif_port: number
  snapshot_url?: string
  stream_type: string
}

// Server types
export interface Server {
  id: number
  device_id: number
  tcp_ports: string
  use_snmp: boolean
}

// Stats types
export interface DeviceStats {
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

export interface LatencyPoint {
  timestamp: string
  latency: number
  status: string
}

// Settings types
export interface AppSettings {
  theme: string
  monitoring_interval: number
  sound_enabled: boolean
  notifications_enabled: boolean
  ping_timeout: number
  snmp_timeout: number
  camera_stream_type: string
}
