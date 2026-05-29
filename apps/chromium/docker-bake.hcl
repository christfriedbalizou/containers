target "docker-metadata-action" {}

variable "APP" {
  default = "chromium"
}

variable "VERSION" {
  default = "2026.5.29"
}

variable "CHROMIUM_VERSION" {
  // renovate: datasource=docker depName=linuxserver/chromium
  default = "b0ddd401-ls35"
}

variable "SOURCE" {
  default = "https://github.com/linuxserver/docker-chromium"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    CHROMIUM_VERSION = "${CHROMIUM_VERSION}"
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
