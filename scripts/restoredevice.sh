#!/bin/sh

DEVICE="$1"

if [ -z "${DEVICE}" ]; then
	echo "No device given"
	exit 1
fi

set -e

ifup ${DEVICE} || true
