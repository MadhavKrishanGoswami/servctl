# Sample Docker Compose File for the infra
```yaml
name: family-cloud

services:
  # --- IMMICH CORE ---
  immich-server:
    container_name: immich_server
    image: ghcr.io/immich-app/immich-server:${IMMICH_VERSION:-release}
    volumes:
      - ${UPLOAD_LOCATION}:/usr/src/app/upload
      - /etc/localtime:/etc/localtime:ro
    env_file: .env
    ports:
      - '2283:2283'
    depends_on: [redis, database]
    restart: always

  immich-machine-learning:
    container_name: immich_machine_learning
    image: ghcr.io/immich-app/immich-machine-learning:${IMMICH_VERSION:-release}
    volumes:
      - model-cache:/cache
    env_file: .env
    restart: always

  # --- NEXTCLOUD ---
  nextcloud:
    image: nextcloud:${NEXTCLOUD_TAG:-apache}
    container_name: nextcloud
    ports:
      - '8080:80'
    depends_on:
      - database
    environment:
      # Database Connection
      - POSTGRES_HOST=${POSTGRES_HOST}
      - POSTGRES_DB=${POSTGRES_DB}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      
      # Admin Auto-Setup
      - NEXTCLOUD_ADMIN_USER=${NEXTCLOUD_ADMIN_USER}
      - NEXTCLOUD_ADMIN_PASSWORD=${NEXTCLOUD_ADMIN_PASSWORD}
      - NEXTCLOUD_DATA_DIR=/var/www/html/data
      
      # NETWORK FIXES (Crucial for "Black Screen")
      - NEXTCLOUD_TRUSTED_DOMAINS=192.168.1.24
      - OVERWRITEHOST=192.168.1.24:8080
      - OVERWRITEPROTOCOL=http
    volumes:
      - ${NEXT_DATA}:/var/www/html/data
      - nextcloud_config:/var/www/html
    restart: always
  # --- INFRASTRUCTURE (On SSD) ---
  redis:
    container_name: immich_redis
    image: docker.io/valkey/valkey:9
    restart: always

  database:
    container_name: immich_postgres
    image: ghcr.io/immich-app/postgres:14-vectorchord0.4.3-pgvectors0.2.0
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_USER: ${DB_USERNAME}
      POSTGRES_DB: ${DB_DATABASE_NAME}
      POSTGRES_INITDB_ARGS: '--data-checksums'
    volumes:
      - ${DB_DATA_LOCATION}:/var/lib/postgresql/data
    shm_size: 128mb
    restart: always
  # --- MONITORING ---
  glances:
    container_name: glances
    image: nicolargo/glances:latest-full  # "Full" includes tools to read HDD temps
    restart: always
    pid: host              # Allows seeing "Real" Host CPU usage
    network_mode: host     # Uses port 61208 directly
    cap_add:
      - SYS_ADMIN          # Required to read HDD SMART data
      - SYS_RAWIO          # Required to read HDD SMART data
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro  # Monitor container stats
      - /:/rootfs:ro                                  # Monitor Host Disk usage
      - /run/udev:/run/udev:ro                        # Hardware device mapping
    environment:
      - TZ=Asia/Kolkata         # Set your timezone
      - GLANCES_OPT=-w          # Run in Web Server mode
  # ---- UPDATE NOTIFIER (NEW) ---
  diun:
    image: crazymax/diun:latest
    container_name: diun
    command: serve
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
      - "diun_data:/data"
    environment:
      - "TZ=Asia/Kolkata"
      - "LOG_LEVEL=info"
      - "LOG_JSON=false"
      # Check every 6 hours
      - "DIUN_WATCH_SCHEDULE=0 8 1,15 * *"
      # Watch all containers by default? (True = Easy Mode)
      - "DIUN_PROVIDERS_DOCKER_WATCHBYDEFAULT=true"
      # Discord Webhook (PASTE YOUR URL HERE or in .env)
      - "DIUN_NOTIF_DISCORD_WEBHOOKURL=https://discord.com/api/webhooks/"
      # Render fancy cards
      - "DIUN_NOTIF_DISCORD_RENDERFIELDS=true"
    restart: always
volumes:
  model-cache:
  nextcloud_config:
  scrutiny_config:
  scrutiny_influxdb:
  diun_data:
``` 

