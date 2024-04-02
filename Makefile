run:
	@echo "Running main application"
	@go run ./cmd/spring-initializer/

compile-current:
	@echo "Compiling for current OS and Platform"
	@go build -o ./bin/spring-initializer ./cmd/spring-initializer/

compile-linux-386:
	@echo "Compiling for linux-386"
	@GOOS=linux GOARCH=386 go build -o bin/spring-initializer-linux-386 ./cmd/spring-initializer/

compile-linux-amd64:
	@echo "Compiling for linux-amd64"
	@GOOS=linux GOARCH=amd64 go build -o bin/spring-initializer-linux-amd64 ./cmd/spring-initializer/

compile-windows-386:
	@echo "Compiling for windows-386"
	@GOOS=windows GOARCH=386 go build -o bin/spring-initializer-windows-386 ./cmd/spring-initializer/

compile-windows-amd64:
	@echo "Compiling for windows-amd64"
	@GOOS=windows GOARCH=amd64 go build -o bin/spring-initializer-windows-amd64 ./cmd/spring-initializer/

compile-freebsd-386:
	@echo "Compiling for freebsd-386"
	@GOOS=freebsd GOARCH=386 go build -o bin/spring-initializer-freebsd-386 ./cmd/spring-initializer/

compile-freebsd-amd64:
	@echo "Compiling for freebsd-amd64"
	@GOOS=freebsd GOARCH=amd64 go build -o bin/spring-initializer-freebsd-amd64 ./cmd/spring-initializer/

compile-all: compile-linux-386 compile-linux-amd64 compile-windows-386 compile-windows-amd64 compile-freebsd-386 compile-freebsd-amd64
