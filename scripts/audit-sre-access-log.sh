#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 || ! -r "$1" ]]; then
  echo "usage: $0 <http-access-log>" >&2
  exit 2
fi

violations=0
while IFS= read -r line; do
  if [[ "$line" =~ \"([^[:space:]]+)[[:space:]]+/api/sre/v1 ]]; then
    method="${BASH_REMATCH[1]}"
    if [[ "$method" != "GET" ]]; then
      echo "$line" >&2
      violations=$((violations + 1))
    fi
  fi
done < "$1"

if (( violations > 0 )); then
  echo "SRE API audit failed: $violations non-GET request(s)" >&2
  exit 1
fi
echo "SRE API audit passed: no non-GET requests"
