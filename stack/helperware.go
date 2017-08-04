//nolint
package stack

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/state"
)

const (
	NameCheck = "check"
	NameGrant = "grant"
)

// CheckMiddleware returns an error if the tx doesn't have auth of this
// Required Actor, otherwise passes along the call untouched
type CheckMiddleware struct {
	Required basecoin.Actor
	PassInitState
	PassInitValidate
}

var _ Middleware = CheckMiddleware{}

func (_ CheckMiddleware) Name() string {
	return NameCheck
}

func (p CheckMiddleware) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	if !ctx.HasPermission(p.Required) {
		return res, errors.ErrUnauthorized()
	}
	return next.CheckTx(ctx, store, tx)
}

func (p CheckMiddleware) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	if !ctx.HasPermission(p.Required) {
		return res, errors.ErrUnauthorized()
	}
	return next.DeliverTx(ctx, store, tx)
}

// GrantMiddleware tries to set the permission to this Actor, which may be prohibited
type GrantMiddleware struct {
	Auth basecoin.Actor
	PassInitState
	PassInitValidate
}

var _ Middleware = GrantMiddleware{}

func (_ GrantMiddleware) Name() string {
	return NameGrant
}

func (g GrantMiddleware) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Checker) (res basecoin.CheckResult, err error) {
	ctx = ctx.WithPermissions(g.Auth)
	return next.CheckTx(ctx, store, tx)
}

func (g GrantMiddleware) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.DeliverResult, err error) {
	ctx = ctx.WithPermissions(g.Auth)
	return next.DeliverTx(ctx, store, tx)
}