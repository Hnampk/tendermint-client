package main

import (
	"fmt"
	"time"

	flogging "github.com/Hnampk/fabric-flogging"
	"github.com/Hnampk/tendermint-client/mock"
	"github.com/Hnampk/tendermint-client/tendermint"
	"github.com/Hnampk/tendermint-client/usecases/abciclient"
	tmconfig "github.com/tendermint/tendermint/config"
	cs "github.com/tendermint/tendermint/consensus"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmevidence "github.com/tendermint/tendermint/evidence"
	tmjson "github.com/tendermint/tendermint/libs/json"
	mempl "github.com/tendermint/tendermint/mempool"
	tmmempool "github.com/tendermint/tendermint/mempool"
	mmock "github.com/tendermint/tendermint/mempool/mock"
	tmnode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/privval"
	tmstate "github.com/tendermint/tendermint/proto/tendermint/state"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	tmproxy "github.com/tendermint/tendermint/proxy"
	sm "github.com/tendermint/tendermint/state"
	tmstore "github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	mainLogger = flogging.MustGetLogger("main")
)

func main() {
	// if err := cmd.RootCmd.Execute(); err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(-1)
	// }

	appProxy := connect()
	if appProxy == nil {
		mainLogger.Panic("connect failed")
	}

	config := tmconfig.DefaultConfig()
	config.BaseConfig.RootDir = "/home/nampkh/.tendermint/"

	genesisDocProvider := tmnode.DefaultGenesisDocProviderFunc(config)
	blockStore, stateDB, err := initDBs(config, tmnode.DefaultDBProvider)
	if err != nil {
		panic(err)
	}
	stateStore := sm.NewStore(stateDB)
	initialState := getInitialState()
	if err := stateStore.Save(initialState); err != nil {
		panic(err)
	}

	genDoc, err := loadGenesisDoc(stateDB)
	if err != nil {
		fmt.Println("heree", err.Error())
		genDoc, err = genesisDocProvider()
		if err != nil {
			panic(err)
		}
		// save genesis doc to prevent a certain class of user errors (e.g. when it
		// was changed, accidentally or not). Also good for audit trail.
		if err := saveGenesisDoc(stateDB, genDoc); err != nil {
			panic(err)
		}
	}

	eventBus, err := createAndStartEventBus()
	if err != nil {
		panic(err)
	}

	handshaker := tendermint.NewHandshaker(stateStore, initialState, blockStore, genDoc)
	handshaker.SetEventBus(eventBus)
	if err := handshaker.DoHandshake(appProxy); err != nil {
		panic(err)
	}

	fmt.Println(blockStore.Height())
	block0 := blockStore.LoadBlock(0)
	fmt.Println(block0)
	block1 := blockStore.LoadBlock(1)
	fmt.Println(block1)

	// newBlock(appProxy)
	firstBlock(appProxy, handshaker.LatestState, stateStore)

}

