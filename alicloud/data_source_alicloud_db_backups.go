package alicloud

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud/connectivity"
)

func dataSourceAlicloudDBBackups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudDBBackupRead,

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"backup_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"backup_status": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateAllowedStringValue([]string{"Success", "Failed"}),
			},
			"backup_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateAllowedStringValue([]string{"Automated", "Manual"}),
			},
			"start_time": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"end_time": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			// Computed values
			"backups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"start_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"end_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_db_names": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_download_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_intranet_download_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"consistent_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_location": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_scale": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"backup_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"store_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudDBBackupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService := RdsService{client}

	args := rds.CreateDescribeBackupsRequest()
	args.RegionId = client.RegionId
	args.DBInstanceId = d.Get("instance_id").(string)
	args.BackupId = d.Get("backup_id").(string)
	args.BackupMode = d.Get("backup_mode").(string)
	args.BackupStatus = d.Get("backup_status").(string)
	args.StartTime = d.Get("start_time").(string)
	args.EndTime = d.Get("end_time").(string)
	args.PageSize = requests.NewInteger(PageSizeLarge)

	backups, err := rdsService.DescribeDBBackups(args)
	if err != nil {
		return err
	}

	return rdsBackupDescription(d, backups)
}

func rdsBackupDescription(d *schema.ResourceData, dbBackups []rds.Backup) error {
	var ids []string
	var s []map[string]interface{}

	for _, item := range dbBackups {
		mapping := map[string]interface{}{
			"id":                           item.BackupId,
			"host_instance_id":             item.HostInstanceID,
			"status":                       item.BackupStatus,
			"start_time":                   item.BackupStartTime,
			"end_time":                     item.BackupEndTime,
			"backup_type":                  item.BackupType,
			"backup_db_names":              item.BackupDBNames,
			"backup_download_url":          item.BackupDownloadURL,
			"backup_intranet_download_url": item.BackupIntranetDownloadURL,
			"consistent_time":              item.ConsistentTime,
			"backup_location":              item.BackupLocation,
			"backup_method":                item.BackupMethod,
			"backup_mode":                  item.BackupMode,
			"backup_scale":                 item.BackupScale,
			"backup_size":                  item.BackupSize,
			"store_status":                 item.StoreStatus,
		}

		ids = append(ids, item.BackupId)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("backups", s); err != nil {
		return err
	}

	// create a json file in current directory and write data source to it
	if output, ok := d.GetOk("output_file"); ok && output.(string) != "" {
		writeToFile(output.(string), s)
	}
	return nil
}
