#!/usr/bin/env bash
set -euo pipefail

if [[ "${SRE_LIVE_VALIDATION_APPROVED:-}" != "yes-for-this-session" ]]; then
  echo "blocked: set SRE_LIVE_VALIDATION_APPROVED=yes-for-this-session only after explicit human approval" >&2
  exit 2
fi

base_url="${SRE_BASE_URL:-http://127.0.0.1:8080}"
case "$base_url" in http://*|https://*) ;; *) echo "SRE_BASE_URL must be http(s)" >&2; exit 2;; esac

for path in status sources summary zones vms shoots observations; do
  curl --fail --silent --show-error --request GET "${base_url}/api/sre/v1/${path}" >/dev/null
done
curl --fail --silent --show-error --request GET --get --data-urlencode "q=vm" --data-urlencode "limit=1" "${base_url}/api/sre/v1/search" >/dev/null
echo "approved GET-only live validation completed; perform the documented manual sample comparison before sign-off"
