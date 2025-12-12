//go:build windows

package main

import (
	_ "embed"
	"time"

	"netvisionmonitor/internal/logger"

	"github.com/energye/systray"
)

//go:embed build/windows/icon.ico
var iconData []byte

var (
	trayApp     *App
	mShow       *systray.MenuItem
	mStatus     *systray.MenuItem
	mQuit       *systray.MenuItem
	trayRunning bool
)

// InitTray initializes system tray icon
func InitTray(app *App) {
	trayApp = app
	trayRunning = true

	// Left click shows context menu
	systray.SetOnClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})

	// Right click also shows context menu
	systray.SetOnRClick(func(menu systray.IMenu) {
		menu.ShowMenu()
	})

	// Double click to show window
	systray.SetOnDClick(func(menu systray.IMenu) {
		if trayApp != nil {
			trayApp.ShowFromTray()
		}
	})

	go systray.Run(onTrayReady, onTrayExit)
}

func onTrayReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("NetVisionMonitor")
	systray.SetTooltip("NetVisionMonitor - Мониторинг сети")

	mShow = systray.AddMenuItem("Открыть NetVisionMonitor", "Показать главное окно")
	systray.AddSeparator()
	mStatus = systray.AddMenuItem("Статус: загрузка...", "Текущий статус мониторинга")
	mStatus.Disable()
	systray.AddSeparator()
	mQuit = systray.AddMenuItem("Выход", "Полностью закрыть приложение")

	mShow.Click(func() {
		if trayApp != nil {
			trayApp.ShowFromTray()
		}
	})

	mQuit.Click(func() {
		if trayApp != nil {
			logger.Info("Quit requested from tray menu")
			trayRunning = false
			// First quit the app, then stop systray
			trayApp.QuitApp()
		}
	})

	go updateTrayStatus()

	logger.Info("System tray initialized")
}

func onTrayExit() {
	logger.Info("System tray exiting")
	trayRunning = false
}

func updateTrayStatus() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Initial update
	doUpdateStatus()

	for trayRunning {
		select {
		case <-ticker.C:
			doUpdateStatus()
		}
	}
}

func doUpdateStatus() {
	if trayApp == nil || mStatus == nil {
		return
	}

	status := trayApp.GetTrayStatus()
	online, _ := status["online"].(int)
	offline, _ := status["offline"].(int)
	total, _ := status["total"].(int)
	monitoring, _ := status["monitoring"].(bool)

	var statusText string
	if monitoring {
		statusText = "Мониторинг: ✓ | "
	} else {
		statusText = "Мониторинг: ✗ | "
	}
	statusText += "Онлайн: " + intToStr(online) + "/" + intToStr(total)
	if offline > 0 {
		statusText += " | Офлайн: " + intToStr(offline)
	}

	mStatus.SetTitle(statusText)
	systray.SetTooltip("NetVisionMonitor\n" + statusText)
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	negative := false
	if i < 0 {
		negative = true
		i = -i
	}
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}

// StopTray stops the system tray
func StopTray() {
	trayRunning = false
	systray.Quit()
}
