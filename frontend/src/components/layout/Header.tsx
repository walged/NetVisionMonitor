import { useState, useEffect, useCallback } from "react"
import { Button } from "@/components/ui/button"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { ScrollArea } from "@/components/ui/scroll-area"
import { PanelLeftClose, PanelLeft, Bell, RefreshCw, AlertCircle, AlertTriangle, Info, CheckCircle, Trash2 } from "lucide-react"
import { GetRecentEvents } from "../../../wailsjs/go/main/App"
import { EventsOn } from "../../../wailsjs/runtime/runtime"

interface Event {
  id: number
  device_id?: number
  type: string
  level: string
  message: string
  created_at: string
}

interface HeaderProps {
  title: string
  sidebarCollapsed: boolean
  onToggleSidebar: () => void
}

function getLevelIcon(level: string) {
  switch (level) {
    case 'error':
      return <AlertCircle className="h-4 w-4 text-destructive" />
    case 'warning':
      return <AlertTriangle className="h-4 w-4 text-yellow-500" />
    case 'success':
      return <CheckCircle className="h-4 w-4 text-green-500" />
    default:
      return <Info className="h-4 w-4 text-blue-500" />
  }
}

function formatTime(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return 'только что'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} мин. назад`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} ч. назад`
  return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
}

export function Header({ title, sidebarCollapsed, onToggleSidebar }: HeaderProps) {
  const [events, setEvents] = useState<Event[]>([])
  const [unreadCount, setUnreadCount] = useState(0)
  const [isOpen, setIsOpen] = useState(false)
  const [clearedAt, setClearedAt] = useState<number | null>(null)

  const loadEvents = useCallback(async () => {
    try {
      const data = await GetRecentEvents(5) // Show only last 5 notifications
      // If cleared, only show events after clear time
      if (clearedAt) {
        const filtered = (data || []).filter(e => new Date(e.created_at).getTime() > clearedAt)
        setEvents(filtered)
        const hourAgo = Date.now() - 3600000
        const unread = filtered.filter(e => new Date(e.created_at).getTime() > hourAgo).length
        setUnreadCount(unread)
      } else {
        setEvents(data || [])
        // Count events from last hour as "unread"
        const hourAgo = Date.now() - 3600000
        const unread = (data || []).filter(e => new Date(e.created_at).getTime() > hourAgo).length
        setUnreadCount(unread)
      }
    } catch (err) {
      console.error('Failed to load events:', err)
    }
  }, [clearedAt])

  const clearNotifications = () => {
    setClearedAt(Date.now())
    setEvents([])
    setUnreadCount(0)
  }

  useEffect(() => {
    loadEvents()

    // Listen for new events
    const unsubscribe = EventsOn('event:new', () => {
      loadEvents()
    })

    // Refresh periodically
    const interval = setInterval(loadEvents, 30000)

    return () => {
      unsubscribe()
      clearInterval(interval)
    }
  }, [loadEvents])

  const handleOpen = (open: boolean) => {
    setIsOpen(open)
    if (open) {
      setUnreadCount(0)
    }
  }

  return (
    <TooltipProvider>
      <header className="flex items-center justify-between h-14 px-4 border-b bg-card">
        <div className="flex items-center gap-3">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                onClick={onToggleSidebar}
              >
                {sidebarCollapsed ? (
                  <PanelLeft className="h-5 w-5" />
                ) : (
                  <PanelLeftClose className="h-5 w-5" />
                )}
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{sidebarCollapsed ? "Развернуть панель" : "Свернуть панель"}</p>
            </TooltipContent>
          </Tooltip>
          <h1 className="text-lg font-semibold">{title}</h1>
        </div>

        <div className="flex items-center gap-2">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon" onClick={loadEvents}>
                <RefreshCw className="h-5 w-5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>Обновить данные</p>
            </TooltipContent>
          </Tooltip>

          <Popover open={isOpen} onOpenChange={handleOpen}>
            <PopoverTrigger asChild>
              <Button variant="ghost" size="icon" className="relative">
                <Bell className="h-5 w-5" />
                {unreadCount > 0 && (
                  <span className="absolute top-1 right-1 min-w-[16px] h-4 px-1 text-xs bg-destructive text-destructive-foreground rounded-full flex items-center justify-center">
                    {unreadCount > 9 ? '9+' : unreadCount}
                  </span>
                )}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-0" align="end">
              <div className="p-3 border-b flex items-center justify-between">
                <h4 className="font-semibold">Уведомления</h4>
                {events.length > 0 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 text-xs text-muted-foreground hover:text-destructive"
                    onClick={clearNotifications}
                  >
                    <Trash2 className="h-3 w-3 mr-1" />
                    Очистить
                  </Button>
                )}
              </div>
              <ScrollArea className="h-[250px]">
                {events.length === 0 ? (
                  <div className="p-4 text-center text-muted-foreground">
                    Нет новых уведомлений
                  </div>
                ) : (
                  <div className="divide-y">
                    {events.map((event) => (
                      <div key={event.id} className="p-3 hover:bg-muted/50 cursor-pointer">
                        <div className="flex items-start gap-3">
                          {getLevelIcon(event.level)}
                          <div className="flex-1 min-w-0">
                            <p className="text-sm truncate">{event.message}</p>
                            <p className="text-xs text-muted-foreground mt-1">
                              {formatTime(event.created_at)}
                            </p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </ScrollArea>
            </PopoverContent>
          </Popover>
        </div>
      </header>
    </TooltipProvider>
  )
}
