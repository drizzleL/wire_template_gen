// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package template

import (
	"github.com/drizzle/wire_template_gen/iface"
	"github.com/drizzle/wire_template_gen/impl"
	"github.com/google/wire"
)

// Injectors from m_wire.go:

func NewUserGetterImpl() *impl.UserGetterImpl {
	userGetterImpl := &impl.UserGetterImpl{}
	return userGetterImpl
}

func NewAddressGetterImpl() *impl.AddressGetterImpl {
	addressGetterImpl := &impl.AddressGetterImpl{}
	return addressGetterImpl
}

// m_wire.go:

var NewUserGetter = wire.NewSet(wire.Bind(new(iface.UserGetter), new(*impl.UserGetterImpl)), NewUserGetterImpl)

var NewAddressGetter = wire.NewSet(wire.Bind(new(iface.AddressGetter), new(*impl.AddressGetterImpl)), NewAddressGetterImpl)
