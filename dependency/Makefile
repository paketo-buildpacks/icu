.PHONY: retrieve

retrieve:
	@go run ./retrieval \
		--buildpack-toml-path "${buildpackTomlPath}" \
		--output "${output}"

test:
	@cd test; \
	./test.sh \
		--tarballPath $(tarballPath) \
		--expectedVersion $(version)
