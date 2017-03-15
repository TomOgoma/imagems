#!/usr/bin/env bash

NAME="imagems"
APP_FILE="/usr/local/bin/${NAME}"
UNIT_FILE="/etc/systemd/system/${NAME}.service"
CONF_DIR="/etc/${NAME}"
CONF_FILE="${CONF_DIR}/${NAME}.conf.yaml"
EXIT_CODE_FAIL=1

echo "Begin uninstall"
if [ -f "$UNIT_FILE" ]; then
	systemctl stop  "${NAME}.service" >/dev/null
	rm -f "${UNIT_FILE}" || exit ${EXIT_CODE_FAIL}
    systemctl daemon-reload
fi
if [ -f "$APP_FILE" ]; then
	rm -f "${APP_FILE}" || exit ${EXIT_CODE_FAIL}
fi
if [ -f "$CONF_FILE" ]; then
    echo "config file at '${CONF_FILE}' left intact intentionally"
fi
echo "Uninstall complete"
