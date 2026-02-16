data "oci_objectstorage_namespace" "ns" {
  compartment_id = var.compartment_id
}

data "oci_core_instances" "debug_nodes" {
  compartment_id = var.compartment_id
  # We can't easily filter by "created by node pool" without knowing the pool ID or tags exactly, but usually all instances in compartment is a good start if it's a clean compartment.
}

output "debug_node_public_ips" {
  value = [for i in data.oci_core_instances.debug_nodes.instances : { name = i.display_name, public_ip = i.public_ip }]
}
