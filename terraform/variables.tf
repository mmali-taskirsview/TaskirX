# TaskirX v3 - Terraform Variables
# AWS Production Infrastructure Configuration

variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"

  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-\\d{1}$", var.aws_region))
    error_message = "AWS region must be a valid region like us-east-1."
  }
}

variable "project_name" {
  description = "Project name (used for resource naming)"
  type        = string
  default     = "taskir"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]*[a-z0-9]$", var.project_name))
    error_message = "Project name must start with lowercase letter, contain only lowercase, numbers, and hyphens."
  }
}

variable "environment" {
  description = "Environment name (prod, staging, dev)"
  type        = string
  default     = "prod"

  validation {
    condition     = contains(["prod", "staging", "dev"], var.environment)
    error_message = "Environment must be one of: prod, staging, dev."
  }
}

variable "availability_zones" {
  description = "Number of availability zones to use"
  type        = number
  default     = 3

  validation {
    condition     = var.availability_zones >= 2 && var.availability_zones <= 3
    error_message = "Must use 2-3 availability zones for high availability."
  }
}

# ============================================================================
# VPC & NETWORKING
# ============================================================================

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"

  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "VPC CIDR must be valid CIDR notation."
  }
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]

  validation {
    condition     = length(var.public_subnet_cidrs) >= 2
    error_message = "Must have at least 2 public subnets."
  }
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.11.0/24", "10.0.12.0/24", "10.0.13.0/24"]

  validation {
    condition     = length(var.private_subnet_cidrs) >= 2
    error_message = "Must have at least 2 private subnets."
  }
}

variable "single_nat_gateway" {
  description = "Use single NAT gateway (false for HA with one per AZ)"
  type        = bool
  default     = false
}

# ============================================================================
# RDS CONFIGURATION
# ============================================================================

variable "database_name" {
  description = "Name of the initial database"
  type        = string
  default     = "taskir"
  sensitive   = true
}

variable "database_user" {
  description = "Master username for RDS"
  type        = string
  default     = "postgres"

  validation {
    condition     = length(var.database_user) >= 1 && length(var.database_user) <= 63
    error_message = "Database user must be 1-63 characters."
  }
}

variable "database_password" {
  description = "Master password for RDS (min 8 chars, must include uppercase, lowercase, number, special char)"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.database_password) >= 8 && can(regex("[A-Z]", var.database_password)) && can(regex("[a-z]", var.database_password)) && can(regex("[0-9]", var.database_password)) && can(regex("[!@#$%^&*]", var.database_password))
    error_message = "Password must be 8+ chars with uppercase, lowercase, number, and special character."
  }
}

variable "database_port" {
  description = "Port for RDS database"
  type        = number
  default     = 5432

  validation {
    condition     = var.database_port >= 1024 && var.database_port <= 65535
    error_message = "Database port must be between 1024 and 65535."
  }
}

variable "rds_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15.3"
}

variable "rds_instance_class" {
  description = "RDS instance class for production"
  type        = string
  default     = "db.r6i.2xlarge" # 8 vCPU, 64GB RAM

  validation {
    condition     = can(regex("^db\\.[a-z0-9]+\\.[a-z0-9]+$", var.rds_instance_class))
    error_message = "RDS instance class must be valid AWS instance type."
  }
}

variable "rds_allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 500

  validation {
    condition     = var.rds_allocated_storage >= 20 && var.rds_allocated_storage <= 65536
    error_message = "Allocated storage must be 20-65536 GB."
  }
}

variable "rds_max_allocated_storage" {
  description = "Maximum allocated storage for auto-scaling (GB)"
  type        = number
  default     = 1000

  validation {
    condition     = var.rds_max_allocated_storage >= 100 && var.rds_max_allocated_storage <= 65536
    error_message = "Max allocated storage must be 100-65536 GB."
  }
}

# ============================================================================
# REDIS / ELASTICACHE CONFIGURATION
# ============================================================================

variable "redis_engine_version" {
  description = "Redis engine version"
  type        = string
  default     = "7.0"

  validation {
    condition     = can(regex("^[0-9]+\\.[0-9]+$", var.redis_engine_version))
    error_message = "Redis engine version must be in format X.Y"
  }
}

variable "redis_node_type" {
  description = "Redis node type for production"
  type        = string
  default     = "cache.r6g.xlarge" # 4 vCPU, 26GB

  validation {
    condition     = can(regex("^cache\\.[a-z0-9]+\\.[a-z0-9]+$", var.redis_node_type))
    error_message = "Redis node type must be valid AWS cache node type."
  }
}

variable "redis_num_cache_nodes" {
  description = "Number of Redis cache nodes (for cluster mode)"
  type        = number
  default     = 3

  validation {
    condition     = var.redis_num_cache_nodes >= 2 && var.redis_num_cache_nodes <= 500
    error_message = "Number of cache nodes must be 2-500."
  }
}

variable "redis_auth_token" {
  description = "AUTH token for Redis (min 16 chars)"
  type        = string
  sensitive   = true

  validation {
    condition     = length(var.redis_auth_token) >= 16
    error_message = "Redis AUTH token must be at least 16 characters."
  }
}

