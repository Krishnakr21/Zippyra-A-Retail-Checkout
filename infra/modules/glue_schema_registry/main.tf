variable "environment" { type = string }

resource "aws_glue_registry" "zippyra" {
  registry_name = "zippyra-${var.environment}-schema-registry"
  description   = "Schema registry for Kafka topic schema evolution safety"

  tags = {
    Name        = "zippyra-${var.environment}-schema-registry"
    Environment = var.environment
  }
}

# Schema definitions for critical event topics that analytics-service
# and payment/order-service outbox patterns depend on
locals {
  schemas = {
    "order.completed" = {
      description = "Order completed event schema"
      schema = jsonencode({
        type = "record"
        name = "OrderCompleted"
        fields = [
          { name = "event_id", type = "string" },
          { name = "order_id", type = "string" },
          { name = "store_id", type = "string" },
          { name = "chain_id", type = "string" },
          { name = "customer_id", type = "string" },
          { name = "total_paise", type = "long" },
          { name = "item_count", type = "int" },
          { name = "payment_mode", type = "string" },
          { name = "completed_at", type = "string" },
        ]
      })
    }
    "payment.confirmed" = {
      description = "Payment confirmed event schema"
      schema = jsonencode({
        type = "record"
        name = "PaymentConfirmed"
        fields = [
          { name = "event_id", type = "string" },
          { name = "payment_id", type = "string" },
          { name = "order_id", type = "string" },
          { name = "store_id", type = "string" },
          { name = "amount_paise", type = "long" },
          { name = "payment_mode", type = "string" },
          { name = "gateway_payment_id", type = ["null", "string"] },
          { name = "confirmed_at", type = "string" },
        ]
      })
    }
    "exit.validated" = {
      description = "Exit validation event schema"
      schema = jsonencode({
        type = "record"
        name = "ExitValidated"
        fields = [
          { name = "event_id", type = "string" },
          { name = "exit_token_id", type = "string" },
          { name = "store_id", type = "string" },
          { name = "customer_id", type = "string" },
          { name = "validated_at", type = "string" },
        ]
      })
    }
    "inventory.stock_updated" = {
      description = "Inventory stock update event schema"
      schema = jsonencode({
        type = "record"
        name = "InventoryStockUpdated"
        fields = [
          { name = "event_id", type = "string" },
          { name = "store_id", type = "string" },
          { name = "barcode", type = "string" },
          { name = "previous_qty", type = "int" },
          { name = "new_qty", type = "int" },
          { name = "reason", type = "string" },
          { name = "updated_at", type = "string" },
        ]
      })
    }
  }
}

resource "aws_glue_schema" "event_schemas" {
  for_each = local.schemas

  schema_name       = replace(each.key, ".", "_")
  registry_arn      = aws_glue_registry.zippyra.arn
  data_format       = "JSON"
  compatibility     = "BACKWARD"
  description       = each.value.description
  schema_definition = each.value.schema
}

output "registry_arn" {
  value = aws_glue_registry.zippyra.arn
}
