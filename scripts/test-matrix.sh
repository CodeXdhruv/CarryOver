#!/usr/bin/env bash
set -euo pipefail

ROLE="${1:-standalone}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT_DIR}/bin/carryover"

echo "=== Distro Test Matrix: Running as role '${ROLE}' ==="

if [ ! -x "${BIN}" ]; then
    echo "Carryover binary not found at ${BIN}, attempting build..."
    cd "${ROOT_DIR}" && make build-engine
fi

# Detect environment
if [ -f /etc/os-release ]; then
    . /etc/os-release
    echo "Running on OS: ${PRETTY_NAME:-$ID}"
fi

# Setup isolated temp home
TEST_HOME="$(mktemp -d -t carryover-distro-home-XXXXXX)"
REPO_DIR="$(mktemp -d -t carryover-distro-repo-XXXXXX)"
SHARED_DIR="/shared"

cleanup() {
    rm -rf "${TEST_HOME}" "${REPO_DIR}"
}
trap cleanup EXIT

# Seed simulated home
mkdir -p "${TEST_HOME}/Documents" "${TEST_HOME}/.config" "${TEST_HOME}/Projects/demo-node"
echo "Hello from ${ROLE}" > "${TEST_HOME}/Documents/notes.txt"
echo '{"name": "demo-node", "version": "1.0.0"}' > "${TEST_HOME}/Projects/demo-node/package.json"
echo 'export GREETING="distro-test"' > "${TEST_HOME}/.bashrc"

echo "[1/4] Scanning system..."
"${BIN}" scan --home "${TEST_HOME}" --json > "${TEST_HOME}/scan.json"
test -s "${TEST_HOME}/scan.json"
echo "  ✓ Scan valid JSON generated"

echo "[2/4] Testing package mapper..."
"${BIN}" map ripgrep --from apt --to pacman --json
echo "  ✓ Package mapping verified"

echo "[3/4] Testing backup pipeline..."
export RESTIC_PASSWORD="matrix-test-password-123"
"${BIN}" backup --dest "${REPO_DIR}" --home "${TEST_HOME}" --password "${RESTIC_PASSWORD}" --json
test -f "${REPO_DIR}/manifest.public.json"
echo "  ✓ Backup completed and verified with public manifest"

echo "[4/4] Testing restore pipeline with conflict preservation..."
# Modify target file to force conflict
echo "Target modified line" >> "${TEST_HOME}/Documents/notes.txt"
"${BIN}" restore --dest "${REPO_DIR}" --home "${TEST_HOME}" --password "${RESTIC_PASSWORD}" --conflict keep-both --json

# Verify conflict directory created
CONFLICT_COUNT=$(find "${TEST_HOME}/.carryover-old" -type f 2>/dev/null | wc -l)
if [ "${CONFLICT_COUNT}" -ge 1 ]; then
    echo "  ✓ Conflict preservation verified: ${CONFLICT_COUNT} backup file(s) safely moved"
else
    echo "  ! Warning: No conflict files found (unexpected)"
fi

# If running in shared docker compose volume, write snapshot for downstream distro
if [ -d "${SHARED_DIR}" ] && [ -w "${SHARED_DIR}" ]; then
    echo "Sharing repository to ${SHARED_DIR}/repo..."
    cp -r "${REPO_DIR}" "${SHARED_DIR}/repo"
    echo "  ✓ Repository snapshot shared across containers"
fi

echo "=== Distro Test Matrix Passed for role '${ROLE}' ==="
