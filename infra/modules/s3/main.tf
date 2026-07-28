variable "environment" { type = string }
variable "aws_region" { type = string; default = "ap-south-1" }
variable "dr_region" { type = string; default = "ap-south-2" }

# -------------------------------------------------------------------
# S3 Buckets: invoices (private/versioned/Glacier), products (OAI), media
# -------------------------------------------------------------------

# Invoices bucket — private, versioned, Glacier lifecycle after 30 days
resource "aws_s3_bucket" "invoices" {
  bucket = "zippyra-${var.environment}-invoices"

  tags = {
    Name        = "zippyra-${var.environment}-invoices"
    Environment = var.environment
  }
}

resource "aws_s3_bucket_versioning" "invoices" {
  bucket = aws_s3_bucket.invoices.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "invoices_lifecycle" {
  bucket = aws_s3_bucket.invoices.id

  rule {
    id     = "glacier-after-30-days"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "GLACIER"
    }

    noncurrent_version_expiration {
      noncurrent_days = 365
    }
  }
}

resource "aws_s3_bucket_public_access_block" "invoices_block" {
  bucket                  = aws_s3_bucket.invoices.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_server_side_encryption_configuration" "invoices_encryption" {
  bucket = aws_s3_bucket.invoices.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "aws:kms"
    }
  }
}

# Products bucket — private, served via CloudFront OAI only (NO direct public access)
resource "aws_s3_bucket" "products" {
  bucket = "zippyra-${var.environment}-products"

  tags = {
    Name        = "zippyra-${var.environment}-products"
    Environment = var.environment
  }
}

resource "aws_s3_bucket_public_access_block" "products_block" {
  bucket                  = aws_s3_bucket.products.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Media bucket
resource "aws_s3_bucket" "media" {
  bucket = "zippyra-${var.environment}-media"

  tags = {
    Name        = "zippyra-${var.environment}-media"
    Environment = var.environment
  }
}

resource "aws_s3_bucket_public_access_block" "media_block" {
  bucket                  = aws_s3_bucket.media.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Cross-region replication to ap-south-2 (Hyderabad) for DR
resource "aws_iam_role" "replication_role" {
  name = "zippyra-${var.environment}-s3-replication-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "s3.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy" "replication_policy" {
  name = "s3-replication"
  role = aws_iam_role.replication_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetReplicationConfiguration",
          "s3:ListBucket",
          "s3:GetObjectVersionForReplication",
          "s3:GetObjectVersionAcl",
          "s3:GetObjectVersionTagging",
        ]
        Resource = [
          aws_s3_bucket.invoices.arn,
          "${aws_s3_bucket.invoices.arn}/*",
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:ReplicateObject",
          "s3:ReplicateDelete",
          "s3:ReplicateTags",
        ]
        Resource = "arn:aws:s3:::zippyra-${var.environment}-invoices-dr/*"
      }
    ]
  })
}

resource "aws_s3_bucket_replication_configuration" "invoices_replication" {
  depends_on = [aws_s3_bucket_versioning.invoices]
  role       = aws_iam_role.replication_role.arn
  bucket     = aws_s3_bucket.invoices.id

  rule {
    id     = "invoices-dr-replication"
    status = "Enabled"

    destination {
      bucket        = "arn:aws:s3:::zippyra-${var.environment}-invoices-dr"
      storage_class = "STANDARD_IA"
    }
  }
}

output "invoices_bucket_arn" { value = aws_s3_bucket.invoices.arn }
output "products_bucket_arn" { value = aws_s3_bucket.products.arn }
output "products_bucket_domain" { value = aws_s3_bucket.products.bucket_regional_domain_name }
output "media_bucket_arn" { value = aws_s3_bucket.media.arn }
