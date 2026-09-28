target "docker-metadata-action" {}

variable "APP" {
  default = "translator"
}

variable "VERSION" {
  // renovate: datasource=forgejo-tags depName=christfried.balizou/translator
  default = "v0.2.0"
}

variable "SOURCE" {
  default = "https://git.${SECRET_DOMAIN}/christfried.balizou/translator"
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
    translator-source = "${SOURCE_DIR}"
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
  // Match the platform qualified in the upstream operating instructions.
  platforms = ["linux/amd64"]
}
