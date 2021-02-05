package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSetPrivate{}, "profile/SetPrivate", nil)
	cdc.RegisterConcrete(MsgSetPublic{}, "profile/SetPublic", nil)
}

// ModuleCdc is the codec for the module
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}
