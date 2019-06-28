# Copyright 2018 The Kubernetes Authors.
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

# This is a rough transliteration of build/copy-host-source.sh, using the
# Bazel workspace dependencies instead of the _output directory.

_separator = "================================================================================"

def _header_cmd(pkg_name):
    return "echo -e '%s\n= %s licensed under: =\n' >> $@" % (_separator, pkg_name)

def _footer_cmd():
    return "echo -e '%s\n' >> $@" % _separator

# Creates a single file named LICENSES containing the licenses of all vendored
# Go dependencies, glibc, and the Go std library and BoringCrypto / BoringSSL.
def gen_licenses(**kwargs):
    srcs = [
        "//:Godeps/LICENSES",
        "@glibc_src//:debian-copyright",
        "@go_src//file",
    ]

    cmds = ["cat $(location //:Godeps/LICENSES) > $@"]

    # If you change this list, also be sure to change build/copy-host-source.sh.
    cmds.extend([
        _header_cmd("GNU C Library"),
        "cat $(location @glibc_src//:debian-copyright) >> $@",
        _footer_cmd(),
        _header_cmd("Go standard library"),
        "tar -Oxf $(location @go_src//file) go/LICENSE >> $@",
        _footer_cmd(),
        _header_cmd("Go BoringCrypto library"),
        "tar -Oxf $(location @go_src//file) go/src/crypto/internal/boring/LICENSE >> $@",
        _footer_cmd(),
    ])

    native.genrule(
        name = "gen_licenses",
        srcs = srcs,
        outs = ["LICENSES"],
        cmd = ";".join(cmds),
        **kwargs
    )
