package skynet

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"skynet/x/skynet/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "ListMerchant",
					Use:       "list-merchant",
					Short:     "List all merchant",
				},
				{
					RpcMethod:      "GetMerchant",
					Use:            "get-merchant [id]",
					Short:          "Gets a merchant",
					Alias:          []string{"show-merchant"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				{
					RpcMethod: "ListPayment",
					Use:       "list-payment",
					Short:     "List all payment",
				},
				{
					RpcMethod:      "GetPayment",
					Use:            "get-payment [id]",
					Short:          "Gets a payment by id",
					Alias:          []string{"show-payment"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "CreateMerchant",
					Use:            "create-merchant [index] [name] [category] [active]",
					Short:          "Create a new merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}, {ProtoField: "name"}, {ProtoField: "category"}, {ProtoField: "active"}},
				},
				{
					RpcMethod:      "UpdateMerchant",
					Use:            "update-merchant [index] [name] [category] [active]",
					Short:          "Update merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}, {ProtoField: "name"}, {ProtoField: "category"}, {ProtoField: "active"}},
				},
				{
					RpcMethod:      "DeleteMerchant",
					Use:            "delete-merchant [index]",
					Short:          "Delete merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "index"}},
				},
				{
					RpcMethod:      "PayMerchant",
					Use:            "pay-merchant [merchant] [amount] [denom] [memo]",
					Short:          "Send a pay-merchant tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "merchant"}, {ProtoField: "amount"}, {ProtoField: "denom"}, {ProtoField: "memo"}},
				},
			},
		},
	}
}
