package models

import "time"

type Schema struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	BackgroundImage string    `json:"background_image"`
	CreatedAt       time.Time `json:"created_at"`
}

type SchemaItem struct {
	ID       int64   `json:"id"`
	DeviceID int64   `json:"device_id"`
	SchemaID int64   `json:"schema_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	// Device info (populated from JOIN)
	DeviceName   string `json:"device_name,omitempty"`
	DeviceType   string `json:"device_type,omitempty"`
	DeviceStatus string `json:"device_status,omitempty"`
	DeviceIP     string `json:"device_ip,omitempty"`
}

type SchemaWithItems struct {
	Schema
	Items []SchemaItemWithDevice `json:"items"`
}

type SchemaItemWithDevice struct {
	SchemaItem
	Device *Device `json:"device,omitempty"`
}