# Sample .env File
```env
# ==========================================
# 1. SHARED DATABASE (Postgres)
# ==========================================
# This connects Immich & Nextcloud to the SAME server
DB_USERNAME=postgres
DB_PASSWORD=madhav8224
DB_DATABASE_NAME=immich
DB_DATA_LOCATION=/var/lib/docker/volumes/pgdata

# ==========================================
# 2. IMMICH CONFIG
# ==========================================
IMMICH_VERSION=release
UPLOAD_LOCATION=/mnt/data/gallery

# ==========================================
# 3. NEXTCLOUD CONFIG
# ==========================================
NEXTCLOUD_TAG=apache
NEXT_DATA=/mnt/data/cloud/data

# --- AUTO CONFIGURATION ---
# These tell Nextcloud how to talk to the DB we just created in Step 1
POSTGRES_HOST=database
POSTGRES_DB=nextcloud
POSTGRES_USER=postgres
POSTGRES_PASSWORD=madhav8224

# --- ADMIN USER ---
NEXTCLOUD_ADMIN_USER=madhav
NEXTCLOUD_ADMIN_PASSWORD=madhav8224
```

madhav@nas:~/infra/scripts$ cat daily_backup.sh disk_alert.sh smart_alert.sh weekly_cleanup.sh 
#!/bin/bash

# --- CONFIGURATION ---
SOURCE="/mnt/disk1/"
DEST="/mnt/backup/"
LOGFILE="/var/log/daily_backup.log"
# PASTE YOUR WEBHOOK URL HERE
WEBHOOK_URL="https://discord.com/api/"

echo "[$(date)] Starting Backup..." >> $LOGFILE

# --- RUN RSYNC ---
rsync -av --delete $SOURCE $DEST >> $LOGFILE 2>&1
EXIT_CODE=$?

# --- GET DISK STATS ---
DISK1_USAGE=$(df -h /mnt/disk1 | awk 'NR==2 {print $3 "/" $2 " (" $5 ")"}')
BACKUP_USAGE=$(df -h /mnt/backup | awk 'NR==2 {print $3 "/" $2 " (" $5 ")"}')

# --- NOTIFICATION LOGIC ---
if [ $EXIT_CODE -eq 0 ]; then
    COLOR=3066993  # GREEN
    TITLE="✅ NAS Backup: Success"
    DESC="The nightly sync completed successfully."
else
    COLOR=15158332 # RED
    TITLE="🚨 NAS Backup: FAILED"
    DESC="Check the logs immediately. Exit Code: $EXIT_CODE"
fi

# --- CONSTRUCT JSON PAYLOAD ---
# We use a heredoc to create clean JSON for Discord
generate_post_data() {
  cat <<EOF
{
  "username": "NAS Guardian",
  "embeds": [{
    "title": "$TITLE",
    "description": "$DESC",
    "color": $COLOR,
    "fields": [
      {
        "name": "📦 Main Pool",
        "value": "$DISK1_USAGE",
        "inline": true
      },
      {
        "name": "🔒 Vault",
        "value": "$BACKUP_USAGE",
        "inline": true
      }
    ],
    "footer": {
      "text": "Log: $LOGFILE • $(date)"
    }
  }]
}
EOF
}

# --- SEND TO DISCORD ---
curl -H "Content-Type: application/json" \
     -X POST \
     -d "$(generate_post_data)" \
     $WEBHOOK_URL >> $LOGFILE 2>&1

echo "[$(date)] Backup Finished (Exit Code: $EXIT_CODE)." >> $LOGFILE
#!/bin/bash

