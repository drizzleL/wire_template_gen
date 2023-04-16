package iface

type UserGetter interface {
	GetUser() (user *User, err error)
}

type AddressGetter interface {
	GetAddress() (address *Address, err error)
}

type User struct{}

type Address struct{}
