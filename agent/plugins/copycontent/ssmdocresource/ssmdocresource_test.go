// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the
// License is located at
//
// http://aws.amazon.com/apache2.0/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package ssmdocresource

import (
	"github.com/aws/amazon-ssm-agent/agent/appconfig"
	filemock "github.com/aws/amazon-ssm-agent/agent/fileutil/filemanager/mock"
	"github.com/aws/amazon-ssm-agent/agent/log"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

var logMock = log.NewMockLog()

func TestSSMDocResource_ValidateLocationInfo(t *testing.T) {

	locationInfo := `{
		"name": "AWS-ExecuteCommand",
		"version": ""
	}`

	ssmresource, _ := NewSSMDocResource(locationInfo)
	_, err := ssmresource.ValidateLocationInfo()

	assert.NoError(t, err)
}

func TestSSMDocResource_FullARNNameInput(t *testing.T) {
	depMock := new(ssmDocDepMock)
	fileMock := filemock.FileSystemMock{}

	locationInfo := `{
		"name": "arn:aws:ssm:us-east-1:1234567890:document/mySharedDocument",
		"version": ""
	}`

	content := "content"
	docOutput := ssm.GetDocumentOutput{
		Content: &content,
	}
	ssmresource, _ := NewSSMDocResource(locationInfo)
	dir := "destination"
	depMock.On("GetDocument", logMock, ssmresource.Info.DocName, ssmresource.Info.DocVersion).Return(&docOutput, nil)
	fileMock.On("MakeDirs", dir).Return(nil)
	fileMock.On("WriteFile", filepath.Join(dir, "mySharedDocument.json"), content).Return(nil)

	ssmdocdep = depMock

	err := ssmresource.Download(logMock, fileMock, "destination")

	assert.NoError(t, err)
	depMock.AssertExpectations(t)
	fileMock.AssertExpectations(t)
}

func TestSSMDocResource_ValidateLocationInfoNoName(t *testing.T) {

	locationInfo := `{
		"name": "",
		"version": "10"
	}`

	ssmresource, _ := NewSSMDocResource(locationInfo)
	_, err := ssmresource.ValidateLocationInfo()

	assert.Error(t, err)
	assert.Equal(t, "SSM Document name in LocationType must be specified", err.Error())
}

func TestSSMDocResource_Download(t *testing.T) {
	depMock := new(ssmDocDepMock)
	fileMock := filemock.FileSystemMock{}

	locationInfo := `{
		"name": "AWS-ExecuteCommand",
		"version": "10"
	}`
	content := "content"
	docOutput := ssm.GetDocumentOutput{
		Content: &content,
	}
	ssmresource, _ := NewSSMDocResource(locationInfo)
	dir := "destination"
	depMock.On("GetDocument", logMock, ssmresource.Info.DocName, ssmresource.Info.DocVersion).Return(&docOutput, nil)
	fileMock.On("MakeDirs", dir).Return(nil)
	fileMock.On("WriteFile", filepath.Join(dir, ssmresource.Info.DocName+".json"), content).Return(nil)

	ssmdocdep = depMock

	err := ssmresource.Download(logMock, fileMock, "destination")

	assert.NoError(t, err)
	depMock.AssertExpectations(t)
	fileMock.AssertExpectations(t)
}

func TestSSMDocResource_DownloadNoDestination(t *testing.T) {
	depMock := new(ssmDocDepMock)
	fileMock := filemock.FileSystemMock{}

	locationInfo := `{
 		"name": "AWS-ExecuteCommand",
 		"version": "10"
 	}`
	content := "content"
	docOutput := ssm.GetDocumentOutput{
		Content: &content,
	}
	ssmresource, _ := NewSSMDocResource(locationInfo)
	dir := appconfig.DownloadRoot
	depMock.On("GetDocument", logMock, ssmresource.Info.DocName, ssmresource.Info.DocVersion).Return(&docOutput, nil)
	fileMock.On("MakeDirs", strings.TrimSuffix(dir, "/")).Return(nil)
	fileMock.On("WriteFile", filepath.Join(dir, ssmresource.Info.DocName+".json"), content).Return(fmt.Errorf("Error"))

	ssmdocdep = depMock

	err := ssmresource.Download(logMock, fileMock, "")

	assert.Error(t, err, "Error")
	depMock.AssertExpectations(t)
	fileMock.AssertExpectations(t)
}

type ssmDocDepMock struct {
	mock.Mock
}

func (s ssmDocDepMock) GetDocument(log log.T, docName string, docVersion string) (response *ssm.GetDocumentOutput, err error) {
	args := s.Called(log, docName, docVersion)
	return args.Get(0).(*ssm.GetDocumentOutput), args.Error(1)
}