func getInitialState() sm.State {
	PrivValidatorKeyFile := "/home/nampkh/.tendermint/config/priv_validator_key.json"
	PrivValidatorStateFile := "/home/nampkh/.tendermint/data/priv_validator_state.json"
	filePV := privval.LoadOrGenFilePV(PrivValidatorKeyFile, PrivValidatorStateFile)
	pubkey, _ := filePV.GetPubKey()

	validator := &tmtypes.Validator{
		Address:          filePV.GetAddress(),
		PubKey:           pubkey,
		VotingPower:      100,
		ProposerPriority: 0,
	}

	return sm.State{
		Version: tmstate.Version{ // same between empty blocks
			Consensus: tmversion.Consensus{
				Block: 11,
				App:   0,
			},
			Software: "unreleased-freezed/v0.34.19-2f231ceb952a2426cf3c0abaf0b455aadd11e5b2",
		},
		ChainID:         "checkers", // same between empty blocks
		InitialHeight:   1,          // same between empty blocks
		LastBlockHeight: 0,
		LastBlockID: tmtypes.BlockID{
			Hash: []byte{},
			PartSetHeader: tmtypes.PartSetHeader{
				Total: 0,
				Hash:  []byte{},
			},
		},
		LastBlockTime: time.Date(2022, 11, 21, 15, 28, 49, 558699629, time.UTC),
		NextValidators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{validator},
			Proposer:   validator,
		},
		Validators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{validator},
			Proposer:   validator,
		},
		LastValidators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{},
			Proposer:   &tmtypes.Validator{},
		},
		LastHeightValidatorsChanged: 1, // same between empty blocks
		ConsensusParams: tmproto.ConsensusParams{ // same between empty blocks
			Block: tmproto.BlockParams{
				MaxBytes:   22020096,
				MaxGas:     -1,
				TimeIotaMs: 1000,
			},
			Evidence: tmproto.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  48 * time.Hour,
				MaxBytes:        1048576,
			},
			Validator: tmproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
			Version: tmproto.VersionParams{
				AppVersion: 0,
			},
		},
		LastHeightConsensusParamsChanged: 1,
		LastResultsHash:                  merkle.HashFromByteSlices(nil), // todo: find me
		AppHash:                          []byte{},                       // todo: find me
	}
}

func firstBlock(appProxy *abciclient.AppProxy, state sm.State, stateStore sm.Store) {
	blockHeight := int64(1)

	lastCommit := tmtypes.NewCommit(0, 0, tmtypes.BlockID{
		Hash: []byte{},
		PartSetHeader: tmtypes.PartSetHeader{
			Total: 0,
			Hash:  []byte{},
		},
	}, []tmtypes.CommitSig{})

	block, partSet := tendermint.MakeBlock(state, blockHeight, mock.MakeTxs(blockHeight), lastCommit, nil, state.Validators.GetProposer().Address)
	blockID := tmtypes.BlockID{Hash: block.Hash(), PartSetHeader: partSet.Header()}
	// thisParts := block.MakePartSet(tmtypes.BlockPartSizeBytes)
	// block.b
	fmt.Println(block)
	fmt.Println(blockID.IsZero())
	fmt.Println(block.LastBlockID)
	fmt.Println(block.LastBlockID.IsZero())
	fmt.Println(block.Hash())
	fmt.Println(state.AppHash)
	fmt.Println("1")

	blockExec := sm.NewBlockExecutor(stateStore, nil, appProxy.Consensus(),
		mmock.Mempool{}, sm.EmptyEvidencePool{})
	fmt.Println("2")
	state, retainHeight, err := blockExec.ApplyBlock(state, blockID, block)
	if err != nil {
		mainLogger.Errorf("error ApplyBlock: %s", err.Error())
		return
	}
	fmt.Println("3")

	mainLogger.Infof("state: %+v", state)
	mainLogger.Infof("retainHeight: %+v", retainHeight)

}

func NewConsensusState(config *tmconfig.Config, appProxy *abciclient.AppProxy, state sm.State, stateDB dbm.DB, blockStore *tmstore.BlockStore, stateStore sm.Store) {

	mempool := mempl.NewCListMempool(
		config.Mempool,
		appProxy.Mempool(),
		state.LastBlockHeight,
	)
	evidenceDB, err := tmnode.DefaultDBProvider(&tmnode.DBContext{"evidence", config})
	if err != nil {
		panic(err)
	}
	evidencePool, err := tmevidence.NewPool(evidenceDB, sm.NewStore(stateDB), blockStore)
	if err != nil {
		panic(err)
	}

	// make block executor for consensus and blockchain reactors to execute blocks
	blockExec := sm.NewBlockExecutor(
		stateStore,
		nil,
		appProxy.Consensus(),
		mempool,
		evidencePool,
	)
	consensusState := cs.NewState(
		config.Consensus,
		state.Copy(),
		blockExec,
		blockStore,
		mempool,
		evidencePool,
	)

	consensusState.OnStart()
}

