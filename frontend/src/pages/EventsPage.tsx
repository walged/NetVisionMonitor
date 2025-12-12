import { useState, useEffect, useCallback, useRef } from 'react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Pagination } from '@/components/ui/pagination'
import {
  Download,
  Filter,
  ScrollText,
  AlertCircle,
  AlertTriangle,
  Info,
  Trash2,
} from 'lucide-react'
import { GetEventsPaginated, ClearEvents } from '../../wailsjs/go/main/App'
import { useMonitoringEvents } from '@/hooks/useMonitoring'
import { models } from '../../wailsjs/go/models'
import { useTranslation } from '@/i18n'

type Event = models.Event

export function EventsPage() {
  const { t, i18n } = useTranslation()
  const [events, setEvents] = useState<Event[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [levelFilter, setLevelFilter] = useState<string>('all')
  const [autoScroll, setAutoScroll] = useState(true)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Pagination state
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(50)
  const [totalItems, setTotalItems] = useState(0)
  const [totalPages, setTotalPages] = useState(1)

  const loadEvents = useCallback(async () => {
    try {
      setIsLoading(true)
      const result = await GetEventsPaginated({
        level: levelFilter !== 'all' ? levelFilter : undefined,
        page: currentPage,
        page_size: pageSize,
        limit: 0,
        offset: 0,
      })
      setEvents(result?.events || [])
      setTotalItems(result?.total || 0)
      setTotalPages(result?.total_pages || 1)
    } finally {
      setIsLoading(false)
    }
  }, [levelFilter, currentPage, pageSize])

  useEffect(() => {
    loadEvents()
  }, [loadEvents])

  // Reset to page 1 when filter changes
  useEffect(() => {
    setCurrentPage(1)
  }, [levelFilter])

  const handlePageChange = (page: number) => {
    setCurrentPage(page)
  }

  const handlePageSizeChange = (size: number) => {
    setPageSize(size)
    setCurrentPage(1)
  }

  const handleLevelFilterChange = (value: string) => {
    setLevelFilter(value)
  }

  // Listen for new events
  useMonitoringEvents(
    useCallback((event) => {
      setEvents((prev) => [event as Event, ...prev].slice(0, 500))
      if (autoScroll && scrollRef.current) {
        scrollRef.current.scrollTop = 0
      }
    }, [autoScroll])
  )

  const handleClear = async () => {
    await ClearEvents()
    setEvents([])
  }

  const handleExport = () => {
    const headers = i18n.language === 'ru' ? 'Время,Уровень,Тип,Сообщение' : 'Time,Level,Type,Message'
    const csv = [
      headers,
      ...events.map(
        (e) =>
          `"${e.created_at}","${e.level}","${e.type}","${e.message.replace(/"/g, '""')}"`
      ),
    ].join('\n')

    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `events_${new Date().toISOString().slice(0, 10)}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  const getLevelIcon = (level: string) => {
    switch (level) {
      case 'error':
        return <AlertCircle className="h-4 w-4 text-destructive" />
      case 'warn':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />
      default:
        return <Info className="h-4 w-4 text-blue-500" />
    }
  }

  const getLevelBadge = (level: string) => {
    switch (level) {
      case 'error':
        return <Badge variant="destructive">{t('events.levels.error')}</Badge>
      case 'warn':
        return <Badge variant="warning">{t('events.levels.warning')}</Badge>
      default:
        return <Badge variant="secondary">{t('events.levels.info')}</Badge>
    }
  }

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr)
    const locale = i18n.language === 'ru' ? 'ru-RU' : 'en-US'
    return date.toLocaleString(locale, {
      day: '2-digit',
      month: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>{t('events.title')}</CardTitle>
            <CardDescription>
              {t('events.subtitle')} ({totalItems})
            </CardDescription>
          </div>
          <div className="flex gap-2">
            <Select value={levelFilter} onValueChange={handleLevelFilterChange}>
              <SelectTrigger className="w-[150px]">
                <Filter className="h-4 w-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{t('common.all')}</SelectItem>
                <SelectItem value="error">{t('events.levels.error')}</SelectItem>
                <SelectItem value="warn">{t('events.levels.warning')}</SelectItem>
                <SelectItem value="info">{t('events.levels.info')}</SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" onClick={handleExport}>
              <Download className="h-4 w-4 mr-2" />
              {t('settings.data.export')}
            </Button>
            <Button variant="outline" onClick={handleClear}>
              <Trash2 className="h-4 w-4 mr-2" />
              {t('events.clearEvents')}
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {events.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
              <ScrollText className="h-12 w-12 mb-4 opacity-50" />
              <p className="text-lg font-medium">{t('events.noEvents')}</p>
              <p className="text-sm">
                {t('events.subtitle')}
              </p>
            </div>
          ) : (
            <ScrollArea className="h-[500px]" ref={scrollRef}>
              <div className="space-y-2">
                {events.map((event) => (
                  <div
                    key={event.id}
                    className="flex items-start gap-3 p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
                  >
                    <div className="mt-0.5">{getLevelIcon(event.level)}</div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        {getLevelBadge(event.level)}
                        <span className="text-xs text-muted-foreground">
                          {formatTime(event.created_at)}
                        </span>
                      </div>
                      <p className="text-sm">{event.message}</p>
                    </div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          )}
          {/* Pagination */}
          {totalPages > 1 && (
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              pageSize={pageSize}
              totalItems={totalItems}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          )}
        </CardContent>
      </Card>
    </div>
  )
}
