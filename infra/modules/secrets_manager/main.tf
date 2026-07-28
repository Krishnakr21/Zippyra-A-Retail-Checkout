variable "environment" { type = string }

locals {
  services = [
    "auth-service",
    "admin-auth-service",
    "retailer-auth-service",
    "chain-hq-auth-service",
    "store-service",
    "catalog-service",
    "inventory-service",
    "cart-service",
    "payment-service",
    "order-service",
    "exit-validation-service",
    "loyalty-service",
    "notification-service",
    "analytics-service",
    "compliance-service",
    "device-mgmt-service",
    "warehouse-service",
    "chain-hq-service",
    "integration-service",
    "customer-support-service",
    "staffing-service",
    "audit-service",
  ]
}

# One secret path per service per environment
# Matching prefix: AWS_SECRETS_MANAGER_PREFIX=zippyra/{env}/
resource "aws_secretsmanager_secret" "service_secrets" {
  for_each = toset(local.services)

  name        = "zippyra/${var.environment}/${each.key}"
  description = "Environment secrets for ${each.key} in ${var.environment}"

  tags = {
    Service     = each.key
    Environment = var.environment
  }
}

# Shared secrets accessible by all services (JWT signing key, DB common creds)
resource "aws_secretsmanager_secret" "shared_secrets" {
  name        = "zippyra/${var.environment}/shared"
  description = "Shared secrets accessible by all services (JWT key, common DB creds)"

  tags = {
    Environment = var.environment
  }
}

output "secret_arns" {
  value = { for k, v in aws_secretsmanager_secret.service_secrets : k => v.arn }
}
