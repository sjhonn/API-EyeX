variable "project_id" {
  description = "Google Cloud project ID"
  type        = string
}

variable "region" {
  description = "Google Cloud region"
  type        = string
  default     = "southamerica-west1"
}

variable "cluster_name" {
  description = "GKE cluster name"
  type        = string
  default     = "eyex"
}
