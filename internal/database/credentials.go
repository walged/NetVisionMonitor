package database

import (
	"database/sql"
	"fmt"
	"time"

	"netvisionmonitor/internal/encryption"
	"netvisionmonitor/internal/models"
)

// CredentialRepository handles credential database operations
type CredentialRepository struct {
	db *sql.DB
}

// NewCredentialRepository creates a new credential repository
func NewCredentialRepository(db *sql.DB) *CredentialRepository {
	return &CredentialRepository{db: db}
}

// Create inserts a new credential with encrypted password
func (r *CredentialRepository) Create(cred *models.Credential) error {
	encryptedPassword, err := encryption.EncryptIfNotEmpty(cred.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	encryptedUsername, err := encryption.EncryptIfNotEmpty(cred.Username)
	if err != nil {
		return fmt.Errorf("failed to encrypt username: %w", err)
	}

	result, err := r.db.Exec(`
		INSERT INTO credentials (name, type, username, password, note, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		cred.Name, cred.Type, encryptedUsername, encryptedPassword, cred.Note, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	cred.ID = id
	return nil
}

// GetByID retrieves a credential by ID (without decrypting password)
func (r *CredentialRepository) GetByID(id int64) (*models.Credential, error) {
	cred := &models.Credential{}
	var encryptedUsername, encryptedPassword string

	err := r.db.QueryRow(`
		SELECT id, name, type, username, password, note, created_at, updated_at
		FROM credentials WHERE id = ?`, id,
	).Scan(
		&cred.ID, &cred.Name, &cred.Type, &encryptedUsername, &encryptedPassword,
		&cred.Note, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Decrypt username
	cred.Username, err = encryption.DecryptIfNotEmpty(encryptedUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt username: %w", err)
	}

	// Don't decrypt password for normal retrieval
	// Password is only decrypted when explicitly needed

	return cred, nil
}

// GetByIDWithPassword retrieves a credential with decrypted password
func (r *CredentialRepository) GetByIDWithPassword(id int64) (*models.Credential, error) {
	cred := &models.Credential{}
	var encryptedUsername, encryptedPassword string

	err := r.db.QueryRow(`
		SELECT id, name, type, username, password, note, created_at, updated_at
		FROM credentials WHERE id = ?`, id,
	).Scan(
		&cred.ID, &cred.Name, &cred.Type, &encryptedUsername, &encryptedPassword,
		&cred.Note, &cred.CreatedAt, &cred.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}

	// Decrypt username
	cred.Username, err = encryption.DecryptIfNotEmpty(encryptedUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt username: %w", err)
	}

	// Decrypt password
	cred.Password, err = encryption.DecryptIfNotEmpty(encryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return cred, nil
}

// GetAll retrieves all credentials (without passwords)
func (r *CredentialRepository) GetAll() ([]models.Credential, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type, username, note, created_at, updated_at
		FROM credentials ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query credentials: %w", err)
	}
	defer rows.Close()

	var credentials []models.Credential
	for rows.Next() {
		var c models.Credential
		var encryptedUsername string
		err := rows.Scan(
			&c.ID, &c.Name, &c.Type, &encryptedUsername, &c.Note, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}

		c.Username, err = encryption.DecryptIfNotEmpty(encryptedUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt username: %w", err)
		}

		credentials = append(credentials, c)
	}
	return credentials, nil
}

// GetByType retrieves credentials by type
func (r *CredentialRepository) GetByType(credType models.CredentialType) ([]models.Credential, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type, username, note, created_at, updated_at
		FROM credentials WHERE type = ? ORDER BY name`, credType)
	if err != nil {
		return nil, fmt.Errorf("failed to query credentials: %w", err)
	}
	defer rows.Close()

	var credentials []models.Credential
	for rows.Next() {
		var c models.Credential
		var encryptedUsername string
		err := rows.Scan(
			&c.ID, &c.Name, &c.Type, &encryptedUsername, &c.Note, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}

		c.Username, err = encryption.DecryptIfNotEmpty(encryptedUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt username: %w", err)
		}

		credentials = append(credentials, c)
	}
	return credentials, nil
}

// Update updates an existing credential
func (r *CredentialRepository) Update(cred *models.Credential) error {
	encryptedUsername, err := encryption.EncryptIfNotEmpty(cred.Username)
	if err != nil {
		return fmt.Errorf("failed to encrypt username: %w", err)
	}

	// If password is provided, encrypt and update it
	if cred.Password != "" {
		encryptedPassword, err := encryption.EncryptIfNotEmpty(cred.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}

		_, err = r.db.Exec(`
			UPDATE credentials SET name = ?, type = ?, username = ?, password = ?, note = ?, updated_at = ?
			WHERE id = ?`,
			cred.Name, cred.Type, encryptedUsername, encryptedPassword, cred.Note, time.Now(), cred.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update credential: %w", err)
		}
	} else {
		// Don't update password
		_, err = r.db.Exec(`
			UPDATE credentials SET name = ?, type = ?, username = ?, note = ?, updated_at = ?
			WHERE id = ?`,
			cred.Name, cred.Type, encryptedUsername, cred.Note, time.Now(), cred.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update credential: %w", err)
		}
	}

	return nil
}

// Delete removes a credential by ID
func (r *CredentialRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM credentials WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	return nil
}

// Count returns the total number of credentials
func (r *CredentialRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM credentials").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count credentials: %w", err)
	}
	return count, nil
}
