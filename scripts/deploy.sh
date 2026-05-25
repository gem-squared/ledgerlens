#!/usr/bin/env bash
# Atomic deploy of bin/ledgerlens-linux to the LedgerLens VPS.
#
# Caller responsibility:
#   - bin/ledgerlens-linux must already be built (Makefile depends `make deploy` on `make build-linux`).
#   - These env vars may be overridden; defaults match the production VPS:
#       LINUX_BIN      path to the linux/amd64 binary (default: bin/ledgerlens-linux)
#       VPS_HOST       ssh user@host           (default: root@173.199.92.236)
#       VPS_DIR        install dir on VPS      (default: /opt/gem2-ledgerlens)
#       VPS_SERVICE    systemd unit name       (default: gem2-ledgerlens)
#       SERVICE_URL    public health URL       (default: https://ledgerlens.gemsquared.ai)
#       SSH_KEY        path to deploy ssh key  (default: ~/.ssh/id_ed25519_aio_deploy)
#
# Failure modes guarded against:
#   - Missing or wrong-arch binary (pre-flight rejects)
#   - Binary missing embedded _next assets (pre-flight rejects)
#   - SSH unreachable (pre-flight rejects)
#   - Upload truncated (post-upload ELF check on VPS)
#   - Service fails to come up (post-restart systemctl is-active)
#   - Service comes up but health endpoint stays 5xx (rollback to .prev)
#
# A previous deploy run that mv'd current→.prev then failed to install new
# left no current binary at all — see 2026-05-25 outage. This script copies
# (cp) current→.prev BEFORE installing the new file, so a mid-deploy failure
# preserves the live binary.
set -euo pipefail

# ── config ─────────────────────────────────────────────────────────────────
LINUX_BIN="${LINUX_BIN:-bin/ledgerlens-linux}"
VPS_HOST="${VPS_HOST:-root@173.199.92.236}"
VPS_DIR="${VPS_DIR:-/opt/gem2-ledgerlens}"
VPS_SERVICE="${VPS_SERVICE:-gem2-ledgerlens}"
SERVICE_URL="${SERVICE_URL:-https://ledgerlens.gemsquared.ai}"
SSH_KEY="${SSH_KEY:-$HOME/.ssh/id_ed25519_aio_deploy}"

SSH="ssh -i $SSH_KEY -o StrictHostKeyChecking=no -o ConnectTimeout=10"
SCP="scp -i $SSH_KEY -o StrictHostKeyChecking=no -o ConnectTimeout=10"

# ── pre-flight ─────────────────────────────────────────────────────────────
echo "==> pre-flight"

# Binary exists
if [ ! -f "$LINUX_BIN" ]; then
  echo "FATAL: $LINUX_BIN missing — run 'make build-linux' first (from project root)" >&2
  exit 1
fi

# Linux ELF executable
if ! file "$LINUX_BIN" | grep -qE "ELF.*(LSB.*executable|x86-64|amd64)"; then
  echo "FATAL: $LINUX_BIN is not a Linux ELF executable" >&2
  file "$LINUX_BIN" >&2
  exit 1
fi

# Size sanity — current binary is ~27 MB; reject < 20 MB or > 60 MB
SIZE=$(stat -f %z "$LINUX_BIN" 2>/dev/null || stat -c %s "$LINUX_BIN")
if [ "$SIZE" -lt 20000000 ] || [ "$SIZE" -gt 60000000 ]; then
  echo "FATAL: $LINUX_BIN suspicious size: $SIZE bytes (expected 20-60 MB)" >&2
  exit 1
fi
echo "  bin:   $LINUX_BIN ($SIZE bytes)"

# Embedded _next assets — proves web-export ran and embed succeeded.
# Use `grep -aF` directly on the binary; piping `strings | grep -q` plays
# poorly with `set -o pipefail` because grep -q's early-exit SIGPIPEs the
# upstream strings process, which pipefail reports as a pipeline failure
# even on a successful match.
if ! grep -aFq "_next/static/css/" "$LINUX_BIN"; then
  echo "FATAL: $LINUX_BIN missing embedded _next/static/css assets" >&2
  echo "       (did 'make web-export' run? does cmd/ledgerlens/web_static/_next/ exist?)" >&2
  exit 1
fi
if ! grep -aFq "_next/static/chunks/" "$LINUX_BIN"; then
  echo "FATAL: $LINUX_BIN missing embedded _next/static/chunks assets" >&2
  exit 1
