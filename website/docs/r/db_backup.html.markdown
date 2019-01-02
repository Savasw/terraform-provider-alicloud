---
layout: "alicloud"
page_title: "Alicloud: alicloud_db_backup"
sidebar_current: "docs-alicloud-resource-db-backup"
description: |-
  Provides an RDS backup resource.
---

# alicloud\_db\_backup

Provides an RDS backup resource to create a backup set for an instance,
and an instance can create no more than 10 backup sets in a day

## Example Usage

```
resource "alicloud_db_backup" "default" {
	instance_id = "rm-iert46599euyt679"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) The DB Instance ID.
* `backup_method` - (Optional) Backup type. For details, see [CreateBackup](https://www.alibabacloud.com/help/doc-detail/26272.htm).
* `backup_strategy` - (Optional) Range of logical backup (db or instance). This parameter is valid only when value of `backup_method` is Logical.
* `db_name` - (Optional) Name of the database for single-database logical backup. This parameter is valid only when `backup_method` is Logical and `backup_strategy` is db.
* `backup_type` - (Optional) Auto or FullBackup. For details, see [CreateBackup](https://www.alibabacloud.com/help/doc-detail/26272.htm).

## Attributes Reference

The following attributes are exported:

* `backup_id` - The RDS backup ID.

## Import

RDS instance can be imported using the id, e.g.

```
$ terraform import alicloud_db_backup.example backup_id:db_instance_id
```