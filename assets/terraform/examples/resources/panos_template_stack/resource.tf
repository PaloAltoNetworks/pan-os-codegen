resource "panos_template_stack" "example" {

  location = {
    panorama = {}
  }
  name        = "tempalte-stack-example"
  description = "example template stack"

  templates = [
    panos_template.example.name
  ]

}

resource "panos_template" "example" {

  location = {
    panorama = {}
  }
  name        = "template-example"
  description = "example template stack"

}
