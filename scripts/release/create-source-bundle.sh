#!/usr/bin/env bash

set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

output_dir="$(default_dist_dir)"
require_daemon_tag=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output-dir)
      output_dir="$2"
      shift 2
      ;;
    --allow-untagged-daemon)
      require_daemon_tag=0
      shift
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

require_command git
require_command go
require_command tar

if (( require_daemon_tag )); then
  "$(dirname "${BASH_SOURCE[0]}")/validate-manifest.sh" --require-daemon-tag
else
  "$(dirname "${BASH_SOURCE[0]}")/validate-manifest.sh"
fi

mkdir -p "${output_dir}"

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

bundle_dir="${tmpdir}/$(source_bundle_dirname)"
daemon_dir="${tmpdir}/daemon"
mkdir -p "${bundle_dir}" "${bundle_dir}/third_party"

git -C "${REPO_ROOT}" archive --format=tar HEAD | tar -xf - -C "${bundle_dir}"
(
  cd "${bundle_dir}"
  GOWORK=off go mod vendor
)

git clone --filter=blob:none --no-checkout "$(daemon_repo)" "${daemon_dir}" >/dev/null 2>&1
git -C "${daemon_dir}" checkout --detach "$(daemon_commit)" >/dev/null 2>&1

if [[ -n "$(daemon_tag)" ]]; then
  if ! git -C "${daemon_dir}" tag --points-at "$(daemon_commit)" | grep -Fxq "$(daemon_tag)"; then
    echo "daemon_tag $(daemon_tag) does not point at $(daemon_commit)" >&2
    exit 1
  fi
fi

if [[ -f "${daemon_dir}/go.mod" ]]; then
  (
    cd "${daemon_dir}"
    GOWORK=off go mod vendor
  )
fi

mkdir -p "${bundle_dir}/third_party/go-librespot"
tar -C "${daemon_dir}" --exclude=.git -cf - . | tar -xf - -C "${bundle_dir}/third_party/go-librespot"

cat > "${bundle_dir}/packaging/source-bundle-metadata.env" <<EOF
LAZYSPOTIFY_VERSION=$(release_version)
LAZYSPOTIFY_COMMIT=$(current_commit)
DAEMON_REPO=$(daemon_repo)
DAEMON_TAG=$(daemon_tag)
DAEMON_COMMIT=$(daemon_commit)
BUNDLE_VERSION=$(bundle_version)
EOF

tar -C "${tmpdir}" -czf "${output_dir}/$(source_bundle_filename)" "$(source_bundle_dirname)"
printf '%s\n' "${output_dir}/$(source_bundle_filename)"
