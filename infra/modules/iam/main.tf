variable "environment" { type = string }
variable "oidc_provider_arn" { type = string }
variable "oidc_provider_url" { type = string }

# -------------------------------------------------------------------
# Per-service IAM roles (least-privilege IRSA)
# Each service can ONLY access its own Secrets Manager path
# -------------------------------------------------------------------

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

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "service_roles" {
  for_each = toset(local.services)

  name = "zippyra-${var.environment}-${each.key}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Federated = var.oidc_provider_arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "${var.oidc_provider_url}:sub" = "system:serviceaccount:zippyra-${var.environment}:${each.key}"
          "${var.oidc_provider_url}:aud" = "sts.amazonaws.com"
        }
      }
    }]
  })

  tags = {
    Service     = each.key
    Environment = var.environment
  }
}

# Least-privilege Secrets Manager access — each service reads ONLY its own path
resource "aws_iam_role_policy" "secrets_access" {
  for_each = toset(local.services)

  name = "secrets-access"
  role = aws_iam_role.service_roles[each.key].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
        ]
        Resource = "arn:aws:secretsmanager:ap-south-1:${data.aws_caller_identity.current.account_id}:secret:zippyra/${var.environment}/${each.key}*"
      },
      {
        # Shared secrets readable by all services (e.g., JWT signing key)
        Effect = "Allow"
        Action = ["secretsmanager:GetSecretValue"]
        Resource = "arn:aws:secretsmanager:ap-south-1:${data.aws_caller_identity.current.account_id}:secret:zippyra/${var.environment}/shared*"
      }
    ]
  })
}

# Additional per-service policies for services that need specific AWS resources
resource "aws_iam_role_policy" "payment_service_extra" {
  name = "payment-service-extra"
  role = aws_iam_role.service_roles["payment-service"].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["kms:Decrypt", "kms:Encrypt"]
      Resource = "*"
      Condition = {
        StringEquals = {
          "kms:ViaService" = "secretsmanager.ap-south-1.amazonaws.com"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "compliance_service_s3" {
  name = "compliance-service-s3"
  role = aws_iam_role.service_roles["compliance-service"].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:PutObject", "s3:GetObject"]
      Resource = "arn:aws:s3:::zippyra-${var.environment}-invoices/*"
    }]
  })
}

resource "aws_iam_role_policy" "catalog_service_s3" {
  name = "catalog-service-s3"
  role = aws_iam_role.service_roles["catalog-service"].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:PutObject", "s3:GetObject", "s3:DeleteObject"]
      Resource = [
        "arn:aws:s3:::zippyra-${var.environment}-products/*",
        "arn:aws:s3:::zippyra-${var.environment}-media/*",
      ]
    }]
  })
}

output "service_role_arns" {
  value = { for k, v in aws_iam_role.service_roles : k => v.arn }
}
