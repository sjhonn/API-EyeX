provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_container_cluster" "eyex" {
  name     = var.cluster_name
  location = var.region

  deletion_protection = false
  enable_autopilot    = true

  release_channel {
    channel = "REGULAR"
  }
}
