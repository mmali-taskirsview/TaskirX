# Minimal Security Groups Module Placeholder

variable "project_name" {}
variable "environment" {}
variable "vpc_id" {}
variable "vpc_cidr" {}
variable "tags" {}

output "rds_security_group_id" {
  value = "sg-rds"
}

output "redis_security_group_id" {
  value = "sg-redis"
}
