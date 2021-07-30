DOCKER_REPOSITORY ?= ghcr.io/
DOCKER_IMAGE ?= effxhq/cluster-agent

docker:
	docker build . -t $(DOCKER_REPOSITORY)$(DOCKER_IMAGE)
