load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["file.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//api:galley/v1",  # keep
        "@com_github_ghodss_yaml//:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_golang_protobuf//ptypes/struct:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    size = "small",
    srcs = ["file_test.go"],
    library = ":go_default_library",
    deps = [
        "@com_github_ghodss_yaml//:go_default_library",
        "@com_github_golang_protobuf//jsonpb:go_default_library",
        "@com_github_golang_protobuf//proto:go_default_library",
        "@com_github_golang_protobuf//ptypes/struct:go_default_library",
    ],
)
