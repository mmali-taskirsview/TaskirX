# Add explicit CSI drive support for OKE
resource "oci_identity_policy" "oke_volume_policy" {
  name           = "oke-volume-policy"
  description    = "Allow OKE to manage block volumes"
  compartment_id = var.compartment_id
  
  statements = [
    "Allow service OKE to manage volume-family in compartment id ${var.compartment_id}",
    "Allow service OKE to use virtual-network-family in compartment id ${var.compartment_id}"
  ]
}

resource "oci_identity_policy" "oke_node_pull" {
  name           = "oke-node-pull-policy"
  description    = "Allow OKE nodes to pull images from OCIR"
  compartment_id = var.compartment_id
  
  statements = [
    # Simplified for now to unblock deployment
    "Allow any-user to read repos in compartment id ${var.compartment_id}"
  ]
}
