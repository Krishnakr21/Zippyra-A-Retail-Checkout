variable "environment" { type = string }
variable "alb_arn" { type = string; default = "" }

# AWS WAF v2 Web ACL
resource "aws_wafv2_web_acl" "main" {
  name        = "zippyra-${var.environment}-waf"
  description = "WAF for Zippyra ${var.environment} ALB"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  # Rule 1: AWS Managed Common Rule Set
  rule {
    name     = "AWSManagedRulesCommonRuleSet"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        vendor_name = "AWS"
        name        = "AWSManagedRulesCommonRuleSet"

        # Path-specific exception: catalog CSV import needs >1MB bodies (up to 25MB)
        scope_down_statement {
          not_statement {
            statement {
              byte_match_statement {
                field_to_match {
                  uri_path {}
                }
                positional_constraint = "STARTS_WITH"
                search_string         = "/v1/catalog/import"
                text_transformation {
                  priority = 0
                  type     = "NONE"
                }
              }
            }
          }
        }
      }
    }

    visibility_config {
      sampled_requests_enabled   = true
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedRulesCommonRuleSet"
    }
  }

  # Rule 2: SQL Injection Protection
  rule {
    name     = "AWSManagedRulesSQLiRuleSet"
    priority = 2

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        vendor_name = "AWS"
        name        = "AWSManagedRulesSQLiRuleSet"
      }
    }

    visibility_config {
      sampled_requests_enabled   = true
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedRulesSQLiRuleSet"
    }
  }

  # Rule 3: Auth endpoint rate limiting (2000 req / 5 min per IP)
  rule {
    name     = "AuthEndpointRateLimit"
    priority = 3

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 2000
        aggregate_key_type = "IP"

        scope_down_statement {
          or_statement {
            statement {
              byte_match_statement {
                field_to_match { uri_path {} }
                positional_constraint = "STARTS_WITH"
                search_string         = "/v1/auth/"
                text_transformation {
                  priority = 0
                  type     = "NONE"
                }
              }
            }
            statement {
              byte_match_statement {
                field_to_match { uri_path {} }
                positional_constraint = "STARTS_WITH"
                search_string         = "/v1/admin-auth/"
                text_transformation {
                  priority = 0
                  type     = "NONE"
                }
              }
            }
            statement {
              byte_match_statement {
                field_to_match { uri_path {} }
                positional_constraint = "STARTS_WITH"
                search_string         = "/v1/retailer-auth/"
                text_transformation {
                  priority = 0
                  type     = "NONE"
                }
              }
            }
            statement {
              byte_match_statement {
                field_to_match { uri_path {} }
                positional_constraint = "STARTS_WITH"
                search_string         = "/v1/chain-hq-auth/"
                text_transformation {
                  priority = 0
                  type     = "NONE"
                }
              }
            }
          }
        }
      }
    }

    visibility_config {
      sampled_requests_enabled   = true
      cloudwatch_metrics_enabled = true
      metric_name                = "AuthEndpointRateLimit"
    }
  }

  # Rule 4: Large body blocking (>1MB except catalog CSV import at 25MB)
  rule {
    name     = "LargeBodyBlock"
    priority = 4

    action {
      block {}
    }

    statement {
      and_statement {
        statement {
          size_constraint_statement {
            field_to_match { body {} }
            comparison_operator = "GT"
            size                = 1048576 # 1MB
            text_transformation {
              priority = 0
              type     = "NONE"
            }
          }
        }
        statement {
          not_statement {
            statement {
              byte_match_statement {
                field_to_match { uri_path {} }
                positional_constraint = "STARTS_WITH"
                search_string         = "/v1/catalog/import"
                text_transformation {
                  priority = 0
                  type     = "NONE"
                }
              }
            }
          }
        }
      }
    }

    visibility_config {
      sampled_requests_enabled   = true
      cloudwatch_metrics_enabled = true
      metric_name                = "LargeBodyBlock"
    }
  }

  # Rule 5: Razorpay webhook IP allowlist (Gap #26)
  # Payment-service and integration-service webhook endpoints should ONLY accept
  # from Razorpay's IP ranges
  rule {
    name     = "RazorpayWebhookIPAllowlist"
    priority = 5

    action {
      block {}
    }

    statement {
      and_statement {
        statement {
          byte_match_statement {
            field_to_match { uri_path {} }
            positional_constraint = "STARTS_WITH"
            search_string         = "/v1/payment/webhook"
            text_transformation {
              priority = 0
              type     = "NONE"
            }
          }
        }
        statement {
          not_statement {
            statement {
              ip_set_reference_statement {
                arn = aws_wafv2_ip_set.razorpay_ips.arn
              }
            }
          }
        }
      }
    }

    visibility_config {
      sampled_requests_enabled   = true
      cloudwatch_metrics_enabled = true
      metric_name                = "RazorpayWebhookIPAllowlist"
    }
  }

  visibility_config {
    sampled_requests_enabled   = true
    cloudwatch_metrics_enabled = true
    metric_name                = "ZippyraWAF"
  }

  tags = {
    Name        = "zippyra-${var.environment}-waf"
    Environment = var.environment
  }
}

# Razorpay IP Set (update with actual Razorpay IP ranges)
resource "aws_wafv2_ip_set" "razorpay_ips" {
  name               = "razorpay-webhook-ips"
  description        = "Razorpay payment gateway webhook source IPs (Gap #26)"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"

  # Razorpay published webhook IP ranges — update as they publish new ranges
  addresses = [
    "52.66.166.0/24",
    "13.235.0.0/16",
  ]
}

# Associate WAF with ALB (if ALB ARN provided)
resource "aws_wafv2_web_acl_association" "alb_association" {
  count        = var.alb_arn != "" ? 1 : 0
  resource_arn = var.alb_arn
  web_acl_arn  = aws_wafv2_web_acl.main.arn
}

output "web_acl_arn" {
  value = aws_wafv2_web_acl.main.arn
}
