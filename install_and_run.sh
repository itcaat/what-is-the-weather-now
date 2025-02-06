#!/bin/bash

REPO="itcaat/what-is-the-weather-now"
INSTALL_DIR="$HOME/what-is-the-weather-now"
BIN_NAME="what-is-the-weather-now-linux-arm64"
ZIP_FILE="$BIN_NAME.zip"
PID_FILE="$INSTALL_DIR/app.pid"
VERSION_FILE="$INSTALL_DIR/version"
FORCE_UPDATE=false

# Check if --force is used
if [[ "$1" == "--force" ]]; then
    FORCE_UPDATE=true
    echo "Force update enabled. Killing all instances and reinstalling..."
fi

# Function to stop the application
stop_application() {
    if [[ -f "$PID_FILE" ]]; then
        APP_PID=$(cat "$PID_FILE")
        if ps -p "$APP_PID" > /dev/null 2>&1; then
            echo "Stopping running application (PID: $APP_PID)..."
            kill "$APP_PID"
            sleep 2
        fi
        rm -f "$PID_FILE"
    fi
}

# Force stop all instances if --force is enabled
if [[ "$FORCE_UPDATE" == true ]]; then
    pkill -f "$BIN_NAME" 2>/dev/null
    echo "Killed all running instances of $BIN_NAME."
    rm -f "$PID_FILE" "$VERSION_FILE"
fi

# 1. Get the latest release tag from GitHub API
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | jq -r '.tag_name')
if [[ -z "$LATEST_RELEASE" || "$LATEST_RELEASE" == "null" ]]; then
    echo "Error: Failed to fetch the latest release."
    exit 1
fi

echo "Latest release: $LATEST_RELEASE"

# 2. Check if the installed version is already the latest
if [[ -f "$VERSION_FILE" && "$FORCE_UPDATE" == false ]]; then
    INSTALLED_VERSION=$(cat "$VERSION_FILE")
    if [[ "$INSTALLED_VERSION" == "$LATEST_RELEASE" ]]; then
        echo "You already have the latest version ($INSTALLED_VERSION). Checking if it's running..."
        if [[ -f "$PID_FILE" ]]; then
            APP_PID=$(cat "$PID_FILE")
            if ps -p "$APP_PID" > /dev/null 2>&1; then
                echo "Application is already running with PID $APP_PID. Exiting."
                exit 0
            else
                echo "Application is not running. Restarting..."
            fi
        fi
    fi
fi

# 3. Stop the running application if needed
stop_application

# 4. Construct the download URL
ASSET_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$ZIP_FILE"

echo "Downloading: $ASSET_URL"

# 5. Create the installation directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# 6. Download and extract the new version
curl -L "$ASSET_URL" -o "$INSTALL_DIR/$ZIP_FILE"
if [[ $? -ne 0 ]]; then
    echo "Error: Failed to download the file."
    exit 1
fi

echo "Extracting to $INSTALL_DIR"
unzip -o "$INSTALL_DIR/$ZIP_FILE" -d "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/$BIN_NAME"

# 7. Remove the ZIP file after extraction
rm "$INSTALL_DIR/$ZIP_FILE"

# 8. Store the new version number
echo "$LATEST_RELEASE" > "$VERSION_FILE"

# 9. Start the updated application in the background
echo "Starting the updated application..."
nohup "$INSTALL_DIR/$BIN_NAME" > "$INSTALL_DIR/output.log" 2>&1 &
echo $! > "$PID_FILE"

echo "Application is running in the background (PID: $(cat $PID_FILE)). Logs: $INSTALL_DIR/output.log"
