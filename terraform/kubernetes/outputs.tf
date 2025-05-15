output "namespace" {
  value = kubernetes_namespace.copypastabot
}

output "external_secret" {
  value = kubernetes_manifest.external_secret.manifest
}