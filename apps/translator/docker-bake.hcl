target "docker-metadata-action" {}

variable "APP" {
  default = "translator"
}

variable "VERSION" {
  // renovate: datasource=forgejo-tags depName=christfried.balizou/translator
  default = "v0.11.1"
}

variable "SOURCE" {
  // Public packaging source. Private checkout uses SECRET_DOMAIN in the runner.
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
    translator-source = "${SOURCE_DIR}"
  }
  secret = ["id=translator-public-key,env=TRANSLATOR_GPG_PUBLIC_KEY"]
  // Secret contents do not invalidate BuildKit cache; always honor key rotation.
  no-cache-filter = ["encrypted"]
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
