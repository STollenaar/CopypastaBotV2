variable "functions" {
  type = map(object({
    description                    = string
    handler                        = string
    runtime                        = string
    timeout                        = number
    memory_size                    = number
    image_uri                      = optional(string, null)
    layers                         = optional(list(string), [])
    override_zip_location          = optional(string, null)
    reserved_concurrent_executions = optional(number, -1)
    tags                           = optional(map(string), null)
    environment_variables          = optional(map(string), null)
    enable_alarm                   = optional(bool, false)
    buildArgs                      = optional(string, "")
    extra_permissions              = optional(list(string), [])
  }))
  description = "The functions, mapped by the folder name under lambda and the description"
}

variable "project" {}
