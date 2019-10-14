// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package docker

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-init/ecs-init/config"
	"github.com/aws/amazon-ecs-init/ecs-init/gpu"
	godocker "github.com/fsouza/go-dockerclient"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// expectedAgentBinds is the total number of agent host config binds.
// Note: Change this value every time when a new bind mount is added to
// agent for the tests to pass
const (
	expectedAgentBindsUnspecifiedPlatform = 17
	expectedAgentBindsSuseUbuntuPlatform  = 15
)

var expectedAgentBinds = expectedAgentBindsUnspecifiedPlatform

func TestIsAgentImageLoadedListFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListImages(godocker.ListImagesOptions{All: true}).Return(nil, errors.New("test error"))

	client := &Client{
		docker: mockDocker,
	}
	loaded, err := client.IsAgentImageLoaded()
	if err == nil {
		t.Error("Error should be returned")
	}
	if loaded {
		t.Error("IsImageLoaded should return false")
	}
}

func TestIsAgentImageLoadedNoMatches(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListImages(godocker.ListImagesOptions{All: true}).Return(
		append(make([]godocker.APIImages, 0), godocker.APIImages{
			RepoTags: append(make([]string, 0), ""),
		}), nil)

	client := &Client{
		docker: mockDocker,
	}
	loaded, err := client.IsAgentImageLoaded()
	if err != nil {
		t.Error("Error should not be returned")
	}
	if loaded {
		t.Error("IsImageLoaded should return false")
	}
}

func TestIsImageLoadedMatches(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListImages(godocker.ListImagesOptions{All: true}).Return(
		append(make([]godocker.APIImages, 0), godocker.APIImages{
			RepoTags: append(make([]string, 0), config.AgentImageName),
		}), nil)

	client := &Client{
		docker: mockDocker,
	}
	loaded, err := client.IsAgentImageLoaded()
	if err != nil {
		t.Error("Error should not be returned")
	}
	if !loaded {
		t.Error("IsImageLoaded should return true")
	}
}

func TestLoadImage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().LoadImage(godocker.LoadImageOptions{})

	client := &Client{
		docker: mockDocker,
	}
	client.LoadImage(nil)
}

func TestRemoveExistingAgentContainerListContainersFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListContainers(godocker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"status": []string{},
		},
	}).Return(nil, errors.New("test error"))

	client := &Client{
		docker: mockDocker,
	}
	err := client.RemoveExistingAgentContainer()
	if err == nil {
		t.Error("Error should be returned")
	}
}

func TestRemoveExistingAgentContainerNoneFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListContainers(godocker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"status": []string{},
		},
	})

	client := &Client{
		docker: mockDocker,
	}
	err := client.RemoveExistingAgentContainer()
	if err != nil {
		t.Error("Error should not be returned")
	}
}

func TestRemoveExistingAgentContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListContainers(godocker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"status": []string{},
		},
	}).Return([]godocker.APIContainers{
		godocker.APIContainers{
			Names: []string{"/" + config.AgentContainerName},
			ID:    "id",
		},
	}, nil)
	mockDocker.EXPECT().RemoveContainer(godocker.RemoveContainerOptions{
		ID:    "id",
		Force: true,
	})

	client := &Client{
		docker: mockDocker,
	}
	err := client.RemoveExistingAgentContainer()
	if err != nil {
		t.Error("Error should not be returned")
	}
}

