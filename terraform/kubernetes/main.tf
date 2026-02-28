locals {
  name = "copypastabotv2"
  #   used_profile = data.awsprofiler_list.list_profiles.profiles[try(index(data.awsprofiler_list.list_profiles.profiles.*.name, "personal"), 0)]
}

resource "kubernetes_persistent_volume_claim_v1" "copypastabot" {
  metadata {
    name      = "copypastabot"
    namespace = data.terraform_remote_state.kubernetes_cluster.outputs.discordbots.namespace.metadata.0.name
  }
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        "storage" = "3Gi"
      }
    }
  }
}
