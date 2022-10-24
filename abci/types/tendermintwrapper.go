package types

import (
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type RequestInfo struct {
	abciTypes.RequestInfo

	ChainID   int
	ChainType ChainType
}

type RequestCheckTx struct {
	abciTypes.RequestCheckTx

	ChainID   int
	ChainType ChainType
}

type RequestInitChain struct {
}
