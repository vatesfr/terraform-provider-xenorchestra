# Get all pools
data "xenorchestra_pools" "all" {
}

# Get pools with specific tags
data "xenorchestra_pools" "tagged" {
  tags = ["production", "web"]
}

# Get all pools sorted by name
data "xenorchestra_pools" "sorted" {
  sort_by = "name_label"
  sort_order = "asc"
}

# Use the pools data source to get pool IDs
data "xenorchestra_pools" "example" {
  tags = ["terraform-managed"]
}

data "xenorchestra_sr" "local_storage" {
  name_label = "Local storage"
  pool_id = data.xenorchestra_pools.example.pools[0].id
}
