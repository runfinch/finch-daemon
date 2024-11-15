#!/bin/bash

#   Copyright The Finch Daemon Authors.

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

# A script to generate release artifacts.
# This will create a folder in your project root called release.
# This will contain the dynamic + static binaries
# as well as their respective sha256 checksums.
# NOTE: this will mutate your $FINCH_DAEMON_PROJECT_ROOT/out folder.

set -eux -o pipefail

CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
FINCH_DAEMON_PROJECT_ROOT="$(cd -- "$CUR_DIR"/.. && pwd)"
OUT_DIR="${FINCH_DAEMON_PROJECT_ROOT}/bin"
RELEASE_DIR="${FINCH_DAEMON_PROJECT_ROOT}/release"
LICENSE_FILE=${FINCH_DAEMON_PROJECT_ROOT}/THIRD_PARTY_LICENSES
TAG_REGEX="v[0-9]+.[0-9]+.[0-9]+"

ARCH="${TARGET_ARCH:-}"

if [ -z "$ARCH" ]; then
    case $(uname -m) in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *) echo "Error: Unsupported arch $(uname -m)"; exit 1 ;;
    esac
fi

echo "Using ARCH: $ARCH"

if [ "$#" -lt 1 ]; then
    echo "Expected 1 parameter (release_tag), got $#."
    echo "Usage: $0 [architecture] [release_tag]"
    echo "Supported architectures: amd64, arm64"
    exit 1
fi

if ! [[ "$1" =~ $TAG_REGEX ]]; then
    echo "Improper tag format. Format should match regex $TAG_REGEX"
    exit 1
fi

release_version=${1/v/} # Remove v from tag name
shift # Remove the release version argument

if [ -d "$RELEASE_DIR" ]; then
    rm -rf "${RELEASE_DIR:?}"/*
else
    mkdir "$RELEASE_DIR"
fi

dynamic_binary_name=finch-daemon-${release_version}-linux-${ARCH}.tar.gz
static_binary_name=finch-daemon-${release_version}-linux-${ARCH}-static.tar.gz

# Build for the selected architecture
GOARCH=$ARCH  make build
cp "$LICENSE_FILE" "${OUT_DIR}"
pushd "$OUT_DIR"
tar -czvf "$RELEASE_DIR"/"$dynamic_binary_name" -- *
popd
rm -rf "{$OUT_DIR:?}"/*

STATIC=1 GOARCH=$ARCH make build
cp "$LICENSE_FILE" "${OUT_DIR}"
pushd "$OUT_DIR"
tar -czvf "$RELEASE_DIR"/"$static_binary_name" -- *
popd
rm -rf "{$OUT_DIR:?}"/*

# Create checksums
pushd "$RELEASE_DIR"
sha256sum "$dynamic_binary_name" > "$RELEASE_DIR"/"$dynamic_binary_name".sha256sum
sha256sum "$static_binary_name" > "$RELEASE_DIR"/"$static_binary_name".sha256sum
popd
