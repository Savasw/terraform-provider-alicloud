---
layout: "alicloud"
page_title: "Alicloud: alicloud_db_clone"
sidebar_current: "docs-alicloud-resource-db-clone"
description: |-
  Provides an RDS instance clone resource.
---

# alicloud\_db\_clone

Provides an RDS instance clone resource. A DB instance is an isolated database
environment in the cloud. A DB instance can contain multiple user-created
databases.

## Example Usage

```
resource "alicloud_db_clone" "default" {
	instance_id = "rm-iert46599euyt679"
	backup_id = "789965432ty678uuhh"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) The DB Instance ID.
* `backup_id` - (Optional) The Backup set ID. At least one of `backup_id` and `restore_time` is required.
* `restore_time` - (Optional) The user can specify any point in the backup retention period, such as 2011-06-11T16:00:00Z. At least one of `backup_id` and `restore_time` is required.
* `instance_type` - (Optional) DB Instance type. For details, see [Instance type table](https://www.alibabacloud.com/help/doc-detail/26312.htm).
* `instance_storage` - (Required) User-defined DB instance storage space. Value range:
    - [5, 2000] for MySQL HA dual node edition;
    - [20,1000] for MySQL 5.7 basic single node edition;
    Increase progressively at a rate of 5 GB. For details, see [Instance type table](https://www.alibabacloud.com/help/doc-detail/26312.htm).
    
* `instance_charge_type` - (Optional) Valid values are `Prepaid`, `Postpaid`, Default to `Postpaid`.
* `period` - (Optional) The duration that you will buy DB instance (in month). It is valid when instance_charge_type is `PrePaid`. Valid values: [1~9], 12, 24, 36. Default to 1.
If it is a multi-zone and `vswitch_id` is specified, the vswitch must in the one of them.
The multiple zone ID can be retrieved by setting `multi` to "true" in the data source `alicloud_zones`.
* `vswitch_id` - (Optional) The virtual switch ID to launch DB instances in one VPC.

~> **NOTE:** Because of data backup and migration, change DB instance type and storage would cost 15~20 minutes. Please make full preparation before changing them.

## Attributes Reference

The following attributes are exported:

* `id` - The Cloned RDS instance ID.
* `instance_charge_type` - The instance charge type.
* `period` - The DB instance using duration.
* `engine` - Database type.
* `engine_version` - The database engine version.
* `instance_type` - The RDS instance type.
* `instance_storage` - The RDS instance storage space.
* `instance_name` - The name of DB instance.
* `port` - RDS database connection port.
* `connection_string` - RDS database connection string.
* `zone_id` - The zone ID of the RDS instance.
* `security_ips` - Security ips of instance whitelist.
* `vswitch_id` - If the rds instance created in VPC, then this value is virtual switch ID.