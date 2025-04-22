package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type EnvConfig struct {
	EnvNames     []string
	DefaultValue string
	IsRequired   bool
}

var pipelineVisibilityEnvConfigs = map[string]EnvConfig{
	"pipeline_end": {
		EnvNames:     []string{"PLUGIN_BUILD_FINISHED", "DRONE_BUILD_FINISHED", "CI_BUILD_FINISHED"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"pipeline_start": {
		EnvNames:     []string{"PLUGIN_BUILD_STARTED", "DRONE_BUILD_STARTED", "CI_BUILD_STARTED"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"git_author_email": {
		EnvNames:     []string{"PLUGIN_COMMIT_AUTHOR_EMAIL", "DRONE_COMMIT_AUTHOR_EMAIL", "CI_COMMIT_AUTHOR_EMAIL"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"git_author_name": {
		EnvNames:     []string{"PLUGIN_COMMIT_AUTHOR", "DRONE_COMMIT_AUTHOR", "CI_COMMIT_AUTHOR"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"git_branch": {
		EnvNames:     []string{"PLUGIN_BRANCH", "DRONE_BRANCH"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"git_commit_message": {
		EnvNames:     []string{"PLUGIN_COMMIT_MESSAGE", "DRONE_COMMIT_MESSAGE", "CI_COMMIT_MESSAGE"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"git_repository_url": {
		EnvNames:     []string{"PLUGIN_REPO_REMOTE", "DRONE_GIT_HTTP_URL", "CI_REPO_REMOTE"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"git_commit_sha": {
		EnvNames:     []string{"PLUGIN_COMMIT_SHA", "DRONE_COMMIT_SHA", "CI_COMMIT_SHA"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"git_tag": {
		EnvNames:     []string{"PLUGIN_TAG", "DRONE_TAG"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"pipeline_is_manual": {
		EnvNames:     []string{"PLUGIN_BUILD_TRIGGER", "DRONE_BUILD_TRIGGER"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"pipeline_level": {
		EnvNames:     []string{""},
		DefaultValue: "pipeline",
		IsRequired:   true,
	},
	"pipeline_name": {
		EnvNames:     []string{"PLUGIN_PIPELINE_ID", "HARNESS_PIPELINE_ID"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"node_hostname": {
		EnvNames:     []string{"PLUGIN_SYSTEM_HOST", "DRONE_SYSTEM_HOSTNAME", "DRONE_SYSTEM_HOST"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"node_name": {
		EnvNames:     []string{"PLUGIN_SYSTEM_HOST", "DRONE_SYSTEM_HOSTNAME", "DRONE_SYSTEM_HOST"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"node_workspace": {
		EnvNames:     []string{"PLUGIN_WORKSPACE", "DRONE_WORKSPACE", "HARNESS_WORKSPACE"},
		DefaultValue: "",
		IsRequired:   false,
	},
	"pipeline_url": {
		EnvNames:     []string{"PLUGIN_BUILD_LINK", "DRONE_BUILD_LINK", "CI_BUILD_LINK"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"pipeline_unique_id": {
		EnvNames:     []string{"PLUGIN_PIPELINE_ID", "HARNESS_PIPELINE_ID"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"pipeline_status": {
		EnvNames:     []string{"PLUGIN_BUILD_STATUS", "DRONE_BUILD_STATUS", "CI_BUILD_STATUS"},
		DefaultValue: "",
		IsRequired:   true,
	},
	"pipeline_type": {
		EnvNames:     []string{""},
		DefaultValue: "cipipeline_resource_request",
		IsRequired:   false,
	},
}

func getPipelineVisibilityEnvValue(env string) string {
	for _, env := range pipelineVisibilityEnvConfigs[env].EnvNames {
		if value := os.Getenv(env); value != "" {
			return value
		}
	}
	return pipelineVisibilityEnvConfigs[env].DefaultValue
}

func ConvertUnixToRFC3339(timestamp int64) string {
	if timestamp == 0 {
		return "" // Return an empty string if the timestamp is invalid
	}
	t := time.Unix(timestamp, 0).UTC()
	return t.Format(time.RFC3339)
}

func parseUnix(value string) int64 {
	ts, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0 // Return 0 on error
	}
	return ts
}

func ValidateRequiredEnvVars() error {
	var missing []string

	for key, conf := range pipelineVisibilityEnvConfigs {
		if conf.IsRequired {
			val := getPipelineVisibilityEnvValue(key)
			if val == "" {
				missing = append(missing, conf.EnvNames...)
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}

func SendDatadogPipelineEvent(cfg *Config) error {
	requestBody := map[string]interface{}{
		"data": map[string]interface{}{
			"attributes": map[string]interface{}{
				"resource": map[string]interface{}{
					"level":         getPipelineVisibilityEnvValue("pipeline_level"),
					"unique_id":     getPipelineVisibilityEnvValue("pipeline_unique_id"),
					"name":          getPipelineVisibilityEnvValue("pipeline_unique_id"),
					"url":           getPipelineVisibilityEnvValue("pipeline_url"),
					"start":         ConvertUnixToRFC3339(parseUnix(getPipelineVisibilityEnvValue("pipeline_start"))),
					"end":           ConvertUnixToRFC3339(parseUnix(getPipelineVisibilityEnvValue("pipeline_end"))),
					"status":        strings.ToLower(getPipelineVisibilityEnvValue("pipeline_status")),
					"partial_retry": false,
					"is_manual":     getPipelineVisibilityEnvValue("pipeline_is_manual") == "manual",
					"git": map[string]interface{}{
						"repository_url": getPipelineVisibilityEnvValue("git_repository_url"),
						"sha":            getPipelineVisibilityEnvValue("git_commit_sha"),
						"author_email":   getPipelineVisibilityEnvValue("git_author_email"),
						"author_name":    getPipelineVisibilityEnvValue("git_author_name"),
					},
					"node": map[string]interface{}{
						"hostname":  getPipelineVisibilityEnvValue("node_hostname"),
						"name":      getPipelineVisibilityEnvValue("node_name"),
						"workspace": getPipelineVisibilityEnvValue("node_workspace"),
					},
				},
			},
			"type": getPipelineVisibilityEnvValue("pipeline_type"),
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error encoding ci pipeline visibility: %s", err)
	}

	if cfg.DryRun {
		log.Println("Dry run, logging payload:")
		log.Println(string(payload))
		return nil
	}

	url := fmt.Sprintf("https://api.%s.datadoghq.com/api/v2/ci/pipeline", cfg.Region)
	log.Printf("url: %s", url)
	log.Printf("Request Payload: %s", string(payload))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("DD-API-KEY", cfg.APIKey)
	if res, err := http.DefaultClient.Do(req); err == nil {
		log.Println("Response")
		log.Println(res.StatusCode)
		resp, _ := io.ReadAll(res.Body)
		log.Println(string(resp))
		if res.StatusCode >= 300 {
			return fmt.Errorf("server responded with: %s", res.Status)
		}
	} else {
		return fmt.Errorf("unable to send data: %s", err)
	}
	return nil
}
