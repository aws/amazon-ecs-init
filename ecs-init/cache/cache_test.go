// Copyright 2014-2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package cache

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-init/ecs-init/config"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	remoteTarballKey       string
	remoteTarballSHA256Key string
)

func init() {
	// Load up the architecture's S3 tarball key for use in this
	// package's tests; unconfigured architectures will result in
	// failing tests.
	agentS3Key, err := config.AgentRemoteTarballKey()
	if err == nil {
		remoteTarballKey = agentS3Key
		remoteTarballSHA256Key, _ = config.AgentRemoteTarballSHA256Key()
	} else {
		log.Println("Warning: this architecture does not support downloading of agent")
	}
}

func TestIsAgentCachedFalseMissingState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().Stat(config.CacheState()).Return(nil, errors.New("test error"))

	d := &Downloader{
		fs: mockFS,
	}

	assert.False(t, d.IsAgentCached(), "expect d.IsAgentCached() to be false")
}

func TestIsAgentCachedFalseEmptyState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)
	mockFSInfo := NewMockfileSizeInfo(mockCtrl)
	mockFS.EXPECT().Stat(config.CacheState()).Return(mockFSInfo, nil)
	mockFSInfo.EXPECT().Size().Return(int64(0))
	mockFS.EXPECT().Open(gomock.Any()).Times(0)

	d := &Downloader{
		fs: mockFS,
	}

	assert.False(t, d.IsAgentCached(), "expect d.IsAgentCached() to be false")
}

func TestIsAgentCachedFalseMissingTarball(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)
	mockFSInfo := NewMockfileSizeInfo(mockCtrl)
	mockFS.EXPECT().Stat(config.CacheState()).Return(mockFSInfo, nil)
	mockFSInfo.EXPECT().Size().Return(int64(1))
	mockFS.EXPECT().Stat(config.AgentTarball()).Return(nil, errors.New("test error"))

	d := &Downloader{
		fs: mockFS,
	}

	assert.False(t, d.IsAgentCached(), "expect d.IsAgentCached() to be false")
}

func TestIsAgentCachedTrue(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	file := ioutil.NopCloser(bytes.NewBufferString(fmt.Sprintf("%d", StatusCached)))
	mockFS := NewMockfileSystem(mockCtrl)
	mockFSInfo := NewMockfileSizeInfo(mockCtrl)
	mockFS.EXPECT().Stat(config.CacheState()).Return(mockFSInfo, nil)
	mockFS.EXPECT().Stat(config.AgentTarball()).Return(mockFSInfo, nil)
	mockFSInfo.EXPECT().Size().Return(int64(1)).Times(2)
	mockFS.EXPECT().Open(config.CacheState()).Return(file, nil)

	d := &Downloader{
		fs: mockFS,
	}

	assert.True(t, d.IsAgentCached(), "expect d.IsAgentCached() to be true")
}

func TestAgentCacheStatus(t *testing.T) {
	var cases = []struct {
		data     string
		expected CacheStatus
	}{
		// Expected states:
		{"0", StatusUncached},
		{"1", StatusCached},
		{"2", StatusReloadNeeded},
		{"1\n", StatusCached},
		// Invalid states:
		{"spurious", StatusUncached},
		{" ", StatusUncached},
		{"256", StatusUncached},
	}

	for _, testcase := range cases {
		t.Run(string(testcase.data), func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			file := ioutil.NopCloser(bytes.NewBufferString(testcase.data))
			mockFS := NewMockfileSystem(mockCtrl)
			mockFSInfo := NewMockfileSizeInfo(mockCtrl)

			mockFS.EXPECT().Stat(config.CacheState()).Return(mockFSInfo, nil)
			mockFS.EXPECT().Stat(config.AgentTarball()).Return(mockFSInfo, nil)
			mockFSInfo.EXPECT().Size().Return(int64(1)).Times(2)
			mockFS.EXPECT().Open(config.CacheState()).Return(file, nil)

			d := &Downloader{fs: mockFS}

			actual := d.AgentCacheStatus()
			assert.Equal(t, testcase.expected, actual, "expected output %d to match %d for input %s", actual, testcase.expected, testcase.data)
		})
	}
}

func TestGetPartitionBucketRegion(t *testing.T) {
	d := &Downloader{}

	var cases = []struct {
		region         string
		expectedResult string
	}{
		{config.DefaultRegionName, config.DefaultRegionName},
		{"cn-north-1", "cn-north-1"},
		{"missing-region", config.DefaultRegionName},
	}

	for _, c := range cases {
		t.Run(c.region, func(t *testing.T) {
			d.region = c.region
			result := d.getPartitionBucketRegion()
			assert.Equal(t, c.expectedResult, result, "expected getPartitionBucketRegion to give result %s", result)
		})
	}
}

func TestDownloadAgentMkdirFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)
	mockS3Downloader := NewMocks3DownloaderAPI(mockCtrl)
	mockMetadata := NewMockinstanceMetadata(mockCtrl)

	mockFS.EXPECT().MkdirAll(config.CacheDirectory(), os.ModeDir|0700).Return(errors.New("test error"))

	d := &Downloader{
		s3Downloader: mockS3Downloader,
		fs:           mockFS,
		metadata:     mockMetadata,
		region:       config.DefaultRegionName,
	}

	d.DownloadAgent()
}

