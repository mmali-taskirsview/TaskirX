resource "oci_artifacts_container_repository" "taskir_repos" {
  for_each       = toset(["taskir-nestjs", "taskir-go-bidding", "taskir-dashboard", "taskir-ad-matching", "taskir-fraud-detection", "taskir-bid-optimization"])
  compartment_id = var.compartment_id
  display_name   = each.key
  is_public      = true # Usually false for production, but using true for easier initial testing if permissible
  readme {
    content = "TaskirX Microservices Repository: ${each.key}"
    format  = "TEXT_MARKDOWN"
  }
}

output "registry_namespace" {
  value = data.oci_objectstorage_namespace.ns.namespace
}
