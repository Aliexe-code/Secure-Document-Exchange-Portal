package models

import (
	"time"

	"Secure-Document-Exchange-Portal/internal/database"
)

// Cache-friendly DTOs that avoid pgtype marshaling issues

// UserCache represents a cached user object
type UserCache struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	FullName     string    `json:"full_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

// FromDatabaseUser converts database.User to UserCache
func FromDatabaseUser(user *database.User) *UserCache {
	if user == nil {
		return nil
	}
	return &UserCache{
		ID:           user.ID.String(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
		IsActive:     user.IsActive.Bool,
	}
}

// DocumentCache represents a cached document object
type DocumentCache struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Filename     string    `json:"filename"`
	FilePath     string    `json:"file_path"`
	EncryptedKey string    `json:"encrypted_key"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	Checksum     string    `json:"checksum"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FromDatabaseDocument converts database.Document to DocumentCache
func FromDatabaseDocument(doc *database.Document) *DocumentCache {
	if doc == nil {
		return nil
	}
	return &DocumentCache{
		ID:           doc.ID.String(),
		UserID:       doc.UserID.String(),
		Filename:     doc.Filename,
		FilePath:     doc.FilePath,
		EncryptedKey: doc.EncryptedKey,
		FileSize:     doc.FileSize,
		MimeType:     doc.MimeType,
		Checksum:     doc.Checksum,
		CreatedAt:    doc.CreatedAt.Time,
		UpdatedAt:    doc.UpdatedAt.Time,
	}
}

// ShareCache represents a cached share object with joined document info
type ShareCache struct {
	ID           string    `json:"id"`
	DocumentID   string    `json:"document_id"`
	ShareToken   string    `json:"share_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	MaxAccess    int32     `json:"max_access"`
	AccessCount  int32     `json:"access_count"`
	PasswordHash *string   `json:"password_hash,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`

	// Joined document information for share access
	Filename     string `json:"filename"`
	FilePath     string `json:"file_path"`
	FileSize     int64  `json:"file_size"`
	MimeType     string `json:"mime_type"`
}

// FromDatabaseShare converts database.Share to ShareCache (without document info)
func FromDatabaseShare(share *database.Share) *ShareCache {
	if share == nil {
		return nil
	}

	var passwordHash *string
	if share.PasswordHash.Valid {
		passwordHash = &share.PasswordHash.String
	}

	return &ShareCache{
		ID:           share.ID.String(),
		DocumentID:   share.DocumentID.String(),
		ShareToken:   share.ShareToken,
		ExpiresAt:    share.ExpiresAt.Time,
		MaxAccess:    share.MaxAccess.Int32,
		AccessCount:  share.AccessCount.Int32,
		PasswordHash: passwordHash,
		CreatedAt:    share.CreatedAt.Time,
		CreatedBy:    share.CreatedBy.String(),
	}
}

// DocumentListCache represents a cached list of documents for a user
type DocumentListCache struct {
	Documents []DocumentCache `json:"documents"`
	CachedAt  time.Time       `json:"cached_at"`
}
