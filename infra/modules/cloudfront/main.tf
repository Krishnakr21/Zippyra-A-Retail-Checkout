variable "environment" { type = string }
variable "products_bucket_domain" { type = string }
variable "products_bucket_arn" { type = string }

# CloudFront Origin Access Identity for products bucket
resource "aws_cloudfront_origin_access_identity" "products_oai" {
  comment = "OAI for zippyra-${var.environment}-products bucket"
}

# Grant CloudFront OAI read access to products bucket
resource "aws_s3_bucket_policy" "products_cf_policy" {
  bucket = "zippyra-${var.environment}-products"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "AllowCloudFrontOAI"
      Effect    = "Allow"
      Principal = { AWS = aws_cloudfront_origin_access_identity.products_oai.iam_arn }
      Action    = "s3:GetObject"
      Resource  = "${var.products_bucket_arn}/*"
    }]
  })
}

# CloudFront Distribution — products assets + static web
resource "aws_cloudfront_distribution" "products_cdn" {
  enabled             = true
  is_ipv6_enabled     = true
  comment             = "Zippyra ${var.environment} product images & static assets CDN"
  default_root_object = "index.html"
  price_class         = "PriceClass_200" # Includes India edge locations

  origin {
    domain_name = var.products_bucket_domain
    origin_id   = "S3-products"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.products_oai.cloudfront_access_identity_path
    }
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-products"
    viewer_protocol_policy = "redirect-to-https"
    compress               = true

    # 24h TTL matching the platform's documented caching strategy
    min_ttl     = 0
    default_ttl = 86400  # 24 hours
    max_ttl     = 604800 # 7 days

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
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
    Name        = "zippyra-${var.environment}-cdn"
    Environment = var.environment
  }
}

output "cdn_domain_name" {
  value = aws_cloudfront_distribution.products_cdn.domain_name
}

output "cdn_distribution_id" {
  value = aws_cloudfront_distribution.products_cdn.id
}
