package abci

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/Hnampk/tendermint-client/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
	dbm "github.com/tendermint/tm-db"
)

type Application struct {
	abciTypes.Application
	*baseapp.BaseApp
}

var (
	app *Application
)

func init() {
	if app != nil {
		app = &Application{}
	}
}

func GetApp() *Application {
	return app
}

func NewBaseApp(
	name string, logger log.Logger, db dbm.DB, txDecoder sdk.TxDecoder, options ...func(*baseapp.BaseApp),
) {
	bapp := baseapp.NewBaseApp(name, logger, db, txDecoder, options...)
	app.BaseApp = bapp

}

// ----- Info/Query Connection

// Info return application info
func (a *Application) Info(requestInfo types.RequestInfo) abciTypes.ResponseInfo {
	switch requestInfo.ChainType {
	case types.ChainType_Cosmos:
		return a.BaseApp.Info(requestInfo.RequestInfo)
	default:
	}

	return abciTypes.ResponseInfo{
		Data:             "",
		Version:          "",
		AppVersion:       0,
		LastBlockHeight:  0,
		LastBlockAppHash: []byte{},
	}
}

// SetOption set application option
func (a *Application) SetOption(requestSetOption abciTypes.RequestSetOption) abciTypes.ResponseSetOption {
	return abciTypes.ResponseSetOption{
		Code: 0,
		Log:  "",
		Info: "",
	}
}

// Query query for state
func (a *Application) Query(requestQuery abciTypes.RequestQuery) abciTypes.ResponseQuery {

	return abciTypes.ResponseQuery{
		Code:      0,
		Log:       "",
		Info:      "",
		Index:     0,
		Key:       []byte{},
		Value:     []byte{},
		ProofOps:  &crypto.ProofOps{},
		Height:    0,
		Codespace: "",
	}
}

// ----- Mempool Connection

// CheckTx validate a tx for the mempool
func (a *Application) CheckTx(requestInfo types.RequestCheckTx) abciTypes.ResponseCheckTx {
	switch requestInfo.ChainType {
	case types.ChainType_Cosmos:
		return a.BaseApp.CheckTx(requestInfo.RequestCheckTx)
	default:
	}

	return abciTypes.ResponseCheckTx{}
}

// ----- Consensus Connection

// InitChain initialize blockchain w validators/other info from TendermintCore
func (a *Application) InitChain(requestInitChain abciTypes.RequestInitChain) abciTypes.ResponseInitChain {
	return abciTypes.ResponseInitChain{}
}

// BeginBlock signals the beginning of a block
func (a *Application) BeginBlock(requestBeginBlock abciTypes.RequestBeginBlock) abciTypes.ResponseBeginBlock {
	return abciTypes.ResponseBeginBlock{}
}

// DeliverTx deliver a tx for full processing
func (a *Application) DeliverTx(requestDeliverTx abciTypes.RequestDeliverTx) abciTypes.ResponseDeliverTx {
	return abciTypes.ResponseDeliverTx{}
}

// EndBlock signals the end of a block, returns changes to the validator set
func (a *Application) EndBlock(requestEndBlock abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	return abciTypes.ResponseEndBlock{}
}

// Commit commit the state and return the application Merkle root hash
func (a *Application) Commit() abciTypes.ResponseCommit {
	return abciTypes.ResponseCommit{
		Data:         []byte{},
		RetainHeight: 0,
	}
}

// ----- State Sync Connection

// ListSnapshots list available snapshots
func (a *Application) ListSnapshots(requestListSnapshots abciTypes.RequestListSnapshots) abciTypes.ResponseListSnapshots {
	return abciTypes.ResponseListSnapshots{}
}

// OfferSnapshot offer a snapshot to the application
func (a *Application) OfferSnapshot(requestOfferSnapshot abciTypes.RequestOfferSnapshot) abciTypes.ResponseOfferSnapshot {
	return abciTypes.ResponseOfferSnapshot{}
}

// LoadSnapshotChunk load a snapshot chunk
func (a *Application) LoadSnapshotChunk(requestLoadSnapshotChunk abciTypes.RequestLoadSnapshotChunk) abciTypes.ResponseLoadSnapshotChunk {
	return abciTypes.ResponseLoadSnapshotChunk{}
}

// ApplySnapshotChunk apply a shapshot chunk
func (a *Application) ApplySnapshotChunk(requestApplySnapshotChunk abciTypes.RequestApplySnapshotChunk) abciTypes.ResponseApplySnapshotChunk {
	return abciTypes.ResponseApplySnapshotChunk{}
}
