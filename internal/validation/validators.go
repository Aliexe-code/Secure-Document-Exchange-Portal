package validation

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"
)

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if an email is in valid format
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if len(email) > 255 {
		return fmt.Errorf("email is too long (max 255 characters)")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePassword checks password strength
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 12 {
		return fmt.Errorf("password must be at least 12 characters long")
	}
	if len(password) > 128 {
		return fmt.Errorf("password is too long (max 128 characters)")
	}

	// Check for at least one uppercase, lowercase, digit, and special char
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password)

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateFullName checks if full name is valid
func ValidateFullName(name string) error {
	if name == "" {
		return fmt.Errorf("full name is required")
	}
	if len(name) < 2 {
		return fmt.Errorf("full name must be at least 2 characters")
	}
	if len(name) > 255 {
		return fmt.Errorf("full name is too long (max 255 characters)")
	}

	// Only allow letters, spaces, hyphens, and apostrophes
	validName := regexp.MustCompile(`^[a-zA-Z\s\-']+$`).MatchString(name)
	if !validName {
		return fmt.Errorf("full name can only contain letters, spaces, hyphens, and apostrophes")
	}

	return nil
}

// Allowed MIME types for file uploads
var allowedMimeTypes = map[string]bool{
	// Documents
	"application/pdf":        true,
	"application/msword":     true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel":                                                true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	"application/vnd.ms-powerpoint":                                           true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain":             true,
	"text/csv":               true,

	// Images
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/svg+xml": true,

	// Archives
	"application/zip":        true,
	"application/x-zip-compressed": true,
	"application/x-rar-compressed": true,
	"application/x-7z-compressed":  true,
	"application/gzip":       true,
	"application/x-tar":      true,

	// Code
	"text/html":        true,
	"text/css":         true,
	"text/javascript":  true,
	"application/json": true,
	"application/xml":  true,
}

// ValidateFile checks file size and type
func ValidateFile(file *multipart.FileHeader) error {
	// Check file size (100 MB max)
	maxSize := int64(100 * 1024 * 1024) // 100 MB
	if file.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed (100 MB)")
	}

	if file.Size == 0 {
		return fmt.Errorf("file is empty")
	}

	// Check MIME type
	contentType := file.Header.Get("Content-Type")
	if !allowedMimeTypes[contentType] {
		return fmt.Errorf("file type not allowed: %s", contentType)
	}

	// Validate filename
	if err := ValidateFilename(file.Filename); err != nil {
		return err
	}

	return nil
}

// ValidateFilename checks for malicious or problematic filenames
func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}

	if len(filename) > 255 {
		return fmt.Errorf("filename is too long (max 255 characters)")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename contains invalid characters")
	}

	// Check for null bytes
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("filename contains null bytes")
	}

	// Ensure it has a valid extension
	ext := filepath.Ext(filename)
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	// Check for dangerous extensions
	dangerousExts := []string{".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar", ".sh"}
	lowerExt := strings.ToLower(ext)
	for _, dangerous := range dangerousExts {
		if lowerExt == dangerous {
			return fmt.Errorf("file type not allowed: %s", ext)
		}
	}

	return nil
}

// ValidateShareExpiration validates expiration days and hours
func ValidateShareExpiration(days, hours int) error {
	if days < 0 || hours < 0 {
		return fmt.Errorf("expiration values cannot be negative")
	}

	totalHours := (days * 24) + hours
	if totalHours == 0 {
		return fmt.Errorf("expiration time must be greater than 0")
	}

	// Maximum 30 days
	if totalHours > (30 * 24) {
		return fmt.Errorf("expiration time cannot exceed 30 days")
	}

	return nil
}

// ValidateShareMaxAccess validates max access count
func ValidateShareMaxAccess(maxAccess int) error {
	if maxAccess < -1 {
		return fmt.Errorf("max access must be -1 (unlimited) or a positive number")
	}

	if maxAccess > 10000 {
		return fmt.Errorf("max access cannot exceed 10,000")
	}

	return nil
}

// ValidateSharePassword validates share password
func ValidateSharePassword(password string) error {
	if password == "" {
		return nil // Password is optional
	}

	if len(password) < 6 {
		return fmt.Errorf("share password must be at least 6 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("share password is too long (max 128 characters)")
	}

	return nil
}
