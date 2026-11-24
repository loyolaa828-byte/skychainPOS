package keeper

import (
	"context"
	"fmt"

	"skynet/x/skynet/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (k msgServer) PayMerchant(ctx context.Context, msg *types.MsgPayMerchant) (*types.MsgPayMerchantResponse, error) {
	payerBytes, err := k.addressCodec.StringToBytes(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid payer address")
	}

	merchantBytes, err := k.addressCodec.StringToBytes(msg.Merchant)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid merchant address")
	}

	if msg.Amount == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "payment amount must be greater than zero")
	}

	merchant, err := k.Merchant.Get(ctx, msg.Merchant)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrNotFound, "merchant %s is not registered", msg.Merchant)
	}
	if !merchant.Active {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "merchant %s is inactive", msg.Merchant)
	}

	denom := msg.Denom
	if denom == "" {
		denom = "usky"
	}

	payerAddr := sdk.AccAddress(payerBytes)
	merchantAddr := sdk.AccAddress(merchantBytes)

	// Tokenomics split: 98.5% Merchant, 1.0% Staker Yield (fee_collector), 0.5% Burned
	stakerCut := msg.Amount / 100
	burnCut := msg.Amount / 200
	merchantCut := msg.Amount - stakerCut - burnCut

	if merchantCut > 0 {
		merchantCoins := sdk.NewCoins(sdk.NewCoin(denom, math.NewIntFromUint64(merchantCut)))
		if err := k.bankKeeper.SendCoins(ctx, payerAddr, merchantAddr, merchantCoins); err != nil {
			return nil, errorsmod.Wrap(err, "failed to pay merchant")
		}
	}

	if stakerCut > 0 {
		stakerCoins := sdk.NewCoins(sdk.NewCoin(denom, math.NewIntFromUint64(stakerCut)))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payerAddr, authtypes.FeeCollectorName, stakerCoins); err != nil {
			return nil, errorsmod.Wrap(err, "failed to distribute staker yield")
		}
	}

	if burnCut > 0 {
		burnCoins := sdk.NewCoins(sdk.NewCoin(denom, math.NewIntFromUint64(burnCut)))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payerAddr, types.ModuleName, burnCoins); err != nil {
			return nil, errorsmod.Wrap(err, "failed to escrow burn coins")
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
			return nil, errorsmod.Wrap(err, "failed to burn supply")
		}
	}

	paymentID, err := k.PaymentSeq.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get next payment ID")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	payment := types.Payment{
		Id:        paymentID,
		Merchant:  msg.Merchant,
		Payer:     msg.Creator,
		Amount:    msg.Amount,
		Denom:     denom,
		Memo:      msg.Memo,
		Timestamp: sdkCtx.BlockTime().Unix(),
	}

	if err := k.Payment.Set(ctx, paymentID, payment); err != nil {
		return nil, errorsmod.Wrap(err, "failed to store payment receipt")
	}

	totalCoins := sdk.NewCoins(sdk.NewCoin(denom, math.NewIntFromUint64(msg.Amount)))
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"pos_payment",
			sdk.NewAttribute("payment_id", fmt.Sprintf("%d", paymentID)),
			sdk.NewAttribute("merchant", msg.Merchant),
			sdk.NewAttribute("payer", msg.Creator),
			sdk.NewAttribute("amount", totalCoins.String()),
			sdk.NewAttribute("staker_yield", fmt.Sprintf("%d%s", stakerCut, denom)),
			sdk.NewAttribute("burned", fmt.Sprintf("%d%s", burnCut, denom)),
			sdk.NewAttribute("memo", msg.Memo),
		),
	)

	return &types.MsgPayMerchantResponse{PaymentId: paymentID}, nil
}
