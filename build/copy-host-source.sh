#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Tars up host source dependences and licenses into the _output directory.

set -o errexit
set -o nounset
set -o pipefail

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${KUBE_ROOT}/build/common.sh"

mkdir -p "${KUBE_OUTPUT}/src"
rm -f "${KUBE_OUTPUT}/src/*.tar"
tar cf "${KUBE_OUTPUT}/src/glibc.tar" -C /usr/src glibc

format_license() {
  local -r dep=$1
  local -r src_license=$2
  echo
  echo "================================================================================"
  echo "= ${dep} licensed under: ="
  echo
  cat "${src_license}"
  echo "================================================================================"
  echo
}

LICENSES_FILE="${KUBE_OUTPUT}/src/HOST_LICENSES"
rm -f "${LICENSES_FILE}"
# If you change this list, also be sure to change build/licenses.bzl.
(
  format_license "GNU C Library" /usr/src/glibc/debian/copyright
  format_license "Go standard library" /usr/local/go/LICENSE
  format_license "Go BoringCrypto library" /usr/local/go/src/crypto/internal/boring/LICENSE
) >"${LICENSES_FILE}"
