#!/bin/sh

DEVICE="$1"

if [ -z "$DEVICE" ]; then
	echo "Dummy no device given"
	exit 1
fi

cleanup () {
	trap "" INT QUIT TERM EXIT
	echo "Dummy stopping ..."
	echo "Dummy done."
	exit
}
trap "cleanup" INT QUIT TERM EXIT

while true; do
	echo "Dummy hotspot setup ${DEVICE}"
	sleep 5
done