func connect() *abciclient.AppProxy {
	appName := "checkers"
	if err := abciclient.NewGRPCAppProxy(appName, "tcp://0.0.0.0:1234"); err != nil {
		mainLogger.Errorf("error while NewGRPCAppProxy: %s", err.Error())
		return nil
	}

	checkersApp, err := abciclient.GetApp(appName)
	if err != nil {
		mainLogger.Errorf("error while GetApp %s: %s", err.Error())
		return nil
	}

	return checkersApp
}

func initDBs(config *tmconfig.Config, dbProvider tmnode.DBProvider) (blockStore *tmstore.BlockStore, stateDB dbm.DB, err error) {
	var blockStoreDB dbm.DB
	blockStoreDB, err = dbProvider(&tmnode.DBContext{"blockstore", config})
	if err != nil {
		return
	}
	blockStore = tmstore.NewBlockStore(blockStoreDB)

	stateDB, err = dbProvider(&tmnode.DBContext{"state", config})
	if err != nil {
		return
	}

	return
}

func saveGenesisDoc(db dbm.DB, genDoc *tmtypes.GenesisDoc) error {
	b, err := tmjson.Marshal(genDoc)
	if err != nil {
		return fmt.Errorf("failed to save genesis doc due to marshaling error: %w", err)
	}
	if err := db.SetSync(genesisDocKey, b); err != nil {
		return err
	}

	return nil
}

var (
	genesisDocKey = []byte("genesisDoc")
)

// panics if failed to unmarshal bytes
func loadGenesisDoc(db dbm.DB) (*tmtypes.GenesisDoc, error) {
	b, err := db.Get(genesisDocKey)
	if err != nil {
		panic(err)
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("genesis doc not found")
	}
	var genDoc *tmtypes.GenesisDoc
	err = tmjson.Unmarshal(b, &genDoc)
	if err != nil {
		panic(fmt.Sprintf("Failed to load genesis doc due to unmarshaling error: %v (bytes: %X)", err, b))
	}
	return genDoc, nil
}
func createAndStartEventBus() (*tmtypes.EventBus, error) {
	eventBus := tmtypes.NewEventBus()
	if err := eventBus.Start(); err != nil {
		return nil, err
	}
	return eventBus, nil
}

func newBlock(appProxy *abciclient.AppProxy, state sm.State) {
	latestHeight := int64(1914)
	state, stateDB, _ := mock.MakeState(1, latestHeight)
	stateStore := sm.NewStore(stateDB)

	blockExec := sm.NewBlockExecutor(stateStore, nil, appProxy.Consensus(),
		mmock.Mempool{}, sm.EmptyEvidencePool{})

	prevHash := state.LastBlockID.Hash
	prevParts := tmtypes.PartSetHeader{}
	prevBlockID := tmtypes.BlockID{Hash: prevHash, PartSetHeader: prevParts}
	lastCommitSigs := []tmtypes.CommitSig{}
	lastCommit := tmtypes.NewCommit(1, 0, prevBlockID, lastCommitSigs)

	block, _ := tendermint.MakeBlock(state, int64(latestHeight+1), mock.MakeTxs(0), lastCommit, nil, state.Validators.GetProposer().Address)

	blockID := tmtypes.BlockID{Hash: block.Hash(), PartSetHeader: block.MakePartSet(mock.TestPartSize).Header()}

	state, retainHeight, err := blockExec.ApplyBlock(state, blockID, block)
	if err != nil {
		mainLogger.Errorf("error ApplyBlock: %s", err.Error())
		return
	}

	mainLogger.Infof("state: %+v", state)
	mainLogger.Infof("retainHeight: %+v", retainHeight)
}

