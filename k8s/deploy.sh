#!/usr/bin/env bash
set -euo pipefail

namespace="shortcut"
release="shortcut"
timeout="20m"
wait_flags=(--wait)
extra_values=()
dry_run=()

usage() {
  echo "usage: $0 [-n namespace] [-r release] [-f values.yaml] ... [--no-wait] [--dry-run] [--timeout DURATION]" >&2
  exit 1
}

chart_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -n)
      namespace="$2"
      shift 2
      ;;
    -r)
      release="$2"
      shift 2
      ;;
    -f)
      extra_values+=("$2")
      shift 2
      ;;
    --no-wait)
      wait_flags=()
      shift
      ;;
    --dry-run)
      dry_run=(--dry-run=client)
      shift
      ;;
    --timeout)
      timeout="$2"
      shift 2
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "unknown option: $1" >&2
      usage
      ;;
  esac
done

helm dependency update "${chart_dir}"

kubectl create namespace "${namespace}" --dry-run=client -o yaml | kubectl apply -f -

helm_args=(
  upgrade --install "${release}" "${chart_dir}"
  --namespace "${namespace}"
  --values "${chart_dir}/values.yaml"
)

for vf in "${extra_values[@]}"; do
  helm_args+=(--values "${vf}")
done

helm_args+=("${wait_flags[@]}")
helm_args+=(--timeout "${timeout}")
helm_args+=("${dry_run[@]}")

helm "${helm_args[@]}"
