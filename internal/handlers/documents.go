package handlers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"Secure-Document-Exchange-Portal/internal/auth"
	"Secure-Document-Exchange-Portal/internal/database"
	"Secure-Document-Exchange-Portal/internal/services"
	"Secure-Document-Exchange-Portal/internal/validation"
	"Secure-Document-Exchange-Portal/templates"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
)

type DocumentHandler struct {
	db      *database.Queries
	storage services.StorageService
	cache   *services.CachedRepository
	// encryption services.EncryptionService // TODO: add when implemented
}

func NewDocumentHandler(db *database.Queries, storage services.StorageService, cache *services.CachedRepository) *DocumentHandler {
	return &DocumentHandler{
		db:      db,
		storage: storage,
		cache:   cache,
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

	// Validate file size and type
	if err := validation.ValidateFile(file); err != nil {
		if c.Get("HX-Request") == "true" {
			errorMsg := fmt.Sprintf(`<div class="mb-4 p-4 bg-red-100 border border-red-400 text-red-700 rounded">
				<p class="font-semibold">✗ Upload failed</p>
				<p class="text-sm mt-1">%s</p>
			</div>`, err.Error())
			return c.Status(fiber.StatusBadRequest).SendString(errorMsg)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
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

	// Invalidate user's document list cache
	h.cache.InvalidateUserDocuments(c.Context(), userID)

	// Check if request is from HTMX
	if c.Get("HX-Request") == "true" {
		// Return user-friendly HTML message and trigger document list refresh
		successMsg := fmt.Sprintf(`<div class="mb-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
			<p class="font-semibold">✓ File uploaded successfully!</p>
			<p class="text-sm mt-1">%s (%.2f MB)</p>
		</div>`, doc.Filename, float64(doc.FileSize)/1024/1024)
		c.Set("Content-Type", "text/html")
		c.Set("HX-Trigger", "documentUploaded")
		return c.Status(fiber.StatusCreated).SendString(successMsg)
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

	// Use cached repository for document list
	docs, err := h.cache.ListDocumentsByUser(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list documents"})
	}

	// Check if request accepts HTML (HTMX request)
	if c.Get("Accept") == "text/html" || c.Get("HX-Request") == "true" {
		// Return HTML template
		var templateDocs []templates.Document
		for _, doc := range docs {
			templateDocs = append(templateDocs, templates.Document{
				ID:        doc.ID,
				Filename:  doc.Filename,
				FileSize:  doc.FileSize,
				MimeType:  doc.MimeType,
				CreatedAt: doc.CreatedAt.Format(time.RFC3339),
			})
		}
		c.Set("Content-Type", "text/html")
		return templates.DocumentList(templateDocs).Render(c.Context(), c.Response().BodyWriter())
	}

	// Default JSON response
	var result []fiber.Map
	for _, doc := range docs {
		result = append(result, fiber.Map{
			"id":         doc.ID,
			"filename":   doc.Filename,
			"file_size":  doc.FileSize,
			"mime_type":  doc.MimeType,
			"created_at": doc.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(result)
}

func (h *DocumentHandler) Download(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authentication failed"})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to download file from storage"})
	}
	defer obj.Close()

	// Stream the file directly to response
	c.Set("Content-Type", doc.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.Filename))

	_, err = io.Copy(c.Response().BodyWriter(), obj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send file"})
	}

	return nil
}

func (h *DocumentHandler) View(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authentication failed"})
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

	// Check if it's an image type
	mimeType := strings.ToLower(doc.MimeType)
	isImage := strings.HasPrefix(mimeType, "image/")

	if isImage {
		// For images, return inline preview
		obj, err := h.storage.Download(c.Context(), "documents", doc.FilePath, minio.GetObjectOptions{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to download file from storage"})
		}
		defer obj.Close()

		c.Set("Content-Type", doc.MimeType)
		_, err = io.Copy(c.Response().BodyWriter(), obj)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send file"})
		}
		return nil
	} else {
		// For other files, show preview modal with download link
		html := fmt.Sprintf(`
		<div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center" hx-target="this" hx-swap="outerHTML">
			<div class="bg-white p-8 rounded-lg shadow-lg max-w-md w-full mx-4">
				<h3 class="text-lg font-semibold mb-4">File Preview</h3>
				<div class="mb-4">
					<p class="text-sm text-gray-600 mb-2"><strong>Filename:</strong> %s</p>
					<p class="text-sm text-gray-600 mb-2"><strong>Type:</strong> %s</p>
					<p class="text-sm text-gray-600 mb-2"><strong>Size:</strong> %.2f MB</p>
					<p class="text-sm text-gray-600 mb-4"><strong>Uploaded:</strong> %s</p>
				</div>
				<div class="flex space-x-2">
					<a href="/api/documents/%s/download" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700" target="_blank">
						Download File
					</a>
					<button onclick="this.closest('[hx-target]').style.display='none'" class="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700">
						Close
					</button>
				</div>
			</div>
		</div>`, doc.Filename, doc.MimeType, float64(doc.FileSize)/1024/1024, doc.CreatedAt.Time.Format("2006-01-02 15:04:05"), docIDStr)

		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}

func (h *DocumentHandler) Delete(c *fiber.Ctx) error {
	userID, err := auth.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authentication failed"})
	}

	docIDStr := c.Params("id")
	docID, err := uuid.Parse(docIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid document ID"})
	}

	// Get document info first (for storage deletion)
	doc, err := h.db.GetDocumentByID(c.Context(), pgtype.UUID{Bytes: docID, Valid: true})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Document not found"})
	}

	// Check ownership
	if !bytes.Equal(doc.UserID.Bytes[:], userID[:]) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
	}

	// Delete from storage first
	err = h.storage.Delete(c.Context(), "documents", doc.FilePath, minio.RemoveObjectOptions{})
	if err != nil {
		// Log error but continue with DB deletion
		fmt.Printf("Failed to delete file from storage: %v\n", err)
	}

	// Delete from database (which checks ownership via user_id)
	err = h.db.DeleteDocument(c.Context(), database.DeleteDocumentParams{
		ID:     pgtype.UUID{Bytes: docID, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Document not found or access denied"})
	}

	// Invalidate document cache
	h.cache.InvalidateDocument(c.Context(), docID, userID)

	// Check if request expects HTML (HTMX)
	if c.Get("Accept") == "text/html" || c.Get("HX-Request") == "true" {
		// Trigger document list refresh after deletion
		c.Set("HX-Trigger", "documentUploaded")
		return c.SendStatus(fiber.StatusOK)
	}

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

	// Parse form fields for expiration, max_access, password
	expireDaysStr := c.FormValue("expire_days")
	expireHoursStr := c.FormValue("expire_hours")
	maxAccessStr := c.FormValue("max_access")
	password := c.FormValue("password")

	// Calculate expiration time (default: 24 hours)
	expiresAt := time.Now().Add(24 * time.Hour)

	var expireDays, expireHours int
	if expireDaysStr != "" {
		fmt.Sscanf(expireDaysStr, "%d", &expireDays)
	}
	if expireHoursStr != "" {
		fmt.Sscanf(expireHoursStr, "%d", &expireHours)
	}

	// Validate expiration values
	if expireDays > 0 || expireHours > 0 {
		if err := validation.ValidateShareExpiration(expireDays, expireHours); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		totalHours := (expireDays * 24) + expireHours
		expiresAt = time.Now().Add(time.Duration(totalHours) * time.Hour)
	}

	// Parse and validate max access count
	maxAccess := -1
	if maxAccessStr != "" {
		fmt.Sscanf(maxAccessStr, "%d", &maxAccess)
		if err := validation.ValidateShareMaxAccess(maxAccess); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	// Validate share password if provided
	if err := validation.ValidateSharePassword(password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Handle password - hash it using bcrypt
	var passwordHash *string
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to process password"})
		}
		hashStr := string(hashedPassword)
		passwordHash = &hashStr
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

	// Check if request expects HTML (HTMX)
	if c.Get("Accept") == "text/html" || c.Get("HX-Request") == "true" {
		c.Set("Content-Type", "text/html")
		
		// Format expiration info
		duration := time.Until(share.ExpiresAt.Time)
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24
		
		var expiryText string
		if days > 0 && hours > 0 {
			expiryText = fmt.Sprintf("%d day(s) and %d hour(s)", days, hours)
		} else if days > 0 {
			expiryText = fmt.Sprintf("%d day(s)", days)
		} else {
			expiryText = fmt.Sprintf("%d hour(s)", hours)
		}
		
		accessInfo := ""
		if maxAccess > 0 {
			accessInfo = fmt.Sprintf("<p class=\"text-sm\">Max accesses: %d</p>", maxAccess)
		}
		
		return c.SendString(fmt.Sprintf(`<div class="p-4 bg-green-100 border border-green-400 text-green-700 rounded">
		<p class="font-semibold">✓ Share link created successfully!</p>
		<p class="text-sm mt-1">Expires in: %s</p>
		%s
		<div class="mt-3 p-2 bg-white rounded border border-green-300">
			<p class="text-xs text-gray-600 mb-1">Share URL:</p>
			<p class="text-sm font-mono break-all">/api/share/%s</p>
		</div>
		<button type="button" onclick="navigator.clipboard.writeText(window.location.origin + '/api/share/%s'); this.textContent='✓ Copied!'; setTimeout(() => this.textContent='Copy Link', 2000)" class="mt-3 bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700">Copy Link</button>
	</div>`, expiryText, accessInfo, share.ShareToken, share.ShareToken))
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"share_token": share.ShareToken,
		"expires_at":  share.ExpiresAt.Time.Format(time.RFC3339),
		"max_access":  share.MaxAccess.Int32,
	})
}

func (h *DocumentHandler) GetShareForm(c *fiber.Ctx) error {
	docID := c.Params("id")
	c.Set("Content-Type", "text/html")
	return templates.ShareForm(docID).Render(c.Context(), c.Response().BodyWriter())
}
