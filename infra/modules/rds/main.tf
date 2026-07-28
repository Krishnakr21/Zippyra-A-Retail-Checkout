variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "database_subnet_ids" { type = list(string) }
variable "vpc_cidr" { type = string }
variable "instance_class" { type = string; default = "db.t4g.large" }
variable "allocated_storage" { type = number; default = 50 }

resource "aws_db_subnet_group" "rds" {
  name       = "zippyra-${var.environment}-db-subnet-group"
  subnet_ids = var.database_subnet_ids

  tags = {
    Name = "zippyra-${var.environment}-db-subnet-group"
  }
}

resource "aws_security_group" "rds_sg" {
  name        = "zippyra-${var.environment}-rds-sg"
  description = "Security group for PostgreSQL RDS"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
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

resource "aws_db_instance" "postgres" {
  identifier                  = "zippyra-${var.environment}-postgres"
  engine                      = "postgres"
  engine_version              = "16.1"
  instance_class              = var.instance_class
  allocated_storage           = var.allocated_storage
  max_allocated_storage       = 500
  storage_type                = "gp3"
  multi_az                    = true
  db_name                     = "zippyra"
  username                    = "zippyra_admin"
  manage_master_user_password = true
  db_subnet_group_name        = aws_db_subnet_group.rds.name
  vpc_security_group_ids      = [aws_security_group.rds_sg.id]

  backup_retention_period = 7
  backup_window           = "19:30-20:30" # 01:00-02:00 IST window (UTC)
  maintenance_window      = "Sun:21:00-Sun:22:00"

  performance_insights_enabled = true
  iam_database_authentication_enabled = true
  deletion_protection          = false

  tags = {
    Name        = "zippyra-${var.environment}-postgres"
    Environment = var.environment
  }
}

output "db_instance_id" {
  value = aws_db_instance.postgres.id
}

output "db_endpoint" {
  value = aws_db_instance.postgres.endpoint
}

output "db_security_group_id" {
  value = aws_security_group.rds_sg.id
}
