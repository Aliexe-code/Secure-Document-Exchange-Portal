# Database Schema Design

## Overview
The database schema for the Secure Document Exchange Portal consists of four main tables: users, documents, shares, and sessions. The schema is designed to support secure document storage, sharing, and user management.

## Tables

### users
Stores user account information and authentication data.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique user identifier |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email address |
| password_hash | VARCHAR(255) | NOT NULL | Bcrypt hashed password |
| full_name | VARCHAR(255) | NOT NULL | User's full name |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Account creation time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |
| is_active | BOOLEAN | DEFAULT TRUE | Account status |

### documents
Stores metadata about uploaded documents.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique document identifier |
| user_id | UUID | NOT NULL, FOREIGN KEY(users.id) | Owner of the document |
| filename | VARCHAR(255) | NOT NULL | Original filename |
| file_path | VARCHAR(500) | NOT NULL | S3/MinIO storage path |
| encrypted_key | TEXT | NOT NULL | Encrypted encryption key |
| file_size | BIGINT | NOT NULL | File size in bytes |
| mime_type | VARCHAR(100) | NOT NULL | MIME type |
| checksum | VARCHAR(128) | NOT NULL | SHA-256 checksum |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Upload time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update time |

### shares
Manages document sharing links and access control.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique share identifier |
| document_id | UUID | NOT NULL, FOREIGN KEY(documents.id) | Shared document |
| share_token | VARCHAR(255) | UNIQUE, NOT NULL | Secure share token |
| expires_at | TIMESTAMP | NOT NULL | Link expiration time |
| max_access | INTEGER | DEFAULT -1 | Maximum access count (-1 = unlimited) |
| access_count | INTEGER | DEFAULT 0 | Current access count |
| password_hash | VARCHAR(255) | NULL | Optional password protection |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Share creation time |
| created_by | UUID | NOT NULL, FOREIGN KEY(users.id) | User who created share |

### sessions
Tracks active user sessions for JWT management.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | Unique session identifier |
| user_id | UUID | NOT NULL, FOREIGN KEY(users.id) | Associated user |
| token | VARCHAR(500) | UNIQUE, NOT NULL | JWT token |
| expires_at | TIMESTAMP | NOT NULL | Token expiration time |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Session creation time |
| ip_address | INET | NULL | Client IP address |
| user_agent | TEXT | NULL | Client user agent |

## Indexes
- users.email (UNIQUE)
- users.created_at
- documents.user_id
- documents.created_at
- shares.document_id
- shares.share_token (UNIQUE)
- shares.expires_at
- shares.created_by
- sessions.user_id
- sessions.token (UNIQUE)
- sessions.expires_at

## Relationships
- users.id → documents.user_id (1:N)
- users.id → shares.created_by (1:N)
- users.id → sessions.user_id (1:N)
- documents.id → shares.document_id (1:N)

## Constraints
- Documents can only be accessed by their owner or through valid shares
- Share links expire automatically and have access limits
- Sessions are invalidated on logout or expiration
- All foreign key relationships enforce referential integrity

## Extensions Required
- `uuid-ossp` for UUID generation
- `pgcrypto` for encryption functions (optional)

## Migration Strategy
Use goose for database migrations with up/down scripts for each schema change.