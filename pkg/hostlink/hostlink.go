//go:generate mockgen -source=hostlink.go -destination=mocks/hostlink_mock.go -package=mocks Link
package hostlink

import "github.com/vishvananda/netlink"

type NetlinkWrapper interface {
	LinkByIndex(index int) (netlink.Link, error)
	LinkByName(name string) (netlink.Link, error)
}

type DefaultNetlinkWrapper struct{}

func (d *DefaultNetlinkWrapper) LinkByIndex(index int) (netlink.Link, error) {
	return netlink.LinkByIndex(index)
}

func (d *DefaultNetlinkWrapper) LinkByName(name string) (netlink.Link, error) {
	return netlink.LinkByName(name)
}

type Link interface {
	Index() int
	Name() string
}

type LinkFactory interface {
	NewLinkWithIndex(index int) (Link, error)
	NewLinkWithName(name string) (Link, error)
}

type HostLink struct {
	index int
	name  string
}

func NewHostLinkFactory() LinkFactory {
	return &hostLinkFactory{
		nlw: &DefaultNetlinkWrapper{},
	}
}

type hostLinkFactory struct {
	nlw NetlinkWrapper
}

func (f *hostLinkFactory) NewLinkWithIndex(index int) (Link, error) {
	link, err := f.nlw.LinkByIndex(index)
	if err != nil {
		return nil, err
	}
	return &HostLink{
		index: index,
		name:  link.Attrs().Name,
	}, nil
}

func (f *hostLinkFactory) NewLinkWithName(name string) (Link, error) {
	link, err := f.nlw.LinkByName(name)
	if err != nil {
		return nil, err
	}
	return &HostLink{
		index: link.Attrs().Index,
		name:  name,
	}, nil
}

func (h *HostLink) Index() int {
	return h.index
}

func (h *HostLink) Name() string {
	return h.name
}
