#!/bin/bash
set -euo pipefail

# https://stackoverflow.com/questions/59895/get-the-source-directory-of-a-bash-script-from-within-the-script-itself
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "${SCRIPT_DIR}/.."

go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
rm -rf coverage.out

# The HTML file can either be opened locally or served through HTTP using a CLI tool like miniserve
# https://github.com/svenstaro/miniserve. The latter is great if you are work
# e.g `miniserve coverage.html`