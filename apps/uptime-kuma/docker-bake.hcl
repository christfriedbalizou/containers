target "docker-metadata-action" {}

variable "APP" {
  default = "uptime-kuma"
}

variable "VERSION" {
  // renovate: datasource=docker depName=ghcr.io/louislam/uptime-kuma
  default = "2.0.2"
}

variable "SOURCE" {
  default = "https://github.com/louislam/uptime-kuma"
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