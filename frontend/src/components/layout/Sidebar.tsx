import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  Monitor,
  Network,
  Map,
  LayoutGrid,
  ScrollText,
  Settings,
  Info,
  Minimize2,
  LogOut,
} from "lucide-react"
import { MinimizeToTray, QuitApp } from "../../../wailsjs/go/main/App"

interface NavItem {
  id: string
  label: string
  icon: React.ReactNode
}

const mainNavItems: NavItem[] = [
  { id: "devices", label: "Устройства", icon: <Monitor className="h-5 w-5" /> },
  { id: "networkmap", label: "Карта сети", icon: <Map className="h-5 w-5" /> },
  { id: "schema", label: "Схема", icon: <LayoutGrid className="h-5 w-5" /> },
  { id: "events", label: "События", icon: <ScrollText className="h-5 w-5" /> },
]

const bottomNavItems: NavItem[] = [
  { id: "settings", label: "Настройки", icon: <Settings className="h-5 w-5" /> },
  { id: "about", label: "О программе", icon: <Info className="h-5 w-5" /> },
]

interface SidebarProps {
  currentPage: string
  onNavigate: (page: string) => void
  collapsed?: boolean
}

export function Sidebar({ currentPage, onNavigate, collapsed = false }: SidebarProps) {
  return (
    <TooltipProvider delayDuration={0}>
      <div
        className={cn(
          "flex flex-col h-full bg-card border-r transition-all duration-300",
          collapsed ? "w-16" : "w-56"
        )}
      >
        {/* Logo */}
        <div className="flex items-center h-14 px-4 border-b">
          <Network className="h-6 w-6 text-primary" />
          {!collapsed && (
            <span className="ml-3 font-semibold text-lg">NetVision</span>
          )}
        </div>

        {/* Main Navigation */}
        <ScrollArea className="flex-1 py-2">
          <nav className="space-y-1 px-2">
            {mainNavItems.map((item) => (
              <NavButton
                key={item.id}
                item={item}
                isActive={currentPage === item.id}
                collapsed={collapsed}
                onClick={() => onNavigate(item.id)}
              />
            ))}
          </nav>
        </ScrollArea>

        <Separator />

        {/* Bottom Navigation */}
        <div className="py-2 px-2 space-y-1">
          {bottomNavItems.map((item) => (
            <NavButton
              key={item.id}
              item={item}
              isActive={currentPage === item.id}
              collapsed={collapsed}
              onClick={() => onNavigate(item.id)}
            />
          ))}
        </div>

        <Separator />

        {/* Tray & Exit */}
        <div className="py-2 px-2 space-y-1">
          <ActionButton
            icon={<Minimize2 className="h-5 w-5" />}
            label="Свернуть в трей"
            collapsed={collapsed}
            onClick={() => MinimizeToTray()}
          />
          <ActionButton
            icon={<LogOut className="h-5 w-5" />}
            label="Выход"
            collapsed={collapsed}
            onClick={() => QuitApp()}
            variant="destructive"
          />
        </div>
      </div>
    </TooltipProvider>
  )
}

interface NavButtonProps {
  item: NavItem
  isActive: boolean
  collapsed: boolean
  onClick: () => void
}

function NavButton({ item, isActive, collapsed, onClick }: NavButtonProps) {
  const button = (
    <Button
      variant={isActive ? "secondary" : "ghost"}
      className={cn(
        "w-full justify-start",
        collapsed && "justify-center px-2"
      )}
      onClick={onClick}
    >
      {item.icon}
      {!collapsed && <span className="ml-3">{item.label}</span>}
    </Button>
  )

  if (collapsed) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>{button}</TooltipTrigger>
        <TooltipContent side="right">
          <p>{item.label}</p>
        </TooltipContent>
      </Tooltip>
    )
  }

  return button
}

interface ActionButtonProps {
  icon: React.ReactNode
  label: string
  collapsed: boolean
  onClick: () => void
  variant?: "default" | "destructive"
}

function ActionButton({ icon, label, collapsed, onClick, variant = "default" }: ActionButtonProps) {
  const button = (
    <Button
      variant="ghost"
      className={cn(
        "w-full justify-start",
        collapsed && "justify-center px-2",
        variant === "destructive" && "text-destructive hover:text-destructive hover:bg-destructive/10"
      )}
      onClick={onClick}
    >
      {icon}
      {!collapsed && <span className="ml-3">{label}</span>}
    </Button>
  )

  if (collapsed) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>{button}</TooltipTrigger>
        <TooltipContent side="right">
          <p>{label}</p>
        </TooltipContent>
      </Tooltip>
    )
  }

  return button
}
