# Nyaa API — Go rewrite

Lightweight rewrite of [Nyaa-Api-Ts](https://github.com/Yash-Garg/Nyaa-Api-Ts) in Go.  
Targets Raspberry Pi (~5-15 MB RAM idle, single compiled binary, no runtime needed).
Built this for Meoko app to run on a pi, works fine for me so far

## Requirements

- Go 1.21+

## Build & run

```bash
# Download dependencies
go mod download

# Build for your current system
go build -o nyaa-api .

# Build for other systems
# Linux 64 bit
GOOS=linux GOARCH=amd64 go build -o nyaa-api-linux-x86_64 .

# Windows 64 bit
GOOS=windows GOARCH=amd64 go build -o nyaa-api-windows-x86_64.exe .

# Pi 3/3+/Zero 2W/Pi 4/400/CM4/Pi 5/CM5 (64-bit OS)
GOOS=linux GOARCH=arm64 go build -o nyaa-api-pi-arm64 .

# Pi 2/Pi 3/3+/Zero 2W (32-bit OS)
GOOS=linux GOARCH=arm GOARM=7 go build -o nyaa-api-pi-armv7 .

# Pi 1/Zero/Zero W
GOOS=linux GOARCH=arm GOARM=6 go build -o nyaa-api-pi-armv6 .


# Run via CMD
# Linux
./nyaa-api

# Windows
nyaa-api.exe
```

The server starts on port `3000` by default.  
Override with the `PORT` environment variable:

```bash
# Linux
PORT=8080 ./nyaa-api

# Windows
set PORT=8080 && nyaa-api.exe
```

## Endpoints

All endpoints mirror the original TypeScript API exactly.

| Endpoint | Description |
|---|---|
| `GET /` | Health check |
| `GET /id/:id` | Single torrent detail |
| `GET /user/:username` | Uploads by user |
| `GET /:category` | Torrent listing by category |
| `GET /:category/:subcategory` | Torrent listing by subcategory |

### Categories & subcategories

| Category | Subcategories |
|---|---|
| `all` | — |
| `anime` | `amv`, `eng`, `non-eng`, `raw` |
| `audio` | `lossless`, `lossy` |
| `manga` | `eng`, `non-eng`, `raw` |
| `live_action` | `eng`, `promo`, `non-eng`, `raw` |
| `pictures` | `graphics`, `photos` |
| `software` | `applications`, `games` |

### Query parameters

| Param | Description |
|---|---|
| `q` | Search query |
| `s` | Sort: `size`, `seeders`, `leechers`, `date`, `downloads` |
| `o` | Order: `asc`, `desc` |
| `p` | Page number |
| `f` | Filter: `1` = No Remakes, `2` = Trusted Only |

### Examples for manual queries

```
GET /anime?q=one+piece&s=seeders&o=desc
GET /anime/eng?q=attack+on+titan&p=2
GET /id/1234567
GET /user/SubsPlease?q=one+piece
```

## Running as a Linux service

Copy the binary to a stable location and create a systemd service so it starts automatically on boot.

```bash
# Copy binary to /usr/local/bin
sudo cp nyaa-api /usr/local/bin/nyaa-api
sudo chmod +x /usr/local/bin/nyaa-api
```

Create the service file:

```bash
sudo nano /etc/systemd/system/nyaa-api.service
```

Paste the following:

```ini
[Unit]
Description=Nyaa API
After=network.target

[Service]
ExecStart=/usr/local/bin/nyaa-api
#ExecStart=/usr/local/bin/nyaa-api --log-ips # To log IPs incase you feel the service might be being abused
Restart=on-failure
User=nobody
Environment=PORT=3000

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
# Reload systemd so it picks up the new service file
sudo systemctl daemon-reload

# Enable — starts automatically on boot
sudo systemctl enable nyaa-api

# Start now
sudo systemctl start nyaa-api
```

Check the service is running:

```bash
sudo systemctl status nyaa-api
```

Useful commands:

```bash
# Stop the service
sudo systemctl stop nyaa-api

# Restart after replacing the binary
sudo systemctl restart nyaa-api

# View live logs
sudo journalctl -u nyaa-api -f
```

To change the port, edit the `Environment` line in the service file then reload:

```bash
sudo nano /etc/systemd/system/nyaa-api.service
# Change Environment=PORT=3000 to your preferred port

sudo systemctl daemon-reload
sudo systemctl restart nyaa-api
```

To add IP logging, its the same as above except just uncomment the # and comment out the standard ExecStart.
