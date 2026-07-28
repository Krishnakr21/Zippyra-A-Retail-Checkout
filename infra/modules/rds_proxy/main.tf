variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "db_security_group_id" { type = string }
variable "db_instance_id" { type = string }

resource "aws_db_proxy" "rds_proxy" {
  name                   = "zippyra-${var.environment}-rds-proxy"
  debug_logging          = false
  engine_family          = "POSTGRESQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.rds_proxy_role.arn
  vpc_security_group_ids = [aws_security_group.rds_proxy_sg.id]
  vpc_subnet_ids         = var.private_subnet_ids

  auth {
    auth_scheme = "SECRETS"
    iam_auth    = "REQUIRED"
    secret_arn  = aws_secretsmanager_secret.rds_proxy_creds.arn
  }

  tags = {
    Name        = "zippyra-${var.environment}-rds-proxy"
    Environment = var.environment
    Purpose     = "22 services × 3 pods × 10 conns = 660 connections pooled through proxy against ~320 RDS limit"
  }
}

resource "aws_db_proxy_default_target_group" "default" {
  db_proxy_name = aws_db_proxy.rds_proxy.name

  connection_pool_config {
    max_connections_percent      = 90
    max_idle_connections_percent = 50
    connection_borrow_timeout    = 120
  }
}

resource "aws_db_proxy_target" "rds_target" {
  db_proxy_name          = aws_db_proxy.rds_proxy.name
  target_group_name      = aws_db_proxy_default_target_group.default.name
  db_instance_identifier = var.db_instance_id
}

resource "aws_security_group" "rds_proxy_sg" {
  name        = "zippyra-${var.environment}-rds-proxy-sg"
  description = "RDS Proxy security group"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.db_security_group_id]
  }
}

resource "aws_iam_role" "rds_proxy_role" {
  name = "zippyra-${var.environment}-rds-proxy-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "rds.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy" "rds_proxy_secrets" {
  name = "rds-proxy-secrets"
  role = aws_iam_role.rds_proxy_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["secretsmanager:GetSecretValue", "secretsmanager:GetResourcePolicy", "secretsmanager:DescribeSecret"]
      Resource = [aws_secretsmanager_secret.rds_proxy_creds.arn]
    }]
  })
}

resource "aws_secretsmanager_secret" "rds_proxy_creds" {
  name        = "zippyra/${var.environment}/rds-proxy-credentials"
  description = "RDS Proxy credentials for PostgreSQL"
}

output "proxy_endpoint" {
  value = aws_db_proxy.rds_proxy.endpoint
}
