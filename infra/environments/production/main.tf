terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "zippyra-terraform-state"
    key            = "production/terraform.tfstate"
    region         = "ap-south-1"
    dynamodb_table = "zippyra-terraform-locks"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  type    = string
  default = "ap-south-1"
  validation {
    condition     = contains(["ap-south-1", "ap-south-2"], var.aws_region)
    error_message = "Must use an India region for DPDP data localization."
  }
}

variable "environment" {
  type    = string
  default = "production"
}

# -------------------------------------------------------------------
# Production sizing: r6g instances, Redis cluster-mode, MSK m5.large
# -------------------------------------------------------------------

module "vpc" {
  source      = "../../modules/vpc"
  aws_region  = var.aws_region
  environment = var.environment
  vpc_cidr    = "10.1.0.0/16"
}

module "iam" {
  source            = "../../modules/iam"
  environment       = var.environment
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.oidc_provider_url
}

module "rds" {
  source              = "../../modules/rds"
  environment         = var.environment
  vpc_id              = module.vpc.vpc_id
  database_subnet_ids = module.vpc.database_subnet_ids
  vpc_cidr            = "10.1.0.0/16"
  instance_class      = "db.r6g.xlarge"
  allocated_storage   = 200
}

module "rds_proxy" {
  source               = "../../modules/rds_proxy"
  environment          = var.environment
  vpc_id               = module.vpc.vpc_id
  private_subnet_ids   = module.vpc.private_subnet_ids
  db_security_group_id = module.rds.db_security_group_id
  db_instance_id       = module.rds.db_instance_id
}

module "elasticache" {
  source             = "../../modules/elasticache"
  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  vpc_cidr           = "10.1.0.0/16"
  node_type          = "cache.r6g.large"
}

module "msk" {
  source               = "../../modules/msk"
  environment          = var.environment
  vpc_id               = module.vpc.vpc_id
  private_subnet_ids   = module.vpc.private_subnet_ids
  vpc_cidr             = "10.1.0.0/16"
  broker_instance_type = "kafka.m5.large"
  broker_count         = 3
}

module "glue_schema_registry" {
  source      = "../../modules/glue_schema_registry"
  environment = var.environment
}

module "eks" {
  source             = "../../modules/eks"
  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  instance_types     = ["m5.xlarge", "m5.large"]
  desired_size       = 6
  min_size           = 4
  max_size           = 20
  spot_percentage    = 60
}

module "secrets_manager" {
  source      = "../../modules/secrets_manager"
  environment = var.environment
}

module "s3" {
  source      = "../../modules/s3"
  environment = var.environment
}

module "cloudfront" {
  source                 = "../../modules/cloudfront"
  environment            = var.environment
  products_bucket_domain = module.s3.products_bucket_domain
  products_bucket_arn    = module.s3.products_bucket_arn
}

module "waf" {
  source      = "../../modules/waf"
  environment = var.environment
}

output "vpc_id" { value = module.vpc.vpc_id }
output "eks_cluster_name" { value = module.eks.cluster_name }
output "rds_endpoint" { value = module.rds.db_endpoint }
output "rds_proxy_endpoint" { value = module.rds_proxy.proxy_endpoint }
output "redis_endpoints" { value = module.elasticache.redis_endpoints }
output "kafka_brokers" { value = module.msk.bootstrap_brokers_tls }
output "cdn_domain" { value = module.cloudfront.cdn_domain_name }
