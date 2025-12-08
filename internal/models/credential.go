package models

import "time"

type CredentialType string

const (
	CredentialTypeSNMP  CredentialType = "snmp"
	CredentialTypeRTSP  CredentialType = "rtsp"
	CredentialTypeONVIF CredentialType = "onvif"
	CredentialTypeSSH   CredentialType = "ssh"
)

type Credential struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Type      CredentialType `json:"type"`
	Username  string         `json:"username"`
	Password  string         `json:"-"` // Never expose in JSON
	Note      string         `json:"note"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// CredentialInput is used for creating/updating credentials
type CredentialInput struct {
	Name     string         `json:"name"`
	Type     CredentialType `json:"type"`
	Username string         `json:"username"`
	Password string         `json:"password"`
	Note     string         `json:"note"`
}
