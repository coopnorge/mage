resource "azurerm_managed_disk" "vm" {
  # tfsec:ignore:AZU003
  count                = length(var.data_disk_names)
  name                 = "${var.vm_name}-${var.data_disk_names[count.index % length(var.data_disk_names)]}"
  location             = var.location
  resource_group_name  = var.resource_group_name
  storage_account_type = var.data_disk_storage_account_type[count.index % length(var.data_disk_storage_account_type)]
  create_option        = "Empty"
  disk_size_gb         = var.data_disk_sizes[count.index % length(var.data_disk_sizes)]
}
resource "azurerm_virtual_machine_data_disk_attachment" "vm" {
  count              = length(var.data_disk_names)
  managed_disk_id    = azurerm_managed_disk.vm[count.index % length(azurerm_managed_disk.vm)].id
  virtual_machine_id = var.virtual_machine_id
  lun                = count.index + 10
  caching            = "ReadWrite"
}  