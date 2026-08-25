# Get all storage repositories
data "xenorchestra_srs" "all" {
}

# Get storage repositories of a specific pool
data "xenorchestra_srs" "pool_srs" {
  pool_id = "your-pool-id"
}

# Get the host-local storage repository of a single host
data "xenorchestra_srs" "host_local" {
  name_label = "Local storage"
  host_id    = "your-host-id"
}

# Get storage repositories by name and tags, sorted by size
data "xenorchestra_srs" "sorted" {
  name_label = "Local storage"
  tags       = ["terraform", "test"]
  sort_by    = "size"
  sort_order = "desc"
}

# Get the storage repositories of a specific type in a pool
data "xenorchestra_srs" "iso_srs" {
  pool_id = "your-pool-id"
  sr_type = "iso"
}
