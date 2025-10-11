# Secure Document Exchange Portal

## Overview
A secure web application for exchanging documents with end-to-end encryption, JWT authentication, and cloud storage integration. Built with Go and modern web technologies.

## System Architecture

### High-Level Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │────│   Fiber API     │────│   PostgreSQL    │
│                 │    │   (Go Backend)  │    │   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              │
                       ┌──────┴──────┐
                       │   Services  │
                       │             │
                       │ • Auth (JWT)│
                       │ • Encryption│
                       │ • Storage   │
                       │ • Caching   │
                       │ • Jobs      │
                       └────────────┘
                              │
                       ┌──────┴──────┐
                       │   External  │
                       │   Services  │
                       │             │
                       │ • MinIO S3  │
                       │ • Redis     │
                       │ • Vault/age │
                       └────────────┘
```

### Components

#### Backend (Go + Fiber)
- **Web Framework**: Fiber - Fast HTTP web framework for Go
- **Database**: PostgreSQL with pgx driver
- **ORM/Code Gen**: sqlc for type-safe SQL queries
- **Authentication**: JWT tokens
- **Encryption**: Hashicorp Vault or age for key management
- **Storage**: AWS S3-compatible (MinIO)
- **Caching**: Redis for metadata caching
- **Background Jobs**: Asynq for asynchronous processing

#### Database Schema
- Users table (id, email, password_hash, created_at, updated_at)
- Documents table (id, user_id, filename, file_path, encrypted_key, metadata, created_at)
- Shares table (id, document_id, share_token, expires_at, access_count, max_access, created_at)
- Sessions table (id, user_id, token, expires_at, created_at)

#### Security Features
- JWT-based authentication
- Document encryption at rest
- Secure sharing links with expiration
- Rate limiting
- Input validation and sanitization

## Technology Stack

### Core Technologies
- **Language**: Go 1.25.1
- **Web Framework**: Fiber v2
- **Database**: PostgreSQL
- **Database Driver**: pgx (connection pooling)
- **SQL Code Generation**: sqlc
- **Migration Tool**: goose
- **Hot Reload**: air

### Storage & External Services
- **Object Storage**: MinIO (S3-compatible)
- **Cache**: Redis
- **Key Management**: Hashicorp Vault or age
- **Job Queue**: Asynq

### Development Tools
- **Build Tool**: Makefile
- **Hot Reload**: air
- **Database Migrations**: goose
- **Code Generation**: sqlc

## Development Setup

### Prerequisites
- Go 1.25.1+
- PostgreSQL
- Redis
- MinIO (or AWS S3)
- Hashicorp Vault (optional, or age)

### Installation
```bash
# Clone repository
git clone <repository-url>
cd secure-document-exchange-portal

# Install dependencies
go mod download

# Setup database
make db-setup

# Run migrations
make migrate-up

# Start development server
make dev
```

### Environment Variables
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=sdep

# JWT
JWT_SECRET=your-secret-key

# MinIO S3
S3_ENDPOINT=http://localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=documents

# Redis
REDIS_URL=redis://localhost:6379

# Vault (optional)
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=your-vault-token
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/refresh` - Refresh JWT token

### Documents
- `POST /api/documents` - Upload document
- `GET /api/documents` - List user documents
- `GET /api/documents/:id` - Get document info
- `DELETE /api/documents/:id` - Delete document

### Sharing
- `POST /api/documents/:id/share` - Create share link
- `GET /api/share/:token` - Access shared document (public)
- `GET /api/share/:token/download` - Download shared document

## Security Considerations
- All documents encrypted before storage
- JWT tokens with expiration
- Share links with configurable expiration and access limits
- Rate limiting on API endpoints
- Input validation and SQL injection prevention
- Secure headers (CSP, HSTS, etc.)

## Performance Optimizations
- Database connection pooling
- Redis caching for metadata
- Background job processing for uploads
- CDN integration for static assets (future)

## Deployment
- Docker containers for all services
- Kubernetes manifests for orchestration
- CI/CD pipeline with automated testing
- Monitoring with Prometheus/Grafana

## Contributing
1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit pull request

## License
MIT License