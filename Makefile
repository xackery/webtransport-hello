NAME := hello

run: cert-gen
	@echo "Running $(NAME)..."
	cd bin && go run ../app/client/main.go cacert.pem cakey.pem

run-server: cert-gen
	@echo "Running $(NAME)..."
	cd bin && go run ../app/server/main.go cacert.pem cakey.pem

cert-gen:
	@mkdir -p bin
ifeq (,$(wildcard ./bin/cacert.pem))
	@echo "Generating certs..."
	openssl genrsa -out bin/cakey.pem 2048
	openssl req -new -x509 -key bin/cakey.pem -out bin/cacert.pem -days 1095 -subj "/C=US/ST=CA/L=San Francisco/O=My Company/OU=Org/CN=example.com"
endif