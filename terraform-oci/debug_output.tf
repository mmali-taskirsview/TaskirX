output "node_pool_options_sources" {
  value = [for x in data.oci_containerengine_node_pool_option.taskir_node_pool_option.sources : x.source_name if length(regexall(".*1\\.31\\.10.*", x.source_name)) > 0]
}
