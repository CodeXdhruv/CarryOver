#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${ROOT_DIR}/dist"
PKG_DIR="${ROOT_DIR}/packaging"
VERSION="1.0.0"

mkdir -p "${DIST_DIR}"

echo "==> Building binary..."
cd "${ROOT_DIR}"
make build-engine

BIN="${ROOT_DIR}/bin/carryover"
if [ ! -f "${BIN}" ]; then
    echo "Binary ${BIN} not found!"
    exit 1
fi

echo "==> Building Debian (.deb) package..."
DEB_STAGING="${PKG_DIR}/deb"
mkdir -p "${DEB_STAGING}/usr/bin"
mkdir -p "${DEB_STAGING}/usr/share/applications"
mkdir -p "${DEB_STAGING}/usr/share/icons/hicolor/scalable/apps"

cp "${BIN}" "${DEB_STAGING}/usr/bin/carryover"
chmod 755 "${DEB_STAGING}/usr/bin/carryover"
cp "${PKG_DIR}/carryover.desktop" "${DEB_STAGING}/usr/share/applications/"
cp "${PKG_DIR}/carryover.svg" "${DEB_STAGING}/usr/share/icons/hicolor/scalable/apps/"

dpkg-deb --build --root-owner-group "${DEB_STAGING}" "${DIST_DIR}/carryover_${VERSION}_amd64.deb"
echo "Created: ${DIST_DIR}/carryover_${VERSION}_amd64.deb"

echo "==> Packaging AppImage root structure..."
APPIMAGE_DIR="${DIST_DIR}/AppDir"
rm -rf "${APPIMAGE_DIR}"
mkdir -p "${APPIMAGE_DIR}/usr/bin"
mkdir -p "${APPIMAGE_DIR}/usr/share/applications"
mkdir -p "${APPIMAGE_DIR}/usr/share/icons/hicolor/scalable/apps"

cp "${BIN}" "${APPIMAGE_DIR}/usr/bin/carryover"
chmod 755 "${APPIMAGE_DIR}/usr/bin/carryover"
cp "${PKG_DIR}/carryover.desktop" "${APPIMAGE_DIR}/"
cp "${PKG_DIR}/carryover.desktop" "${APPIMAGE_DIR}/usr/share/applications/"
cp "${PKG_DIR}/carryover.svg" "${APPIMAGE_DIR}/"
cp "${PKG_DIR}/carryover.svg" "${APPIMAGE_DIR}/usr/share/icons/hicolor/scalable/apps/"
cp "${PKG_DIR}/appimage/AppRun" "${APPIMAGE_DIR}/AppRun"
chmod 755 "${APPIMAGE_DIR}/AppRun"

tar -czf "${DIST_DIR}/carryover_${VERSION}_AppDir.tar.gz" -C "${DIST_DIR}" AppDir
echo "Created: ${DIST_DIR}/carryover_${VERSION}_AppDir.tar.gz"

echo "==> Packaging artifacts ready in ${DIST_DIR}"
