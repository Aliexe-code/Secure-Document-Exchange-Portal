package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Secure-Document-Exchange-Portal/internal/database"
	"Secure-Document-Exchange-Portal/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

// Cache key patterns
const (
	CacheKeyUserByID       = "user:id:%s"           // user:id:{uuid}
	CacheKeyUserByEmail    = "user:email:%s"        // user:email:{email}
	CacheKeyDocument       = "document:id:%s"       // document:id:{uuid}
	CacheKeyDocumentsList  = "documents:user:%s"    // documents:user:{userID}
	CacheKeyShare          = "share:token:%s"       // share:token:{token}
	CacheKeyShareByID      = "share:id:%s"          // share:id:{uuid}

	// Cache TTLs
	CacheTTLUser          = 30 * time.Minute
	CacheTTLUserByEmail   = 15 * time.Minute
	CacheTTLDocument      = 1 * time.Hour
	CacheTTLDocumentsList = 5 * time.Minute
	CacheTTLShare         = 1 * time.Hour
)

// CachedRepository provides caching layer for database operations
type CachedRepository struct {
	db    *database.Queries
	cache CacheService
}

// NewCachedRepository creates a new cached repository
func NewCachedRepository(db *database.Queries, cache CacheService) *CachedRepository {
	return &CachedRepository{
		db:    db,
		cache: cache,
	}
}

// GetUserByID retrieves a user with caching
func (r *CachedRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.UserCache, error) {
	cacheKey := fmt.Sprintf(CacheKeyUserByID, userID.String())

	// Try cache first
	var cachedUser models.UserCache
	err := r.cache.Get(ctx, cacheKey, &cachedUser)
	if err == nil {
		// Cache hit
		return &cachedUser, nil
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		// Log but continue on cache errors
		fmt.Printf("Cache error for user %s: %v\n", userID, err)
	}

	// Cache miss - query database
	user, err := r.db.GetUserByID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	// Convert to cache-friendly format
	userCache := models.FromDatabaseUser(&user)

	// Store in cache (ignore errors)
	_ = r.cache.Set(ctx, cacheKey, userCache, CacheTTLUser)

	return userCache, nil
}

// GetUserByEmail retrieves a user by email with caching
func (r *CachedRepository) GetUserByEmail(ctx context.Context, email string) (*models.UserCache, error) {
	cacheKey := fmt.Sprintf(CacheKeyUserByEmail, email)

	// Try cache first
	var cachedUser models.UserCache
	err := r.cache.Get(ctx, cacheKey, &cachedUser)
	if err == nil {
		// Cache hit
		return &cachedUser, nil
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		fmt.Printf("Cache error for user email %s: %v\n", email, err)
	}

	// Cache miss - query database
	user, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Convert to cache-friendly format
	userCache := models.FromDatabaseUser(&user)

	// Store in cache
	_ = r.cache.Set(ctx, cacheKey, userCache, CacheTTLUserByEmail)

	return userCache, nil
}

// InvalidateUser removes all user cache entries
func (r *CachedRepository) InvalidateUser(ctx context.Context, userID uuid.UUID, email string) {
	_ = r.cache.Delete(ctx, fmt.Sprintf(CacheKeyUserByID, userID.String()))
	_ = r.cache.Delete(ctx, fmt.Sprintf(CacheKeyUserByEmail, email))
}

// GetDocumentByID retrieves a document with caching
func (r *CachedRepository) GetDocumentByID(ctx context.Context, docID uuid.UUID) (*models.DocumentCache, error) {
	cacheKey := fmt.Sprintf(CacheKeyDocument, docID.String())

	// Try cache first
	var cachedDoc models.DocumentCache
	err := r.cache.Get(ctx, cacheKey, &cachedDoc)
	if err == nil {
		return &cachedDoc, nil
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		fmt.Printf("Cache error for document %s: %v\n", docID, err)
	}

	// Cache miss - query database
	doc, err := r.db.GetDocumentByID(ctx, pgtype.UUID{Bytes: docID, Valid: true})
	if err != nil {
		return nil, err
	}

	// Convert to cache-friendly format
	docCache := models.FromDatabaseDocument(&doc)

	// Store in cache
	_ = r.cache.Set(ctx, cacheKey, docCache, CacheTTLDocument)

	return docCache, nil
}

