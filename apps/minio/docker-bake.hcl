target "docker-metadata-action" {}

variable "APP" {
  default = "minio"
}

variable "VERSION" {
  // This is the date of the pinned Silo main revision below.
  default = "2026.9.13"
}

variable "SILO_REVISION" {
  // renovate: datasource=git-refs depName=pgsty/silo packageName=https://github.com/pgsty/silo currentValue=main
  default = "89637554d60c27cfc51d2281d0a4fe15e415f06d"
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
    SILO_REVISION = "${SILO_REVISION}"
  }
  labels = {
    "org.opencontainers.image.source" = "${SOURCE}"
    "org.opencontainers.image.licenses" = "AGPL-3.0-or-later"
    "io.silo.source" = "https://github.com/pgsty/silo"
    "io.silo.revision" = "${SILO_REVISION}"
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