# --- CONFIGURATION ---
THRESHOLD=90
# Change to the partition you want to watch (usually / or /mnt/disk1)
PARTITION="/mnt/disk1"
WEBHOOK_URL="https://discord.com/api/"

# Get usage percentage (numbers only)
USAGE=$(df -h "$PARTITION" | awk 'NR==2 {print $5}' | sed 's/%//g')

# --- CHECK LOGIC ---
if [ "$USAGE" -gt "$THRESHOLD" ]; then
    
    # JSON Payload for Discord
    # We use RED color (15158332) for danger
    json_payload=$(cat <<EOF
{
  "username": "Server Alerter",
  "embeds": [{
    "title": "🚨 CRITICAL: DISK FULL",
    "description": "Storage is running out! Server functionality may break soon.",
    "color": 15158332,
    "fields": [
      {
        "name": "Partition",
        "value": "$PARTITION",
        "inline": true
      },
      {
        "name": "Current Usage",
        "value": "${USAGE}%",
        "inline": true
      }
    ]
  }]
}
EOF
)

    # Send Alert
    curl -H "Content-Type: application/json" \
         -X POST \
         -d "$json_payload" \
         $WEBHOOK_URL
fi
#!/bin/bash

# --- CONFIGURATION ---
# List ALL physical drives you found in lsblk
DRIVES=("/dev/sda" "/dev/sdb" "/dev/sdc")

# Your Webhook URL
WEBHOOK_URL=""

# --- LOOP THROUGH DRIVES ---
for DRIVE in "${DRIVES[@]}"; do
    
    # 1. Get Health Status
    # We use 'tr -d' to clean up spaces
    HEALTH=$(sudo smartctl -H $DRIVE | grep "overall-health" | awk -F: '{print $2}' | tr -d ' ')

    # 2. Check for Failure
    # If output is empty (command failed) or NOT "PASSED", we alert.
    if [ "$HEALTH" != "PASSED" ]; then
        
        # PREPARE ALERT
        TITLE="🚨 DRIVE FAILURE: $DRIVE"
        DESC="Physical drive $DRIVE is failing S.M.A.R.T. checks. Status: ${HEALTH:-UNKNOWN}"
        COLOR=15158332 # RED
        
        # JSON PAYLOAD
        json_payload=$(cat <<EOF
{
  "username": "Disk Doctor",
  "embeds": [{
    "title": "$TITLE",
    "description": "$DESC",
    "color": $COLOR,
    "fields": [
      { "name": "Drive", "value": "$DRIVE", "inline": true },
      { "name": "Health Status", "value": "${HEALTH:-CRITICAL}", "inline": true }
    ]
  }]
}
EOF
)
        # SEND TO DISCORD
        curl -H "Content-Type: application/json" -X POST -d "$json_payload" $WEBHOOK_URL
    fi
done
#!/bin/bash

# --- CONFIGURATION ---
LOGFILE="/var/log/weekly_cleanup.log"
WEBHOOK_URL="https://discord.com/api/webhooks/"

echo "[$(date)] Starting Cleanup..." > $LOGFILE

# 1. CLEAN APT (System Packages)
# Delete cached .deb files (saves space)
sudo apt-get clean
# Remove dependencies that are no longer needed
sudo apt-get autoremove -y >> $LOGFILE 2>&1

# 2. CLEAN DOCKER (The Safe Way)
# Only remove "dangling" images (images that have no name/tag and aren't used)
# We do NOT use -a because that deletes images you might want to use later.
docker image prune -f >> $LOGFILE 2>&1

# 3. CLEAN LOGS (Optional - prevent huge logs)
# Truncate logs larger than 50MB
find /var/log -type f -name "*.log" -size +50M -exec truncate -s 0 {} \;

# --- REPORT TO DISCORD ---
curl -H "Content-Type: application/json" -X POST -d "{\"username\": \"Janitor\", \"content\": \"🧹 **Weekly Cleanup Complete.** System is shiny and chrome.\"}" $WEBHOOK_URL

echo "[$(date)] Cleanup Finished." >> $LOGFILE