func TestStartAgentNoEnvFile(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	containerID := "container id"

	mockFS := NewMockfileSystem(mockCtrl)
	mockDocker := NewMockdockerclient(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return(nil, errors.New("not found")).AnyTimes()
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return(nil, errors.New("test error")).AnyTimes()
	mockDocker.EXPECT().CreateContainer(gomock.Any()).Do(func(opts godocker.CreateContainerOptions) {
		validateCommonCreateContainerOptions(opts, t)
	}).Return(&godocker.Container{
		ID: containerID,
	}, nil)
	mockDocker.EXPECT().StartContainer(containerID, nil)
	mockDocker.EXPECT().WaitContainer(containerID)

	client := &Client{
		docker: mockDocker,
		fs:     mockFS,
	}

	_, err := client.StartAgent()
	if err != nil {
		t.Error("Error should not be returned")
	}
}

func validateCommonCreateContainerOptions(opts godocker.CreateContainerOptions, t *testing.T) {
	if opts.Name != "ecs-agent" {
		t.Errorf("Expected container Name to be %s but was %s", "ecs-agent", opts.Name)
	}
	cfg := opts.Config

	if len(cfg.Env) < 3 {
		t.Errorf("Expected at least 3 elements to be in Env, but was %d", len(cfg.Env))
	}
	envVariables := make(map[string]struct{})
	for _, envVar := range cfg.Env {
		envVariables[envVar] = struct{}{}
	}
	expectKey("ECS_DATADIR=/data", envVariables, t)
	expectKey("ECS_LOGFILE=/log/"+config.AgentLogFile, envVariables, t)
	expectKey("ECS_AGENT_CONFIG_FILE_PATH="+config.AgentJSONConfigFile(), envVariables, t)
	expectKey("ECS_UPDATE_DOWNLOAD_DIR="+config.CacheDirectory(), envVariables, t)
	expectKey("ECS_UPDATES_ENABLED=true", envVariables, t)
	expectKey(`ECS_AVAILABLE_LOGGING_DRIVERS=["json-file","syslog","awslogs","none"]`,
		envVariables, t)
	expectKey("ECS_ENABLE_TASK_IAM_ROLE=true", envVariables, t)
	expectKey("ECS_ENABLE_TASK_IAM_ROLE_NETWORK_HOST=true", envVariables, t)
	expectKey("ECS_ENABLE_TASK_ENI=true", envVariables, t)
	expectKey("ECS_ENABLE_AWSLOGS_EXECUTIONROLE_OVERRIDE=true", envVariables, t)

	if cfg.Image != config.AgentImageName {
		t.Errorf("Expected image to be %s", config.AgentImageName)
	}

	hostCfg := opts.HostConfig

	// for hosts that do not have cert directories explicity mounted, ignore
	// host cert directory configuration.
	// TODO (adnxn): ideally, these should be behind build flags.
	// https://github.com/aws/amazon-ecs-init/issues/131
	if certDir := config.HostPKIDirPath(); certDir == "" {
		expectedAgentBinds = expectedAgentBindsSuseUbuntuPlatform
	}

	if len(hostCfg.Binds) != expectedAgentBinds {
		t.Errorf("Expected exactly %d elements to be in Binds, but was %d", expectedAgentBinds, len(hostCfg.Binds))
	}
	binds := make(map[string]struct{})
	for _, binding := range hostCfg.Binds {
		binds[binding] = struct{}{}
	}
	defaultDockerSocket, _ := config.DockerUnixSocket()
	expectKey(defaultDockerSocket+":"+defaultDockerSocket, binds, t)
	expectKey(config.LogDirectory()+":/log", binds, t)
	expectKey(config.AgentDataDirectory()+":/data", binds, t)
	expectKey(config.AgentConfigDirectory()+":"+config.AgentConfigDirectory(), binds, t)
	expectKey(config.CacheDirectory()+":"+config.CacheDirectory(), binds, t)
	expectKey(config.ProcFS+":"+hostProcDir+":ro", binds, t)
	expectKey(iptablesUsrLibDir+":"+iptablesUsrLibDir+":ro", binds, t)
	expectKey(iptablesLibDir+":"+iptablesLibDir+":ro", binds, t)
	expectKey(iptablesUsrLib64Dir+":"+iptablesUsrLib64Dir+":ro", binds, t)
	expectKey(iptablesLib64Dir+":"+iptablesLib64Dir+":ro", binds, t)
	expectKey(iptablesExecutableDir+":"+iptablesExecutableDir+":ro", binds, t)
	for _, pluginDir := range pluginDirs {
		expectKey(pluginDir+":"+pluginDir+readOnly, binds, t)
	}

	if hostCfg.NetworkMode != networkMode {
		t.Errorf("Expected network mode to be %s, got %s", networkMode, hostCfg.NetworkMode)
	}

	if len(hostCfg.CapAdd) != 2 {
		t.Error("Mismatch detected in added host config capabilities")
	}

	capNetAdminFound := false
	capSysAdminFound := false
	for _, cap := range hostCfg.CapAdd {
		if cap == CapNetAdmin {
			capNetAdminFound = true
		}
		if cap == CapSysAdmin {
			capSysAdminFound = true
		}
	}
	if !capNetAdminFound {
		t.Errorf("Missing %s from host config capabilities", CapNetAdmin)
	}
	if !capSysAdminFound {
		t.Errorf("Missing %s from host config capabilities", CapSysAdmin)
	}

	if hostCfg.Init != true {
		t.Error("Incorrect host config. Expected Init to be true")
	}
}

func expectKey(key string, input map[string]struct{}, t *testing.T) {
	if _, ok := input[key]; !ok {
		t.Errorf("Expected %s to be defined", key)
	}
}

func TestStartAgentEnvFile(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	envFile := "\nAGENT_TEST_VAR=val\nAGENT_TEST_VAR2=val2\n"
	containerID := "container id"

	mockFS := NewMockfileSystem(mockCtrl)
	mockDocker := NewMockdockerclient(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return(nil, errors.New("not found")).AnyTimes()
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return([]byte(envFile), nil).AnyTimes()
	mockDocker.EXPECT().CreateContainer(gomock.Any()).Do(func(opts godocker.CreateContainerOptions) {
		validateCommonCreateContainerOptions(opts, t)
		cfg := opts.Config

		envVariables := make(map[string]struct{})
		for _, envVar := range cfg.Env {
			envVariables[envVar] = struct{}{}
		}
		expectKey("AGENT_TEST_VAR=val", envVariables, t)
		expectKey("AGENT_TEST_VAR2=val2", envVariables, t)
	}).Return(&godocker.Container{
		ID: containerID,
	}, nil)
	mockDocker.EXPECT().StartContainer(containerID, nil)
	mockDocker.EXPECT().WaitContainer(containerID)

	client := &Client{
		docker: mockDocker,
		fs:     mockFS,
	}

	_, err := client.StartAgent()
	if err != nil {
		t.Error("Error should not be returned")
	}
}
func TestStartAgentWithGPUConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	envFile := "\nECS_ENABLE_GPU_SUPPORT=true\n"
	containerID := "container id"
	expectedAgentBinds += 1

	defer func() {
		MatchFilePatternForGPU = FilePatternMatchForGPU
		expectedAgentBinds = expectedAgentBindsUnspecifiedPlatform
	}()
	MatchFilePatternForGPU = func(pattern string) ([]string, error) {
		return []string{"/dev/nvidia0", "/dev/nvidia1"}, nil
	}

	mockFS := NewMockfileSystem(mockCtrl)
	mockDocker := NewMockdockerclient(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return([]byte(envFile), nil).AnyTimes()
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return(nil, errors.New("not found")).AnyTimes()
	mockDocker.EXPECT().CreateContainer(gomock.Any()).Do(func(opts godocker.CreateContainerOptions) {
		validateCommonCreateContainerOptions(opts, t)
		var found bool
		for _, bind := range opts.HostConfig.Binds {
			if bind == gpu.GPUInfoDirPath+":"+gpu.GPUInfoDirPath {
				found = true
				break
			}
		}
		assert.True(t, found)

		cfg := opts.Config

		envVariables := make(map[string]struct{})
		for _, envVar := range cfg.Env {
			envVariables[envVar] = struct{}{}
		}
	}).Return(&godocker.Container{
		ID: containerID,
	}, nil)
	mockDocker.EXPECT().StartContainer(containerID, nil)
	mockDocker.EXPECT().WaitContainer(containerID)

	client := &Client{
		docker: mockDocker,
		fs:     mockFS,
	}

	_, err := client.StartAgent()
	assert.NoError(t, err)
}

func TestStartAgentWithGPUConfigNoDevices(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	envFile := "\nECS_ENABLE_GPU_SUPPORT=true\n"
	containerID := "container id"

	defer func() {
		MatchFilePatternForGPU = FilePatternMatchForGPU
	}()
	MatchFilePatternForGPU = func(pattern string) ([]string, error) {
		// matches is nil
		return nil, nil
	}

	mockFS := NewMockfileSystem(mockCtrl)
	mockDocker := NewMockdockerclient(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return([]byte(envFile), nil).AnyTimes()
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return(nil, errors.New("not found")).AnyTimes()
	mockDocker.EXPECT().CreateContainer(gomock.Any()).Do(func(opts godocker.CreateContainerOptions) {
		validateCommonCreateContainerOptions(opts, t)
		cfg := opts.Config

		envVariables := make(map[string]struct{})
		for _, envVar := range cfg.Env {
			envVariables[envVar] = struct{}{}
		}
	}).Return(&godocker.Container{
		ID: containerID,
	}, nil)
	mockDocker.EXPECT().StartContainer(containerID, nil)
	mockDocker.EXPECT().WaitContainer(containerID)

	client := &Client{
		docker: mockDocker,
		fs:     mockFS,
	}

	_, err := client.StartAgent()
	assert.NoError(t, err)
}

func TestGetContainerConfigWithFileOverrides(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	envFile := "\nECS_UPDATES_ENABLED=false\nAGENT_TEST_VAR2=\"val2\"=val2\n"

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return(nil, errors.New("not found"))
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return([]byte(envFile), nil)

	client := &Client{
		fs: mockFS,
	}
	cfg := client.getContainerConfig()

	envVariables := make(map[string]struct{})
	for _, envVar := range cfg.Env {
		envVariables[envVar] = struct{}{}
	}
	expectKey("ECS_UPDATES_ENABLED=false", envVariables, t)
	expectKey("AGENT_TEST_VAR2=\"val2\"=val2", envVariables, t)
	if _, ok := envVariables["ECS_UPDATES_ENABLED=true"]; ok {
		t.Errorf("Did not expect ECS_UPDATES_ENABLED=true to be defined")
	}

}

func TestGetInstanceConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	envFile := "\nECS_ENABLE_GPU_SUPPORT=true\n"
	defer func() {
		MatchFilePatternForGPU = FilePatternMatchForGPU
	}()
	MatchFilePatternForGPU = func(pattern string) ([]string, error) {
		// GPU device present
		return []string{"/dev/nvidia0"}, nil
	}

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return([]byte(envFile), nil)
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return(nil, errors.New("not found"))

	client := &Client{
		fs: mockFS,
	}
	cfg := client.getContainerConfig()

	envVariables := make(map[string]struct{})
	for _, envVar := range cfg.Env {
		envVariables[envVar] = struct{}{}
	}
	expectKey("ECS_ENABLE_GPU_SUPPORT=true", envVariables, t)
}

func TestGetNonGPUInstanceConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	envFile := "\nECS_ENABLE_GPU_SUPPORT=true\n"
	defer func() {
		MatchFilePatternForGPU = FilePatternMatchForGPU
	}()
	MatchFilePatternForGPU = func(pattern string) ([]string, error) {
		// matches is nil
		return nil, nil
	}

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return([]byte(envFile), nil)
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return(nil, errors.New("not found"))

	client := &Client{
		fs: mockFS,
	}
	cfg := client.getContainerConfig()

	envVariables := make(map[string]struct{})
	for _, envVar := range cfg.Env {
		envVariables[envVar] = struct{}{}
	}
	if _, ok := envVariables["ECS_ENABLE_GPU_SUPPORT=true"]; ok {
		t.Errorf("Expected ECS_ENABLE_GPU_SUPPORT=true to be not present")
	}
}

func TestGetConfigOverrides(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	instanceEnvFile := "\nECS_ENABLE_GPU_SUPPORT=true\n"
	userEnvFile := "\nECS_ENABLE_GPU_SUPPORT=false\n"

	mockFS := NewMockfileSystem(mockCtrl)

	mockFS.EXPECT().ReadFile(config.InstanceConfigFile()).Return([]byte(instanceEnvFile), nil)
	mockFS.EXPECT().ReadFile(config.AgentConfigFile()).Return([]byte(userEnvFile), nil)

	client := &Client{
		fs: mockFS,
	}
	cfg := client.getContainerConfig()

	envVariables := make(map[string]struct{})
	for _, envVar := range cfg.Env {
		envVariables[envVar] = struct{}{}
	}
	expectKey("ECS_ENABLE_GPU_SUPPORT=false", envVariables, t)
}

func TestStopAgentError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListContainers(godocker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"status": []string{},
		},
	}).Return(nil, errors.New("test error"))

	client := &Client{
		docker: mockDocker,
	}

	err := client.StopAgent()
	if err == nil {
		t.Error("Error should be returned")
	}
}

