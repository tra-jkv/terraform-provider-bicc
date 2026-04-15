variable "bicc_username" {
  description = "BICC username for authentication"
  type        = string
  sensitive   = true
}

variable "bicc_password" {
  description = "BICC password for authentication"
  type        = string
  sensitive   = true
}

variable "bicc_host" {
  description = "BICC host"
  type        = string
  sensitive   = true
}
