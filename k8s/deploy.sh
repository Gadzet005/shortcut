#!/usr/bin/env bash
set -eo pipefail

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

rm -rf "${chart_dir}/charts" "${chart_dir}/Chart.lock"

helm dependency build "${chart_dir}"

# unpack .tgz 
if [[ -d "${chart_dir}/charts" ]]; then
  echo "Extracting chart archives..."
  for tgz in "${chart_dir}"/charts/*.tgz; do
    if [[ -f "$tgz" ]]; then
      echo "Extracting: $(basename "$tgz")"
      tar -xzf "$tgz" -C "${chart_dir}/charts/"
      # Удаляем архив после распаковки
      rm -f "$tgz"
    fi
  done
  echo "All charts extracted successfully"
fi

kubectl create namespace "${namespace}" --dry-run=client -o yaml | kubectl apply -f -

helm_args=(
  upgrade --install "${release}" "${chart_dir}"
  --namespace "${namespace}"
  --values "${chart_dir}/values.yaml"
)

# Add extra values files if provided
if [[ ${#extra_values[@]} -gt 0 ]]; then
  for vf in "${extra_values[@]}"; do
    helm_args+=(--values "${vf}")
  done
fi

# Add wait flags if not disabled
if [[ ${#wait_flags[@]} -gt 0 ]]; then
  helm_args+=("${wait_flags[@]}")
fi

helm_args+=(--timeout "${timeout}")

# Add dry-run if specified
if [[ ${#dry_run[@]} -gt 0 ]]; then
  helm_args+=("${dry_run[@]}")
fi

# Execute helm command
helm "${helm_args[@]}"