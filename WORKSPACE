workspace(
  name = "tbc",
  # Map the @npm bazel workspace to the node_modules directory.
  # This lets Bazel use the same node_modules as other local tooling.
  managed_directories = {"@npm": ["node_modules"]},
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
  name = "build_bazel_rules_nodejs",
  sha256 = "8a7c981217239085f78acc9898a1f7ba99af887c1996ceb3b4504655383a2c3c",
  urls = ["https://github.com/bazelbuild/rules_nodejs/releases/download/4.0.0/rules_nodejs-4.0.0.tar.gz"],
)

load("@build_bazel_rules_nodejs//:index.bzl", "npm_install")
npm_install(
  # Name this npm so that Bazel Label references look like @npm//package
  name = "npm",
  package_json = "//:package.json",
  package_lock_json = "//:package-lock.json",
)
