---
layout: "alicloud"
page_title: "Alicloud: alicloud_db_backups"
sidebar_current: "docs-alicloud-datasource-db-backups"
description: |-
    Provides a collection of RDS instance backups according to the specified filters.
---

# alicloud\_db\_backups

The `alicloud_db_backups` data source provides a collection of RDS instance backups available in Alibaba Cloud account.

## Example Usage

```
data "alicloud_db_backups" "db_backups" {
  instance_id = "rm-iert46599euyt679"
  status     = "Success"

}

output "first_db_backup_id" {
  value = "${data.alicloud_db_backups.db_backups.backups.0.id}"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) DB instance id to filter results by instance id.
* `backup_id` - (Optional) Backup set ID to filter a particular backup set.
* `backup_status` - (Optional) Status of the backup set to filter by. `Success` for successful backups, `Failed` for failed backups.
* `backup_mode` - (Optional) Type of Backup. `Automated` for regular task, `Manual` for Temporary task.
* `start_time` - (Optional) Query start time, for example, 2011-06-11T15:00Z
* `end_time` - (Optional) Query end time, which must be later than the query start time, for example, 2011-06-11T16:00Z.
* `output_file` - (Optional) File name where to save data source results (after running `terraform plan`).

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `backups` - A list of RDS instance backups. Each element contains the following attributes:
  * `id` - The ID of the RDS instance backup set.
  * `host_instance_id` - ID of the instance that generates the backup set.
  * `status` - Status of the backup set.
  * `start_time` - Start time of this backup.
  * `end_time` - Start time of this backup.
  * `backup_type` - Backup type.
  * `backup_db_names` - Backup Database Names.
  * `backup_download_url` - OSS backup download link. It is empty if the current backup cannot be downloaded.
  * `backup_intranet_download_url` - Download URL. It is empty if the current backup cannot be downloaded.
  * `consistent_time` - The consistency time point of the backup set.
  * `backup_location` - The location where the backup file is stored in the OSS.
  * `backup_method` - Backup Method.
  * `backup_mode` - Backup Mode.
  * `backup_scale` - When the backup type is Logical
  * `backup_size` - Data file size. Unit: byte.
  * `store_status` - Data backup storage status.
