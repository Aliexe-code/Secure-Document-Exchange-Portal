# Secure Document Exchange Portal - API Testing Guide

This guide provides comprehensive examples for testing all API endpoints using REST API clients like Postman, Insomnia, or curl commands.

## Base URL
```
http://localhost:8080/api
```

## Environment Setup

### Prerequisites
1. PostgreSQL database running
2. Redis server running
3. MinIO server running
4. Environment variables set:
   ```bash
   export DATABASE_URL="postgresql://user:password@localhost:5432/sdep"
   export JWT_SECRET="your-secret-key"
   ```

### Database Setup
```bash
# Run migrations
goose -dir migrations postgres "$DATABASE_URL" up

# Generate SQL code
sqlc generate
```

### Start Server
```bash
go run main.go
```

## Authentication Endpoints

### 1. User Registration
**Endpoint:** `POST /api/auth/register`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "full_name": "John Doe"
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "full_name": "John Doe",
  "created_at": "2025-01-11T10:00:00Z"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "full_name": "John Doe"
  }'
```

### 2. User Login
**Endpoint:** `POST /api/auth/login`

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "full_name": "John Doe",
    "created_at": "2025-01-11T10:00:00Z"
  },
  "expires_at": "2025-01-11T22:00:00Z"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 3. Refresh Token
**Endpoint:** `POST /api/auth/refresh`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-01-11T22:00:00Z"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Document Management Endpoints

All document endpoints require authentication. Include the JWT token in the Authorization header.

### 4. Upload Document
**Endpoint:** `POST /api/documents`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Body (Form Data):**
- Key: `file`
- Type: File
- Value: Select your file

**⚠️ Important Postman Setup:**
1. Go to **Headers** tab:
   - Add: `Authorization` = `Bearer your-jwt-token`
   - **DO NOT** manually set `Content-Type` header

2. Go to **Body** tab:
   - Select **"form-data"** (not "binary" or "raw")
   - Add a key named `file`
   - Change the type from "Text" to "File"
   - Click "Select Files" and choose your file

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "filename": "document.pdf",
  "file_size": 1024000,
  "mime_type": "application/pdf",
  "created_at": "2025-01-11T10:30:00Z"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/documents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F "file=@/path/to/document.pdf"
```

### 5. List Documents
**Endpoint:** `GET /api/documents`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Response (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "filename": "document.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "created_at": "2025-01-11T10:30:00Z"
  },
  {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "filename": "image.jpg",
    "file_size": 2048000,
    "mime_type": "image/jpeg",
    "created_at": "2025-01-11T11:00:00Z"
  }
]
```

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/documents \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 6. Download Document
**Endpoint:** `GET /api/documents/{id}`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Response (200 OK):** Binary file content

**cURL Example:**
```bash
curl -X GET http://localhost:8080/api/documents/550e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -o downloaded_file.pdf
```

### 7. Delete Document
**Endpoint:** `DELETE /api/documents/{id}`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Response (204 No Content):**

**cURL Example:**
```bash
curl -X DELETE http://localhost:8080/api/documents/550e8400-e29b-41d4-a716-446655440001 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Document Sharing Endpoints

### 8. Create Share Link
**Endpoint:** `POST /api/documents/{id}/share`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
Content-Type: application/json
```

**Request Body (optional, but must be valid JSON):**
```json
{
  "expires_at": "2025-01-12T10:00:00Z",
  "max_access": 10,
  "password": "sharepassword"
}
```

**⚠️ Important:** Even if you don't want to set any options, you must send an empty JSON object `{}` or valid JSON. Do not send an empty body.

**Response (201 Created):**
```json
{
  "share_token": "abc123def456",
  "expires_at": "2025-01-12T10:00:00Z",
  "max_access": 10
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/documents/550e8400-e29b-41d4-a716-446655440001/share \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "expires_at": "2025-01-12T10:00:00Z",
    "max_access": 10,
    "password": "sharepassword"
  }'
