-include .env

.PHONY: run

run:
	@go run . \
		--file test_data/test_linux_arm64.deb \
		--bucket=apt \
		--access-key=$(MINIO_ACCESS_KEY_ID) \
		--secret-key=$(MINIO_SECRET_ACCESS_KEY) \
		--secure=false \
		--endpoint=localhost:9000 \
		--component=main \
		--origin="Apt Repository" \
		--label="Apt Repo" \
		--arch=arm64 \
		--archive=stable


run.s3:
	@go run . \
		--file dist/aptforge_0.0.3_linux_arm64.deb \
		--bucket=aircast-apt \
		--access-key=$(AWS_ACCESS_KEY_ID) \
		--secret-key=$(AWS_SECRET_ACCESS_KEY) \
		--secure=false \
		--endpoint=s3.amazonaws.com \
		--component=main \
		--origin="Apt Repository" \
		--label="Apt Repo" \
		--arch=arm64 \
		--archive=stable


test:
	@echo "Testing..."
	grc go test ./... -v



# Use := instead of = for variable assignment
ACTIONS_RUNTIME_URL := http://host.docker.internal:4322/
ARTIFACT_SERVER_ADDR := [::0]
ARTIFACT_SERVER_PORT := 4322
ARTIFACT_SERVER_PATH := out
GITHUB_TOKEN := $(GITHUB_TOKEN)

# Use .PHONY to declare targets that are not files
.PHONY: ci.piagent.build.stage ci.piagent.build.dev ci.piagent.release.dev ci.piagent.release.stage ci.piagent.release.prod


# Extract shared command to a single one
define act_command
	ACTIONS_RUNTIME_URL=$(ACTIONS_RUNTIME_URL) act \
		--artifact-server-addr "$(ARTIFACT_SERVER_ADDR)" \
		--artifact-server-port $(ARTIFACT_SERVER_PORT) \
		--artifact-server-path $(ARTIFACT_SERVER_PATH) \
		-j $(1) -s GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-j $(1) -s AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		-j $(1) -s AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		-W .github/workflows/$(2)
endef

ci.release: ; \
	$(call act_command,release,release.yml)
