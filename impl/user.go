package impl

import (
	"github.com/drizzle/wire_template_gen/iface"
)

type UserGetterImpl struct{}

func (UserGetterImpl) GetUser() (user *iface.User, err error) {
	panic("not implemented") // TODO: Implement
}

type AddressGetterImpl struct{}

func (AddressGetterImpl) GetAddress() (address *iface.Address, err error) {
	panic("not implemented") // TODO: Implement
}
