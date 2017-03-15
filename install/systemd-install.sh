#!/usr/bin/env bash

# TODO SEEDMS change seedms the microservice's name (prefer it similar to name constant in main.go)
NAME="seedms"
BUILD_NAME="${NAME}-installer-version"
CONF_DIR="/etc/${NAME}"
CONF_FILE="${CONF_DIR}/${NAME}.conf.yaml"
INSTALL_DIR="/usr/local/bin"
UNIT_FILE="/etc/systemd/system/${NAME}.service"

./systemd-uninstall.sh || exit 1
echo "Begin install"
mkdir -p "${CONF_DIR}" || exit 1
if [ ! -f "${CONF_FILE}" ]; then
    cp "${NAME}.conf.yaml" "${CONF_FILE}" || exit 1
fi
mkdir -p "${INSTALL_DIR}" || exit 1
cp -f "${BUILD_NAME}" "${INSTALL_DIR}/${NAME}" || exit 1
cp -f "${NAME}.service" "${UNIT_FILE}" || exit 1
systemctl enable "${NAME}.service"
echo "Config file is at '${CONF_FILE}'"
echo "Install complete"
