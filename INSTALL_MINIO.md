# Installing MinIO (Without Docker)

Since you don't want to use Docker, here's how to install and run MinIO directly on your system.

## Quick Installation

### 1. Download MinIO

```bash
cd ~/Downloads

# For Linux
wget https://dl.min.io/server/minio/release/linux-amd64/minio
chmod +x minio
sudo mv minio /usr/local/bin/

# Verify installation
minio --version
```

### 2. Create Data Directory

```bash
mkdir -p ~/minio-data
```

### 3. Start MinIO Server

```bash
# Set credentials (optional, defaults to minioadmin/minioadmin)
export MINIO_ROOT_USER=minioadmin
export MINIO_ROOT_PASSWORD=minioadmin

# Start MinIO
minio server ~/minio-data --console-address ":9001"
```

You should see output like:
```
API: http://192.168.1.100:9000  http://127.0.0.1:9000
Console: http://192.168.1.100:9001 http://127.0.0.1:9001

Documentation: https://docs.min.io
```

### 4. Access MinIO Console

Open browser: http://localhost:9001
- Username: `minioadmin`
- Password: `minioadmin`

### 5. Create the Bucket

You have two options:

#### Option A: Using the Web Console
1. Login to http://localhost:9001
2. Click "Buckets" in the left sidebar
3. Click "Create Bucket"
4. Name it: `documents`
5. Click "Create"

#### Option B: Using MinIO Client (mc)

```bash
# Download mc
wget https://dl.min.io/client/mc/release/linux-amd64/mc
chmod +x mc
sudo mv mc /usr/local/bin/

# Configure
mc alias set local http://localhost:9000 minioadmin minioadmin

# Create bucket
mc mb local/documents

# Verify
mc ls local/
```

## Running MinIO as a Service (Optional)

To run MinIO automatically on system startup:

### Create systemd service

```bash
sudo nano /etc/systemd/system/minio.service
```

Paste this content:

```ini
[Unit]
Description=MinIO
Documentation=https://docs.min.io
Wants=network-online.target
After=network-online.target

[Service]
User=YOUR_USERNAME
Group=YOUR_USERNAME

Environment="MINIO_ROOT_USER=minioadmin"
Environment="MINIO_ROOT_PASSWORD=minioadmin"

ExecStart=/usr/local/bin/minio server /home/YOUR_USERNAME/minio-data --console-address ":9001"

# Let systemd restart this service always
Restart=always

# Specifies the maximum file descriptor number that can be opened by this process
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Replace `YOUR_USERNAME` with your actual username.

Then enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable minio
sudo systemctl start minio
sudo systemctl status minio
```

## Testing the Installation

```bash
# Test API endpoint
curl http://localhost:9000/minio/health/live

# Should return empty 200 OK response

# List buckets
mc ls local/
```

## Restart Your Application

After MinIO is running:

```bash
# In your SDEP directory
air
# or
go run main.go
```

You should now see:
```
✓ Using MinIO storage
```

## Troubleshooting

### Port already in use
If port 9000 is in use:
```bash
# Find what's using it
sudo lsof -i :9000
# or
sudo netstat -tlnp | grep 9000

# Kill the process or use a different port
minio server ~/minio-data --address ":9090" --console-address ":9091"
```

Then update your `.env` file:
```env
S3_ENDPOINT=http://localhost:9090
```

### Can't connect from the app
- Check MinIO is running: `ps aux | grep minio`
- Check ports are listening: `ss -tlnp | grep 9000`
- Check firewall: `sudo ufw status` (if using ufw)
- Try: `curl http://localhost:9000/minio/health/live`

## Using Local Storage Instead

If you prefer not to install MinIO, the application will **automatically fall back to local file storage**. Just restart your app and it will use `./storage/documents/` directory.

```bash
# Clean restart
pkill -f "go run main.go" || pkill -f "air"
air
```

You should see:
```
MinIO not available, falling back to local storage
✓ Using local file storage at ./storage
```
