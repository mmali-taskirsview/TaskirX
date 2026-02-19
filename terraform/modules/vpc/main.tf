# Minimal VPC Module Placeholder

variable "project_name" {}
variable "environment" {}
variable "vpc_cidr" {}
variable "availability_zones" {}
variable "private_subnet_cidr" {}
variable "public_subnet_cidr" {}
variable "enable_nat_gateway" {}
variable "single_nat_gateway" {}
variable "tags" {}

output "vpc_id" {
  value = "vpc-12345678"
}

output "db_subnet_group_name" {
    value = "default-db-subnet-group"
}
