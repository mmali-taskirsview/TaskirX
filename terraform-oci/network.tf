resource "oci_core_vcn" "taskir_vcn" {
  cidr_block     = "10.0.0.0/16"
  compartment_id = var.compartment_id
  display_name   = "taskir-vcn"
  dns_label      = "taskirvcn"
}

resource "oci_core_internet_gateway" "ig" {
  compartment_id = var.compartment_id
  display_name   = "internet-gateway"
  vcn_id         = oci_core_vcn.taskir_vcn.id
  enabled        = true
}



resource "oci_core_subnet" "public_subnet" {
  cidr_block        = "10.0.1.0/24"
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.taskir_vcn.id
  display_name      = "public-subnet-lb"
  route_table_id    = oci_core_route_table.public_rt.id
  security_list_ids = [oci_core_security_list.public_sl.id]
}

resource "oci_core_subnet" "node_subnet" {
  cidr_block        = "10.0.2.0/24"
  compartment_id    = var.compartment_id
  vcn_id            = oci_core_vcn.taskir_vcn.id
  display_name      = "node-subnet"
  route_table_id    = oci_core_route_table.node_rt.id 
  security_list_ids = [oci_core_security_list.node_sl.id]
}

resource "oci_core_security_list" "public_sl" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.taskir_vcn.id
  display_name   = "public-security-list"

  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 80
      max = 80
    }
  }
  
  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 443
      max = 443
    }
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    description = "Kubernetes API Endpoint"
    tcp_options {
      min = 6443
      max = 6443
    }
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "10.0.2.0/24"
    description = "Allow Worker Nodes to talk to API Endpoint"
    tcp_options {
      min = 6443
      max = 6443
    }
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "10.0.2.0/24"
    description = "Allow Worker Nodes to talk to API Endpoint (12250)"
    tcp_options {
      min = 12250
      max = 12250
    }
  }
  
  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "0.0.0.0/0"
    description = "Allow SSH to Jumpbox"
    tcp_options {
      min = 22
      max = 22
    }
  }

  ingress_security_rules {
    protocol    = "1" # ICMP
    source      = "0.0.0.0/0"
    icmp_options {
      type = 3
      code = 4
    }
  }

  ingress_security_rules {
    protocol    = "1" # ICMP
    source      = "0.0.0.0/0"
    icmp_options {
      type = 8 # Echo Request
    }
  }
}

resource "oci_core_security_list" "node_sl" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.taskir_vcn.id
  display_name   = "node-security-list"

  egress_security_rules {
    destination = "0.0.0.0/0"
    protocol    = "all"
  }
  
  ingress_security_rules {
    protocol = "1" # ICMP
    source   = "0.0.0.0/0"
    icmp_options {
      type = 3
      code = 4
    }
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"
    tcp_options {
      min = 22
      max = 22
    }
  }

  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "10.0.0.0/16" 
    description = "Allow all internal VCN traffic (including Jumpbox to Node)"
    tcp_options {
      min = 22
      max = 22
    }
  }

  ingress_security_rules {
    protocol = "all"
    source   = "10.0.0.0/16"
  }

  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "0.0.0.0/0"
    description = "Allow Kubelet API & Health Check from Control Plane (Temporary)"
    tcp_options {
      min = 10250
      max = 10256
    }
  }

  ingress_security_rules {
    protocol    = "6" # TCP
    source      = "10.0.0.0/16"
  }

  lifecycle {
    ignore_changes = [
      ingress_security_rules,
      egress_security_rules
    ]
  }
}

resource "oci_core_service_gateway" "sg" {
  compartment_id = var.compartment_id
  display_name   = "service-gateway"
  vcn_id         = oci_core_vcn.taskir_vcn.id
  services {
    service_id = data.oci_core_services.all_services.services[0].id
  }
}

data "oci_core_services" "all_services" {
  filter {
    name   = "name"
    values = ["All .* Services In Oracle Services Network"]
    regex  = true
  }
}

// Update route table to include SGW
resource "oci_core_route_table" "public_rt" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.taskir_vcn.id
  display_name   = "public-route-table"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.ig.id
  }
}

data "oci_core_services" "object_storage_service" {
  filter {
    name   = "name"
    values = ["OCI .* Object Storage"]
    regex  = true
  }
}

resource "oci_core_route_table" "node_rt" {
  compartment_id = var.compartment_id
  vcn_id         = oci_core_vcn.taskir_vcn.id
  display_name   = "node-route-table"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_nat_gateway.nzw.id
  }

  route_rules {
    destination       = data.oci_core_services.all_services.services[0].cidr_block
    destination_type  = "SERVICE_CIDR_BLOCK"
    network_entity_id = oci_core_service_gateway.sg.id
  }
}

resource "oci_core_nat_gateway" "nzw" {
  compartment_id = var.compartment_id
  display_name   = "nat-gateway"
  vcn_id         = oci_core_vcn.taskir_vcn.id
}
