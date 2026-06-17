target "docker-metadata-action" {}

variable "APP" {
  default = "chrome"
}

variable "VERSION" {
  // renovate: datasource=docker depName=linuxserver/chrome versioning=loose
  default = "149.0.7827.155-1-ls100"
}

variable "SOURCE" {
  default = "https://github.com/linuxserver/docker-chrome"
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
    "linux/amd64"
  ]
}
