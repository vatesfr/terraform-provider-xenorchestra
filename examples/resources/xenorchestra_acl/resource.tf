data "xenorchestra_pool" "pool" {
  name_label = "Your pool"
}

data "xenorchestra_user" "user" {
  username = "my-username"
}

resource "xenorchestra_acl" "acl" {
  subject = data.xenorchestra_user.user.id
  object = data.xenorchestra_pool.pool.id
  action = "operator"
}
