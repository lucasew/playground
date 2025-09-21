variable "trigger_error" {
  default = true
  type = bool
}

output "teste" {
  value = <<EOF
    Teste eoq trabson
    Teste eoq trabson
    Teste eoq trabson
    Teste eoq trabson
  EOF

  precondition {
    condition = var.trigger_error
    error_message = "teste"
  }
}
