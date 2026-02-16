// resource "cloudflare_record" "api" {
//   zone_id = var.zone_id
//   name    = "api"
//   value   = "192.0.2.1" # Placeholder, will be updated with LB IP manually or via script
//   type    = "A"
//   proxied = true
// }
//
// # --- CLOUDFLARE SSL/TLS SETTINGS ---
// # Enforce Full (Strict) SSL Mode for end-to-end encryption to OCI LB
// // resource "cloudflare_zone_settings_override" "adx_settings" {
// //   zone_id = var.zone_id
// //   settings {
// //     ssl = "strict"
// //     min_tls_version = "1.2"
// //     always_use_https = "on"
// //     brotli = "on"
// //     http3 = "on"
// //   }
// // }
//
// # --- PAGE RULES FOR CACHING STATIC CONTENT ---
// resource "cloudflare_page_rule" "cache_static" {
//   zone_id = var.zone_id
//   target  = "*.taskir.com/assets/*"
//   priority = 1
//   status  = "active"
//   actions {
//     cache_level = "cache_everything"
//     edge_cache_ttl = 86400  # 1 day
//   }
// }
