# Changelog

## [Unreleased] - 2025-10-20

### ✨ Features Added

#### Authentication & User Management
- User registration with full name, email, and password
- User login with JWT authentication
- Password hashing with bcrypt
- Form-based and JSON API support
- Session management with HTTP-only cookies
- Success messages after registration
- Dark mode support for auth pages

#### Document Management
- File upload with multiple format support
- File download functionality
- Document listing with metadata (size, type, creation date)
- Delete documents with confirmation
- Automatic document list refresh after upload/delete
- User-friendly upload success messages
- Dark mode support for all document pages

#### File Sharing System
- Create shareable links for documents
- Configurable expiration (days + hours format)
- Optional access count limits
- Optional password protection
- Copy-to-clipboard functionality
- Public access (no login required for recipients)
- Password entry form for protected shares
- Beautiful error pages (expired, not found, limit reached)

#### Dark Theme
- Toggle button with moon/sun icon
- Persistent theme preference (localStorage)
- System preference detection
- Smooth color transitions
- Full coverage (all pages, forms, modals)
- Proper contrast ratios

#### Storage Options
- MinIO object storage integration
- Automatic fallback to local file storage
- Configurable storage backend
- Storage directory auto-creation

### 🔧 Technical Improvements

#### Backend
- Go 1.25+ with Fiber web framework
- PostgreSQL database with connection pooling
- SQLC for type-safe queries
- HTMX for dynamic UI updates
- Templ for Go templates
- Middleware for authentication
- Form data parsing support

#### Frontend
- Tailwind CSS with dark mode support
- HTMX for seamless interactions
- Responsive design
- Modern UI components
- Modal dialogs for actions

#### Developer Experience
- Hot reload with Air
- Docker Compose support
- MinIO installation scripts
- Comprehensive documentation
- .env.example template
- Security guidelines

### 📚 Documentation
- README.md with architecture overview
- SETUP.md for quick start
- INSTALL_MINIO.md for storage setup
- SECURITY.md with security best practices
- .env.example for configuration
- API testing guide

### 🔒 Security
- Environment variable configuration
- .gitignore properly configured
- No secrets in code
- Bcrypt password hashing
- JWT token-based auth
- HTTP-only cookies
- Input validation
- Security documentation

### 🐛 Bug Fixes
- Fixed registration form data parsing
- Fixed document list auto-refresh
- Fixed share modal visibility
- Fixed password-protected share access
- Fixed form encoding for share requests
- Fixed dark mode icon toggle

### ⚠️ Known Limitations
- Share passwords stored as plaintext (TODO: hash them)
- Files not encrypted at rest (TODO: implement encryption)
- No rate limiting (TODO: add for production)
- No HTTPS in development (required for production)
- CORS allows all origins (TODO: restrict in production)

### 🚀 Future Enhancements (Planned)
- File encryption at rest
- Share password hashing
- Rate limiting
- Email verification
- Password reset
- Two-factor authentication
- Virus scanning for uploads
- Advanced sharing permissions
- File versioning
- User quotas

---

**Version**: Development
**Status**: Not production-ready (see SECURITY.md)
**Contributors**: Built with AI assistance
