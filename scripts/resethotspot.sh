#!/bin/sh

LEDCONTROL=$(which ledcontrol)

set -e

ledsignal() {
	if [ ! -x "$LEDCONTROL" ]; then
		return
	fi
	local args="del"
	if [ "$1" = "on" ]; then
		args="preset wlan-hotspot"
	fi
	$LEDCONTROL -id="spreedbox-wlan-hotspot" -slot=2 $args || true
}

ledsignal off

