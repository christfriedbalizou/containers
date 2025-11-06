target "docker-metadata-action" {}

variable "APP" {
  default = "minio"
}

variable "VERSION" {
  // renovate: datasource=github-release depName=minio/minio
  default = "RELEASE.2025-10-15T17-29-55Z"
}

variable "SOURCE" {
  default = "https://github.com/minio/minio"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    VERSION = "${VERSION}"
  }
  labels = {
    "org.opencontainers.image.source" = "${SOURCE}"
  }
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
  tags = ["${APP}:${VERSION}"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
}