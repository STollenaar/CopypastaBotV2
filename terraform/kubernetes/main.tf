locals {
  name = "copypastabotv2"
  #   used_profile = data.awsprofiler_list.list_profiles.profiles[try(index(data.awsprofiler_list.list_profiles.profiles.*.name, "personal"), 0)]
}

resource "kubernetes_namespace" "copypastabot" {
  metadata {
    name = local.name
  }
}
