#!/bin/bash
set -euo pipefail

# https://stackoverflow.com/questions/59895/get-the-source-directory-of-a-bash-script-from-within-the-script-itself
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "${SCRIPT_DIR}/.."

IMAGE="ebs-bootstrap"
ARCHS=("amd64" "arm64")

function map_depedencies() {    
    # Check if Mac-GNU alternative binaries are installed
    getopt_cmd="getopt"

    if [[ "$(uname)" == "Darwin" ]] ; then
        getopt_cmd="$(brew --prefix)/opt/gnu-getopt/bin/getopt"
        if [[ ! -x "$(type -P "${getopt_cmd}")" ]] ; then
            echo >&2 "
                ERROR: GNU-enhanced version of getopt not installed
                    Run \"brew install gnu-getopt\""
            exit 2
        fi
    fi
}

function get_docker_platform() {
    arch="${1:-}"
    if [ "${arch}" = 'arm64' ]; then
        echo "linux/arm64"
    elif [ "${arch}" = 'amd64' ]; then
        echo "linux/amd64"
    else
        >&2 echo "🔴 Unsupported architecture: ${arch}"; exit 1
    fi
}

function get_binary_name() {
    arch="${1:-}"
    if [ "${arch}" = 'arm64' ]; then
        echo "ebs-bootstrap-linux-arm64"
    elif [ "${arch}" = 'amd64' ]; then
        echo "ebs-bootstrap-linux-x86_64"
    else
        >&2 echo "🔴 Unsupported architecture: ${arch}"; exit 1
    fi
}

function docker_build() {
    for arch in "${ARCHS[@]}"
    do
        docker_platform="$(get_docker_platform "${arch}")"
        docker build . -t "${IMAGE}:${arch}" --platform "${docker_platform}" --no-cache
        echo "🟢 Built image: ${IMAGE}:${arch}"
    done
}

function copy_binaries() {
    for arch in "${ARCHS[@]}"
    do
        name="$(get_binary_name "${arch}")"
        id=$(docker create "${IMAGE}:${arch}")
        # docker cp produces a tar stream
        docker cp "$id:/app/ebs-bootstrap" - | tar xf - --transform "s/ebs-bootstrap/${name}/"
        docker rm -v "$id"
        echo "🟢 Built and copied binary: ${name}"
    done
}

function main() {
    docker_build
    copy_binaries
}

map_depedencies

ARGUMENT_LIST=(
  "architecture"
)

opts=$("${getopt_cmd}" \
  --longoptions "$(printf "%s:," "${ARGUMENT_LIST[@]}")" \
  --name "$(basename "$0")" \
  --options "" \
  -- "$@"
)

eval set --"$opts"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --architecture)
      ARCHS=("${2}")
      shift 2
      ;;

    *)
      break
      ;;
  esac
done

main
