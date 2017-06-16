load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["context.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//model:go_default_library",
        "@com_github_golang_protobuf//ptypes:go_default_library",
        "@io_istio_api//:proxy/v1/config",  # keep
    ],
)
