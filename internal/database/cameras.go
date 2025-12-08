package database

import (
	"database/sql"
	"fmt"

	"netvisionmonitor/internal/encryption"
	"netvisionmonitor/internal/models"
)

// CameraRepository handles camera-specific database operations
type CameraRepository struct {
	db *sql.DB
}

// NewCameraRepository creates a new camera repository
func NewCameraRepository(db *sql.DB) *CameraRepository {
	return &CameraRepository{db: db}
}

// Create inserts camera-specific data
func (r *CameraRepository) Create(cam *models.Camera) error {
	encryptedRTSP, err := encryption.EncryptIfNotEmpty(cam.RTSPURL)
	if err != nil {
		return fmt.Errorf("failed to encrypt RTSP URL: %w", err)
	}

	_, err = r.db.Exec(`
		INSERT INTO cameras (device_id, rtsp_url, onvif_port, snapshot_url, stream_type)
		VALUES (?, ?, ?, ?, ?)`,
		cam.DeviceID, encryptedRTSP, cam.ONVIFPort, cam.SnapshotURL, cam.StreamType,
	)
	if err != nil {
		return fmt.Errorf("failed to create camera: %w", err)
	}
	return nil
}

// GetByDeviceID retrieves camera data by device ID
func (r *CameraRepository) GetByDeviceID(deviceID int64) (*models.Camera, error) {
	cam := &models.Camera{}
	var encryptedRTSP string

	err := r.db.QueryRow(`
		SELECT device_id, rtsp_url, onvif_port, snapshot_url, stream_type
		FROM cameras WHERE device_id = ?`, deviceID,
	).Scan(&cam.DeviceID, &encryptedRTSP, &cam.ONVIFPort, &cam.SnapshotURL, &cam.StreamType)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get camera: %w", err)
	}

	cam.RTSPURL, err = encryption.DecryptIfNotEmpty(encryptedRTSP)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt RTSP URL: %w", err)
	}

	return cam, nil
}

// Update updates camera-specific data
func (r *CameraRepository) Update(cam *models.Camera) error {
	encryptedRTSP, err := encryption.EncryptIfNotEmpty(cam.RTSPURL)
	if err != nil {
		return fmt.Errorf("failed to encrypt RTSP URL: %w", err)
	}

	_, err = r.db.Exec(`
		UPDATE cameras SET rtsp_url = ?, onvif_port = ?, snapshot_url = ?, stream_type = ?
		WHERE device_id = ?`,
		encryptedRTSP, cam.ONVIFPort, cam.SnapshotURL, cam.StreamType, cam.DeviceID,
	)
	if err != nil {
		return fmt.Errorf("failed to update camera: %w", err)
	}
	return nil
}

// Delete removes camera data by device ID
func (r *CameraRepository) Delete(deviceID int64) error {
	_, err := r.db.Exec("DELETE FROM cameras WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete camera: %w", err)
	}
	return nil
}
