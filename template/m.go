package template

import (
	"github.com/drizzle/wire_template_gen/iface"
	"github.com/drizzle/wire_template_gen/impl"
)

// 在这里我们绑定接口与实现
// inf_mapping
var (
	_ iface.UserGetter    = (*impl.UserGetterImpl)(nil)
	_ iface.AddressGetter = (*impl.AddressGetterImpl)(nil)
)
