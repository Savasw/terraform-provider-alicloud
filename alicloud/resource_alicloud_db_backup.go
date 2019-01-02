package alicloud

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"

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
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
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
	log.Printf("[DEBUG] old backups: %#v", len(dbBackupsBeforeCreate))

	raw, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
		return rdsClient.CreateBackup(request)
	})

	if err != nil {
		return fmt.Errorf("Error creating Alicloud db backup: %#v", err)
	}

	backupJobId := raw.(*rds.CreateBackupResponse).BackupJobId

	//Get Backup Task Status
	describeTaskReq := rds.CreateDescribeBackupTasksRequest()
	describeTaskReq.DBInstanceId = request.DBInstanceId
	describeTaskReq.BackupJobId = backupJobId

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"NoStart", "Preparing", "Waiting", "Uploading", "Checking"},
		Target:     []string{"Finished"},
		Refresh:    waitForBackupFinish(client, describeTaskReq),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for Backup task to finish: %s", stateErr)
	}

	var dbBackupsAfterCreate []rds.Backup

	//as it takes sometime for backup to appear in describe after finished
	for {
		dbBackupsAfterCreate, err = rdsService.DescribeDBBackups(getRequest)
		if err != nil {
			return err
		}

		if len(dbBackupsAfterCreate) > len(dbBackupsBeforeCreate) {
			break
		}
	}

	log.Printf("[DEBUG] new backups: %#v", len(dbBackupsAfterCreate))

	newlyCreatedBackups := getRecentBackups(dbBackupsBeforeCreate, dbBackupsAfterCreate)

	log.Printf("[DEBUG] recent backups: %#v", newlyCreatedBackups)

	if len(newlyCreatedBackups) != 1 {
		return fmt.Errorf("Error retreiving newly created backup")
	}

	backup := newlyCreatedBackups[0]

	d.SetId(backup.BackupId + COLON_SEPARATED + request.DBInstanceId)
	//d.SetId(backupId)

	log.Printf("[INFO] Created Backup : %#v", backup)

	if backup.BackupStatus != "Success" {
		return fmt.Errorf(
			"Error in Created Backup Status : %s", backup.BackupStatus)
	}

	return resourceAlicloudDBBackupRead(d, meta)
}

func resourceAlicloudDBBackupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	rdsService := RdsService{client}

	backupId, instanceId, _ := getBackupIdAndInstanceId(d, meta)

	getRequest := rds.CreateDescribeBackupsRequest()
	getRequest.DBInstanceId = instanceId
	//getRequest.BackupId = backupId
	getRequest.PageSize = requests.NewInteger(PageSizeLarge)
	dbBackups, err := rdsService.DescribeDBBackups(getRequest)

	if err != nil {
		if rdsService.NotFoundDBInstance(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Describe DB Instance Backup: %#v", err)
	}

	dbBackup, err := getBackupFromList(dbBackups, backupId)

	if err != nil {
		d.SetId("")
		return nil
	}

	d.Set("backup_method", dbBackup.BackupMethod)
	d.Set("backup_type", dbBackup.BackupType)
	d.Set("backup_id", dbBackup.BackupId)

	return nil
}

func resourceAlicloudDBBackupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	rdsService := RdsService{client}

	backupId, instanceId, _ := getBackupIdAndInstanceId(d, meta)

	getRequest := rds.CreateDescribeBackupsRequest()
	getRequest.DBInstanceId = instanceId
	//getRequest.BackupId = backupId
	getRequest.PageSize = requests.NewInteger(PageSizeLarge)
	dbBackups, err := rdsService.DescribeDBBackups(getRequest)

	if err != nil {
		//check if instance is not in existence
		if rdsService.NotFoundDBInstance(err) {
			return nil
		}
		return fmt.Errorf("Error Describe DB Instance Backup: %#v", err)
	}

	dbBackup, err := getBackupFromList(dbBackups, backupId)

	//check if backup is not in existence
	if err != nil {
		return nil
	}

	request := rds.CreateDeleteBackupRequest()
	request.DBInstanceId = instanceId
	request.BackupId = dbBackup.BackupId

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
			return rdsClient.DeleteBackup(request)
		})

		if err != nil {
			if rdsService.NotFoundDBBackup(err) {
				return nil
			}
			return resource.RetryableError(fmt.Errorf("Delete DB instance backup timeout and got an error: %#v.", err))
		}
		dbBackups, err := rdsService.DescribeDBBackups(getRequest)

		if err != nil {
			//check if instance is not in existence
			if rdsService.NotFoundDBInstance(err) {
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error Describe DB Instance Backup: %#v", err))
		}

		_, err = getBackupFromList(dbBackups, backupId)

		//check if backup is not in existence
		if err != nil {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("Delete DB instance backup timeout and got an error: %#v.", err))
	})
}

func waitForBackupFinish(client *connectivity.AliyunClient, request *rds.DescribeBackupTasksRequest) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		raw, err := client.WithRdsClient(func(rdsClient *rds.Client) (interface{}, error) {
			return rdsClient.DescribeBackupTasks(request)
		})
		if err != nil {
			return nil, "", err
		}

		backupJob := raw.(*rds.DescribeBackupTasksResponse).Items.BackupJob[0]
		log.Printf("Backup Job Status: %s", backupJob.BackupStatus)
		return backupJob, backupJob.BackupStatus, nil
	}
}

func getBackupFromList(backupList []rds.Backup, backupId string) (rds.Backup, error) {
	empty := rds.Backup{}

	for _, backup := range backupList {
		if backupId == backup.BackupId {
			return backup, nil
		}
	}

	return empty, fmt.Errorf("DB Instance Backup %s not found", backupId)
}

func getRecentBackups(oldBackups []rds.Backup, newBackups []rds.Backup) []rds.Backup {
	oldBackupIdSet := make(map[string]struct{})
	var recentBackups []rds.Backup
	for _, backup := range oldBackups {
		oldBackupIdSet[backup.BackupId] = struct{}{}
	}

	for _, backup := range newBackups {
		if _, ok := oldBackupIdSet[backup.BackupId]; !ok {
			recentBackups = append(recentBackups, backup)
		}
	}

	return recentBackups

}

func getBackupIdAndInstanceId(d *schema.ResourceData, meta interface{}) (string, string, error) {
	return splitBackupIdAndInstanceId(d.Id())
}

func splitBackupIdAndInstanceId(s string) (string, string, error) {
	parts := strings.Split(s, COLON_SEPARATED)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource id")
	}
	return parts[0], parts[1], nil
}
