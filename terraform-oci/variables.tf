# OCI Configuration
variable "tenancy_ocid" {}
variable "user_ocid" {}
variable "fingerprint" {}
variable "private_key_path" {}
variable "region" {
  default = "us-ashburn-1"
}
variable "compartment_id" {}

# Cloudflare Configuration
variable "cloudflare_api_token" {
  sensitive = true
}
variable "domain_name" {
  default = "taskir.com"
}
variable "zone_id" {
  description = "The Cloudflare Zone ID"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "The Cloudflare Zone ID (Alias for zone_id for consistency)"
  type        = string
  default     = "" # Optional if passed via tfvars
}

# Pinecone Configuration
variable "pinecone_api_key" {
  sensitive = true
}
variable "pinecone_environment" {
  default = "us-east-1"
}

# Cluster Configuration
variable "cluster_name" {
  default = "taskir-oke-cluster"
}
