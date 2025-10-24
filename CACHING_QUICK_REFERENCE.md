# Redis Caching - Developer Quick Reference

## How to Use Caching in Your Code

### 1. Getting User Data with Cache

```go
// Instead of:
user, err := db.GetUserByID(ctx, userID)

// Use cached repository:
user, err := cachedRepo.GetUserByID(ctx, userID)
// Returns *models.UserCache instead of database.User
```

### 2. Getting Documents with Cache

```go
// Get single document:
doc, err := cachedRepo.GetDocumentByID(ctx, docID)

// Get user's document list:
docs, err := cachedRepo.ListDocumentsByUser(ctx, userID)
// Returns []models.DocumentCache
```

### 3. Getting Share Data with Cache

```go
// Most important for performance - share access:
share, err := cachedRepo.GetShareByToken(ctx, token)
// Returns *models.ShareCache with joined document info
```

### 4. Invalidating Cache (Important!)

**After User Updates:**
```go
cachedRepo.InvalidateUser(ctx, userID, email)
```

**After Document Upload/Update:**
```go
cachedRepo.InvalidateUserDocuments(ctx, userID)
```

**After Document Delete:**
```go
cachedRepo.InvalidateDocument(ctx, docID, userID)
// Also invalidates user's document list
```

**After Share Access Count Update:**
```go
cachedRepo.InvalidateShare(ctx, token)
```

**Bulk Share Invalidation:**
```go
cachedRepo.InvalidateAllShares(ctx)
```

## Cache Keys Reference

| Operation | Cache Key Pattern | TTL |
|-----------|------------------|-----|
| User by ID | `user:id:{uuid}` | 30 min |
| User by email | `user:email:{email}` | 15 min |
| Document | `document:id:{uuid}` | 1 hour |
| Document list | `documents:user:{userID}` | 5 min |
| Share | `share:token:{token}` | 1 hour |

## Adding New Cached Operations

### Step 1: Create Cache DTO (if needed)

In `internal/models/cache.go`:

```go
type MyEntityCache struct {
    ID        string    `json:"id"`
    Data      string    `json:"data"`
    CreatedAt time.Time `json:"created_at"`
}

func FromDatabaseMyEntity(entity *database.MyEntity) *MyEntityCache {
    return &MyEntityCache{
        ID:        entity.ID.String(),
        Data:      entity.Data,
        CreatedAt: entity.CreatedAt.Time,
    }
}
```

### Step 2: Add Cache Method

In `internal/services/cache_helpers.go`:

```go
const (
    CacheKeyMyEntity = "myentity:id:%s"
    CacheTTLMyEntity = 30 * time.Minute
)

func (r *CachedRepository) GetMyEntityByID(ctx context.Context, id uuid.UUID) (*models.MyEntityCache, error) {
    cacheKey := fmt.Sprintf(CacheKeyMyEntity, id.String())

    // Try cache first
    var cached models.MyEntityCache
    err := r.cache.Get(ctx, cacheKey, &cached)
    if err == nil {
        return &cached, nil
    }

    if err != nil && !errors.Is(err, redis.Nil) {
        fmt.Printf("Cache error: %v\n", err)
    }

    // Cache miss - query database
    entity, err := r.db.GetMyEntity(ctx, pgtype.UUID{Bytes: id, Valid: true})
    if err != nil {
        return nil, err
    }

    // Convert to cache format
    cached = models.FromDatabaseMyEntity(&entity)

    // Store in cache
    _ = r.cache.Set(ctx, cacheKey, cached, CacheTTLMyEntity)

    return &cached, nil
}

func (r *CachedRepository) InvalidateMyEntity(ctx context.Context, id uuid.UUID) {
    cacheKey := fmt.Sprintf(CacheKeyMyEntity, id.String())
    _ = r.cache.Delete(ctx, cacheKey)
}
```

### Step 3: Use in Handler

```go
func (h *MyHandler) GetEntity(c *fiber.Ctx) error {
    id, _ := uuid.Parse(c.Params("id"))

    // Use cached repository
    entity, err := h.cache.GetMyEntityByID(c.Context(), id)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Not found"})
    }

    return c.JSON(entity)
}

func (h *MyHandler) UpdateEntity(c *fiber.Ctx) error {
    id, _ := uuid.Parse(c.Params("id"))

    // ... update entity in database ...

    // IMPORTANT: Invalidate cache after update!
    h.cache.InvalidateMyEntity(c.Context(), id)

    return c.JSON(fiber.Map{"status": "updated"})
}
```

## Common Patterns

### Cache-Aside Pattern (Current Implementation)
```go
// 1. Check cache
cached, err := cache.Get(ctx, key, &dest)
if err == nil {
    return cached  // Cache hit
}

// 2. Cache miss - query database
data, err := db.Query(...)
if err != nil {
    return err
}

// 3. Store in cache for next time
_ = cache.Set(ctx, key, data, ttl)

return data
```

