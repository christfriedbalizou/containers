target "docker-metadata-action" {}

variable "APP" {
  default = "vscode"
}

variable "VERSION" {
  // renovate: datasource=github-release depName=coder/code-server
  default = "4.118.0"
}

variable "SOURCE" {
  default = "https://github.com/coder/code-server"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    CODE_SERVER_VERSION = "${VERSION}"
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
