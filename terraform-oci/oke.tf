resource "oci_containerengine_cluster" "taskir_oke" {
  compartment_id     = var.compartment_id
  kubernetes_version = "v1.31.10"
  name               = var.cluster_name
  vcn_id             = oci_core_vcn.taskir_vcn.id
  cluster_pod_network_options {
    cni_type = "FLANNEL_OVERLAY"
  }

  endpoint_config {
    is_public_ip_enabled = true
    subnet_id            = oci_core_subnet.public_subnet.id
  }
  options {
    service_lb_subnet_ids = [oci_core_subnet.public_subnet.id]
    add_ons {
      is_kubernetes_dashboard_enabled = false
      is_tiller_enabled               = false
    }
  }
}

resource "oci_containerengine_node_pool" "taskir_node_pool" {
  cluster_id         = oci_containerengine_cluster.taskir_oke.id
  compartment_id     = var.compartment_id
  kubernetes_version = "v1.31.10"
  name               = "taskir-node-pool"
  node_shape         = "VM.Standard.E3.Flex"
  
    node_shape_config {
        memory_in_gbs = 8
        ocpus         = 1
    }

    # Remove user_data to eliminate potential blocking issues during bootstrap
    # node_metadata = {
    #   user_data = ...
    # }

    ssh_public_key = tls_private_key.oke_node_key.public_key_openssh

  node_source_details {
    # capturing the whole object to debug
    image_id    = local.node_image_id
    source_type = "IMAGE"
  }

  node_config_details {
    placement_configs {
      availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
      subnet_id           = oci_core_subnet.node_subnet.id
    }
    size = 1
  }


}


data "oci_identity_availability_domains" "ads" {
  compartment_id = var.tenancy_ocid
}

data "oci_containerengine_node_pool_option" "taskir_node_pool_option" {
  node_pool_option_id = "all"
  compartment_id      = var.compartment_id
}

locals {
  # Find the image ID for version v1.31.10 match.
  # MUST exclude "aarch64" (ARM) and "GPU" images to ensure compatibility with VM.Standard.E3.Flex (AMD x86_64).
  # capturing the whole object to debug
  node_image_obj = [for x in data.oci_containerengine_node_pool_option.taskir_node_pool_option.sources : x if length(regexall(".*1\\.31\\.10.*", x.source_name)) > 0 && length(regexall(".*aarch64.*", x.source_name)) == 0 && length(regexall(".*GPU.*", x.source_name)) == 0][0]
  node_image_id = local.node_image_obj.image_id
} 

output "debug_node_image_name" {
  value = local.node_image_obj.source_name
}

# We will modify the unexpected output with something simpler since regexall might be tricky if no match found.
# Actually, let's just stick to "Oracle Linux 8" generic image for now, BUT add node_metadata user_data to bootstrap it properly?
# No, OKE node pools should handle bootstrap automatically on generic images IF they are supported.
# But "Oracle Linux 8" generic image IS supported.
# The error "2 nodes(s) register timeout" is 90% likely network.
# Let's verify network again.

# Network: 
# Nodes need to reach API Server (public endpoint).
# Nodes have Public IP?
# Let's ADD "assign_public_ip = true" explicitly to placement_configs just in case.
# But placement_configs doesn't support "assign_public_ip".
# It must be on the subnet.
# Let's check "oci_core_subnet" definition again. It doesn't enable "prohibit_public_ip_on_vnic" (default false).

# MAYBE limits on NAT? No.
# MAYBE Time sync?

# Let's try to use the OKE Optimized Image. 
# I will change the data source to search for "Oracle-Linux-8.*OKE-1.31.1"

