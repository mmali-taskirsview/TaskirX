# Add explicit CSI drive support for OKE
resource "oci_identity_policy" "oke_volume_policy" {
  name           = "oke-volume-policy"
  description    = "Allow OKE to manage block volumes"
  compartment_id = var.compartment_id
  
  statements = [
    "Allow serviceoke to manage volume-family in compartment id ${var.compartment_id}",
    "Allow serviceoke to use virtual-network-family in compartment id ${var.compartment_id}"
  ]
}

resource "oci_identity_policy" "oke_node_pull" {
  name           = "oke-node-pull-policy"
  description    = "Allow OKE nodes to pull images from OCIR"
  compartment_id = var.compartment_id
  
  statements = [
    "Allow any-user to use fls-family in compartment id ${var.compartment_id} where request.principal.type = 'cluster'",
    "Allow any-user to use repos in compartment id ${var.compartment_id} where request.principal.type = 'cluster'"
  ]
}
