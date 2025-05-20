# drone-datadog

[![Build Status](https://cloud.drone.io/api/badges/masci/drone-datadog/status.svg)](https://cloud.drone.io/masci/drone-datadog)

This plugin lets you send events, metrics and CI Pipeline visibility to Datadog from a drone pipeline.

Run the following script to install git-leaks support to this repo.
```
chmod +x ./git-hooks/install.sh
./git-hooks/install.sh
```

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t plugins/datadog  -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm \
  -e PLUGIN_CI_VISIBILITY_TYPE=pipeline \
  -v $(pwd):/drone \
  plugins/datadog
```

## Usage

To send a metric every time a pipeline runs, add this step:

```yml
- name: count-pipeline
  image: masci/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    metrics:
      - type: "count"
        name: "masci.pipelines.count"
        value: 1.0
        tags: ["project:${DRONE_REPO_NAME}", "branch:${DRONE_BRANCH}"]
```

Sending an event is similar, both `metrics` and `events` support the `host` field:

```yml
- name: notify-pipeline
  image: masci/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Building drone-datadog success"
        text: "Version ${DRONE_TAG} is available on Docker Hub"
        alert_type: "info"
        host: ${DRONE_SYSTEM_HOSTNAME}
        priority: "low"
```

You can use events to notify something bad happened:

```yml
- name: notify-pipeline
  image: masci/drone-datadog
  settings:
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Build failure"
        text: "Build ${DRONE_BUILD_NUMBER} has failed"
        alert_type: "error"
        priority: "normal"
  when:
    status:
      - failure
```

You can change the datadog site region to EU (`com` is default)

```yml
- name: notify-pipeline
  image: masci/drone-datadog
  settings:
    region: eu
    api_key:
      from_secret: datadog_api_key
    events:
      - title: "Build failure"
        text: "Build ${DRONE_BUILD_NUMBER} has failed"
        alert_type: "error"
  when:
    status:
      - failure
```
To send CI Pipeline Visibility:

```yml
- step:
  type: Plugin
  name: pipline_visibility_01
  identifier: pipline_visibility_01
  spec:
    connectorRef: senthil_dockerhub_connector
    image: senthilhns/datadog-pipeline-visibility
    settings:
      pipeline_id: datadog_pipeline_visibility_senthil_4353
      build_started: "1743948411"
      build_finished: "1743948415"
      commit_author: John Doe
      commit_author_email: John@test.com
      commit_message: Fix bug in API
      branch: main
      commit_sha: abc123def
      build_link: https://ci.example.com/build/12345
      repo_remote: https://github.com/masci/drone-datadog
      build_status: success
      api_key:
        from_secret: datadog_api_key
      ci_visibility_type: pipeline
      region: us5
      workspace: /harness
 
```


### Pipeline Visibility Environment Configurations
When the env variable names are set by the pipeline they are automatically picked up by the plugin. The following table lists the environment variables that are set by the pipeline and their corresponding yaml step key. User can also set the values in the yaml step key. The plugin will pick up the values from the environment variables if they are set, otherwise it will pick up the values from the yaml step key. 

| Env VariableNames Names                               | yaml step key         | Description                                | Is Required |
|-------------------------------------------------------|------------------------|--------------------------------------------|-------------|
| `DRONE_BUILD_FINISHED`, `CI_BUILD_FINISHED`           | `build_finished`       | Timestamp of when the pipeline finished    | ✅ Yes      |
| `DRONE_BUILD_STARTED`, `CI_BUILD_STARTED`             | `build_started`        | Timestamp of when the pipeline started     | ✅ Yes      |
| `DRONE_COMMIT_AUTHOR_EMAIL`, `CI_COMMIT_AUTHOR_EMAIL` | `commit_author_email`  | Email of the Git commit author             | ✅ Yes      |
| `DRONE_GIT_HTTP_URL`, `CI_REPO_REMOTE`                | `repo_remote`          | Git repository HTTP URL                    | ✅ Yes      |
| `DRONE_COMMIT_SHA`, `CI_COMMIT_SHA`                   | `commit_sha`           | Git commit SHA                             | ✅ Yes      |
| `DRONE_BUILD_LINK`, `CI_BUILD_LINK`                   | `build_link`           | URL to the pipeline or build details       | ✅ Yes      |
| `HARNESS_PIPELINE_ID`                                 | `pipeline_id`          | Unique identifier for the pipeline         | ✅ Yes      |
| `DRONE_BUILD_STATUS`, `CI_BUILD_STATUS`               | `build_status`         | Final build status (e.g., success/failure) | ✅ Yes      |
| `DRONE_COMMIT_AUTHOR`, `CI_COMMIT_AUTHOR`             | `commit_author`        | Name of the Git commit author              | ❌ No       |
| `DRONE_BRANCH`                                        | `branch`               | Git branch name                            | ❌ No       |
| `DRONE_COMMIT_MESSAGE`, `CI_COMMIT_MESSAGE`           | `commit_message`       | Commit message                             | ❌ No       |
| `DRONE_TAG`                                           | `tag`                  | Git tag (if the build was tagged)          | ❌ No       |
| `DRONE_BUILD_TRIGGER`                                 | `build_trigger`        | What triggered the build (e.g., manual)    | ❌ No       |
| `DRONE_SYSTEM_HOSTNAME`, `DRONE_SYSTEM_HOST`          | `system_host`          | Hostname of the node running the build     | ❌ No       |
| `DRONE_SYSTEM_HOSTNAME`, `DRONE_SYSTEM_HOST`          | `system_host`          | Alias for node name                        | ❌ No       |
| `DRONE_WORKSPACE`, `HARNESS_WORKSPACE`                | `workspace`            | Workspace path on the build node           | ❌ No       |

You can look at [this repo .drone.yml](.drone.yml) file for a real world example.
