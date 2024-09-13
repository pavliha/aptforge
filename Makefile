
.PHONY: run

run:
	@go run . \
		--file test_data/test_linux_arm64.deb \
		--bucket apt \
		--access-key REDACTED \
		--secret-key REDACTED \
		--endpoint localhost:9000 \
		--component main \
		--origin "Custom Repository" \
		--label "Custom Repo" \
		--arch arm64 \
		--archive stable