func TestStopAgentNone(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListContainers(godocker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"status": []string{},
		},
	}).Return([]godocker.APIContainers{godocker.APIContainers{}}, nil)

	client := &Client{
		docker: mockDocker,
	}

	err := client.StopAgent()
	if err != nil {
		t.Error("Error should not be returned")
	}
}

func TestStopAgent(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDocker := NewMockdockerclient(mockCtrl)

	mockDocker.EXPECT().ListContainers(godocker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"status": []string{},
		},
	}).Return([]godocker.APIContainers{
		godocker.APIContainers{
			Names: []string{"/" + config.AgentContainerName},
			ID:    "id",
		},
	}, nil)
	mockDocker.EXPECT().StopContainer("id", uint(10))

	client := &Client{
		docker: mockDocker,
	}

	err := client.StopAgent()
	if err != nil {
		t.Error("Error should not be returned")
	}
}

func TestContainerLabels(t *testing.T) {
	testData := `{"test.label.1":"value1","test.label.2":"value2"}`
	out, err := generateLabelMap(testData)
	if err != nil {
		t.Logf("Got an error while decoding labels, error: %s", err)
		t.Fail()
	}
	if out["test.label.1"] != "value1" {
		t.Logf("Label did not give the correct value out.")
		t.Fail()
	}

	for key, value := range out {
		t.Logf("Key: %s %T | Value: %s %T", key, key, value, value)
	}
}

func TestContainerLabelsNoData(t *testing.T) {
	tests := []struct {
		name     string
		testData string
	}{
		{
			name:     "Blank",
			testData: "",
		},
		{
			name:     "Empty JSON",
			testData: `{}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &godocker.Config{}

			setLabels(cfg, test.testData)

			if cfg.Labels != nil {
				t.Logf("Labels are set but we didn't give any. Current labels: %s", cfg.Labels)
				t.Fail()
			}
		})
	}
}

func TestContainerLabelsBadData(t *testing.T) {
	testData := `{"something":[{"test.label.1":"value1"},{"test.label.2":"value2"}]}`
	_, err := generateLabelMap(testData)
	if err == nil {
		t.Logf("Didn't get a error while getting lables on badly formatted data, error: %s", err)
		t.Fail()
	}
}
