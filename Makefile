SKAFFOLD_DEFAULT_REPO ?= ghcr.io/effxhq

deploy: .deploy
.deploy:
	@env SKAFFOLD_DEFAULT_REPO=$(SKAFFOLD_DEFAULT_REPO) skaffold run $(SKAFFOLD_ARGS)

cleanup:
	@env SKAFFOLD_DEFAULT_REPO=$(SKAFFOLD_DEFAULT_REPO) skaffold delete $(SKAFFOLD_ARGS)
