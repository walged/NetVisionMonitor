import { useState, useRef, useEffect, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { Network, Server, Camera, Monitor } from 'lucide-react'
import type { SchemaItem } from '@/types'

interface ConnectionLine {
  fromId: number
  toId: number
  fromX: number
  fromY: number
  toX: number
  toY: number
  type?: 'camera' | 'uplink'
}

interface SchemaCanvasProps {
  backgroundImage?: string
  items: SchemaItem[]
  allItems?: SchemaItem[]  // Full items array for connection lookup
  connections: ConnectionLine[]
  onItemMove: (itemId: number, x: number, y: number) => void
  onItemClick: (item: SchemaItem) => void
  onItemRemove: (itemId: number) => void
  onCanvasClick?: (x: number, y: number) => void
  selectedItemId?: number | null
  editMode?: boolean
  showConnections?: boolean
}

export function SchemaCanvas({
  backgroundImage,
  items,
  allItems,
  connections,
  onItemMove,
  onItemClick,
  onItemRemove,
  onCanvasClick,
  selectedItemId,
  editMode = false,
  showConnections = true,
}: SchemaCanvasProps) {
  // Use allItems for connection lookup if provided, otherwise fall back to items
  const itemsForConnections = allItems || items
  const containerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale] = useState(1)
  const [offset, setOffset] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragItem, setDragItem] = useState<SchemaItem | null>(null)
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 })
  const [isPanning, setIsPanning] = useState(false)
  const [panStart, setPanStart] = useState({ x: 0, y: 0 })

  // Handle wheel zoom - zoom towards mouse cursor position
  const handleWheel = useCallback((e: WheelEvent) => {
    e.preventDefault()

    const container = containerRef.current
    if (!container) return

    const rect = container.getBoundingClientRect()
    // Mouse position relative to container
    const mouseX = e.clientX - rect.left
    const mouseY = e.clientY - rect.top

    // Calculate zoom
    const delta = e.deltaY > 0 ? 0.9 : 1.1

    setScale((prevScale) => {
      const newScale = Math.min(Math.max(prevScale * delta, 0.25), 4)
      const scaleFactor = newScale / prevScale

      // Adjust offset so zoom centers on mouse position
      setOffset((prevOffset) => ({
        x: mouseX - (mouseX - prevOffset.x) * scaleFactor,
        y: mouseY - (mouseY - prevOffset.y) * scaleFactor,
      }))

      return newScale
    })
  }, [])

  useEffect(() => {
    const container = containerRef.current
    if (container) {
      container.addEventListener('wheel', handleWheel, { passive: false })
      return () => container.removeEventListener('wheel', handleWheel)
    }
  }, [handleWheel])

  // Convert screen coordinates to canvas coordinates
  const screenToCanvas = (screenX: number, screenY: number) => {
    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return { x: 0, y: 0 }
    return {
      x: (screenX - rect.left - offset.x) / scale,
      y: (screenY - rect.top - offset.y) / scale,
    }
  }

  // Handle mouse down on item
  const handleItemMouseDown = (e: React.MouseEvent, item: SchemaItem) => {
    e.preventDefault() // Prevent text selection

    if (!editMode) {
      onItemClick(item)
      return
    }

    e.stopPropagation()
    // Select the item in edit mode
    onItemClick(item)

    const canvasPos = screenToCanvas(e.clientX, e.clientY)
    setDragItem(item)
    setDragOffset({
      x: canvasPos.x - item.x,
      y: canvasPos.y - item.y,
    })
    setIsDragging(true)
  }

  // Handle mouse move
  const handleMouseMove = (e: React.MouseEvent) => {
    if (isDragging && dragItem) {
      const canvasPos = screenToCanvas(e.clientX, e.clientY)
      const newX = Math.max(0, canvasPos.x - dragOffset.x)
      const newY = Math.max(0, canvasPos.y - dragOffset.y)
      onItemMove(dragItem.id, newX, newY)
    } else if (isPanning) {
      const dx = e.clientX - panStart.x
      const dy = e.clientY - panStart.y
      setOffset((o) => ({ x: o.x + dx, y: o.y + dy }))
      setPanStart({ x: e.clientX, y: e.clientY })
    }
  }

  // Handle mouse up
  const handleMouseUp = () => {
    setIsDragging(false)
    setDragItem(null)
    setIsPanning(false)
  }

  // Handle canvas mouse down (panning)
  const handleCanvasMouseDown = (e: React.MouseEvent) => {
    if (e.button === 1 || (e.button === 0 && e.altKey)) {
      // Middle click or Alt+Click for panning
      setIsPanning(true)
      setPanStart({ x: e.clientX, y: e.clientY })
    } else if (e.button === 0 && onCanvasClick && editMode) {
      const canvasPos = screenToCanvas(e.clientX, e.clientY)
      onCanvasClick(canvasPos.x, canvasPos.y)
    }
  }

  // Get status color
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return { bg: 'bg-green-500', border: 'border-green-500', text: 'text-green-500' }
      case 'offline':
        return { bg: 'bg-red-500', border: 'border-red-500', text: 'text-red-500' }
      default:
        return { bg: 'bg-gray-400', border: 'border-gray-400', text: 'text-gray-400' }
    }
  }

  // Get device icon component
  const getDeviceIcon = (type: string, status: string) => {
    const colors = getStatusColor(status)
    const iconClass = `h-6 w-6 ${colors.text}`

    switch (type) {
      case 'switch':
        return <Network className={iconClass} />
      case 'server':
        return <Server className={iconClass} />
      case 'camera':
        return <Camera className={iconClass} />
      default:
        return <Monitor className={iconClass} />
    }
  }

  // Get device background color based on type
  const getDeviceBgColor = (type: string) => {
    switch (type) {
      case 'switch':
        return 'bg-blue-50 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800'
      case 'server':
        return 'bg-green-50 dark:bg-green-950/30 border-green-200 dark:border-green-800'
      case 'camera':
        return 'bg-purple-50 dark:bg-purple-950/30 border-purple-200 dark:border-purple-800'
      default:
        return 'bg-gray-50 dark:bg-gray-900/30 border-gray-200 dark:border-gray-700'
    }
  }

  // Reset view
  const resetView = () => {
    setScale(1)
    setOffset({ x: 0, y: 0 })
  }

  return (
    <div className="relative w-full h-full overflow-hidden bg-muted/30 rounded-lg">
      {/* Toolbar */}
      <div className="absolute top-2 right-2 z-10 flex gap-1">
        <button
          onClick={() => setScale((s) => Math.min(s * 1.2, 4))}
          className="p-2 bg-background/90 rounded-lg shadow hover:bg-accent text-sm font-medium"
          title="Увеличить"
        >
          +
        </button>
        <button
          onClick={() => setScale((s) => Math.max(s * 0.8, 0.25))}
          className="p-2 bg-background/90 rounded-lg shadow hover:bg-accent text-sm font-medium"
          title="Уменьшить"
        >
          −
        </button>
        <button
          onClick={resetView}
          className="p-2 bg-background/90 rounded-lg shadow hover:bg-accent text-xs"
          title="Сбросить"
        >
          ⟲
        </button>
        <span className="p-2 bg-background/90 rounded-lg shadow text-xs font-mono">
          {Math.round(scale * 100)}%
        </span>
      </div>

      {/* Canvas */}
      <div
        ref={containerRef}
        className="w-full h-full cursor-grab active:cursor-grabbing"
        onMouseDown={handleCanvasMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      >
        <div
          className="relative origin-top-left"
          style={{
            transform: `translate(${offset.x}px, ${offset.y}px) scale(${scale})`,
            width: '2000px',
            height: '2000px',
          }}
        >
          {/* Background image */}
          {backgroundImage && (
            <img
              src={backgroundImage}
              alt="Schema background"
              className="absolute inset-0 max-w-none pointer-events-none"
              draggable={false}
            />
          )}

          {/* Grid (when no background) */}
          {!backgroundImage && (
            <div
              className="absolute inset-0 pointer-events-none"
              style={{
                backgroundImage: `
                  linear-gradient(to right, hsl(var(--border)) 1px, transparent 1px),
                  linear-gradient(to bottom, hsl(var(--border)) 1px, transparent 1px)
                `,
                backgroundSize: '50px 50px',
              }}
            />
          )}

          {/* Connection lines (SVG) */}
          {showConnections && connections.length > 0 && (
            <svg
              className="absolute inset-0 pointer-events-none"
              style={{ width: '2000px', height: '2000px', zIndex: 1 }}
            >
              <defs>
                <marker
                  id="arrowhead"
                  markerWidth="10"
                  markerHeight="7"
                  refX="9"
                  refY="3.5"
                  orient="auto"
                >
                  <polygon
                    points="0 0, 10 3.5, 0 7"
                    fill="hsl(var(--muted-foreground))"
                    opacity="0.5"
                  />
                </marker>
              </defs>
              {connections.map((conn, idx) => {
                // Find the items to get their current positions
                const fromItem = itemsForConnections.find(i => i.device_id === conn.fromId)
                const toItem = itemsForConnections.find(i => i.device_id === conn.toId)

                if (!fromItem || !toItem) return null

                const fromX = fromItem.x + (fromItem.width || 70) / 2
                const fromY = fromItem.y + (fromItem.height || 70) / 2
                const toX = toItem.x + (toItem.width || 70) / 2
                const toY = toItem.y + (toItem.height || 70) / 2

                const isUplink = conn.type === 'uplink'

                return (
                  <g key={idx}>
                    {/* Connection line */}
                    <line
                      x1={fromX}
                      y1={fromY}
                      x2={toX}
                      y2={toY}
                      stroke={isUplink ? '#f59e0b' : '#3b82f6'}
                      strokeWidth={isUplink ? 3 : 2}
                      strokeDasharray={isUplink ? '0' : '6 3'}
                      opacity={isUplink ? 0.9 : 0.7}
                    />
                    {/* Circle at from device (switch) */}
                    <circle
                      cx={fromX}
                      cy={fromY}
                      r={isUplink ? 6 : 5}
                      fill={isUplink ? '#f59e0b' : '#3b82f6'}
                      opacity="0.8"
                    />
                    {/* Circle at to device (camera or linked switch) */}
                    <circle
                      cx={toX}
                      cy={toY}
                      r={isUplink ? 6 : 5}
                      fill={isUplink ? '#f59e0b' : '#a855f7'}
                      opacity="0.8"
                    />
                  </g>
                )
              })}
            </svg>
          )}

          {/* Device items */}
          {items.map((item) => {
            const statusColors = getStatusColor(item.device_status || 'unknown')

            return (
              <div
                key={item.id}
                className={cn(
                  'absolute flex flex-col items-center justify-center rounded-xl border-2 shadow-lg cursor-pointer transition-all duration-150 select-none',
                  getDeviceBgColor(item.device_type || ''),
                  selectedItemId === item.id && 'ring-2 ring-primary ring-offset-2',
                  editMode && 'hover:ring-2 hover:ring-primary/50 hover:shadow-xl',
                  isDragging && dragItem?.id === item.id && 'cursor-grabbing opacity-80'
                )}
                style={{
                  left: item.x,
                  top: item.y,
                  width: item.width || 70,
                  height: item.height || 70,
                  zIndex: 2,
                }}
                onMouseDown={(e) => handleItemMouseDown(e, item)}
                draggable={false}
              >
                {/* Status indicator */}
                <div
                  className={cn(
                    'absolute -top-1.5 -right-1.5 w-4 h-4 rounded-full border-2 border-background shadow-sm',
                    statusColors.bg
                  )}
                />

                {/* Icon */}
                <div className="mb-0.5 pointer-events-none">
                  {getDeviceIcon(item.device_type || '', item.device_status || 'unknown')}
                </div>

                {/* Name */}
                <span className="text-[9px] text-center leading-tight px-1 truncate max-w-full font-medium pointer-events-none">
                  {item.device_name}
                </span>

                {/* Remove button (edit mode) */}
                {editMode && selectedItemId === item.id && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      onItemRemove(item.id)
                    }}
                    className="absolute -top-2 -left-2 w-5 h-5 bg-destructive text-destructive-foreground rounded-full text-xs flex items-center justify-center hover:bg-destructive/80 shadow"
                  >
                    ×
                  </button>
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* Instructions */}
      {editMode && (
        <div className="absolute bottom-2 left-2 text-xs text-muted-foreground bg-white dark:bg-zinc-900 px-2 py-1 rounded border shadow">
          Перетащите устройства • Колёсико мыши для масштаба • Alt+клик для перемещения
        </div>
      )}
    </div>
  )
}