func TestDownloadAgentDownloadSHA256Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)
	mockS3Downloader := NewMocks3DownloaderAPI(mockCtrl)
	mockMetadata := NewMockinstanceMetadata(mockCtrl)

	gomock.InOrder(
		mockFS.EXPECT().MkdirAll(config.CacheDirectory(), os.ModeDir|0700),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballSHA256Key).Return("", errors.New("test error")),
	)

	d := &Downloader{
		s3Downloader: mockS3Downloader,
		fs:           mockFS,
		metadata:     mockMetadata,
		region:       config.DefaultRegionName,
	}

	d.DownloadAgent()
}

func TestDownloadAgentReadPublishedSha256Failure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)
	mockS3Downloader := NewMocks3DownloaderAPI(mockCtrl)
	mockMetadata := NewMockinstanceMetadata(mockCtrl)

	tempSHA256File, err := ioutil.TempFile("", "sha256-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempSHA256File.Close()

	gomock.InOrder(
		mockFS.EXPECT().MkdirAll(config.CacheDirectory(), os.ModeDir|0700),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballSHA256Key).Return(tempSHA256File.Name(), nil),
		mockFS.EXPECT().Open(tempSHA256File.Name()).Return(tempSHA256File, nil),
		mockFS.EXPECT().ReadAll(tempSHA256File).Return(nil, errors.New("test error")),
		mockFS.EXPECT().Remove(tempSHA256File.Name()),
	)

	d := &Downloader{
		s3Downloader: mockS3Downloader,
		fs:           mockFS,
		metadata:     mockMetadata,
		region:       config.DefaultRegionName,
	}

	d.DownloadAgent()
}

func TestDownloadAgentDownloadTarballFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha256sum := "sha256sum"

	mockFS := NewMockfileSystem(mockCtrl)
	mockS3Downloader := NewMocks3DownloaderAPI(mockCtrl)
	mockMetadata := NewMockinstanceMetadata(mockCtrl)

	tempSHA256File, err := ioutil.TempFile("", "sha256-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempSHA256File.Close()

	tempAgentFile, err := ioutil.TempFile("", "agent-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempAgentFile.Close()

	gomock.InOrder(
		mockFS.EXPECT().MkdirAll(config.CacheDirectory(), os.ModeDir|0700),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballSHA256Key).Return(tempSHA256File.Name(), nil),
		mockFS.EXPECT().Open(tempSHA256File.Name()).Return(tempSHA256File, nil),
		mockFS.EXPECT().ReadAll(tempSHA256File).Return([]byte(sha256sum), nil),
		mockFS.EXPECT().Remove(tempSHA256File.Name()),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballKey).Return("", errors.New("test error")),
	)

	d := &Downloader{
		s3Downloader: mockS3Downloader,
		fs:           mockFS,
		metadata:     mockMetadata,
		region:       config.DefaultRegionName,
	}

	d.DownloadAgent()
}

func TestDownloadAgentCopyFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha256sum := "sha256sum"

	mockFS := NewMockfileSystem(mockCtrl)
	mockS3Downloader := NewMocks3DownloaderAPI(mockCtrl)
	mockMetadata := NewMockinstanceMetadata(mockCtrl)

	tempSHA256File, err := ioutil.TempFile("", "sha256-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempSHA256File.Close()

	tempAgentFile, err := ioutil.TempFile("", "agent-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempAgentFile.Close()

	tempReader := ioutil.NopCloser(&bytes.Buffer{})

	gomock.InOrder(
		mockFS.EXPECT().MkdirAll(config.CacheDirectory(), os.ModeDir|0700),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballSHA256Key).Return(tempSHA256File.Name(), nil),
		mockFS.EXPECT().Open(tempSHA256File.Name()).Return(tempSHA256File, nil),
		mockFS.EXPECT().ReadAll(tempSHA256File).Return([]byte(sha256sum), nil),
		mockFS.EXPECT().Remove(tempSHA256File.Name()),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballKey).Return(tempAgentFile.Name(), nil),
		mockFS.EXPECT().Open(tempAgentFile.Name()).Return(tempReader, nil),
		mockFS.EXPECT().Copy(gomock.Any(), tempReader).Return(int64(0), errors.New("test error")),
		mockFS.EXPECT().Stat(tempAgentFile.Name()).Return(nil, nil),
		mockFS.EXPECT().Remove(tempAgentFile.Name()),
	)

	d := &Downloader{
		s3Downloader: mockS3Downloader,
		fs:           mockFS,
		metadata:     mockMetadata,
		region:       config.DefaultRegionName,
	}

	d.DownloadAgent()
}

