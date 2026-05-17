.SILENT:

lint:
	golangci-lint run ./...

unit:
	go test -race -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html

tests:
	@for dir in ./testdata/*/; do \
		echo "Testing $$dir"; \
		go run . -events "$$dir/input" -config "$$dir/config.json" | diff -y --suppress-common-lines "$$dir/output" -; \
		echo; \
	done