#!/bin/sh

DEVICE="$1"
PSKFILE="$2"
SSID="spreedbox"
HOSTNAME="spreedbox.local"
NETWORK_PREFIX="192.168.43"

if [ -z "${DEVICE}" ]; then
	echo "No device given"
	exit 1
fi

LEDCONTROL=$(which ledcontrol)

set -e

XUDNSD=$(which xudnsd)
HOSTAPD=$(which hostapd)
UDHCPD=$(which udhcpd)

TMPDIR=$(mktemp -d)
XUDNSD_PID=
UDHCPD_PID=

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

cleanup () {
	trap "" INT QUIT TERM EXIT
	echo "hotspot stopping ..."
	if [ -e hostapd.pid ]; then
		kill -TERM $(cat hostapd.pid) 2>/dev/null || true
	fi
	kill -TERM ${UDHCPD_PID} 2>/dev/null || true
	kill -TERM ${XUDNSD_PID} 2>/dev/null || true
	flushdevice
	rm -rf ${TMPDIR}
	ledsignal off
	echo "hotspot done."
	exit
}
trap "cleanup" INT QUIT TERM EXIT

flushdevice () {
	echo "hotspot flushing device ${DEVICE} ..."
	ifconfig ${DEVICE} down || true
	ip addr flush dev ${DEVICE} || true
}

startdevice () {
	echo "hotspot starting device ${DEVICE} ..."
	wpa_action ${DEVICE} down || true
	ifdown --force ${DEVICE} || true
	sleep 2
	ifconfig ${DEVICE} ${NETWORK_PREFIX}.1/24 up
}

nextcloud () {
	if [ -x /etc/network/if-up.d/spreedbox-nextcloud-ifup ]; then
		echo "hotspot updating Nextcloud settings ..."
		/etc/network/if-up.d/spreedbox-nextcloud-ifup || true
	fi
}

xudnsd () {
	echo "hotspot starting xudnsd ..."
	${XUDNSD} -ip=${NETWORK_PREFIX}.1 -name=${HOSTNAME}. &
	XUDNSD_PID=$!
}

hostapd () {
	cat >hostapd.conf <<EOL
interface=${DEVICE}
ssid=${SSID}
driver=rtl871xdrv
hw_mode=g
channel=11
auth_algs=3
wmm_enabled=1
ap_isolate=1
EOL
	if [ -n "${PSKFILE}" ]; then
		cat >>hostapd.conf <<EOL
wpa=2
wpa_key_mgmt=WPA-PSK
rsn_pairwise=CCMP
wpa_psk_file=${PSKFILE}
EOL
	fi
	echo "hotspot starting hostapd ..."
	${HOSTAPD} -B -P hostapd.pid hostapd.conf
}

udhcpd () {
	cat >udhcpd.conf <<EOL
start		${NETWORK_PREFIX}.20
end			${NETWORK_PREFIX}.254
interface	${DEVICE}
lease_file	udhcpd.leases
auto_time	0
opt		domain	local
opt		subnet	255.255.255.0
opt		router	${NETWORK_PREFIX}.1
opt		dns		${NETWORK_PREFIX}.1
EOL
	echo "hotspot starting udhcpd ..."
	touch udhcpd.leases
	${UDHCPD} -f udhcpd.conf &
	UDHCPD_PID=$!
}

cd ${TMPDIR}

ledsignal on
flushdevice
startdevice
nextcloud
xudnsd
hostapd
udhcpd

echo "hotspot running ${UDHCPD_PID} ..."
wait ${UDHCPD_PID}
