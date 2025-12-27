#!/bin/bash
set -e

APP_NAME="HabitTracker"
APP_BUNDLE="${APP_NAME}.app"
BINARY_NAME="habit-tracker"
ICON_SOURCE="packaging/icon.png"

echo "Building binary..."
go build -o "${BINARY_NAME}" main.go

echo "Creating app bundle structure..."
rm -rf "${APP_BUNDLE}"
mkdir -p "${APP_BUNDLE}/Contents/MacOS"
mkdir -p "${APP_BUNDLE}/Contents/Resources"

echo "Copying binary and config..."
cp "${BINARY_NAME}" "${APP_BUNDLE}/Contents/Resources/"
cp config.json "${APP_BUNDLE}/Contents/Resources/"

echo "Creating startup wrapper..."
cat <<EOF > "${APP_BUNDLE}/Contents/Resources/start.sh"
#!/bin/bash
cd "\$(dirname "\$0")"
./${BINARY_NAME}
EOF
chmod +x "${APP_BUNDLE}/Contents/Resources/start.sh"

echo "Creating launcher script..."
cat <<EOF > "${APP_BUNDLE}/Contents/MacOS/launcher"
#!/bin/bash
DIR="\$(cd "\$(dirname "\$0")" && pwd)"
RESOURCE_DIR="\$(dirname "\$DIR")/Resources"
# Open Terminal and run the wrapper script
open -a Terminal "\$RESOURCE_DIR/start.sh"
EOF
chmod +x "${APP_BUNDLE}/Contents/MacOS/launcher"

echo "Creating Info.plist..."
cat <<EOF > "${APP_BUNDLE}/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>launcher</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.user.habittracker</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.10</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

if [ -f "${ICON_SOURCE}" ]; then
    echo "Processing icon..."
    ICONSET_DIR="packaging/AppIcon.iconset"
    mkdir -p "${ICONSET_DIR}"
    
    sips -z 16 16     -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_16x16.png"
    sips -z 32 32     -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_16x16@2x.png"
    sips -z 32 32     -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_32x32.png"
    sips -z 64 64     -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_32x32@2x.png"
    sips -z 128 128   -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_128x128.png"
    sips -z 256 256   -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_128x128@2x.png"
    sips -z 256 256   -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_256x256.png"
    sips -z 512 512   -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_256x256@2x.png"
    sips -z 512 512   -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_512x512.png"
    sips -z 1024 1024 -s format png "${ICON_SOURCE}" --out "${ICONSET_DIR}/icon_512x512@2x.png"
    
    iconutil -c icns "${ICONSET_DIR}" -o "${APP_BUNDLE}/Contents/Resources/AppIcon.icns"
    rm -rf "${ICONSET_DIR}"
else
    echo "Warning: Icon file not found at ${ICON_SOURCE}. Generating bundle without custom icon."
fi

echo "Done! You can drag ${APP_BUNDLE} to your Applications folder."
