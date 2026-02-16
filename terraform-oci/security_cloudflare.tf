resource "oci_core_network_security_group" "cloudflare_only" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.taskir_vcn.id
  display_name   = "Cloudflare-Only-Ingress"
}

data "cloudflare_ip_ranges" "cloudflare" {}

resource "oci_core_network_security_group_security_rule" "allow_cloudflare_https" {
  for_each                  = toset(data.cloudflare_ip_ranges.cloudflare.ipv4_cidr_blocks)
  network_security_group_id = oci_core_network_security_group.cloudflare_only.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source_type               = "CIDR_BLOCK"
  source                    = each.value
  tcp_options {
    destination_port_range {
      max = 443
      min = 443
    }
  }
}

resource "oci_core_network_security_group_security_rule" "allow_cloudflare_http" {
  for_each                  = toset(data.cloudflare_ip_ranges.cloudflare.ipv4_cidr_blocks)
  network_security_group_id = oci_core_network_security_group.cloudflare_only.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source_type               = "CIDR_BLOCK"
  source                    = each.value
  tcp_options {
    destination_port_range {
      max = 80
      min = 80
    }
  }
}