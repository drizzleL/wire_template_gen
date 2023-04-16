package template

import (
	"github.com/drizzle/wire_template_gen/iface"
	"github.com/drizzle/wire_template_gen/impl"
	"github.com/google/wire"
)

var NewUserGetter = wire.NewSet(
	wire.Bind(new(iface.UserGetter), new(*impl.UserGetterImpl)),
	NewUserGetterImpl,
)

var NewAddressGetter = wire.NewSet(
	wire.Bind(new(iface.AddressGetter), new(*impl.AddressGetterImpl)),
	NewAddressGetterImpl,
)

func NewUserGetterImpl() *impl.UserGetterImpl {
	panic(wire.Build(
		wire.Struct(new(impl.UserGetterImpl), "*"),
	))
}
func NewAddressGetterImpl() *impl.AddressGetterImpl {
	panic(wire.Build(
		wire.Struct(new(impl.AddressGetterImpl), "*"),
	))
}
