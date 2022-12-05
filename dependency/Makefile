.PHONY: test retrieve

retrieve:
	@cd retrieval; \
	go run main.go \
		--buildpack-toml-path "${buildpackTomlPath}" \
		--output "${output}"

test:
	@cd test; \
	./test.sh \
		--tarballPath ${tarballPath} \
		--expectedVersion ${version}
