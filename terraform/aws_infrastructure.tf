# Terraform configuration for TaskirX infrastructure on AWS
# Main configuration file

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }

  backend "s3" {
    bucket         = "taskir-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "taskir-terraform-locks"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "TaskirX"
      ManagedBy   = "Terraform"
    }
  }
}

provider "kubernetes" {
  host                   = aws_eks_cluster.taskir.endpoint
  cluster_ca_certificate = base64decode(aws_eks_cluster.taskir.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.taskir.token
}

# Data source for EKS cluster auth
data "aws_eks_cluster_auth" "taskir" {
  name = aws_eks_cluster.taskir.name
}

# ============================================================================
# VPC and Networking
# ============================================================================

module "vpc" {
  source = "./modules/vpc"

  environment = var.environment
  cidr_block  = "10.0.0.0/16"
  
  private_subnet_cidrs = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnet_cidrs  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = true

  tags = {
    Name = "taskir-vpc"
  }
}

# ============================================================================
# EKS Cluster
# ============================================================================

resource "aws_eks_cluster" "taskir" {
  name            = "taskir-cluster-${var.environment}"
  version         = "1.27"
  role_arn        = aws_iam_role.eks_cluster_role.arn
  
  vpc_config {
    subnet_ids              = concat(module.vpc.private_subnet_ids, module.vpc.public_subnet_ids)
    endpoint_private_access = true
    endpoint_public_access  = true
    security_group_ids      = [aws_security_group.eks_cluster_sg.id]
  }

  enabled_cluster_log_types = [
    "api",
    "audit",
    "authenticator",
    "controllerManager",
    "scheduler"
  ]

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
    aws_iam_role_policy_attachment.eks_vpc_cni_policy
  ]

  tags = {
    Name = "taskir-cluster-${var.environment}"
  }
}

# ============================================================================
# EKS Node Groups
# ============================================================================

resource "aws_eks_node_group" "taskir_nodes" {
  cluster_name    = aws_eks_cluster.taskir.name
  node_group_name = "taskir-node-group-${var.environment}"
  node_role_arn   = aws_iam_role.eks_node_role.arn
  subnet_ids      = module.vpc.private_subnet_ids

  scaling_config {
    desired_size = var.node_group_desired_size
    max_size     = var.node_group_max_size
    min_size     = var.node_group_min_size
  }

  instance_types = ["t3.medium"]
  capacity_type  = "ON_DEMAND"

  disk_size = 50

  tags = {
    Name = "taskir-node-group-${var.environment}"
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_node_policy,
    aws_iam_role_policy_attachment.eks_cni_policy,
    aws_iam_role_policy_attachment.eks_registry_policy
  ]
}

# ============================================================================
# RDS Database
# ============================================================================

module "rds" {
  source = "./modules/rds"

  environment = var.environment
  
  db_name     = "taskir"
  db_user     = "taskir"
  db_password = random_password.db_password.result
  
  allocated_storage     = var.db_allocated_storage
  storage_type          = "gp3"
  engine_version        = "14.7"
  instance_class        = var.db_instance_class
  multi_az              = var.db_multi_az
  
  vpc_id             = module.vpc.vpc_id
  db_subnet_group_id = module.vpc.db_subnet_group_id
  
  backup_retention_period = 30
  backup_window           = "03:00-04:00"
  maintenance_window      = "sun:04:00-sun:05:00"
  
  enable_encryption = true
  kms_key_id        = aws_kms_key.rds.arn
  
  tags = {
    Name = "taskir-postgres-${var.environment}"
  }
}

# ============================================================================
# ElastiCache Redis
# ============================================================================

module "redis" {
  source = "./modules/redis"

  environment = var.environment
  
  engine_version = "7.0"
  node_type      = var.redis_node_type
  num_cache_nodes = var.redis_num_nodes
  
  vpc_id             = module.vpc.vpc_id
  subnet_ids         = module.vpc.private_subnet_ids
  security_group_ids = [aws_security_group.redis_sg.id]
  
  automatic_failover_enabled = true
  multi_az_enabled           = true
  
  at_rest_encryption_enabled = true
  in_transit_encryption_enabled = true
  
  maintenance_window = "sun:03:00-sun:04:00"
  
  tags = {
    Name = "taskir-redis-${var.environment}"
  }
}

# ============================================================================
# S3 for static assets and backups
# ============================================================================

