resource "docker_image" "nginx-image" {
  name         = "nginx:latest"
  keep_locally = false
}

resource "docker_container" "nginx-container" {
  image = docker_image.nginx-image.latest
  name  = "tutorial"
  ports {
    internal = 80
    external = 8000
  }
}
