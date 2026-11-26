variable "review_password" {
  type      = string
  sensitive = true
  ephemeral = true
}

# Supply an ephemeral secret from your credential manager. Increment the version
# to rotate it. The role must already exist (or reference a panos_admin_role).
resource "panos_administrator" "policy_review" {
  location            = { panorama = {} }
  name                = "network-policy-review"
  password_wo         = var.review_password
  password_wo_version = 1
  permissions = {
    role_based = {
      custom = { profile = "NetworkPolicyReview" }
    }
  }
}
