REPO_ROOT:=$(shell git rev-parse --show-toplevel

help:
	@echo "This is a helper makefile for oapi-codegen"
	@echo "Targets:"
	@echo "    generate:    regenerate all generated files"
	@echo "    test:        run all tests"
	@echo "    gin_example  generate gin example server code"
	@echo "    tidy         tidy go mod"

boilerplate:: ##@local Generates a swagger client and docs if a swagger file exists
ifneq ( oa-sis-oas.yaml ,"")
	@echo "Generating client/server/models for os-sis"
	@[[ -d "./gen" ]] || mkdir "./gen"

	@echo "- Removing any old client/server/models"
	@rm -rf ./gen/oa-sis
	@mkdir ./gen/oa-sis

	@echo "- Generating the client/server/models"
	@oapi-codegen -package OAsis -generate client ${REPO_ROOT}/pkg/oa-sis-oas.yaml > ${REPO_ROOT}/pkg/gen/OAsis/oa-sis.client.gen.go
	@oapi-codegen -package OAsis -generate types ${REPO_ROOT}/pkg/oa-sis-oas.yaml > ${REPO_ROOT}/pkg/gen/OAsis/oa-sis.types.gen.go
	@oapi-codegen -package OAsis -generate server,spec ${REPO_ROOT}/pkg/oa-sis-oas.yaml > ${REPO_ROOT}/pkg/gen/OAsis/oa-sis.server.gen.go

	@echo "+ Complete."
else
	@echo "No oas spec yaml found. Should be of the format <service_name>-oas.yaml"
endif

generate:
	go generate ./...

test:
	go test -cover ./...

tidy:
	@echo "tidy..."
	go mod tidy
