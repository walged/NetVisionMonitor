package main

import (
	"fmt"

	"netvisionmonitor/internal/database"
	"netvisionmonitor/internal/models"
)

// CredentialInput is used for creating/updating credentials from frontend
type CredentialInput struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
	Note     string `json:"note"`
}

// GetCredentials returns all credentials (without passwords)
func (a *App) GetCredentials() ([]models.Credential, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	repo := database.NewCredentialRepository(a.db.DB())
	return repo.GetAll()
}

// GetCredentialsByType returns credentials of a specific type
func (a *App) GetCredentialsByType(credType string) ([]models.Credential, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	repo := database.NewCredentialRepository(a.db.DB())
	return repo.GetByType(models.CredentialType(credType))
}

// GetCredential returns a credential by ID (without password)
func (a *App) GetCredential(id int64) (*models.Credential, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	repo := database.NewCredentialRepository(a.db.DB())
	return repo.GetByID(id)
}

// CreateCredential creates a new credential
func (a *App) CreateCredential(input CredentialInput) (*models.Credential, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Validate input
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}

	cred := &models.Credential{
		Name:     input.Name,
		Type:     models.CredentialType(input.Type),
		Username: input.Username,
		Password: input.Password,
		Note:     input.Note,
	}

	repo := database.NewCredentialRepository(a.db.DB())
	if err := repo.Create(cred); err != nil {
		return nil, err
	}

	// Don't return password
	cred.Password = ""
	return cred, nil
}

// UpdateCredential updates an existing credential
func (a *App) UpdateCredential(input CredentialInput) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	if input.ID == 0 {
		return fmt.Errorf("credential ID is required")
	}

	cred := &models.Credential{
		ID:       input.ID,
		Name:     input.Name,
		Type:     models.CredentialType(input.Type),
		Username: input.Username,
		Password: input.Password, // Will be encrypted if not empty
		Note:     input.Note,
	}

	repo := database.NewCredentialRepository(a.db.DB())
	return repo.Update(cred)
}

// DeleteCredential deletes a credential by ID
func (a *App) DeleteCredential(id int64) error {
	if a.db == nil {
		return fmt.Errorf("database not initialized")
	}

	repo := database.NewCredentialRepository(a.db.DB())
	return repo.Delete(id)
}
