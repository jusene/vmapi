package module

import (
	"fmt"
	"libvirt.org/libvirt-go"
)

func LibvirtConnect(targetHost string) (*libvirt.Connect, error) {
	conn, err := libvirt.NewConnect(fmt.Sprintf("qemu+tcp://%s:16509/system", targetHost))
	return conn, err
}
