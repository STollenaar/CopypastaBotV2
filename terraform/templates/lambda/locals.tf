locals {
  files = {
    for key, v in var.functions : key => {
      files    = fileset("${path.root}/../cmd/${key}", "*")
      zips     = [for f in fileset("${path.root}/../cmd/${key}", "*") : f if endswith(f, ".zip")]
      go_files = [for f in fileset("${path.root}/../cmd/${key}", "*") : f if endswith(f, ".go") || endswith(f, ".mod") || endswith(f, ".sum")]
    }
  }

  default_layers = ["arn:aws:lambda:ca-central-1:580247275435:layer:LambdaInsightsExtension:37"]
}
