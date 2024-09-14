
.PHONY: run

run:
	@go run . \
		--file test_data/test_linux_arm64.deb \
		--bucket=apt \
		--access-key=REDACTED \
		--secret-key=REDACTED \
		--secure=false \
		--endpoint=localhost:9000 \
		--component=main \
		--origin="Apt Repository" \
		--label="Apt Repo" \
		--arch=arm64 \
		--archive=stable

test:
	@echo "Testing..."
	grc go test ./... -v
