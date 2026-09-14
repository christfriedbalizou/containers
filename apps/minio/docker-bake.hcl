target "docker-metadata-action" {}

variable "APP" {
  default = "minio"
}

variable "VERSION" {
  // Pin the image and source to the same published Silo release.
  // renovate: datasource=github-releases depName=pgsty/silo versioning=loose
  default = "RELEASE.2026-09-03T13-18-01Z"
}

variable "SOURCE" {
  default = "https://github.com/pgsty/silo"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    VERSION = "${VERSION}"
    SILO_TAG = "${VERSION}"
  }
  labels = {
    "org.opencontainers.image.source" = "${SOURCE}"
    "org.opencontainers.image.licenses" = "AGPL-3.0-or-later"
    "io.silo.source" = "https://github.com/pgsty/silo"
    "io.silo.tag" = "${VERSION}"
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
