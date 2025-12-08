package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SchemaInput is used for creating/updating schemas from frontend
type SchemaInput struct {
	ID              int64  `json:"id,omitempty"`
	Name            string `json:"name"`
	BackgroundImage string `json:"background_image,omitempty"` // base64 or path
}

// SchemaItemInput is used for creating/updating schema items
type SchemaItemInput struct {
	ID       int64   `json:"id,omitempty"`
	DeviceID int64   `json:"device_id"`
	SchemaID int64   `json:"schema_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
}

// GetSchemas returns all schemas
func (a *App) GetSchemas() ([]models.Schema, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaRepository(a.db.DB())
	return repo.GetAll()
}

// GetSchema returns a schema by ID with its items
func (a *App) GetSchema(id int64) (*models.Schema, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaRepository(a.db.DB())
	return repo.GetByID(id)
}

// GetSchemaItems returns all items for a schema
func (a *App) GetSchemaItems(schemaID int64) ([]models.SchemaItem, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaItemRepository(a.db.DB())
	return repo.GetBySchemaID(schemaID)
}

// CreateSchema creates a new schema
func (a *App) CreateSchema(input SchemaInput) (*models.Schema, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Handle background image if provided as base64
	bgPath := ""
	if input.BackgroundImage != "" && strings.HasPrefix(input.BackgroundImage, "data:image") {
		path, err := a.saveBackgroundImage(input.BackgroundImage)
		if err != nil {
			return nil, fmt.Errorf("failed to save background image: %w", err)
		}
		bgPath = path
	} else {
		bgPath = input.BackgroundImage
	}

	schema := &models.Schema{
		Name:            input.Name,
		BackgroundImage: bgPath,
	}

	repo := database.NewSchemaRepository(a.db.DB())
	if err := repo.Create(schema); err != nil {
		return nil, err
	}

	return schema, nil
}

// UpdateSchema updates an existing schema
func (a *App) UpdateSchema(input SchemaInput) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaRepository(a.db.DB())
	existing, err := repo.GetByID(input.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("schema not found")
	}

	// Handle background image
	bgPath := existing.BackgroundImage
	if input.BackgroundImage != "" && strings.HasPrefix(input.BackgroundImage, "data:image") {
		// Delete old image if exists
		if existing.BackgroundImage != "" {
			os.Remove(existing.BackgroundImage)
		}
		path, err := a.saveBackgroundImage(input.BackgroundImage)
		if err != nil {
			return fmt.Errorf("failed to save background image: %w", err)
		}
		bgPath = path
	} else if input.BackgroundImage != "" {
		bgPath = input.BackgroundImage
	}

	existing.Name = input.Name
	existing.BackgroundImage = bgPath

	return repo.Update(existing)
}

// DeleteSchema deletes a schema
func (a *App) DeleteSchema(id int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaRepository(a.db.DB())
	schema, err := repo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete background image if exists
	if schema != nil && schema.BackgroundImage != "" {
		os.Remove(schema.BackgroundImage)
	}

	return repo.Delete(id)
}

// AddDeviceToSchema adds a device to a schema
func (a *App) AddDeviceToSchema(input SchemaItemInput) (*models.SchemaItem, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaItemRepository(a.db.DB())

	// Check if device already exists on schema
	existing, err := repo.GetByDeviceAndSchema(input.DeviceID, input.SchemaID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("device already exists on this schema")
	}

	item := &models.SchemaItem{
		DeviceID: input.DeviceID,
		SchemaID: input.SchemaID,
		X:        input.X,
		Y:        input.Y,
		Width:    input.Width,
		Height:   input.Height,
	}

	if item.Width == 0 {
		item.Width = 60
	}
	if item.Height == 0 {
		item.Height = 60
	}

	if err := repo.Create(item); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateSchemaItemPosition updates a schema item position
func (a *App) UpdateSchemaItemPosition(id int64, x, y float64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaItemRepository(a.db.DB())
	return repo.UpdatePosition(id, x, y)
}

// UpdateSchemaItem updates a schema item
func (a *App) UpdateSchemaItem(input SchemaItemInput) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	item := &models.SchemaItem{
		ID:     input.ID,
		X:      input.X,
		Y:      input.Y,
		Width:  input.Width,
		Height: input.Height,
	}

	repo := database.NewSchemaItemRepository(a.db.DB())
	return repo.Update(item)
}

// RemoveDeviceFromSchema removes a device from a schema
func (a *App) RemoveDeviceFromSchema(itemID int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewSchemaItemRepository(a.db.DB())
	return repo.Delete(itemID)
}

// SelectBackgroundImage opens a file dialog to select background image
func (a *App) SelectBackgroundImage() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Выберите изображение",
		Filters: []runtime.FileFilter{
			{DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.bmp"},
		},
	})
	if err != nil {
		return "", err
	}

	if selection == "" {
		return "", nil
	}

	// Read and convert to base64
	data, err := os.ReadFile(selection)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// Detect mime type
	ext := strings.ToLower(filepath.Ext(selection))
	mimeType := "image/png"
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	case ".bmp":
		mimeType = "image/bmp"
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data), nil
}

// saveBackgroundImage saves a base64 image to disk
func (a *App) saveBackgroundImage(base64Data string) (string, error) {
	// Parse base64 data
	parts := strings.Split(base64Data, ",")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid base64 data")
	}

	// Decode
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Determine extension
	ext := ".png"
	if strings.Contains(parts[0], "jpeg") || strings.Contains(parts[0], "jpg") {
		ext = ".jpg"
	} else if strings.Contains(parts[0], "gif") {
		ext = ".gif"
	}

	// Create schemas directory
	schemasDir := filepath.Join(a.cfg.DataDir, "schemas")
	if err := os.MkdirAll(schemasDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create schemas directory: %w", err)
	}

	// Generate filename
	filename := fmt.Sprintf("bg_%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(schemasDir, filename)

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	return filePath, nil
}

// GetBackgroundImage returns background image as base64
func (a *App) GetBackgroundImage(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	mimeType := "image/png"
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data), nil
}
