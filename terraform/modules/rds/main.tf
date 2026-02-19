# Minimal RDS Module Placeholder

variable "project_name" {}
variable "environment" {}
variable "identifier" {}
variable "instance_class" {}
variable "engine_version" {}
variable "allocated_storage" {}
variable "max_allocated_storage" {}
variable "db_name" {}
variable "username" {}
variable "port" {}
variable "multi_az" {}
variable "backup_retention" {}
variable "backup_window" {}
variable "maintenance_window" {}
variable "storage_encrypted" {}
variable "kms_key_id" {}
variable "performance_insights_enabled" {}
variable "monitoring_interval" {}
variable "monitoring_role_arn" {}
variable "enable_cloudwatch_logs_exports" {}
variable "db_subnet_group_name" {}
variable "vpc_security_group_ids" {}
variable "publicly_accessible" {}
variable "skip_final_snapshot" {}
variable "final_snapshot_identifier_prefix" {}
variable "copy_tags_to_snapshot" {}
variable "deletion_protection" {}
variable "tags" {}

output "endpoint" {
  value = "mock.rds.amazonaws.com:5432"
}
