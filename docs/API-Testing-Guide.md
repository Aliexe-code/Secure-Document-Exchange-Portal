# API Testing Guide

This guide provides comprehensive documentation for testing the Secure Document Exchange Portal API endpoints using Postman or similar API testing tools.

## Base URL

```
http://localhost:8080
```

## Authentication

Most endpoints require authentication via JWT tokens. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Endpoints

### Authentication Endpoints

#### 1. Register User
- **Method**: POST
- **Path**: `/api/auth/register`
- **Content-Type**: application/json

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "full_name": "John Doe"
}
```

**Success Response (201)**:
```json
{
  "id": "uuid-string",
  "email": "user@example.com",
  "full_name": "John Doe",
  "created_at": "2025-01-19T10:00:00Z"
}
```

**Error Response (409)**:
```json
{
  "error": "User already exists"
}
```

#### 2. Login
- **Method**: POST
- **Path**: `/api/auth/login`
- **Content-Type**: application/json

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Success Response (200)**:
```json
{
  "token": "jwt-token-string",
  "user": {
    "id": "uuid-string",
    "email": "user@example.com",
    "full_name": "John Doe",
    "created_at": "2025-01-19T10:00:00Z"
  },
  "expires_at": "2025-01-20T10:00:00Z"
}
```

**Error Response (401)**:
```json
{
  "error": "Invalid credentials"
}
```

#### 3. Refresh Token
- **Method**: POST
- **Path**: `/api/auth/refresh`
- **Headers**:
  - `Authorization: Bearer <current-jwt-token>`

**Success Response (200)**:
```json
{
  "token": "new-jwt-token-string",
  "expires_at": "2025-01-20T10:00:00Z"
}
```

### Document Endpoints

All document endpoints require authentication.

#### 1. Upload Document
- **Method**: POST
- **Path**: `/api/documents`
- **Headers**:
  - `Authorization: Bearer <jwt-token>`
- **Content-Type**: multipart/form-data

**Form Data**:
- `file`: File to upload

**Success Response (201)**:
```json
{
  "id": "document-uuid",
  "filename": "document.pdf",
  "file_size": 1024000,
  "mime_type": "application/pdf",
  "created_at": "2025-01-19T10:00:00Z"
}
```

#### 2. List Documents
- **Method**: GET
- **Path**: `/api/documents`
- **Headers**:
  - `Authorization: Bearer <jwt-token>`

**Success Response (200)**:
```json
[
  {
    "id": "document-uuid",
    "filename": "document.pdf",
    "file_size": 1024000,
    "mime_type": "application/pdf",
    "created_at": "2025-01-19T10:00:00Z"
  }
]
```

#### 3. Download Document
- **Method**: GET
- **Path**: `/api/documents/{document_id}`
- **Headers**:
  - `Authorization: Bearer <jwt-token>`

**Success Response (200)**: File download with appropriate headers

#### 4. Delete Document
- **Method**: DELETE
- **Path**: `/api/documents/{document_id}`
- **Headers**:
  - `Authorization: Bearer <jwt-token>`

**Success Response (204)**: No content

#### 5. Create Share Link
- **Method**: POST
- **Path**: `/api/documents/{document_id}/share`
- **Headers**:
  - `Authorization: Bearer <jwt-token>`
- **Content-Type**: application/json

**Request Body**:
```json
{
  "expires_at": "2025-01-26T10:00:00Z",
  "max_access": 10,
  "password": "optional-password"
}
```

**Success Response (201)**:
```json
{
  "share_token": "share-token-uuid",
  "expires_at": "2025-01-26T10:00:00Z",
  "max_access": 10
}
```

### Public Share Endpoints

#### Access Shared Document
- **Method**: GET
- **Path**: `/api/share/{share_token}`
- **Query Parameters**:
  - `password`: Required if share has password protection

**Success Response (200)**: File download

## Postman Collection

### Environment Variables
Set up these environment variables in Postman:

- `base_url`: `http://localhost:8080`
- `jwt_token`: Set after login

### Sample Collection Structure

```
Secure Document Exchange Portal
├── Authentication
│   ├── Register
│   ├── Login
│   └── Refresh Token
└── Documents
    ├── Upload Document
    ├── List Documents
    ├── Download Document
    ├── Delete Document
    ├── Create Share Link
    └── Access Shared Document (Public)
```

### Pre-request Scripts

For endpoints requiring authentication, add this pre-request script:

```javascript
if (pm.environment.get("jwt_token")) {
    pm.request.headers.add({
        key: "Authorization",
        value: `Bearer ${pm.environment.get("jwt_token")}`
    });
}
```

### Tests

Add these test scripts to relevant requests:

**Login Test**:
```javascript
if (pm.response.code === 200) {
    const response = pm.response.json();
    pm.environment.set("jwt_token", response.token);
}
```

**Registration Test**:
```javascript
pm.test("Status code is 201", function () {
    pm.response.to.have.status(201);
});
```

## Error Codes

- `400`: Bad Request - Invalid input
- `401`: Unauthorized - Invalid credentials or missing token
- `403`: Forbidden - Access denied
- `404`: Not Found - Resource not found
- `409`: Conflict - Resource already exists
- `410`: Gone - Resource expired or access limit exceeded
- `500`: Internal Server Error - Server error

## Rate Limiting

The API implements rate limiting. If you exceed the limits, you'll receive a `429 Too Many Requests` response.

## File Upload Limits

- Maximum file size: 100MB (configurable)
- Supported formats: All file types (MIME type validation can be added)

## Security Notes

- All passwords are hashed using bcrypt
- JWT tokens expire after 24 hours
- File encryption is planned but not yet implemented
- CORS is enabled for web client access