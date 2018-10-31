package iservice

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgSvcDef{}, "iris-hub/iservice/MsgSvcDef", nil)
	cdc.RegisterConcrete(MsgSvcBind{}, "iris-hub/iservice/MsgSvcBinding", nil)
}

var msgCdc = wire.NewCodec()

func init() {
	RegisterWire(msgCdc)
}