```

### 9. Access Shared Document
**Endpoint:** `GET /api/share/{token}`

**Query Parameters (if password protected):**
- `password`: Share password

**Response (200 OK):** Binary file content

**cURL Examples:**

**Without password:**
```bash
curl -X GET http://localhost:8080/api/share/abc123def456 \
  -o shared_file.pdf
```

**With password:**
```bash
curl -X GET "http://localhost:8080/api/share/abc123def456?password=sharepassword" \
  -o shared_file.pdf
```

## Health Check

### 10. Health Check
**Endpoint:** `GET /health`

**Response (200 OK):**
```json
{
  "status": "ok"
}
```

**cURL Example:**
```bash
curl -X GET http://localhost:8080/health
```

## Error Responses

### Common Error Format
```json
{
  "error": "Error message description"
}
```

### HTTP Status Codes
- `200 OK` - Success
- `201 Created` - Resource created
- `204 No Content` - Success with no content
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate email)
- `410 Gone` - Resource expired or deleted
- `500 Internal Server Error` - Server error

## Testing Workflow

### 1. Register a User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123", "full_name": "Test User"}'
```

### 2. Login to Get Token
```bash
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}' \
  | jq -r '.token')
```

### 3. Upload a Document
```bash
DOC_ID=$(curl -X POST http://localhost:8080/api/documents \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test.pdf" \
  | jq -r '.id')
```

### 4. List Documents
```bash
curl -X GET http://localhost:8080/api/documents \
  -H "Authorization: Bearer $TOKEN"
```

### 5. Create Share Link
```bash
SHARE_TOKEN=$(curl -X POST http://localhost:8080/api/documents/$DOC_ID/share \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"expires_at": "2025-01-12T10:00:00Z"}' \
  | jq -r '.share_token')
```

### 6. Access Shared Document
```bash
curl -X GET http://localhost:8080/api/share/$SHARE_TOKEN \
  -o downloaded_file.pdf
```

## Postman Collection

You can import this Postman collection to test all endpoints:

```json
{
  "info": {
    "name": "Secure Document Exchange Portal",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080/api"
    },
    {
      "key": "token",
      "value": ""
    }
  ],
  "item": [
    {
      "name": "Auth",
      "item": [
        {
          "name": "Register",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"user@example.com\",\n  \"password\": \"password123\",\n  \"full_name\": \"John Doe\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/auth/register",
              "host": ["{{base_url}}"],
              "path": ["auth", "register"]
            }
          }
        },
        {
          "name": "Login",
          "event": [
            {
              "listen": "test",
              "script": {
                "exec": [
                  "pm.collectionVariables.set(\"token\", pm.response.json().token);"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Content-Type",
                "value": "application/json"
              }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"email\": \"user@example.com\",\n  \"password\": \"password123\"\n}"
            },
            "url": {
              "raw": "{{base_url}}/auth/login",
              "host": ["{{base_url}}"],
              "path": ["auth", "login"]
            }
          }
        }
      ]
    },
    {
      "name": "Documents",
      "item": [
        {
          "name": "Upload Document",
          "request": {
            "method": "POST",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "body": {
              "mode": "formdata",
              "formdata": [
                {
                  "key": "file",
                  "type": "file",
                  "src": [],
                  "description": "Select a file to upload"
                }
              ]
            },
            "url": {
              "raw": "{{base_url}}/documents",
              "host": ["{{base_url}}"],
              "path": ["documents"]
            },
            "description": "Upload a document file. Make sure to select 'form-data' in body and choose a file for the 'file' key."
          }
        },
        {
          "name": "List Documents",
          "request": {
            "method": "GET",
            "header": [
              {
                "key": "Authorization",
                "value": "Bearer {{token}}"
              }
            ],
            "url": {
              "raw": "{{base_url}}/documents",
              "host": ["{{base_url}}"],
              "path": ["documents"]
            }
          }
        }
      ]
    }
  ]
}
```

## Notes

- All authenticated endpoints require a valid JWT token in the Authorization header
- File uploads use multipart/form-data encoding
- Document downloads return binary content with appropriate Content-Type headers
- Share links are public and don't require authentication
- Password-protected shares require the password as a query parameter
- All timestamps are in RFC3339 format
- UUIDs are used for all resource identifiers