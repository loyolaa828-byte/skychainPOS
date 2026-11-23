package keeper

import (
	"context"

	"skynet/x/skynet/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) PayMerchant(ctx context.Context, msg *types.MsgPayMerchant) (*types.MsgPayMerchantResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgPayMerchantResponse{}, nil
}
