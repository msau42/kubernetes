package(default_visibility = ["//visibility:public"])

load("@io_kubernetes_build//defs:build.bzl", "release_filegroup")
load("@io_kubernetes_build//defs:pkg.bzl", "pkg_tar")

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)

pkg_tar(
    name = "_kubernetes-srcs",
    srcs = ["//:all-srcs"],
    package_dir = "kubernetes",
    strip_prefix = "//",
    tags = ["no-cache"],
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "_glibc-srcs",
    srcs = ["@glibc_src//:srcs"],
    package_dir = "glibc",
    tags = ["no-cache"],
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "kubernetes-src",
    extension = "tar.gz",
    tags = ["no-cache"],
    deps = [
        ":_glibc-srcs",
        ":_kubernetes-srcs",
        "@go_src//file",
    ],
)

# FIXME: this should be configurable/auto-detected
PLATFORM_ARCH_STRING = "linux-amd64"

# Included in node and server tarballs.
filegroup(
    name = "license-targets",
    srcs = [
        ":kubernetes-src.tar.gz",
        "//:LICENSES",
    ],
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "_client-bin",
    srcs = ["//build:client-targets"],
    mode = "0755",
    package_dir = "client/bin",
    tags = ["no-cache"],
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "kubernetes-client-%s" % PLATFORM_ARCH_STRING,
    extension = "tar.gz",
    package_dir = "kubernetes",
    tags = ["no-cache"],
    deps = [
        ":_client-bin",
    ],
)

pkg_tar(
    name = "_node-bin",
    srcs = [
        "//build:client-targets",
        "//build:node-targets",
    ],
    mode = "0755",
    package_dir = "node/bin",
    tags = ["no-cache"],
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "kubernetes-node-%s" % PLATFORM_ARCH_STRING,
    srcs = [":license-targets"],
    extension = "tar.gz",
    mode = "0644",
    package_dir = "kubernetes",
    tags = ["no-cache"],
    deps = [
        ":_node-bin",
    ],
)

pkg_tar(
    name = "_server-bin",
    srcs = [
        "//build:client-targets",
        "//build:docker-artifacts",
        "//build:node-targets",
        "//build:server-targets",
    ],
    mode = "0755",
    package_dir = "server/bin",
    tags = ["no-cache"],
    visibility = ["//visibility:private"],
)

genrule(
    name = "dummy",
    outs = [".dummy"],
    cmd = "touch $@",
    visibility = ["//visibility:private"],
)

# Some of the startup scripts fail if there isn't an addons/ directory in the server tarball.
pkg_tar(
    name = "_server-addons",
    srcs = [
        ":.dummy",
    ],
    package_dir = "addons",
    tags = ["no-cache"],
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "kubernetes-server-%s" % PLATFORM_ARCH_STRING,
    srcs = [":license-targets"],
    extension = "tar.gz",
    mode = "0644",
    package_dir = "kubernetes",
    tags = ["no-cache"],
    deps = [
        ":_server-addons",
        ":_server-bin",
    ],
)

pkg_tar(
    name = "_test-bin",
    srcs = ["//build:test-targets"],
    mode = "0755",
    package_dir = "platforms/" + PLATFORM_ARCH_STRING.replace("-", "/"),
    tags = ["no-cache"],
    # TODO: how to make this multiplatform?
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "kubernetes-test",
    srcs = ["//build:test-portable-targets"],
    extension = "tar.gz",
    package_dir = "kubernetes",
    strip_prefix = "//",
    tags = ["no-cache"],
    deps = [
        # TODO: how to make this multiplatform?
        ":_test-bin",
    ],
)

pkg_tar(
    name = "_full_server",
    srcs = [
        ":kubernetes-manifests.tar.gz",
    ],
    package_dir = "server",
    tags = ["no-cache"],
    visibility = ["//visibility:private"],
)

pkg_tar(
    name = "kubernetes",
    srcs = [
        "//:LICENSES",
        "//:README.md",
        "//:version",
        "//cluster:all-srcs",
        "//docs:all-srcs",
        "//hack/lib:all-srcs",
    ],
    extension = "tar.gz",
    package_dir = "kubernetes",
    strip_prefix = "//",
    tags = ["no-cache"],
    deps = [
        ":_full_server",
    ],
)

pkg_tar(
    name = "kubernetes-manifests",
    extension = "tar.gz",
    tags = ["no-cache"],
    deps = [
        "//cluster:manifests",
    ],
)

release_filegroup(
    name = "release-tars",
    srcs = [
        ":kubernetes.tar.gz",
        ":kubernetes-client-%s.tar.gz" % PLATFORM_ARCH_STRING,
        ":kubernetes-node-%s.tar.gz" % PLATFORM_ARCH_STRING,
        ":kubernetes-server-%s.tar.gz" % PLATFORM_ARCH_STRING,
        ":kubernetes-manifests.tar.gz",
        ":kubernetes-src.tar.gz",
        ":kubernetes-test.tar.gz",
    ],
)