func secondState() sm.State {
	PrivValidatorKeyFile := "/home/nampkh/.tendermint/config/priv_validator_key.json"
	PrivValidatorStateFile := "/home/nampkh/.tendermint/data/priv_validator_state.json"
	filePV := privval.LoadOrGenFilePV(PrivValidatorKeyFile, PrivValidatorStateFile)
	pubkey, _ := filePV.GetPubKey()

	validator := &tmtypes.Validator{
		Address:          filePV.GetAddress(),
		PubKey:           pubkey,
		VotingPower:      100,
		ProposerPriority: 0,
	}

	state := sm.State{
		Version: tmstate.Version{ // same between empty blocks
			Consensus: tmversion.Consensus{
				Block: 11,
				App:   0,
			},
			Software: "unreleased-freezed/v0.34.19-2f231ceb952a2426cf3c0abaf0b455aadd11e5b2",
		},
		ChainID:         "checkers", // same between empty blocks
		InitialHeight:   1,          // same between empty blocks
		LastBlockHeight: 1,
		LastBlockID: tmtypes.BlockID{
			Hash: []byte{},
			PartSetHeader: tmtypes.PartSetHeader{
				Total: 0,
				Hash:  []byte{},
			},
		},
		LastBlockTime: time.Date(2022, 11, 21, 15, 28, 49, 558699629, time.UTC),
		NextValidators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{validator},
			Proposer:   validator,
		},
		Validators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{validator},
			Proposer:   validator,
		},
		LastValidators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{},
			Proposer:   &tmtypes.Validator{},
		},
		LastHeightValidatorsChanged: 1, // same between empty blocks
		ConsensusParams: tmproto.ConsensusParams{ // same between empty blocks
			Block: tmproto.BlockParams{
				MaxBytes:   22020096,
				MaxGas:     -1,
				TimeIotaMs: 1000,
			},
			Evidence: tmproto.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  48 * time.Hour,
				MaxBytes:        1048576,
			},
			Validator: tmproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
			Version: tmproto.VersionParams{
				AppVersion: 0,
			},
		},
		LastHeightConsensusParamsChanged: 1,
		LastResultsHash:                  []byte{}, // todo: find me
		AppHash:                          []byte{}, // todo: find me
	}
	return state
}

