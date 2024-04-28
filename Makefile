run: build
	@./bin/typergame || true

build:
	@go build -o ./bin/typergame .
