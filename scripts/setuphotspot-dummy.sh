#!/bin/sh

DEVICE="$1"
PSKFILE="$2"

if [ -z "${DEVICE}" ]; then
	echo "Dummy no device given"
	exit 1
fi

if [ -n "${PSKFILE}" ]; then
	if [ ! -e "${PSKFILE}" ]; then
		echo "PSK file not found at ${PSKFILE}"
		exit 2
	fi
fi

cleanup () {
	trap "" INT QUIT TERM EXIT
	echo "Dummy stopping ..."
	echo "Dummy done."
	exit
}
trap "cleanup" INT QUIT TERM EXIT

while true; do
	echo "Dummy hotspot running with ${DEVICE}"
	sleep 5
done
