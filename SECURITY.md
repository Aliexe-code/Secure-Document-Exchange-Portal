# Security Notes

## Current Security Status

### ✅ Implemented Security Features

1. **Authentication & Authorization**
   - JWT-based authentication
   - Password hashing with bcrypt (cost 10)
   - Protected API endpoints with middleware
   - Session management via HTTP-only cookies

2. **Secure File Sharing**
   - Time-based expiration for share links
   - Access count limits
   - Password protection (optional)
   - Unique share tokens (UUID v4)

3. **Data Protection**
   - Environment variables for sensitive data
   - .gitignore properly configured
   - No hardcoded secrets in code

4. **Input Validation**
   - Form validation on frontend and backend
   - File upload size limits
   - Email format validation
   - Password minimum length (6 characters)

### ⚠️ Known Security Limitations (To Be Addressed)

1. **Share Password Storage**
   - Currently stored as plaintext in database
   - **ACTION REQUIRED**: Hash share passwords before production use
   - Location: `internal/handlers/documents.go` line ~372

2. **File Encryption**
   - Files are currently stored unencrypted
   - **ACTION REQUIRED**: Implement end-to-end encryption before production
   - Placeholder exists in: `internal/handlers/documents.go`

3. **HTTPS/TLS**
   - Currently runs on HTTP in development
   - **ACTION REQUIRED**: Enable HTTPS in production
   - Use reverse proxy (nginx/caddy) or configure Fiber with TLS

4. **Rate Limiting**
   - No rate limiting implemented
   - **RECOMMENDATION**: Add rate limiting for:
     - Login attempts
     - Registration
     - File uploads
     - Share link access

5. **CORS Configuration**
   - Currently allows all origins
   - **ACTION REQUIRED**: Restrict CORS to specific domains in production

## Production Deployment Checklist

### 🔒 Critical - Must Do Before Production

- [ ] Hash share link passwords using bcrypt
- [ ] Implement file encryption at rest
- [ ] Enable HTTPS/TLS
- [ ] Generate strong JWT_SECRET (32+ characters)
- [ ] Use production database with strong credentials
- [ ] Change default MinIO credentials
- [ ] Set secure CORS origins
- [ ] Add rate limiting middleware

### 🛡️ Recommended - Should Do

- [ ] Implement request logging and monitoring
- [ ] Add input sanitization for XSS prevention
- [ ] Implement CSRF protection
- [ ] Add SQL injection prevention (already using parameterized queries)
- [ ] Set security headers (CSP, X-Frame-Options, etc.)
- [ ] Regular security audits
- [ ] Dependency vulnerability scanning
- [ ] Add file type validation
- [ ] Implement file size limits per user
- [ ] Add virus/malware scanning for uploads

### 📋 Nice to Have

- [ ] Two-factor authentication (2FA)
- [ ] OAuth2 integration
- [ ] Account lockout after failed login attempts
- [ ] Email verification
- [ ] Password reset functionality
- [ ] Audit logging for all operations
- [ ] IP-based access control
- [ ] Geolocation restrictions

## Reporting Security Issues

If you discover a security vulnerability, please email: [your-email@example.com]

**DO NOT** create a public GitHub issue for security vulnerabilities.

## Environment Variables Security

Never commit these files:
- `.env` (contains real secrets)
- `storage/` (contains uploaded files)
- `minio-data/` (contains MinIO data)
- `*.db` files
- Log files

Always use:
- `.env.example` (template without real values)
- Strong, randomly generated secrets
- Different credentials per environment

## Quick Security Commands

```bash
# Generate a strong JWT secret
openssl rand -base64 32

# Check for exposed secrets
git log --all -- .env

# Remove accidentally committed .env
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch .env' \
  --prune-empty --tag-name-filter cat -- --all
```

## Security Best Practices

1. **Never commit sensitive data**
2. **Use environment variables for all secrets**
3. **Keep dependencies updated**
4. **Use HTTPS in production**
5. **Implement proper logging and monitoring**
6. **Regular security audits**
7. **Follow principle of least privilege**
8. **Validate and sanitize all inputs**

---

**Last Updated**: October 2025
**Status**: Development - Not Production Ready
