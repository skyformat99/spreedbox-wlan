#!/bin/sh

DEVICE="$1"

if [ "$#" -ne 1 ]; then
	echo "No device given"
	exit 1
fi

while true; do
	echo "Hotspot setup dummy ${DEVICE}"
	sleep 5
done