# ============================================================================
# EKS CONFIGURATION
# ============================================================================

variable "eks_version" {
  description = "Kubernetes version for EKS"
  type        = string
  default     = "1.28"

  validation {
    condition     = can(regex("^1\\.[0-9]{2}$", var.eks_version))
    error_message = "EKS version must be in format 1.XX"
  }
}

variable "eks_desired_size" {
  description = "Desired number of worker nodes"
  type        = number
  default     = 3

  validation {
    condition     = var.eks_desired_size >= 3 && var.eks_desired_size <= 100
    error_message = "Must have 3-100 worker nodes."
  }
}

variable "eks_min_size" {
  description = "Minimum number of worker nodes"
  type        = number
  default     = 3

  validation {
    condition     = var.eks_min_size >= 3
    error_message = "Minimum 3 worker nodes for HA."
  }
}

variable "eks_max_size" {
  description = "Maximum number of worker nodes"
  type        = number
  default     = 10

  validation {
    condition     = var.eks_max_size >= 5
    error_message = "Max nodes must be at least 5."
  }
}

variable "eks_instance_types" {
  description = "Instance types for EKS worker nodes"
  type        = list(string)
  default     = ["t3.2xlarge", "t3a.2xlarge"]

  validation {
    condition     = length(var.eks_instance_types) >= 1
    error_message = "Must specify at least one instance type."
  }
}

variable "eks_disk_size" {
  description = "EBS volume size for worker nodes in GB"
  type        = number
  default     = 100

  validation {
    condition     = var.eks_disk_size >= 50 && var.eks_disk_size <= 16384
    error_message = "Disk size must be 50-16384 GB."
  }
}

# ============================================================================
# ALB CONFIGURATION
# ============================================================================

variable "alb_enable_deletion_protection" {
  description = "Enable deletion protection for ALB"
  type        = bool
  default     = true
}

variable "alb_enable_http2" {
  description = "Enable HTTP/2 for ALB"
  type        = bool
  default     = true
}

variable "alb_enable_access_logs" {
  description = "Enable ALB access logs"
  type        = bool
  default     = true
}

# ============================================================================
# ROUTE53 & DOMAIN
# ============================================================================

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  # Example: "taskir.app"
}

variable "create_route53_zone" {
  description = "Create Route53 hosted zone"
  type        = bool
  default     = true
}

# ============================================================================
# CLOUDFRONT CDN
# ============================================================================

variable "enable_cloudfront" {
  description = "Enable CloudFront CDN"
  type        = bool
  default     = true
}

variable "cloudfront_ttl" {
  description = "CloudFront default TTL in seconds"
  type        = number
  default     = 3600

  validation {
    condition     = var.cloudfront_ttl >= 0 && var.cloudfront_ttl <= 31536000
    error_message = "TTL must be 0-31536000 seconds."
  }
}

# ============================================================================
# WAF & SECURITY
# ============================================================================

variable "enable_waf" {
  description = "Enable AWS WAF for ALB"
  type        = bool
  default     = true
}

variable "enable_shield_advanced" {
  description = "Enable AWS Shield Advanced (DDoS protection)"
  type        = bool
  default     = false # Requires additional contract
}

# ============================================================================
# BACKUP & DISASTER RECOVERY
# ============================================================================

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 30

  validation {
    condition     = var.backup_retention_days >= 1 && var.backup_retention_days <= 365
    error_message = "Backup retention must be 1-365 days."
  }
}

variable "enable_cross_region_backup" {
  description = "Enable cross-region backup replication"
  type        = bool
  default     = true
}

variable "backup_region" {
  description = "Secondary region for backup replication"
  type        = string
  default     = "us-west-2"

  validation {
    condition     = can(regex("^[a-z]{2}-[a-z]+-\\d{1}$", var.backup_region))
    error_message = "Backup region must be a valid AWS region."
  }
}

# ============================================================================
# MONITORING & LOGGING
# ============================================================================

variable "enable_cloudwatch_logs" {
  description = "Enable CloudWatch logs for all services"
  type        = bool
  default     = true
}

variable "logs_retention_days" {
  description = "CloudWatch logs retention in days"
  type        = number
  default     = 30

  validation {
    condition     = contains([1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1827, 3653], var.logs_retention_days)
    error_message = "Must be a valid CloudWatch retention value."
  }
}

variable "enable_xray" {
  description = "Enable AWS X-Ray tracing"
  type        = bool
  default     = true
}

# ============================================================================
# TAGGING & COST MANAGEMENT
# ============================================================================

variable "enable_cost_allocation_tags" {
  description = "Enable cost allocation tags"
  type        = bool
  default     = true
}

variable "cost_center" {
  description = "Cost center for billing"
  type        = string
  default     = "engineering"
}

variable "team" {
  description = "Team responsible for resources"
  type        = string
  default     = "platform"
}

variable "owner" {
  description = "Owner email for resources"
  type        = string
  default     = "devops@taskir.app"

  validation {
    condition     = can(regex("^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$", var.owner))
    error_message = "Owner must be a valid email address."
  }
}
