data "xenorchestra_pool" "pool" {
  name_label = "Your pool"
  # Or id = "pool-uuid"
}
data "xenorchestra_sr" "local_storage" {
  name_label = "Your storage repository label"
  pool_id    = data.xenorchestra_pool.pool.id
}
