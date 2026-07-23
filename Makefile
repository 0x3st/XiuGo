ROOT_DIR    = $(shell pwd)
NAMESPACE   = "default"
DEPLOY_NAME = "template-single"
DOCKER_NAME = "template-single"

include ./hack/hack-cli.mk
include ./hack/hack.mk

# ---- XiuGo plugins (drop-in under plugins/) ----
.PHONY: plugins plugins-list new-plugin
plugins:
	go generate ./internal/plugin/registry

plugins-list:
	@ls -1 plugins 2>/dev/null | grep -v README || true

# usage: make new-plugin id=my_plugin
new-plugin:
	@test -n "$(id)" || (echo "usage: make new-plugin id=my_plugin" && exit 1)
	./hack/new-plugin.sh $(id)
	$(MAKE) plugins
