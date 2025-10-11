package handlers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"Secure-Document-Exchange-Portal/internal/auth"
	"Secure-Document-Exchange-Portal/internal/database"
	"Secure-Document-Exchange-Portal/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"
)

type DocumentHandler struct {
	db       *database.Queries
	storage  services.StorageService
	// encryption services.EncryptionService // TODO: add when implemented
}

func NewDocumentHandler(db *database.Queries, storage services.StorageService) *DocumentHandler {
	return &DocumentHandler{
		db:      db,
		storage: storage,
	}
}

func (h *DocumentHandler) Upload(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required: " + err.Error()})
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer src.Close()

	// Read file content for checksum and encryption
	fileData, err := io.ReadAll(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file"})
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", sha256.Sum256(fileData))

	// TODO: Encrypt file data
	encryptedData := fileData // Placeholder
	encryptionKey := "placeholder-key" // TODO: generate and encrypt key

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s/%s%s", userID.String(), uuid.New().String(), ext)

	// Upload to storage
	reader := bytes.NewReader(encryptedData)
	_, err = h.storage.Upload(c.Context(), "documents", objectName, reader, int64(len(encryptedData)), minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upload file to storage: " + err.Error()})
	}

	// Save to database
	doc, err := h.db.CreateDocument(c.Context(), database.CreateDocumentParams{
		UserID:       pgtype.UUID{Bytes: userID, Valid: true},
		Filename:     file.Filename,
		FilePath:     objectName,
		EncryptedKey: encryptionKey,
		FileSize:     file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		Checksum:     checksum,
	})
	if err != nil {
		// TODO: delete from storage on error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save document to database: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         doc.ID.String(),
		"filename":   doc.Filename,
		"file_size":  doc.FileSize,
		"mime_type":  doc.MimeType,
		"created_at": doc.CreatedAt.Time.Format(time.RFC3339),
	})
}

func (h *DocumentHandler) List(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	docs, err := h.db.ListDocumentsByUser(c.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list documents"})
	}

	var result []fiber.Map
	for _, doc := range docs {
		result = append(result, fiber.Map{
			"id":         doc.ID.String(),
			"filename":   doc.Filename,
			"file_size":  doc.FileSize,
			"mime_type":  doc.MimeType,
			"created_at": doc.CreatedAt.Time.Format(time.RFC3339),
		})
	}

	return c.JSON(result)
}

func (h *DocumentHandler) Download(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	docIDStr := c.Params("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid document ID"})
	}

	doc, err := h.db.GetDocumentByID(c.Context(), pgtype.UUID{Bytes: docID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Document not found"})
	}

	// Check ownership
	if !bytes.Equal(doc.UserID.Bytes[:], userID[:]) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Download from storage
	obj, err := h.storage.Download(c.Context(), "documents", doc.FilePath, minio.GetObjectOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to download file"})
	}
	defer obj.Close()

	// TODO: Decrypt data
	data, err := io.ReadAll(obj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file"})
	}

	decryptedData := data // Placeholder

	// Send file
	c.Set("Content-Type", doc.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.Filename))
	return c.Send(decryptedData)
}

func (h *DocumentHandler) Delete(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	docIDStr := c.Params("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid document ID"})
	}

	// Delete from database (which checks ownership via user_id)
	err = h.db.DeleteDocument(c.Context(), database.DeleteDocumentParams{
		ID:     pgtype.UUID{Bytes: docID, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Document not found or access denied"})
	}

	// TODO: Delete from storage
	// For now, assume deleted

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *DocumentHandler) CreateShare(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return err
	}

	docIDStr := c.Params("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid document ID"})
	}

	// Check ownership
	doc, err := h.db.GetDocumentByID(c.Context(), pgtype.UUID{Bytes: docID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Document not found"})
	}

	if !bytes.Equal(doc.UserID.Bytes[:], userID[:]) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Parse request for expiration, max_access, password
	var req struct {
		ExpiresAt *time.Time `json:"expires_at"`
		MaxAccess *int       `json:"max_access"`
		Password  *string    `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	expiresAt := time.Now().Add(24 * time.Hour) // default
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}

	maxAccess := -1
	if req.MaxAccess != nil {
		maxAccess = *req.MaxAccess
	}

	var passwordHash *string
	if req.Password != nil {
		// TODO: hash password
		passwordHash = req.Password
	}

	// Generate token
	shareToken := uuid.New().String()

	// Create share
	var passwordText pgtype.Text
	if passwordHash != nil {
		passwordText = pgtype.Text{String: *passwordHash, Valid: true}
	} else {
		passwordText = pgtype.Text{Valid: false}
	}

	share, err := h.db.CreateShare(c.Context(), database.CreateShareParams{
		DocumentID:   pgtype.UUID{Bytes: docID, Valid: true},
		ShareToken:   shareToken,
		ExpiresAt:    pgtype.Timestamptz{Time: expiresAt, Valid: true},
		MaxAccess:    pgtype.Int4{Int32: int32(maxAccess), Valid: true},
		PasswordHash: passwordText,
		CreatedBy:    pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create share"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"share_token": share.ShareToken,
		"expires_at":  share.ExpiresAt.Time.Format(time.RFC3339),
		"max_access":  share.MaxAccess.Int32,
	})
}
