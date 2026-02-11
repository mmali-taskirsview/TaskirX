# TaskirX v3 - ClickHouse Infrastructure
# Provisions ClickHouse cluster using Altinity Operator or Managed Service
# NOTE: For this initial deployment, we are deploying a StatefulSet inside EKS (managed in main.tf helm charts)
# This file reserves resources for future managed ClickHouse Cloud integration.

resource "kubernetes_persistent_volume_claim" "clickhouse_storage" {
  metadata {
    name = "clickhouse-pvc"
    namespace = "taskir"
  }
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "20Gi"
      }
    }
    storage_class_name = "gp3"
  }
  depends_on = [aws_eks_cluster.main]
}