resource "aws_s3_bucket" "taskir_assets" {
  bucket = "taskir-assets-${var.environment}-${data.aws_caller_identity.current.account_id}"

  tags = {
    Name = "taskir-assets-${var.environment}"
  }
}

resource "aws_s3_bucket_versioning" "taskir_assets" {
  bucket = aws_s3_bucket.taskir_assets.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "taskir_assets" {
  bucket = aws_s3_bucket.taskir_assets.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.s3.arn
    }
  }
}

resource "aws_s3_bucket_public_access_block" "taskir_assets" {
  bucket = aws_s3_bucket.taskir_assets.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ============================================================================
# CloudFront CDN
# ============================================================================

resource "aws_cloudfront_distribution" "taskir" {
  origin {
    domain_name = aws_s3_bucket.taskir_assets.bucket_regional_domain_name
    origin_id   = "taskir_assets"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.taskir.cloudfront_access_identity_path
    }
  }

  enabled = true
  is_ipv6_enabled = true
  default_root_object = "index.html"

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "taskir_assets"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = {
    Name = "taskir-cdn-${var.environment}"
  }
}

resource "aws_cloudfront_origin_access_identity" "taskir" {
  comment = "OAI for TaskirX ${var.environment}"
}

# ============================================================================
# KMS Encryption Keys
# ============================================================================

resource "aws_kms_key" "rds" {
  description             = "KMS key for RDS encryption"
  deletion_window_in_days = 10
  enable_key_rotation     = true

  tags = {
    Name = "taskir-rds-key-${var.environment}"
  }
}

resource "aws_kms_key" "s3" {
  description             = "KMS key for S3 encryption"
  deletion_window_in_days = 10
  enable_key_rotation     = true

  tags = {
    Name = "taskir-s3-key-${var.environment}"
  }
}

# ============================================================================
# CloudWatch Monitoring
# ============================================================================

resource "aws_cloudwatch_log_group" "eks" {
  name              = "/aws/eks/taskir-${var.environment}"
  retention_in_days = 30

  tags = {
    Name = "taskir-eks-logs-${var.environment}"
  }
}

resource "aws_cloudwatch_log_group" "application" {
  name              = "/aws/taskir/application-${var.environment}"
  retention_in_days = 30

  tags = {
    Name = "taskir-app-logs-${var.environment}"
  }
}

# ============================================================================
# IAM Roles and Policies
# ============================================================================

resource "aws_iam_role" "eks_cluster_role" {
  name = "taskir-eks-cluster-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster_role.name
}

resource "aws_iam_role_policy_attachment" "eks_vpc_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_cluster_role.name
}

resource "aws_iam_role" "eks_node_role" {
  name = "taskir-eks-node-role-${var.environment}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_node_role.name
}

resource "aws_iam_role_policy_attachment" "eks_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_node_role.name
}

resource "aws_iam_role_policy_attachment" "eks_registry_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_node_role.name
}

# ============================================================================
# Security Groups
# ============================================================================

resource "aws_security_group" "eks_cluster_sg" {
  name        = "taskir-eks-sg-${var.environment}"
  description = "Security group for TaskirX EKS cluster"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "taskir-eks-sg-${var.environment}"
  }
}

resource "aws_security_group" "redis_sg" {
  name        = "taskir-redis-sg-${var.environment}"
  description = "Security group for TaskirX Redis"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.eks_cluster_sg.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "taskir-redis-sg-${var.environment}"
  }
}

# ============================================================================
# Random Password for Database
# ============================================================================

resource "random_password" "db_password" {
  length  = 32
  special = true
}

# ============================================================================
# Data Sources
# ============================================================================

data "aws_caller_identity" "current" {}

# ============================================================================
# Outputs
# ============================================================================

output "eks_cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = aws_eks_cluster.taskir.endpoint
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = aws_eks_cluster.taskir.name
}

output "rds_endpoint" {
  description = "RDS database endpoint"
  value       = module.rds.endpoint
  sensitive   = true
}

output "redis_endpoint" {
  description = "Redis endpoint"
  value       = module.redis.endpoint
}

output "s3_bucket_name" {
  description = "S3 bucket for assets"
  value       = aws_s3_bucket.taskir_assets.id
}

output "cloudfront_domain" {
  description = "CloudFront distribution domain"
  value       = aws_cloudfront_distribution.taskir.domain_name
}