func TestDownloadAgentSHA256Mismatch(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	sha256sum := "sha256sum"

	mockFS := NewMockfileSystem(mockCtrl)
	mockS3Downloader := NewMocks3DownloaderAPI(mockCtrl)
	mockMetadata := NewMockinstanceMetadata(mockCtrl)

	tempSHA256File, err := ioutil.TempFile("", "sha256-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempSHA256File.Close()

	tempAgentFile, err := ioutil.TempFile("", "agent-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempAgentFile.Close()

	tempReader := ioutil.NopCloser(&bytes.Buffer{})

	gomock.InOrder(
		mockFS.EXPECT().MkdirAll(config.CacheDirectory(), os.ModeDir|0700),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballSHA256Key).Return(tempSHA256File.Name(), nil),
		mockFS.EXPECT().Open(tempSHA256File.Name()).Return(tempSHA256File, nil),
		mockFS.EXPECT().ReadAll(tempSHA256File).Return([]byte(sha256sum), nil),
		mockFS.EXPECT().Remove(tempSHA256File.Name()),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballKey).Return(tempAgentFile.Name(), nil),
		mockFS.EXPECT().Open(tempAgentFile.Name()).Return(tempReader, nil),
		mockFS.EXPECT().Copy(gomock.Any(), tempReader).Return(int64(0), nil),
		mockFS.EXPECT().Stat(tempAgentFile.Name()).Return(nil, nil),
		mockFS.EXPECT().Remove(tempAgentFile.Name()),
	)

	d := &Downloader{
		s3Downloader: mockS3Downloader,
		fs:           mockFS,
		metadata:     mockMetadata,
		region:       config.DefaultRegionName,
	}

	d.DownloadAgent()
}

func TestDownloadAgentSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tarballContents := "tarball contents"
	tarballReader := ioutil.NopCloser(bytes.NewBufferString(tarballContents))
	expectedSha256Sum := fmt.Sprintf("%x\n", sha256.Sum256([]byte(tarballContents)))

	tempSHA256File, err := ioutil.TempFile("", "sha256-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempSHA256File.Close()

	tempAgentFile, err := ioutil.TempFile("", "agent-test")
	assert.NoError(t, err, "Expect to successfully create a temporary file")
	defer tempAgentFile.Close()

	mockFS := NewMockfileSystem(mockCtrl)
	mockS3Downloader := NewMocks3DownloaderAPI(mockCtrl)
	mockMetadata := NewMockinstanceMetadata(mockCtrl)

	gomock.InOrder(
		mockFS.EXPECT().MkdirAll(config.CacheDirectory(), os.ModeDir|0700),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballSHA256Key).Return(tempSHA256File.Name(), nil),
		mockFS.EXPECT().Open(tempSHA256File.Name()).Return(tempSHA256File, nil),
		mockFS.EXPECT().ReadAll(tempSHA256File).Return([]byte(expectedSha256Sum), nil),
		mockFS.EXPECT().Remove(tempSHA256File.Name()),
		mockS3Downloader.EXPECT().downloadFile(remoteTarballKey).Return(tempAgentFile.Name(), nil),
		mockFS.EXPECT().Open(tempAgentFile.Name()).Return(tarballReader, nil),
		mockFS.EXPECT().Copy(gomock.Any(), tarballReader).Do(func(writer io.Writer, reader io.Reader) {
			_, err = io.Copy(writer, reader)
			assert.NoError(t, err, "Expect to successfully write to file")
		}),
		mockFS.EXPECT().Rename(tempAgentFile.Name(), config.AgentTarball()),
		mockFS.EXPECT().Stat(tempAgentFile.Name()).Return(nil, errors.New("temp file has been renamed")),
	)

	d := &Downloader{
		s3Downloader: mockS3Downloader,
		fs:           mockFS,
		metadata:     mockMetadata,
		region:       config.DefaultRegionName,
	}

	d.DownloadAgent()
}

func TestLoadDesiredAgentFailOpenDesired(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().Open(config.DesiredImageLocatorFile()).Return(nil, errors.New("test error"))

	d := &Downloader{
		fs: mockFS,
	}

	_, err := d.LoadDesiredAgent()
	assert.Error(t, err, "Expect error to be returned when unable to open desired image file")
}

func TestLoadDesiredAgentFailReadDesired(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().Open(config.DesiredImageLocatorFile()).Return(ioutil.NopCloser(&bytes.Buffer{}), nil)

	d := &Downloader{
		fs: mockFS,
	}

	_, err := d.LoadDesiredAgent()
	assert.Error(t, err, "Expect error to be returned when unable to read desired image file")
}

func TestRecordCachedAgent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().WriteFile(config.CacheState(), []byte("1"), os.FileMode(orwPerm))

	d := &Downloader{
		fs: mockFS,
	}
	d.RecordCachedAgent()
}

func TestLoadDesiredAgent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	desiredImage := "my-new-agent-image"

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().Open(config.DesiredImageLocatorFile()).Return(ioutil.NopCloser(bytes.NewBufferString(desiredImage+"\n")), nil)
	mockFS.EXPECT().Base(gomock.Any()).Return(desiredImage + "\n")
	mockFS.EXPECT().Open(config.CacheDirectory() + "/" + desiredImage)

	d := &Downloader{
		fs: mockFS,
	}

	d.LoadDesiredAgent()
}
