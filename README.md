Spreedbox wlan
==============

Spreedbox wlan is a daemon providing wifi configuration API and events vi NATS to configure wifi network interfaces. Integrates with `spreedbox-network` for interface control and events.

## Automatic Wi-Fi hotspot

To allow configuration with Wi-Fi only, Spreedbox wlan supports starting a wi-fi
hotspot which can then be used to setup the Spreedbox and join an existing wi-fi
network.

The wi-fi hotspot starts under the following conditions:

 - Wi-Fi interface is available.
 - There is no link on any device.
 - There never was a link on any device for longer than 30 seconds since the
   last boot (link-seen flag).
 - Network configuration not changed since one minute (not fresh).
 - The other conditions have not changed for one minute.

This usually means, the Wi-Fi hotspot automatically starts 1 minute after a
fresh boot, if the configured Wi-Fi cannot be joined / none is configured and an
Ethernet link is not detected.

Once the Wi-Fi hotspot is running, the Spreedbox will no longer try to join the
configured Wi-Fi network. The hotspot will keep running until the network
configuration is changed / reapplied or spreedbox-wlan is restarted (eg. by a
reboot).

The Wi-Fi hotspot will automatically stop as soon as any of the above conditions
change.

If the Wi-Fi interface is removed while the hotspot is active, the hotspot will
be stopped and scheduled to be restarted once per minute. This means if the
device becomes available again and none of the conditions above have changed,
the hotspot starts again after at most 1 minute after device availabilty.

To connect to the Spreedbox web interface when using the Wi-Fi hotspot, open up
a web browser and go to https://spreedbox/ or https://192.168.43.1/.

### Wi-Fi hotspot network

The Wi-Fi hotspot provides its own network (192.168.43.1/24) with DHCP and DNS.
The Spreedbox is available by name `spreedbox` and by IP 192.168.43.1.

The network is visible and with network name (SSID) is `spreedbox` on channel 11
(IEEE 802.11g (2.4 GHz)).

The only device which can be reached when connected with the Spreedbox hotspot
is the Spreedbox itself.

### Wi-Fi hotspot security

The Wi-Fi hotspot network is encrypted with a hardware specific device password
using WPA2-PSK (CCMP) security.

The password is printed on the Spreedbox label. You can also show the hardware
Wi-Fi password with `sudo wlanctl defaultpassword` on the Spreedbox (SSH).

## LED signaling Wi-Fi hotspot

Whenever the Wi-Fi hotspot is started, the `wlan-hotspot` LED state is active to
indicate Wi-Fi hotspot availability.

## License

This package uses the AGPL license, see the `LICENSE` file.
