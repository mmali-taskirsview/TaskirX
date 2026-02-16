resource "oci_core_instance" "debug_node_connectivity" {
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  compartment_id      = var.compartment_id
  display_name        = "debug-node-connectivity"
  shape               = "VM.Standard.E3.Flex"

  shape_config {
    ocpus         = 1
    memory_in_gbs = 1
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.node_subnet.id
    display_name     = "primaryvnic"
    assign_public_ip = false
  }

  source_details {
    source_type = "IMAGE"
    source_id   = data.oci_core_images.ol8_images.images[0].id
  }

  metadata = {
    ssh_authorized_keys = tls_private_key.oke_node_key.public_key_openssh
  }
}

output "debug_node_private_ip_addr" {
    value = oci_core_instance.debug_node_connectivity.private_ip
}
