.PHONY: install

build:
	go build -o install/imagems-installer-version

install:
	cd "install" && sudo ./systemd-install.sh

uninstall:
	cd "install" && sudo ./systemd-uninstall.sh

