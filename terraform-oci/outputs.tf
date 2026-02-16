output "k8s_cluster_id" {
  value = oci_containerengine_cluster.taskir_oke.id
}

output "region" {
  value = var.region
}

output "cluster_name" {
  value = var.cluster_name
}

# output "pinecone_host" {
#   value = pinecone_index.taskir_ads.host
# }

output "vcn_id" {
  value = oci_core_vcn.taskir_vcn.id
}

output "lb_subnet_id" {
  value = oci_core_subnet.public_subnet.id
}
