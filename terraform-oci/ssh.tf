resource "tls_private_key" "oke_node_key" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "local_file" "oke_node_key_file" {
  content  = tls_private_key.oke_node_key.private_key_pem
  filename = "${path.module}/id_rsa_oke.pem"
  file_permission = "0600"
}

output "generated_ssh_private_key" {
  value     = tls_private_key.oke_node_key.private_key_pem
  sensitive = true
}

output "generated_ssh_public_key" {
  value = tls_private_key.oke_node_key.public_key_openssh
}
