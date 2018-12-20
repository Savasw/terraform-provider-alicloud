package alicloud

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"log"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud/connectivity"
)

func resourceAlicloudDBBackup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudDBBackupCreate,
		Read:   resourceAlicloudDBBackupRead,
		Delete: resourceAlicloudDBBackupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"backup_method": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"backup_strategy": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"db_name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"backup_type": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"backup_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAlicloudDBBackupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService := RdsService{client}

	request := rds.CreateCreateBackupRequest()
	request.DBInstanceId = Trim(d.Get("instance_id").(string))
	request.BackupMethod = Trim(d.Get("backup_method").(string))
	request.BackupType = Trim(d.Get("backup_type").(string))
	request.DBName = Trim(d.Get("db_name").(string))
	request.BackupStrategy = Trim(d.Get("backup_strategy").(string))

	//Describe backups before creating new so that we are able to figure out newly created backup id
	// as create backup doesn't provide backup id in response
	getRequest := rds.CreateDescribeBackupsRequest()
	getRequest.DBInstanceId = request.DBInstanceId
	getRequest.PageSize = requests.NewInteger(PageSizeLarge)
	dbBackupsBeforeCreate, err := rdsService.DescribeDBBackups(getRequest)
	if err != nil {
		return err
	}

	_, err = client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
		return rdsClient.CreateBackup(request)
	})

	if err != nil {
		return fmt.Errorf("Error creating Alicloud db backup: %#v", err)
	}
	//resp, _ := raw.(*rds.CreateBackupResponse)

	dbBackupsAfterCreate, err := rdsService.DescribeDBBackups(getRequest)
	if err != nil {
		return err
	}

	newlyCreatedBackupIds := getRecentBackups(dbBackupsBeforeCreate, dbBackupsAfterCreate)

	if len(newlyCreatedBackupIds) != 1 {
		return fmt.Errorf("Error fetching created backup : %#v", err)
	}

	backupId := newlyCreatedBackupIds[0]

	d.SetId(backupId + ":" + request.DBInstanceId)
	//d.SetId(backupId)

	log.Printf("[INFO] Backup ID: %s", backupId)

	getRequest.BackupId = backupId

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Creating"},
		Target:     []string{"Success"},
		Refresh:    waitForBackupSuccess(client, getRequest),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for Backup (%s) to become Success: %s",
			backupId, stateErr)
	}
	/*// wait instance status change from Creating to running
	if err := rdsService.WaitForDBInstance(d.Id(), Running, DefaultLongTimeout); err != nil {
		return fmt.Errorf("WaitForInstance %s got error: %#v", Running, err)
	}*/

	return resourceAlicloudDBBackupRead(d, meta)
}

func resourceAlicloudDBBackupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	rdsService := RdsService{client}

	backupId, instanceId, _ := getBackupIdAndInstanceId(d, meta)

	getRequest := rds.CreateDescribeBackupsRequest()
	getRequest.DBInstanceId = instanceId
	getRequest.BackupId = backupId
	getRequest.PageSize = requests.NewInteger(PageSizeLarge)
	dbBackups, err := rdsService.DescribeDBBackups(getRequest)

	if err != nil {
		if rdsService.NotFoundDBInstance(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Describe DB Instance Backup: %#v", err)
	}

	dbBackup := dbBackups[0]

	d.Set("backup_method", dbBackup.BackupMethod)
	d.Set("backup_type", dbBackup.BackupType)
	d.Set("backup_id", dbBackup.BackupId)

	return nil
}

func resourceAlicloudDBBackupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService := RdsService{client}

	instance, err := rdsService.DescribeDBInstanceById(d.Id())
	if err != nil {
		if rdsService.NotFoundDBInstance(err) {
			return nil
		}
		return fmt.Errorf("Error Describe DB Instance Backup: %#v", err)
	}
	if PayType(instance.PayType) == Prepaid {
		return fmt.Errorf("At present, 'Prepaid' instance cannot be deleted and must wait it to be expired and release it automatically.")
	}

	request := rds.CreateDeleteDBInstanceRequest()
	request.DBInstanceId = d.Id()

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
			return rdsClient.DeleteDBInstance(request)
		})

		if err != nil {
			if rdsService.NotFoundDBInstance(err) {
				return nil
			}
			return resource.RetryableError(fmt.Errorf("Delete DB instance timeout and got an error: %#v.", err))
		}

		instance, err := rdsService.DescribeDBInstanceById(d.Id())
		if err != nil {
			if NotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error Describe DB InstanceAttribute: %#v", err))
		}
		if instance == nil {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Delete DB instance timeout and got an error: %#v.", err))
	})
}

func waitForBackupSuccess(client *connectivity.AliyunClient, request *rds.DescribeBackupsRequest) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		rdsService := RdsService{client}
		backups, err := rdsService.DescribeDBBackups(request)
		if err != nil {
			return nil, "", err
		}

		if len(backups) < 1 {
			return nil, "", fmt.Errorf("Not able to retrieve backup %s for status check", request.BackupId)
		}

		backup := backups[0]

		if backup.BackupStatus == "Failed" {
			return nil, "", fmt.Errorf("Backup Status: '%s'", backup.BackupStatus)

		}

		if backup.BackupStatus != "Success" {
			return backup, "Creating", nil
		}

		return backup, backup.BackupStatus, nil
	}
}

/*func describeDBBackups(client *connectivity.AliyunClient, request *rds.DescribeBackupsRequest) ([]rds.Backup, error){


	var dbBackups []rds.Backup

	for {
		raw, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
			return rdsClient.DescribeBackups(request)
		})
		if err != nil {
			return nil, fmt.Errorf("Error describing Alicloud db backups: %#v", err)
		}
		resp, _ := raw.(*rds.DescribeBackupsResponse)
		if resp == nil || len(resp.Items.Backup) < 1 {
			break
		}

		for _, item := range resp.Items.Backup {

			dbBackups = append(dbBackups, item)
		}

		if len(resp.Items.Backup) < PageSizeLarge {
			break
		}

		if page, err := getNextpageNumber(request.PageNumber); err != nil {
			return nil, err
		} else {
			request.PageNumber = page
		}
	}

	return dbBackups, nil
}*/

func getRecentBackups(oldBackups []rds.Backup, newBackups []rds.Backup) []string {
	oldBackupIdSet := make(map[string]struct{})
	var recentBackups []string
	for _, backup := range oldBackups {
		oldBackupIdSet[backup.BackupId] = struct{}{}
	}

	for _, backup := range newBackups {
		if _, ok := oldBackupIdSet[backup.BackupId]; !ok {
			recentBackups = append(recentBackups, backup.BackupId)
		}
	}

	return recentBackups

}

func getBackupIdAndInstanceId(d *schema.ResourceData, meta interface{}) (string, string, error) {
	return splitBackupIdAndInstanceId(d.Id())
}

func splitBackupIdAndInstanceId(s string) (string, string, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource id")
	}
	return parts[0], parts[1], nil
}
