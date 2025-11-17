// Example 1: Device-level proxy settings. This gets applied directly to the
// connected device.
resource "panos_proxy_settings" "panorama_settings" {
  location = {
    system = {}
  }

  lcaas_use_proxy       = true
  secure_proxy_server   = "proxy.example.com"
  secure_proxy_port     = 8080
  secure_proxy_user     = "proxy-user"
  secure_proxy_password = "proxy-password"
}

// Example 2: Proxy settings on a specific NGFW device managed by Panorama.
// The device "some-ngfw-serial" should be a managed device on Panorama.
resource "panos_template" "ngfw_template" {
  location = { panorama = {} }
  name     = "ngfw-proxy-template"
}

resource "panos_proxy_settings" "ngfw_on_panorama_settings" {
  depends_on = [panos_template.ngfw_template]
  location = {
    template = {
      name = panos_template.ngfw_template.name
    }
  }

  lcaas_use_proxy       = true
  secure_proxy_server   = "proxy.example.com"
  secure_proxy_port     = 8080
  secure_proxy_user     = "proxy-user"
  secure_proxy_password = "proxy-password"
}

// Example 3: Proxy settings in a template.
resource "panos_template" "template1" {
  location = { panorama = {} }
  name     = "my-proxy-template"
}

resource "panos_proxy_settings" "template_settings" {
  depends_on = [panos_template.template1]
  location = {
    template = {
      name = panos_template.template1.name
    }
  }

  lcaas_use_proxy       = true
  secure_proxy_server   = "proxy.example.com"
  secure_proxy_port     = 8080
  secure_proxy_user     = "proxy-user"
  secure_proxy_password = "proxy-password"
}

// Example 4: Proxy settings in a template stack.
resource "panos_template_stack" "stack1" {
  location = { panorama = {} }
  name     = "my-proxy-stack"
}

resource "panos_proxy_settings" "stack_settings" {
  depends_on = [panos_template_stack.stack1]
  location = {
    template_stack = {
      name = panos_template_stack.stack1.name
    }
  }

  lcaas_use_proxy       = true
  secure_proxy_server   = "proxy.example.com"
  secure_proxy_port     = 8080
  secure_proxy_user     = "proxy-user"
  secure_proxy_password = "proxy-password"
}
