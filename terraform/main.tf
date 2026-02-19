# TaskirX v3 - AWS Production Infrastructure
# Main Terraform Configuration
# Provisions: VPC, RDS, EKS, ElastiCache, ALB, Route53, CloudFront

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }

  # Backend: Store state in S3 for team collaboration
  backend "s3" {
    bucket         = "taskir-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "taskir-terraform-locks"
  }
}

# AWS Provider
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "TaskirX"
      Environment = var.environment
      ManagedBy   = "Terraform"
      CreatedAt   = timestamp()
    }
  }
}

# Kubernetes Provider (configured after EKS creation)
provider "kubernetes" {
  host                   = aws_eks_cluster.main.endpoint
  cluster_ca_certificate = base64decode(aws_eks_cluster.main.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.main.token
}

# Helm Provider
provider "helm" {
  kubernetes {
    host                   = aws_eks_cluster.main.endpoint
    cluster_ca_certificate = base64decode(aws_eks_cluster.main.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.main.token
  }
}

# Data source for EKS cluster authentication
data "aws_eks_cluster_auth" "main" {
  name = aws_eks_cluster.main.name
}

# Get current AWS account ID
data "aws_caller_identity" "current" {}

# Get availability zones
data "aws_availability_zones" "available" {
  state = "available"
}

# ============================================================================
# LOCALS & VARIABLES
# ============================================================================

locals {
  cluster_name = "${var.project_name}-${var.environment}-cluster"
  
  azs = slice(data.aws_availability_zones.available.names, 0, var.availability_zones)
  
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    Terraform   = "true"
  }
}

# ============================================================================
# VPC & NETWORKING
# ============================================================================

module "vpc" {
  source = "./modules/vpc"

  project_name        = var.project_name
  environment         = var.environment
  vpc_cidr            = var.vpc_cidr
  availability_zones  = local.azs
  private_subnet_cidr = var.private_subnet_cidrs
  public_subnet_cidr  = var.public_subnet_cidrs
  
  enable_nat_gateway = true
  single_nat_gateway = var.single_nat_gateway

  tags = local.common_tags
}

# ============================================================================
# SECURITY GROUPS
# ============================================================================

module "security_groups" {
  source = "./modules/security_groups"

  project_name = var.project_name
  environment  = var.environment
  vpc_id       = module.vpc.vpc_id
  vpc_cidr     = var.vpc_cidr

  tags = local.common_tags
}

# ============================================================================
# RDS POSTGRESQL DATABASE
# ============================================================================

module "rds" {
  source = "./modules/rds"

  project_name           = var.project_name
  environment            = var.environment
  identifier             = "${var.project_name}-${var.environment}-db"
  engine_version         = var.rds_engine_version
  instance_class         = var.rds_instance_class
  allocated_storage      = var.rds_allocated_storage
  max_allocated_storage  = var.rds_max_allocated_storage
  db_name                = var.database_name
  username               = var.database_user
  port                   = var.database_port
  
  # High availability
  multi_az            = true
  backup_retention    = 30
  backup_window       = "03:00-04:00"
  maintenance_window  = "sun:04:00-sun:05:00"
  
  # Security
  storage_encrypted            = true
  kms_key_id                   = aws_kms_key.rds.arn
  performance_insights_enabled = true
  monitoring_interval          = 60
  monitoring_role_arn          = aws_iam_role.rds_monitoring.arn
  enable_cloudwatch_logs_exports = ["postgresql"]
  
  # Networking
  db_subnet_group_name            = module.vpc.db_subnet_group_name
  vpc_security_group_ids          = [module.security_groups.rds_security_group_id]
  publicly_accessible             = false
  skip_final_snapshot             = false
  final_snapshot_identifier_prefix = "taskir-final-snapshot"

  # Replication
  copy_tags_to_snapshot        = true
  # backup_window                = "03:00-04:00"  (Removed duplicate)
  # multi_az                     = true           (Removed duplicate)
  deletion_protection          = true

  tags = local.common_tags

  depends_on = [
    module.vpc,
    module.security_groups
  ]
}

# ============================================================================
# ELASTICACHE - REDIS CACHE
# ============================================================================

module "elasticache" {
  source = "./modules/elasticache"

  project_name              = var.project_name
  environment               = var.environment
  engine_version            = var.redis_engine_version
  node_type                 = var.redis_node_type
  num_cache_nodes           = var.redis_num_cache_nodes
  parameter_group_family    = "redis7"
  port                      = 6379
  
  # High availability
  automatic_failover_enabled = true
  multi_az                   = true
  
  # Security
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = var.redis_auth_token
  security_group_ids         = [module.security_groups.elasticache_security_group_id]
  
  # Networking
  subnet_group_name = module.vpc.elasticache_subnet_group_name
  
  # Maintenance
  maintenance_window = "sun:05:00-sun:07:00"
  notification_topic_arn = ""
  
  # Logging
  log_delivery_configuration = {
    slow_log = {
      cloudwatch_logs_enabled = true
      cloudwatch_logs_log_group = aws_cloudwatch_log_group.redis_slow_log.name
    }
    engine_log = {
      cloudwatch_logs_enabled = true
      cloudwatch_logs_log_group = aws_cloudwatch_log_group.redis_engine_log.name
    }
  }

  tags = local.common_tags

  depends_on = [
    module.vpc,
    module.security_groups
  ]
}

# CloudWatch Log Groups for Redis
resource "aws_cloudwatch_log_group" "redis_slow_log" {
  name              = "/aws/elasticache/${var.project_name}-slow-log"
  retention_in_days = 30

  tags = local.common_tags
}

resource "aws_cloudwatch_log_group" "redis_engine_log" {
  name              = "/aws/elasticache/${var.project_name}-engine-log"
  retention_in_days = 30

  tags = local.common_tags
}

# ============================================================================
# S3 BUCKET - APPLICATION DATA & BACKUPS
# ============================================================================

resource "aws_s3_bucket" "app_storage" {
  bucket = "${var.project_name}-${var.environment}-storage-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.common_tags, {
    Purpose = "Application Storage"
  })
}

