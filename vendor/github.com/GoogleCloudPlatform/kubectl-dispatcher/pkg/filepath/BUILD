load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["filepath.go"],
    importmap = "k8s.io/kubernetes/vendor/github.com/GoogleCloudPlatform/kubectl-dispatcher/pkg/filepath",
    importpath = "github.com/GoogleCloudPlatform/kubectl-dispatcher/pkg/filepath",
    visibility = ["//visibility:public"],
    deps = [
        "//staging/src/k8s.io/apimachinery/pkg/version:go_default_library",
        "//vendor/github.com/GoogleCloudPlatform/kubectl-dispatcher/pkg/util:go_default_library",
    ],
)

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
    visibility = ["//visibility:public"],
)