// ListDocumentsByUser retrieves documents for a user with caching
func (r *CachedRepository) ListDocumentsByUser(ctx context.Context, userID uuid.UUID) ([]models.DocumentCache, error) {
	cacheKey := fmt.Sprintf(CacheKeyDocumentsList, userID.String())

	// Try cache first
	var cachedList models.DocumentListCache
	err := r.cache.Get(ctx, cacheKey, &cachedList)
	if err == nil {
		// Check if cache is not too stale (additional freshness check)
		if time.Since(cachedList.CachedAt) < CacheTTLDocumentsList {
			return cachedList.Documents, nil
		}
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		fmt.Printf("Cache error for user documents %s: %v\n", userID, err)
	}

	// Cache miss - query database
	docs, err := r.db.ListDocumentsByUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	// Convert to cache-friendly format
	var docCaches []models.DocumentCache
	for _, doc := range docs {
		docCaches = append(docCaches, *models.FromDatabaseDocument(&doc))
	}

	// Store in cache
	listCache := models.DocumentListCache{
		Documents: docCaches,
		CachedAt:  time.Now(),
	}
	_ = r.cache.Set(ctx, cacheKey, listCache, CacheTTLDocumentsList)

	return docCaches, nil
}

// InvalidateUserDocuments removes document list cache for a user
func (r *CachedRepository) InvalidateUserDocuments(ctx context.Context, userID uuid.UUID) {
	cacheKey := fmt.Sprintf(CacheKeyDocumentsList, userID.String())
	_ = r.cache.Delete(ctx, cacheKey)
}

// InvalidateDocument removes document cache
func (r *CachedRepository) InvalidateDocument(ctx context.Context, docID uuid.UUID, userID uuid.UUID) {
	_ = r.cache.Delete(ctx, fmt.Sprintf(CacheKeyDocument, docID.String()))
	// Also invalidate the user's document list
	r.InvalidateUserDocuments(ctx, userID)
}

// GetShareByToken retrieves a share by token with document info (for AccessShare)
// This is the most frequently accessed query and benefits most from caching
func (r *CachedRepository) GetShareByToken(ctx context.Context, token string) (*models.ShareCache, error) {
	cacheKey := fmt.Sprintf(CacheKeyShare, token)

	// Try cache first
	var cachedShare models.ShareCache
	err := r.cache.Get(ctx, cacheKey, &cachedShare)
	if err == nil {
		return &cachedShare, nil
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		fmt.Printf("Cache error for share token %s: %v\n", token, err)
	}

	// Cache miss - query database
	// Note: The database query returns joined data with document info
	shareData, err := r.db.GetShareByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Convert to cache-friendly format with document info
	shareCache := &models.ShareCache{
		ID:           shareData.ID.String(),
		DocumentID:   shareData.DocumentID.String(),
		ShareToken:   shareData.ShareToken,
		ExpiresAt:    shareData.ExpiresAt.Time,
		MaxAccess:    shareData.MaxAccess.Int32,
		AccessCount:  shareData.AccessCount.Int32,
		CreatedAt:    shareData.CreatedAt.Time,
		CreatedBy:    shareData.CreatedBy.String(),
		Filename:     shareData.Filename,
		FilePath:     shareData.FilePath,
		FileSize:     shareData.FileSize,
		MimeType:     shareData.MimeType,
	}

	if shareData.PasswordHash.Valid {
		shareCache.PasswordHash = &shareData.PasswordHash.String
	}

	// Store in cache with shorter TTL (to ensure freshness for access counts)
	_ = r.cache.Set(ctx, cacheKey, shareCache, CacheTTLShare)

	return shareCache, nil
}

// InvalidateShare removes share cache
func (r *CachedRepository) InvalidateShare(ctx context.Context, token string) {
	cacheKey := fmt.Sprintf(CacheKeyShare, token)
	_ = r.cache.Delete(ctx, cacheKey)
}

// InvalidateAllShares removes all share caches (useful after access count updates)
func (r *CachedRepository) InvalidateAllShares(ctx context.Context) {
	_ = r.cache.DeletePattern(ctx, "share:*")
}
