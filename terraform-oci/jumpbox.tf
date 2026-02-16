# Latest Oracle Linux 8 Image
data "oci_core_images" "ol8_images" {
  compartment_id           = var.compartment_id
  operating_system         = "Oracle Linux"
  operating_system_version = "8"
  shape                    = "VM.Standard.E3.Flex"
  sort_by                  = "TIMECREATED"
  sort_order               = "DESC"
}

resource "oci_core_instance" "jumpbox" {
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  compartment_id      = var.compartment_id
  display_name        = "taskir-jumpbox"
  shape               = "VM.Standard.E3.Flex"

  shape_config {
    ocpus         = 1
    memory_in_gbs = 1
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.public_subnet.id
    display_name     = "primaryvnic"
    assign_public_ip = true
  }

  source_details {
    source_type = "IMAGE"
    source_id   = data.oci_core_images.ol8_images.images[0].id
  }

  metadata = {
    ssh_authorized_keys = tls_private_key.oke_node_key.public_key_openssh
  }
}


resource "oci_core_instance" "debug_node" {
  compartment_id      = var.compartment_id
  display_name        = "debug-node-instance"
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  shape               = "VM.Standard.E3.Flex"

  shape_config {
    ocpus         = 1
    memory_in_gbs = 4
  }

  create_vnic_details {
    subnet_id        = oci_core_subnet.node_subnet.id
    assign_public_ip = false
  }

  source_details {
    source_type = "image"
    source_id   = data.oci_core_images.ol8_images.images[0].id
  }

  metadata = {
    ssh_authorized_keys = tls_private_key.oke_node_key.public_key_openssh
    user_data           = base64encode(<<-EOF
      #cloud-config
      runcmd:
        - echo "Starting debug checks" > /tmp/debug.log
        - curl -v http://google.com >> /tmp/debug.log 2>&1
        - curl -v https://objectstorage.ap-singapore-1.oraclecloud.com >> /tmp/debug.log 2>&1
    EOF
    )
  }
}


output "jumpbox_public_ip" {
  value = oci_core_instance.jumpbox.public_ip
}

output "debug_node_private_ips" {
  value = [for i in data.oci_core_instances.debug_nodes.instances : { name = i.display_name, private_ip = i.private_ip }]
}

// output "debug_node_private_ip" {
//   value = oci_core_instance.debug_node.private_ip
// }
