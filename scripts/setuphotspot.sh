#!/bin/sh

set -e

DEVICE="$1"
SSID="spreedbox"
HOSTNAME="spreedbox.local"
NETWORK_PREFIX="192.168.43"

if [ "$#" -ne 1 ]; then
	echo "No device given"
	exit 1
fi

XUDNSD=$(which xudnsd)
HOSTAPD=$(which hostapd)
UDHCPD=$(which udhcpd)

TMPDIR=$(mktemp -d)
XUDNSD_PID=
UDHCPD_PID=

cleanup () {
	trap "" INT QUIT TERM EXIT
	echo "Stopping ..."
	if [ -e hostapd.pid ]; then
		kill -TERM $(cat hostapd.pid) 2>/dev/null || true
	fi
	kill -TERM ${UDHCPD_PID} 2>/dev/null || true
	kill -TERM ${XUDNSD_PID} 2>/dev/null || true
	flushdevice
	restartdevice
	rm -rf ${TMPDIR}
	echo "Done."
}
trap "cleanup" INT QUIT TERM EXIT

flushdevice () {
	echo "Flushing device ${DEVICE} ..."
	ifconfig ${DEVICE} down || true
	ip addr flush dev ${DEVICE}
}

restartdevice() {
	echo "Restarting device ${DEVICE} ..."
	sleep 2
	ifup ${DEVICE} || true
}

startdevice () {
	echo "Starting device ${DEVICE} ..."
	ifdown --force ${DEVICE} || true
	sleep 2
	ifconfig ${DEVICE} ${NETWORK_PREFIX}.1/24 up
}

xudnsd () {
	echo "Starting xudnsd ..."
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
	echo "Starting hostapd ..."
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
	echo "Starting udhcpd ..."
	touch udhcpd.leases
	${UDHCPD} -f udhcpd.conf &
	UDHCPD_PID=$!
}

cd ${TMPDIR}

flushdevice
startdevice
xudnsd
hostapd
udhcpd

echo "Running ${UDHCPD_PID} ..."
wait ${UDHCPD_PID}