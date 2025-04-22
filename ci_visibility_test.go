package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGetPipelineVisibilityEnvValue(t *testing.T) {
	// set env and test override
	os.Setenv("PLUGIN_BUILD_STARTED", "1700000000")
	defer os.Unsetenv("PLUGIN_BUILD_STARTED")

	val := getPipelineVisibilityEnvValue("pipeline_start")
	if val != "1700000000" {
		t.Errorf("Expected 1700000000, got %s", val)
	}
}

func TestConvertUnixToRFC3339(t *testing.T) {
	result := ConvertUnixToRFC3339(1700000000)
	expected := "2023-11-14T22:13:20Z"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestSendDatadogPipelineEventDryRun(t *testing.T) {

	setTestEnv()
	cfg, err := parseConfig()
	if err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}

	err = SendDatadogPipelineEvent(cfg)
	if err != nil {
		t.Errorf("Expected no error on dry run, got: %v", err)
	}

}

func setTestEnv() {
	os.Setenv("PLUGIN_DRY_RUN", "true")
	os.Setenv("PLUGIN_API_KEY", "dummy")
	os.Setenv("PLUGIN_REGION", "us")
	os.Setenv("PLUGIN_CI_VISIBILITY_TYPE", "pipeline")

	// set required envs
	os.Setenv("PLUGIN_BUILD_STARTED", "1700000000")
	os.Setenv("PLUGIN_BUILD_FINISHED", "1700003600")
	os.Setenv("PLUGIN_COMMIT_AUTHOR_EMAIL", "user@example.com")
	os.Setenv("PLUGIN_REPO_REMOTE", "https://example.com/repo.git")
	os.Setenv("PLUGIN_COMMIT_SHA", "abcdef")
	os.Setenv("PLUGIN_BUILD_LINK", "https://ci.com/build/123")
	os.Setenv("PLUGIN_BUILD_STATUS", "success")
	os.Setenv("PLUGIN_PIPELINE_ID", "pipe123")
}

func unsetTestEnv() {

	os.Setenv("PLUGIN_DRY_RUN", "false")
	os.Setenv("PLUGIN_API_KEY", "")
	os.Setenv("PLUGIN_REGION", "")
	os.Setenv("PLUGIN_CI_VISIBILITY_TYPE", "")

	// set required envs
	os.Setenv("PLUGIN_BUILD_STARTED", "")
	os.Setenv("PLUGIN_BUILD_FINISHED", "")
	os.Setenv("PLUGIN_COMMIT_AUTHOR_EMAIL", "")
	os.Setenv("PLUGIN_REPO_REMOTE", "")
	os.Setenv("PLUGIN_COMMIT_SHA", "")
	os.Setenv("PLUGIN_BUILD_LINK", "")
	os.Setenv("PLUGIN_BUILD_STATUS", "")
	os.Setenv("PLUGIN_PIPELINE_ID", "")
}

func TestMarshalPipelineRequest(t *testing.T) {
	start := ConvertUnixToRFC3339(1700000000)
	end := ConvertUnixToRFC3339(1700003600)

	body := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"resource": map[string]interface{}{
					"start": start,
					"end":   end,
				},
			},
			"type": "cipipeline_resource_request",
		},
	}

	_, err := json.Marshal(body)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
	}
}
