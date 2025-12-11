package main

import (
	"embed"
	"os"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// checkSingleInstance ensures only one instance of the application is running
func checkSingleInstance() (func(), bool) {
	// Create a named mutex for single instance check
	mutexName, _ := syscall.UTF16PtrFromString("NetVisionMonitor_SingleInstance_Mutex")

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	createMutex := kernel32.NewProc("CreateMutexW")

	handle, _, err := createMutex.Call(0, 1, uintptr(unsafe.Pointer(mutexName)))
	if handle == 0 {
		return nil, false
	}

	// Check if mutex already exists (ERROR_ALREADY_EXISTS = 183)
	if err.(syscall.Errno) == 183 {
		syscall.CloseHandle(syscall.Handle(handle))
		return nil, false
	}

	// Return cleanup function
	cleanup := func() {
		syscall.CloseHandle(syscall.Handle(handle))
	}

	return cleanup, true
}

func main() {
	// Check if another instance is already running
	cleanup, isFirst := checkSingleInstance()
	if !isFirst {
		// Another instance is running, show message and exit
		user32 := syscall.NewLazyDLL("user32.dll")
		messageBox := user32.NewProc("MessageBoxW")

		title, _ := syscall.UTF16PtrFromString("NetVisionMonitor")
		text, _ := syscall.UTF16PtrFromString("Приложение уже запущено!")

		messageBox.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x40) // MB_ICONINFORMATION
		os.Exit(0)
	}
	defer cleanup()

	// Create an instance of the app structure
	app := NewApp()

	// Check if started minimized (autostart)
	startMinimized := IsStartedMinimized()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "NetVisionMonitor",
		Width:             1280,
		Height:            800,
		MinWidth:          1024,
		MinHeight:         700,
		StartHidden:       startMinimized, // Start minimized if --minimized flag
		HideWindowOnClose: true,           // Hide instead of close (minimize to tray)
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1}, // Dark theme background
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			Theme:                windows.Dark,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
