package rds

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DescribeBackups invokes the rds.DescribeBackups API synchronously
// api document: https://help.aliyun.com/api/rds/describebackups.html
func (client *Client) DescribeBackups(request *DescribeBackupsRequest) (response *DescribeBackupsResponse, err error) {
	response = CreateDescribeBackupsResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeBackupsWithChan invokes the rds.DescribeBackups API asynchronously
// api document: https://help.aliyun.com/api/rds/describebackups.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeBackupsWithChan(request *DescribeBackupsRequest) (<-chan *DescribeBackupsResponse, <-chan error) {
	responseChan := make(chan *DescribeBackupsResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeBackups(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DescribeBackupsWithCallback invokes the rds.DescribeBackups API asynchronously
// api document: https://help.aliyun.com/api/rds/describebackups.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeBackupsWithCallback(request *DescribeBackupsRequest, callback func(response *DescribeBackupsResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeBackupsResponse
		var err error
		defer close(result)
		response, err = client.DescribeBackups(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DescribeBackupsRequest is the request struct for api DescribeBackups
type DescribeBackupsRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	BackupId             string           `position:"Query" name:"BackupId"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	EndTime              string           `position:"Query" name:"EndTime"`
	StartTime            string           `position:"Query" name:"StartTime"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	PageNumber           requests.Integer `position:"Query" name:"PageNumber"`
	BackupStatus         string           `position:"Query" name:"BackupStatus"`
	BackupLocation       string           `position:"Query" name:"BackupLocation"`
	PageSize             requests.Integer `position:"Query" name:"PageSize"`
	DBInstanceId         string           `position:"Query" name:"DBInstanceId"`
	BackupMode           string           `position:"Query" name:"BackupMode"`
}

// DescribeBackupsResponse is the response struct for api DescribeBackups
type DescribeBackupsResponse struct {
	*responses.BaseResponse
	RequestId        string                 `json:"RequestId" xml:"RequestId"`
	TotalRecordCount string                 `json:"TotalRecordCount" xml:"TotalRecordCount"`
	PageNumber       string                 `json:"PageNumber" xml:"PageNumber"`
	PageRecordCount  string                 `json:"PageRecordCount" xml:"PageRecordCount"`
	TotalBackupSize  int                    `json:"TotalBackupSize" xml:"TotalBackupSize"`
	Items            ItemsInDescribeBackups `json:"Items" xml:"Items"`
}

// CreateDescribeBackupsRequest creates a request to invoke DescribeBackups API
func CreateDescribeBackupsRequest() (request *DescribeBackupsRequest) {
	request = &DescribeBackupsRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Rds", "2014-08-15", "DescribeBackups", "rds", "openAPI")
	return
}

// CreateDescribeBackupsResponse creates a response to parse from DescribeBackups response
func CreateDescribeBackupsResponse() (response *DescribeBackupsResponse) {
	response = &DescribeBackupsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
