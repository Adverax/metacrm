ROOT=../../../../../../../

.PHONY: rest
rest: ## Generate rest server
	oapi-codegen \
		-generate gin-server \
		-o ./http/server.gen.go \
		-package httpApi \
		${ROOT}/contracts/iam.yml
	oapi-codegen \
		-generate types \
		-o ./http/types.gen.go \
		-package httpApi \
		${ROOT}/contracts/iam.yml

.PHONY: grpc
grpc: ## Generate grpc server
	protoc -I ${ROOT}/contracts \
  	  --go_out=./grpc \
  	  --go_opt=paths=source_relative \
  	  --go-grpc_out=./grpc \
	  --go-grpc_opt=paths=source_relative \
	  ${ROOT}/contracts/iam-permissions.proto
	protoc -I ${ROOT}/contracts \
      	  --go_out=./grpc \
      	  --go_opt=paths=source_relative \
      	  --go-grpc_out=./grpc \
    	  --go-grpc_opt=paths=source_relative \
    	  ${ROOT}/contracts/iam-permissions-sync.proto


.PHONY: hel
help: ## Display the commands
	@grep -hE '(^[a-zA-Z0-9_-]+:.*##.*$$)|(^##)' $(MAKEFILE_LIST) \
	| awk 'BEGIN{FS=":.*## "}{ \
		if ($$0 ~ /^##/) { gsub(/^## ?/,""); printf "\033[33m%s\033[0m\n", $$0; } \
		else { printf "\033[32m%-30s\033[0m %s\n", $$1, $$2; } \
	}'

.DEFAULT_GOAL = help
