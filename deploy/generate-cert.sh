#!/usr/bin/env bash
# Generates a self-signed SSL certificate for nginx HTTPS (443).
# Intended to be run on the deploy host (Linux) with openssl installed.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SSL_DIR="${SCRIPT_DIR}/ssl"

mkdir -p "${SSL_DIR}"

if [[ -f "${SSL_DIR}/zscaler.crt" && -f "${SSL_DIR}/zscaler.key" ]]; then
    echo "[generate-cert] SSL certificate already exists at ${SSL_DIR}, skipping."
    exit 0
fi

echo "[generate-cert] Generating self-signed certificate at ${SSL_DIR}..."

# Resolve native OS paths for the output files. On Git Bash/MSYS2 for Windows,
# cygpath -w yields backslash-free Windows paths (C:\...\ssl\zscaler.key) that
# the Windows openssl binary accepts directly. On native Linux, cygpath is
# absent so we fall back to the POSIX path as-is.
if command -v cygpath >/dev/null 2>&1; then
    KEY_OUT="$(cygpath -w "${SSL_DIR}/zscaler.key")"
    CRT_OUT="$(cygpath -w "${SSL_DIR}/zscaler.crt")"
else
    KEY_OUT="${SSL_DIR}/zscaler.key"
    CRT_OUT="${SSL_DIR}/zscaler.crt"
fi

# On Git Bash/MSYS2 for Windows, args that look like Unix paths get rewritten
# ("/C=US/..." -> "C:/Program Files/Git/C=US/..."), which breaks openssl's
# -subj: the key still gets written but the cert does not. MSYS_NO_PATHCONV=1
# disables that rewriting for the openssl invocation. Because we pass native
# Windows paths to -keyout/-out above, disabling conversion is safe for those
# args too. MSYS_NO_PATHCONV is a no-op on native Linux.
MSYS_NO_PATHCONV=1 openssl req -x509 -nodes -newkey rsa:2048 \
    -keyout "${KEY_OUT}" \
    -out "${CRT_OUT}" \
    -days 365 \
    -subj "/C=US/ST=CA/L=SanJose/O=ZscalerMigration/OU=Platform/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

chmod 600 "${SSL_DIR}/zscaler.key"
chmod 644 "${SSL_DIR}/zscaler.crt"

echo "[generate-cert] Done."
echo "[generate-cert]   cert: ${SSL_DIR}/zscaler.crt"
echo "[generate-cert]   key:  ${SSL_DIR}/zscaler.key"
