# Sentry Terraform Configuration for Mobile Apps & PagerDuty Integration Rules

terraform {
  required_providers {
    sentry = {
      source  = "jessecuff/sentry"
      version = "~> 0.12.0"
    }
  }
}

variable "sentry_organization" {
  type        = string
  default     = "zippyra"
  description = "Sentry Organization Slug"
}

variable "pagerduty_service_integration_key" {
  type        = string
  default     = "pd-integration-key-mobile-crashes"
  sensitive   = true
  description = "PagerDuty Integration Key for Mobile Crash Escalation"
}

# -------------------------------------------------------------------
# Sentry Projects for Mobile Applications
# -------------------------------------------------------------------
resource "sentry_project" "mobile_customer_app" {
  organization = var.sentry_organization
  team         = "mobile-team"
  name         = "Mobile Customer App"
  slug         = "mobile-customer-app"
  platform     = "flutter"
}

resource "sentry_project" "mobile_staff_app" {
  organization = var.sentry_organization
  team         = "mobile-team"
  name         = "Mobile Staff App"
  slug         = "mobile-staff-app"
  platform     = "flutter"
}

locals {
  sentry_mobile_projects = [
    sentry_project.mobile_customer_app.slug,
    sentry_project.mobile_staff_app.slug,
  ]
}

# -------------------------------------------------------------------
# ALERT RULE 1: High-Impact Crash (>10 unique users in 1 hour) -> PagerDuty SEV2
# -------------------------------------------------------------------
resource "sentry_issue_alert" "rule1_high_impact_users" {
  for_each     = toset(local.sentry_mobile_projects)
  organization = var.sentry_organization
  project      = each.key
  name         = "P2 SEV2: High Impact Issue (>10 Users in 1h)"
  action_match = "any"
  filter_match = "all"

  conditions = [
    {
      id    = "sentry.rules.conditions.unique_users_value.UniqueUsersValueCondition"
      value = 10
      interval = "1h"
    }
  ]

  actions = [
    {
      id      = "sentry.rules.actions.notify_event_service.NotifyEventServiceAction"
      service = "pagerduty"
    }
  ]
}

# -------------------------------------------------------------------
# ALERT RULE 2: Immediate PagerDuty for New Issue in Critical Feature (payment, exit, auth)
# -------------------------------------------------------------------
resource "sentry_issue_alert" "rule2_critical_feature_crash" {
  for_each     = toset(local.sentry_mobile_projects)
  organization = var.sentry_organization
  project      = each.key
  name         = "P1 SEV1: New Issue in Critical Feature (payment/exit/auth)"
  action_match = "any"
  filter_match = "all"

  conditions = [
    {
      id = "sentry.rules.conditions.first_seen_event.FirstSeenEventCondition"
    }
  ]

  filters = [
    {
      id    = "sentry.rules.filters.tagged_event.TaggedEventFilter"
      key   = "feature"
      match = "is"
      value = "payment,exit,auth"
    }
  ]

  actions = [
    {
      id      = "sentry.rules.actions.notify_event_service.NotifyEventServiceAction"
      service = "pagerduty"
    }
  ]
}

# -------------------------------------------------------------------
# ALERT RULE 3: Crash-Free Session Rate Drops Below 99% (Release Health)
# -------------------------------------------------------------------
resource "sentry_metric_alert" "rule3_crash_free_rate_drop" {
  for_each     = toset(local.sentry_mobile_projects)
  organization = var.sentry_organization
  project      = each.key
  name         = "P2 SEV2: Crash-Free Session Rate < 99%"

  query        = "crash_free_sessions"
  dataset      = "sessions"
  time_window  = 1440 # 24 Hours in minutes
  aggregate    = "percentage(crash_free_sessions) * 100"

  threshold_type = 1 # Lower than threshold triggers
  resolve_threshold = 99.5

  trigger {
    label          = "Critical Crash-Free Drop"
    alert_threshold = 99.0
    actions = [
      {
        type        = "pagerduty"
        target_type = "specific"
        target_identifier = var.pagerduty_service_integration_key
      }
    ]
  }
}
