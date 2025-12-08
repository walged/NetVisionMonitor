package database

import (
	"database/sql"
	"fmt"
	"time"

	"netvisionmonitor/internal/models"
)

// SchemaRepository handles schema database operations
type SchemaRepository struct {
	db *sql.DB
}

// NewSchemaRepository creates a new schema repository
func NewSchemaRepository(db *sql.DB) *SchemaRepository {
	return &SchemaRepository{db: db}
}

// Create inserts a new schema
func (r *SchemaRepository) Create(schema *models.Schema) error {
	result, err := r.db.Exec(`
		INSERT INTO schemas (name, background_image, created_at)
		VALUES (?, ?, ?)`,
		schema.Name, schema.BackgroundImage, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	schema.ID = id
	return nil
}

// GetByID retrieves a schema by ID
func (r *SchemaRepository) GetByID(id int64) (*models.Schema, error) {
	schema := &models.Schema{}
	err := r.db.QueryRow(`
		SELECT id, name, background_image, created_at
		FROM schemas WHERE id = ?`, id,
	).Scan(&schema.ID, &schema.Name, &schema.BackgroundImage, &schema.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}
	return schema, nil
}

// GetAll retrieves all schemas
func (r *SchemaRepository) GetAll() ([]models.Schema, error) {
	rows, err := r.db.Query(`
		SELECT id, name, background_image, created_at
		FROM schemas ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []models.Schema
	for rows.Next() {
		var s models.Schema
		err := rows.Scan(&s.ID, &s.Name, &s.BackgroundImage, &s.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, s)
	}
	return schemas, nil
}

// Update updates a schema
func (r *SchemaRepository) Update(schema *models.Schema) error {
	_, err := r.db.Exec(`
		UPDATE schemas SET name = ?, background_image = ?
		WHERE id = ?`,
		schema.Name, schema.BackgroundImage, schema.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update schema: %w", err)
	}
	return nil
}

// Delete removes a schema by ID
func (r *SchemaRepository) Delete(id int64) error {
	// Delete schema items first
	_, err := r.db.Exec("DELETE FROM schema_items WHERE schema_id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete schema items: %w", err)
	}

	_, err = r.db.Exec("DELETE FROM schemas WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete schema: %w", err)
	}
	return nil
}

// SchemaItemRepository handles schema item database operations
type SchemaItemRepository struct {
	db *sql.DB
}

// NewSchemaItemRepository creates a new schema item repository
func NewSchemaItemRepository(db *sql.DB) *SchemaItemRepository {
	return &SchemaItemRepository{db: db}
}

// Create inserts a new schema item
func (r *SchemaItemRepository) Create(item *models.SchemaItem) error {
	result, err := r.db.Exec(`
		INSERT INTO schema_items (device_id, schema_id, x, y, width, height)
		VALUES (?, ?, ?, ?, ?, ?)`,
		item.DeviceID, item.SchemaID, item.X, item.Y, item.Width, item.Height,
	)
	if err != nil {
		return fmt.Errorf("failed to create schema item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	item.ID = id
	return nil
}

// GetBySchemaID retrieves all items for a schema
func (r *SchemaItemRepository) GetBySchemaID(schemaID int64) ([]models.SchemaItem, error) {
	rows, err := r.db.Query(`
		SELECT si.id, si.device_id, si.schema_id, si.x, si.y, si.width, si.height,
		       d.name, d.type, d.status, d.ip_address
		FROM schema_items si
		LEFT JOIN devices d ON si.device_id = d.id
		WHERE si.schema_id = ?`, schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema items: %w", err)
	}
	defer rows.Close()

	var items []models.SchemaItem
	for rows.Next() {
		var item models.SchemaItem
		var deviceName, deviceType, deviceStatus, deviceIP sql.NullString
		err := rows.Scan(
			&item.ID, &item.DeviceID, &item.SchemaID, &item.X, &item.Y, &item.Width, &item.Height,
			&deviceName, &deviceType, &deviceStatus, &deviceIP,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schema item: %w", err)
		}
		if deviceName.Valid {
			item.DeviceName = deviceName.String
			item.DeviceType = deviceType.String
			item.DeviceStatus = deviceStatus.String
			item.DeviceIP = deviceIP.String
		}
		items = append(items, item)
	}
	return items, nil
}

// Update updates a schema item position
func (r *SchemaItemRepository) Update(item *models.SchemaItem) error {
	_, err := r.db.Exec(`
		UPDATE schema_items SET x = ?, y = ?, width = ?, height = ?
		WHERE id = ?`,
		item.X, item.Y, item.Width, item.Height, item.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update schema item: %w", err)
	}
	return nil
}

// UpdatePosition updates only the position of a schema item
func (r *SchemaItemRepository) UpdatePosition(id int64, x, y float64) error {
	_, err := r.db.Exec(`
		UPDATE schema_items SET x = ?, y = ? WHERE id = ?`,
		x, y, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update schema item position: %w", err)
	}
	return nil
}

// Delete removes a schema item by ID
func (r *SchemaItemRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM schema_items WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete schema item: %w", err)
	}
	return nil
}

// DeleteByDeviceID removes all schema items for a device
func (r *SchemaItemRepository) DeleteByDeviceID(deviceID int64) error {
	_, err := r.db.Exec("DELETE FROM schema_items WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete device schema items: %w", err)
	}
	return nil
}

// GetByDeviceAndSchema retrieves a schema item by device and schema
func (r *SchemaItemRepository) GetByDeviceAndSchema(deviceID, schemaID int64) (*models.SchemaItem, error) {
	item := &models.SchemaItem{}
	err := r.db.QueryRow(`
		SELECT id, device_id, schema_id, x, y, width, height
		FROM schema_items WHERE device_id = ? AND schema_id = ?`,
		deviceID, schemaID,
	).Scan(&item.ID, &item.DeviceID, &item.SchemaID, &item.X, &item.Y, &item.Width, &item.Height)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get schema item: %w", err)
	}
	return item, nil
}
