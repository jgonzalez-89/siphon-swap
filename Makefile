# Service name:
name := cryptoswap

# Tags:
version := latest

# Paths:
swagger_path := ./docs/api

# Libraries:
codegen := github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen

# Versions:
codegen_version := v2.4.1

build:
	mkdir -p build
	go build -o build/${name} cmd/main.go

docker.clean:
	docker rmi -f ${name}:${version}

docker.build: docker.clean
	docker build -t ${name}:${version} .

down:
	docker compose down --remove-orphans

up: down
	docker compose up -d

dependencies:
	go get -tool ${codegen}@${codegen_version}

generate:
	go generate ./...
	$(call gen_all,handlers)

define gen_protocols
	$(eval $@_RESOURCE = $(1))
	go run ${codegen} \
		--package ${$@_RESOURCE} \
		--generate types \
		-o ./internal/transport/handlers/${$@_RESOURCE}/${$@_RESOURCE}_protocol_gen.go \
		${swagger_path}/${$@_RESOURCE}/$(subst _,-,${$@_RESOURCE}).yml
endef

define gen_handlers
	$(eval $@_RESOURCE = $(1))
	go run ${codegen} \
		--package ${$@_RESOURCE} \
		--config docs/configs/handlers-gen.yml \
		-o ./internal/transport/handlers/${$@_RESOURCE}/${$@_RESOURCE}_handler_gen.go \
		${swagger_path}/${$@_RESOURCE}/$(subst _,-,${$@_RESOURCE}).yml
endef

define gen_all
	$(eval $@_RESOURCE = $(1))
	@$(call gen_protocols,${$@_RESOURCE})
	@$(call gen_handlers,${$@_RESOURCE})
endef
