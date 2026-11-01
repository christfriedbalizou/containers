target "docker-metadata-action" {}

variable "APP" {
  default = "tv-guides"
}

variable "VERSION" {
  // renovate: datasource=forgejo-tags depName=christfried.balizou/tv-guides
  default = "v0.3.0"
}

variable "SOURCE" {
  default = "https://github.com/christfriedbalizou/containers"
}

variable "SOURCE_DIR" {
  default = "./source"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  contexts = {
    tv-guides-source = "${SOURCE_DIR}"
  }
  args = {
    VERSION = "${VERSION}"
  }
  labels = {
    "org.opencontainers.image.source" = "${SOURCE}"
    "org.opencontainers.image.licenses" = "AGPL-3.0-only"
  }
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
  tags = ["${APP}:${VERSION}"]
}

target "image-all" {
  inherits = ["image"]
  platforms = ["linux/amd64"]
}
