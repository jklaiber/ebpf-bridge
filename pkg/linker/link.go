package linker

import "github.com/vishvananda/netlink"

type Link interface {
	GetIfaceNameForIndex(index int32) (string, error)
	GetIndexForIfaceName(name string) (int32, error)
}

type LinkerLinux struct{}

func (l *LinkerLinux) GetIfaceNameForIndex(index int32) (string, error) {
	iface, err := netlink.LinkByIndex(int(index))
	if err != nil {
		return "", err
	}
	return iface.Attrs().Name, nil
}

func (l *LinkerLinux) GetIndexForIfaceName(name string) (int32, error) {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return 0, err
	}
	return int32(iface.Attrs().Index), nil
}
