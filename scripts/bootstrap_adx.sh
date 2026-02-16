#!/bin/bash
# BOOTSTRAP SCRIPT FOR ADX ON OCI COMPUTE
# Run this as user-data or manually.

set -e

APP_DIR="/opt/adx-app"
REPO_URL="https://github.com/taskirkhan20-hue/TaskirX.git" # Example URL
NODE_VERSION="18.x"

# 1. Update System & Install Dependencies
sudo yum update -y
sudo yum install -y git curl

# Install Node.js
curl -sL https://rpm.nodesource.com/setup_${NODE_VERSION} | sudo bash -
sudo yum install -y nodejs

# Install PM2 for process management
sudo npm install pm2 -g

# 2. Clone Repository
if [ ! -d "$APP_DIR" ]; then
    sudo git clone $REPO_URL $APP_DIR
else
    cd $APP_DIR && sudo git pull
fi

cd $APP_DIR/nestjs-backend
sudo npm install
sudo npm run build

# 3. Configure Environment Variables
# Ideally, fetch these from OCI Vault or parameter store!
cat <<EOF | sudo tee .env
PINECONE_API_KEY="YOUR_PINECONE_KEY_HERE"
PINECONE_ENV="us-west1-gcp-free"
DB_HOST="postgres.private.subnet.vcn..."
DB_USER="adx_user"
DB_PASS="secure_password"
PORT=3000
EOF

# 4. Start Application with PM2
pm2 start dist/main.js --name "adx-backend" --instances max

# 5. Persist PM2 across reboot
pm2 startup systemd
pm2 save

echo "ADX Backend Deployed Successfully!"