fi
echo "  embed: _next/static/{css,chunks} present"

# SSH reachability
if ! $SSH "$VPS_HOST" "true" 2>/dev/null; then
  echo "FATAL: cannot SSH to $VPS_HOST using $SSH_KEY" >&2
  exit 1
fi
echo "  ssh:   $VPS_HOST reachable"

# ── upload ─────────────────────────────────────────────────────────────────
echo "==> upload to $VPS_HOST:/tmp/ledgerlens.new"
$SCP "$LINUX_BIN" "$VPS_HOST:/tmp/ledgerlens.new" >/dev/null

# ── install + restart ──────────────────────────────────────────────────────
echo "==> install + restart on VPS"
$SSH "$VPS_HOST" "
  set -euo pipefail
  cd $VPS_DIR

  # Verify upload landed AND is the right shape before touching live state.
  [ -f /tmp/ledgerlens.new ] || { echo 'ABORT: /tmp/ledgerlens.new missing on VPS'; exit 1; }
  REMOTE_SIZE=\$(stat -c %s /tmp/ledgerlens.new)
  [ \"\$REMOTE_SIZE\" -ge 20000000 ] || { echo \"ABORT: uploaded file too small (\$REMOTE_SIZE bytes)\"; exit 1; }
  chmod +x /tmp/ledgerlens.new
  file /tmp/ledgerlens.new | grep -q 'ELF.*executable' \
    || { echo 'ABORT: uploaded file is not an ELF executable'; file /tmp/ledgerlens.new; exit 1; }

  # Rotate backup ladder (mv — old slot can be lost).
  [ -f $VPS_SERVICE.prev4 ] && mv $VPS_SERVICE.prev4 $VPS_SERVICE.prev5 || true
  [ -f $VPS_SERVICE.prev3 ] && mv $VPS_SERVICE.prev3 $VPS_SERVICE.prev4 || true
  [ -f $VPS_SERVICE.prev2 ] && mv $VPS_SERVICE.prev2 $VPS_SERVICE.prev3 || true
  [ -f $VPS_SERVICE.prev  ] && mv $VPS_SERVICE.prev  $VPS_SERVICE.prev2 || true

  # Backup current as .prev with cp (NOT mv) so a failed install leaves the
  # live binary intact. This is the fix for the 2026-05-25 outage class.
  [ -f $VPS_SERVICE ] && cp $VPS_SERVICE $VPS_SERVICE.prev

  # Install candidate (atomic rename within same filesystem if /tmp is on /).
  mv /tmp/ledgerlens.new $VPS_SERVICE
  [ -x $VPS_SERVICE ] || { echo 'ABORT: installed binary not executable'; exit 1; }

  systemctl restart $VPS_SERVICE
  sleep 2
  if ! systemctl is-active $VPS_SERVICE >/dev/null; then
    echo 'ABORT: service not active after restart'
    systemctl status $VPS_SERVICE --no-pager | tail -20
    exit 1
  fi
  echo '  service: active'
"

# ── health check ───────────────────────────────────────────────────────────
echo "==> health check at $SERVICE_URL/api/v1/health"
HEALTH_OK=0
for i in 1 2 3 4 5; do
  HTTP=$(/usr/bin/curl -sS "$SERVICE_URL/api/v1/health" -m 8 -o /tmp/ll-health.json -w "%{http_code}" || echo "000")
  if [ "$HTTP" = "200" ] && grep -q '"status":"ok"' /tmp/ll-health.json 2>/dev/null; then
    echo "  health: OK (attempt $i, http=$HTTP)"
    HEALTH_OK=1
    break
  fi
  echo "  health: retry $i/5 (http=$HTTP)"
  sleep 2
done

if [ $HEALTH_OK -eq 0 ]; then
  echo "FATAL: health check failed — rolling back to .prev" >&2
  $SSH "$VPS_HOST" "
    set -euo pipefail
    cd $VPS_DIR
    if [ ! -f $VPS_SERVICE.prev ]; then
      echo 'ROLLBACK FAILED: no $VPS_SERVICE.prev to restore from' >&2
      exit 2
    fi
    cp $VPS_SERVICE.prev $VPS_SERVICE
    chmod +x $VPS_SERVICE
    systemctl restart $VPS_SERVICE
    sleep 2
    systemctl is-active $VPS_SERVICE && echo '  rolled back to .prev'
  "
  exit 1
fi

echo "==> deploy OK"