### Cache Invalidation Pattern
```go
// After any mutation operation:
func (h *Handler) Update(...) error {
    // 1. Update database
    err := h.db.Update(...)
    if err != nil {
        return err
    }

    // 2. Invalidate cache (critical!)
    h.cache.InvalidateEntity(ctx, id)

    // 3. Return response
    return c.JSON(result)
}
```

## Debugging Cache Issues

### Check if Cache is Enabled
```go
if cache.IsEnabled() {
    log.Println("Cache is working")
} else {
    log.Println("Cache is disabled - check Redis connection")
}
```

### Test Cache Connection
```go
err := cache.Ping(ctx)
if err != nil {
    log.Printf("Cache ping failed: %v", err)
}
```

### Check if Key Exists
```go
exists, err := cache.Exists(ctx, "user:id:123")
if exists {
    log.Println("Key is in cache")
}
```

### Manually Clear Cache (for testing)
```go
// Clear specific key
cache.Delete(ctx, "user:id:123")

// Clear all users
cache.DeletePattern(ctx, "user:*")

// Clear all documents
cache.DeletePattern(ctx, "document:*")

// Clear everything (use with caution!)
cache.DeletePattern(ctx, "*")
```

## Testing Without Redis

The application works without Redis! If Redis is unavailable:
- Cache operations are silently skipped
- All queries fall back to database
- No errors are thrown
- Performance degrades gracefully

## Common Mistakes to Avoid

### ❌ Don't: Forget to invalidate cache after updates
```go
func Update() {
    db.Update(...)  // Updates DB
    // Missing: cache.Invalidate()
    // Result: Stale data in cache!
}
```

### ✅ Do: Always invalidate after mutations
```go
func Update() {
    db.Update(...)
    cache.Invalidate()  // Fresh data on next read
}
```

### ❌ Don't: Cache sensitive data without encryption
```go
cache.Set("user:password", plainPassword, ttl)  // BAD!
```

### ✅ Do: Only cache what's safe to cache
```go
cache.Set("user:id:123", userCache, ttl)  // Password hash, not plain password
```

### ❌ Don't: Ignore cache errors in critical paths
```go
cache.Get(key, &dest)  // What if Redis is down?
return dest  // Might return empty data!
```

### ✅ Do: Handle cache misses gracefully
```go
err := cache.Get(key, &dest)
if err != nil || dest == nil {
    // Fall back to database
    dest = db.Query()
}
return dest
```

## Performance Monitoring

### Redis CLI Commands for Monitoring
```bash
# Connect to Redis
redis-cli

# Check memory usage
INFO memory

# Monitor commands in real-time
MONITOR

# Get all keys (don't use in production with many keys!)
KEYS *

# Get specific pattern
KEYS user:*

# Check cache hit/miss ratio
INFO stats
```

### Key Metrics to Watch
- **keyspace_hits**: Number of cache hits
- **keyspace_misses**: Number of cache misses
- **Hit rate**: `hits / (hits + misses) * 100`
- **Memory usage**: Keep under 80% of maxmemory
- **Evicted keys**: Should be minimal

## Quick Troubleshooting

### Problem: Cache not working
**Check**:
1. Is Redis running? `redis-cli ping`
2. Check logs for "Redis cache unavailable"
3. Verify connection string: `localhost:6379`

### Problem: Stale data
**Solution**:
1. Check if invalidation is called after updates
2. Reduce TTL for that cache key
3. Manually clear cache: `cache.DeletePattern(ctx, "key:*")`

### Problem: High memory usage
**Solution**:
1. Set Redis maxmemory: `maxmemory 256mb`
2. Set eviction policy: `maxmemory-policy allkeys-lru`
3. Reduce TTLs for large objects
4. Monitor with `INFO memory`

### Problem: Slow cache operations
**Check**:
1. Redis server load: `INFO cpu`
2. Network latency: `redis-cli --latency`
3. Large values being cached
4. Too many connections: `INFO clients`

## Best Practices Checklist

- ✅ Always invalidate cache after mutations
- ✅ Use appropriate TTLs (shorter for frequently changing data)
- ✅ Handle cache misses gracefully
- ✅ Log cache errors but don't fail requests
- ✅ Monitor cache hit rates
- ✅ Use connection pooling (already configured)
- ✅ Never cache plain passwords or secrets
- ✅ Test with and without Redis available
- ✅ Use cache keys with clear patterns
- ✅ Document cache dependencies

## Need Help?

Refer to:
- `CACHING_IMPLEMENTATION.md` - Full implementation details
- `internal/services/cache_helpers.go` - Example implementations
- [Redis documentation](https://redis.io/documentation)
- [go-redis documentation](https://redis.uptrace.dev/)