func newState() {
	PrivValidatorKeyFile := "/home/nampkh/.tendermint/config/priv_validator_key.json"
	PrivValidatorStateFile := "/home/nampkh/.tendermint/data/priv_validator_state.json"
	filePV := privval.LoadOrGenFilePV(PrivValidatorKeyFile, PrivValidatorStateFile)
	pubkey, _ := filePV.GetPubKey()

	validator := &tmtypes.Validator{
		Address:          filePV.GetAddress(),
		PubKey:           pubkey,
		VotingPower:      100,
		ProposerPriority: 0,
	}
	state := sm.State{
		Version: tmstate.Version{ // same between empty blocks
			Consensus: tmversion.Consensus{
				Block: 11,
				App:   0,
			},
			Software: "unreleased-freezed/v0.34.19-2f231ceb952a2426cf3c0abaf0b455aadd11e5b2",
		},
		ChainID:         "checkers", // same between empty blocks
		InitialHeight:   1,          // same between empty blocks
		LastBlockHeight: 0,
		LastBlockID: tmtypes.BlockID{
			Hash: []byte{},
			PartSetHeader: tmtypes.PartSetHeader{
				Total: 0,
				Hash:  []byte{},
			},
		},
		LastBlockTime: time.Date(2022, 11, 21, 15, 28, 49, 558699629, time.UTC),
		NextValidators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{validator},
			Proposer:   validator,
		},
		Validators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{validator},
			Proposer:   validator,
		},
		LastValidators: &tmtypes.ValidatorSet{
			Validators: []*tmtypes.Validator{},
			Proposer:   &tmtypes.Validator{},
		},
		LastHeightValidatorsChanged: 1, // same between empty blocks
		ConsensusParams: tmproto.ConsensusParams{ // same between empty blocks
			Block: tmproto.BlockParams{
				MaxBytes:   22020096,
				MaxGas:     -1,
				TimeIotaMs: 1000,
			},
			Evidence: tmproto.EvidenceParams{
				MaxAgeNumBlocks: 100000,
				MaxAgeDuration:  48 * time.Hour,
				MaxBytes:        1048576,
			},
			Validator: tmproto.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
			Version: tmproto.VersionParams{
				AppVersion: 0,
			},
		},
		LastHeightConsensusParamsChanged: 1,
		LastResultsHash:                  []byte{}, // todo: find me
		AppHash:                          []byte{}, // todo: find me
	}
	configConsensus := &tmconfig.ConsensusConfig{
		RootDir:                     "/home/nampkh/.tendermint",
		WalPath:                     "data/cs.wal/wal",
		TimeoutPropose:              1 * time.Second,
		TimeoutProposeDelta:         500 * time.Millisecond,
		TimeoutPrevote:              1 * time.Second,
		TimeoutPrevoteDelta:         500 * time.Millisecond,
		TimeoutPrecommit:            1 * time.Second,
		TimeoutPrecommitDelta:       500 * time.Millisecond,
		TimeoutCommit:               1 * time.Second,
		SkipTimeoutCommit:           false,
		CreateEmptyBlocks:           true,
		CreateEmptyBlocksInterval:   0 * time.Second,
		PeerGossipSleepDuration:     100 * time.Millisecond,
		PeerQueryMaj23SleepDuration: 2 * time.Second,
		DoubleSignCheckHeight:       0,
	}

	stateDB := dbm.NewMemDB()
	stateStore := sm.NewStore(stateDB)
	if err := stateStore.Save(state); err != nil {
		panic(err)
	}

	// Create the proxyApp and establish connections to the ABCI app (consensus, mempool, query).
	// proxyApp, err := createAndStartProxyAppConns(tmproxy.DefaultClientCreator("", "", ""))
	// if err != nil {
	// 	panic(err)
	// }
	// mempoolReactor, mempool := createMempoolAndMempoolReactor(config, proxyApp, state, nil)

	// blockExec := sm.NewBlockExecutor(stateStore, nil, proxyApp,mempool, )

	cs.NewState(
		configConsensus,
		state.Copy(),
		nil, nil, nil, nil,
		// blockExec,
		// blockStore,
		// mempool,
		// evidencePool,
		// cs.StateMetrics(csMetrics),
	)
}

func createAndStartProxyAppConns(clientCreator tmproxy.ClientCreator) (tmproxy.AppConns, error) {
	proxyApp := tmproxy.NewAppConns(clientCreator)
	// proxyApp.SetLogger(logger.With("module", "proxy"))
	if err := proxyApp.Start(); err != nil {
		return nil, fmt.Errorf("error starting proxy app connections: %v", err)
	}
	return proxyApp, nil
}

func createMempoolAndMempoolReactor(config *tmconfig.Config, proxyApp tmproxy.AppConns,
	state sm.State, memplMetrics *tmmempool.Metrics) (*tmmempool.Reactor, *tmmempool.CListMempool) {

	mempool := tmmempool.NewCListMempool(
		config.Mempool,
		proxyApp.Mempool(),
		state.LastBlockHeight,
		tmmempool.WithMetrics(memplMetrics),
		tmmempool.WithPreCheck(sm.TxPreCheck(state)),
		tmmempool.WithPostCheck(sm.TxPostCheck(state)),
	)
	// mempoolLogger := logger.With("module", "mempool")
	mempoolReactor := tmmempool.NewReactor(config.Mempool, mempool)
	// mempoolReactor.SetLogger(mempoolLogger)

	if config.Consensus.WaitForTxs() {
		mempool.EnableTxsAvailable()
	}
	return mempoolReactor, mempool
}
