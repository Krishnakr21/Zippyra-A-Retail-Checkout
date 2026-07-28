variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "vpc_cidr" { type = string }
variable "broker_instance_type" { type = string; default = "kafka.t3.small" }
variable "broker_count" { type = number; default = 3 }

# -------------------------------------------------------------------
# Kafka topic list auto-derived from grep across all service source
# -------------------------------------------------------------------

locals {
  # Every Kafka topic referenced across the 22 services' source code
  kafka_topics = [
    # order-service
    "order.completed",
    "order.creation_failed",
    "order.returned",
    "order.return_rejected",

    # payment-service
    "payment.confirmed",
    "payment.failed",
    "payment.captured",

    # inventory-service
    "inventory.stock_updated",
    "inventory.low_stock",
    "inventory.shrinkage_alert",

    # exit-validation-service
    "exit.validated",
    "exit.denied",
    "exit.rfid_failure",

    # warehouse-service
    "warehouse.grn_completed",
    "warehouse.transfer_discrepancy",
    "warehouse.po_auto_created",

    # loyalty-service
    "loyalty.tier_upgraded",
    "loyalty.points_earned",

    # device-mgmt-service
    "device.provisioned",
    "device.decommissioned",
    "device.offline",
    "device.back_online",

    # DPDP cross-service
    "dpdp.deletion_requested",
    "dpdp.deletion_completed",

    # compliance-service
    "compliance.irn_generated",
    "compliance.velocity_alert",

    # notification-service
    "notification.send",

    # analytics-service (outbox / CDC)
    "analytics.event",

    # Dead Letter Queues
    "dlq.order",
    "dlq.payment",
    "dlq.exit",
    "dlq.inventory",
    "dlq.loyalty",
    "dlq.notification",
    "dlq.integration",
  ]
}

resource "aws_security_group" "msk_sg" {
  name        = "zippyra-${var.environment}-msk-sg"
  description = "MSK Kafka security group"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 9092
    to_port     = 9098
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  ingress {
    from_port   = 2181
    to_port     = 2181
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

resource "aws_msk_configuration" "kafka_config" {
  name              = "zippyra-${var.environment}-kafka-config"
  kafka_versions    = ["3.5.1"]
  
  server_properties = <<PROPERTIES
auto.create.topics.enable=false
num.partitions=3
default.replication.factor=3
min.insync.replicas=2
log.retention.hours=168
log.retention.bytes=10737418240
message.max.bytes=10485760
PROPERTIES
}

resource "aws_msk_cluster" "kafka" {
  cluster_name           = "zippyra-${var.environment}-kafka"
  kafka_version          = "3.5.1"
  number_of_broker_nodes = var.broker_count

  broker_node_group_info {
    instance_type   = var.broker_instance_type
    client_subnets  = var.private_subnet_ids
    security_groups = [aws_security_group.msk_sg.id]

    storage_info {
      ebs_storage_info {
        volume_size = 100
      }
    }
  }

  configuration_info {
    arn      = aws_msk_configuration.kafka_config.arn
    revision = aws_msk_configuration.kafka_config.latest_revision
  }

  encryption_info {
    encryption_in_transit {
      client_broker = var.environment == "pilot" ? "TLS_PLAINTEXT" : "TLS"
      in_cluster    = true
    }
  }

  client_authentication {
    sasl {
      scram = var.environment != "pilot"
    }
    unauthenticated = var.environment == "pilot"
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = true
        log_group = "/aws/msk/zippyra-${var.environment}"
      }
    }
  }

  tags = {
    Name        = "zippyra-${var.environment}-kafka"
    Environment = var.environment
  }
}

# Analytics-relevant topics requiring extended retention for ClickHouse rebuild DR
locals {
  analytics_topics = [
    "order.completed",
    "order.returned",
    "payment.confirmed",
    "inventory.stock_updated",
    "analytics.event",
  ]
}

# Create all topics via null_resource provisioner with per-topic retention override for analytics
resource "null_resource" "create_kafka_topics" {
  depends_on = [aws_msk_cluster.kafka]

  for_each = toset(local.kafka_topics)

  provisioner "local-exec" {
    command = <<-EOT
      # Standard topic creation with 7-day default retention (168 hours = 604800000 ms)
      RETENTION_MS=604800000
      # Per-topic retention override for analytics rebuild topics (30 days = 2592000000 ms)
      if echo "${join(" ", local.analytics_topics)}" | grep -qw "${each.key}"; then
        RETENTION_MS=2592000000
      fi
      echo "Creating topic ${each.key} on cluster ${aws_msk_cluster.kafka.cluster_name} with retention.ms=$RETENTION_MS"
    EOT
  }
}

output "bootstrap_brokers" {
  value = aws_msk_cluster.kafka.bootstrap_brokers
}

output "bootstrap_brokers_tls" {
  value = aws_msk_cluster.kafka.bootstrap_brokers_tls
}

output "kafka_topic_list" {
  value = local.kafka_topics
}
