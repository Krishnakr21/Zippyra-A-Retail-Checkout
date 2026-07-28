variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "vpc_cidr" { type = string }
variable "node_type" { type = string; default = "cache.t4g.medium" }

# -------------------------------------------------------------------
# 6 separate Redis replication groups matching the eviction policies
# specified in the platform's backend env file
# -------------------------------------------------------------------

locals {
  redis_clusters = {
    cart = {
      port            = 6379
      eviction_policy = "volatile-ttl"
      description     = "Cart sessions, offer rules, checkout locks"
    }
    session = {
      port            = 6380
      eviction_policy = "volatile-ttl"
      description     = "Auth sessions (JWT refresh, OTP rate limit)"
    }
    sku_cache = {
      port            = 6381
      eviction_policy = "allkeys-lru"
      description     = "Barcode→SKU lookup hot cache (catalog-service)"
    }
    rate_limit = {
      port            = 6382
      eviction_policy = "volatile-ttl"
      description     = "API rate limiting sliding windows"
    }
    exit_token = {
      port            = 6383
      eviction_policy = "noeviction"
      description     = "EXIT TOKENS — MUST NEVER BE EVICTED under memory pressure. A paid customer's exit token eviction = customer trapped at gate."
    }
    realtime = {
      port            = 6384
      eviction_policy = "volatile-ttl"
      description     = "WebSocket presence, device heartbeats, real-time analytics counters"
    }
  }
}

resource "aws_security_group" "redis_sg" {
  name        = "zippyra-${var.environment}-redis-sg"
  description = "ElastiCache Redis security group"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 6379
    to_port     = 6390
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_subnet_group" "redis" {
  name       = "zippyra-${var.environment}-redis-subnet"
  subnet_ids = var.private_subnet_ids
}

resource "aws_elasticache_parameter_group" "redis_params" {
  for_each = local.redis_clusters

  name        = "zippyra-${var.environment}-${each.key}-params"
  family      = "redis7"
  description = each.value.description

  parameter {
    name  = "maxmemory-policy"
    value = each.value.eviction_policy
  }
}

resource "aws_elasticache_replication_group" "redis" {
  for_each = local.redis_clusters

  replication_group_id = "zippyra-${var.environment}-${each.key}"
  description          = each.value.description
  node_type            = var.node_type
  num_cache_clusters   = 1 # Single node per cluster for pilot (no cluster-mode — Phase 1)
  port                 = each.value.port
  parameter_group_name = aws_elasticache_parameter_group.redis_params[each.key].name
  subnet_group_name    = aws_elasticache_subnet_group.redis.name
  security_group_ids   = [aws_security_group.redis_sg.id]

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true

  automatic_failover_enabled = false # Single-node pilot; enable for production

  tags = {
    Name             = "zippyra-${var.environment}-redis-${each.key}"
    Environment      = var.environment
    EvictionPolicy   = each.value.eviction_policy
    CriticalityNote  = each.key == "exit_token" ? "NON-NEGOTIABLE: noeviction — paid customer exit tokens must never be evicted" : ""
  }
}

output "redis_endpoints" {
  value = { for k, v in aws_elasticache_replication_group.redis : k => v.primary_endpoint_address }
}
