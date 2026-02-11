# Configured manually or via output from OCI Load Balancer
resource "cloudflare_record" "api" {
  zone_id = var.zone_id
  name    = "api"
  value   = "192.0.2.1" # Placeholder: Update via Dashboard or Terraform var after K8s LoadBalancer provisioning
  type    = "A"
  proxied = true
  lifecycle {
    ignore_changes = [value]
  }
}

resource "cloudflare_record" "dashboard" {
  zone_id = var.zone_id
  name    = "dashboard"
  value   = "192.0.2.1" # Placeholder
  type    = "A"
  proxied = true
  lifecycle {
    ignore_changes = [value]
  }
}

resource "cloudflare_record" "bidding" {
  zone_id = var.zone_id
  name    = "bidding"
  value   = "192.0.2.1" # Placeholder
  type    = "A"
  proxied = true
  lifecycle {
    ignore_changes = [value]
  }
}