resource "aws_s3_bucket_versioning" "app_storage" {
  bucket = aws_s3_bucket.app_storage.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "app_storage" {
  bucket = aws_s3_bucket.app_storage.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.s3.arn
    }
  }
}

resource "aws_s3_bucket_public_access_block" "app_storage" {
  bucket = aws_s3_bucket.app_storage.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Backup bucket
resource "aws_s3_bucket" "backups" {
  bucket = "${var.project_name}-${var.environment}-backups-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.common_tags, {
    Purpose = "Database Backups"
  })
}

resource "aws_s3_bucket_versioning" "backups" {
  bucket = aws_s3_bucket.backups.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.s3.arn
    }
  }
}

resource "aws_s3_bucket_public_access_block" "backups" {
  bucket = aws_s3_bucket.backups.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Lifecycle policy for automatic cleanup
resource "aws_s3_bucket_lifecycle_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    id     = "delete-old-backups"
    status = "Enabled"

    expiration {
      days = 90
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

# ============================================================================
# KMS ENCRYPTION KEYS
# ============================================================================

resource "aws_kms_key" "rds" {
  description             = "KMS key for TaskirX RDS database encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = local.common_tags
}

resource "aws_kms_alias" "rds" {
  name          = "alias/${var.project_name}-${var.environment}-rds"
  target_key_id = aws_kms_key.rds.key_id
}

resource "aws_kms_key" "s3" {
  description             = "KMS key for TaskirX S3 encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = local.common_tags
}

resource "aws_kms_alias" "s3" {
  name          = "alias/${var.project_name}-${var.environment}-s3"
  target_key_id = aws_kms_key.s3.key_id
}

resource "aws_kms_key" "eks" {
  description             = "KMS key for TaskirX EKS encryption"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = local.common_tags
}

resource "aws_kms_alias" "eks" {
  name          = "alias/${var.project_name}-${var.environment}-eks"
  target_key_id = aws_kms_key.eks.key_id
}

# ============================================================================
# IAM ROLES & POLICIES
# ============================================================================

# RDS Monitoring Role
resource "aws_iam_role" "rds_monitoring" {
  name = "${var.project_name}-${var.environment}-rds-monitoring"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  role       = aws_iam_role.rds_monitoring.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# ============================================================================
# OUTPUTS
# ============================================================================

output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "VPC ID"
}

output "private_subnet_ids" {
  value       = module.vpc.private_subnet_ids
  description = "Private subnet IDs"
}

output "public_subnet_ids" {
  value       = module.vpc.public_subnet_ids
  description = "Public subnet IDs"
}

output "rds_endpoint" {
  value       = module.rds.endpoint
  description = "RDS database endpoint"
  sensitive   = true
}

output "rds_address" {
  value       = module.rds.address
  description = "RDS database address"
}

output "rds_port" {
  value       = module.rds.port
  description = "RDS database port"
}

output "redis_endpoint" {
  value       = module.elasticache.primary_endpoint_address
  description = "Redis cache endpoint"
}

output "redis_port" {
  value       = module.elasticache.port
  description = "Redis cache port"
}

output "app_storage_bucket" {
  value       = aws_s3_bucket.app_storage.id
  description = "Application storage S3 bucket"
}

output "backups_bucket" {
  value       = aws_s3_bucket.backups.id
  description = "Backups S3 bucket"
}

output "security_groups" {
  value = {
    rds           = module.security_groups.rds_security_group_id
    elasticache   = module.security_groups.elasticache_security_group_id
    alb           = module.security_groups.alb_security_group_id
    eks_nodes     = module.security_groups.eks_nodes_security_group_id
    eks_cluster   = module.security_groups.eks_cluster_security_group_id
  }
  description = "Security group IDs"
}
