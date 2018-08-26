PROVIDER_NAME=terraform-provider-vspheretemplate
PROVIDER_INSTALL_DIR=$(HOME)/.terraform.d/plugins

default: build

build:
	go build -o $(PROVIDER_NAME)

install:
	mkdir -p $(PROVIDER_INSTALL_DIR)
	go build -o $(PROVIDER_INSTALL_DIR)/$(PROVIDER_NAME)

uninstall:
	rm $(PROVIDER_INSTALL_DIR)/$(PROVIDER_NAME)
