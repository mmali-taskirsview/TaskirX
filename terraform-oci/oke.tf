resource "oci_containerengine_cluster" "taskir_oke" {
  compartment_id     = var.compartment_id
  kubernetes_version = "v1.28.2"
  name               = var.cluster_name
  vcn_id             = oci_core_vcn.taskir_vcn.id
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
  kubernetes_version = "v1.28.2"
  name               = "taskir-node-pool"
  node_shape         = "VM.Standard.E4.Flex"
  
  node_shape_config {
    memory_in_gbs = 16
    ocpus         = 2
  }

  node_config_details {
    placement_configs {
      availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
      subnet_id           = oci_core_subnet.node_subnet.id
    }
    size = 3
  }
}

data "oci_identity_availability_domains" "ads" {
  compartment_id = var.tenancy_ocid
}
