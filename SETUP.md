# Setup Guide for SDEP

## Quick Start

The application now supports **automatic fallback to local file storage** if MinIO is not available. This means you can start using the application immediately without setting up external services.

### Running the Application

1. **Start the application**:
   ```bash
   go run main.go
   # or if using air for hot reload:
   air
   ```

2. The application will automatically:
   - Try to connect to MinIO on `localhost:9000`
   - If MinIO is not available, fall back to local file storage in `./storage` directory
   - Files will be stored locally on disk

3. Open your browser and navigate to: `http://localhost:8080`

## Storage Options

### Option 1: Local File Storage (Default Fallback)

No setup required! The application will automatically use local file storage if MinIO is unavailable.

- **Location**: `./storage/documents/`
- **Pros**: No external services needed, simple setup
- **Cons**: Not suitable for production, no distributed storage

### Option 2: MinIO Storage (Recommended for Production)

To use MinIO for scalable object storage:

#### With Docker/Podman

If you have Docker or Podman installed:

```bash
# Using docker
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# Create the bucket
docker exec minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec minio mc mb local/documents
```

#### Using Docker Compose

```bash
# Start all services (MinIO, Redis, PostgreSQL)
docker compose up -d

# Or use the provided script
chmod +x start-services.sh
./start-services.sh
```

#### Standalone MinIO Installation

1. Download MinIO binary:
   ```bash
   wget https://dl.min.io/server/minio/release/linux-amd64/minio
   chmod +x minio
   ```

2. Start MinIO:
   ```bash
   ./minio server ./minio-data --console-address ":9001"
   ```

3. Create bucket:
   ```bash
   # Install mc (MinIO Client)
   wget https://dl.min.io/client/mc/release/linux-amd64/mc
   chmod +x mc
   
   # Configure and create bucket
   ./mc alias set local http://localhost:9000 minioadmin minioadmin
   ./mc mb local/documents
   ```

4. Access MinIO Console: `http://localhost:9001` (minioadmin/minioadmin)

## Environment Variables

Copy `.env.example` to `.env` and update if needed:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=sdep

# JWT
JWT_SECRET=your-secret-key

# MinIO (Optional - will fall back to local storage if unavailable)
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=documents

# Redis (Optional)
REDIS_URL=redis://localhost:6379
```

## Testing the Upload Feature

1. Register a new user at `/register`
2. Login with your credentials
3. Click "Upload Document"
4. Select a file (image, PDF, or any document)
5. Click "Upload"
6. The file will be stored either in MinIO (if available) or in `./storage/documents/` directory

## Troubleshooting

### Upload returns 500 error

**Solution**: The application now automatically falls back to local storage. Just restart your server and it should work with local file storage.

### Want to use MinIO instead of local storage

1. Start MinIO (see options above)
2. Create the `documents` bucket
3. Restart your application - it will detect MinIO and use it automatically

### Files not showing up

- Check that PostgreSQL database is running
- Verify database migrations have been applied
- Check application logs for errors

## Production Deployment

For production, it's recommended to:

1. Use MinIO or AWS S3 for object storage
2. Use Redis for caching (optional but recommended)
3. Set proper JWT secrets
4. Enable HTTPS
5. Configure proper backup strategies

See `README.md` for more detailed production deployment instructions.
