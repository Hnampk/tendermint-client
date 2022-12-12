package abciclient

import (
	"fmt"

	flogging "github.com/Hnampk/fabric-flogging"
	cmap "github.com/orcaman/concurrent-map"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/proxy"
)

type ApplicationManager struct {
	applicationPool cmap.ConcurrentMap
}

type AppProxy struct {
	proxy.AppConns
	name, address string
}

var (
	applicationManager *ApplicationManager

	applicationLogger = flogging.MustGetLogger("usecases.abciclient.app_manager")
)

func init() {
	if applicationManager == nil {
		applicationManager = &ApplicationManager{
			applicationPool: cmap.New(),
		}
	}
}

func GetApp(name string) (appProxy *AppProxy, err error) {
	if appProxyData, existed := applicationManager.applicationPool.Get(name); !existed {
		wrappedErr := fmt.Errorf("app proxy %s not exists", name)
		applicationLogger.Errorf(wrappedErr.Error())
		return nil, wrappedErr
	} else {
		appProxy = appProxyData.(*AppProxy)
	}

	return
}

// NewAppProxy create a new tendermint client using GRPC
func NewGRPCAppProxy(name, address string) error {
	if _, existed := applicationManager.applicationPool.Get(name); !existed {
		applicationLogger.Infof("Create NewGRPCAppProxy %s", name)

		clientCreator := proxy.NewRemoteClientCreator(address, "grpc", true)
		proxyApp := proxy.NewAppConns(clientCreator)
		applicationLogger.Infof("Created Start proxyApp %s", name)

		if err := proxyApp.Start(); err != nil {
			wrappedErr := fmt.Errorf("error while connect to ABCIServer %s: %w", address, err)
			applicationLogger.Errorf(wrappedErr.Error())
			return wrappedErr
		}

		applicationManager.applicationPool.Set(name, &AppProxy{
			AppConns: proxyApp,
			name:     name,
			address:  address,
		})
		applicationLogger.Infof("App proxy %s connected to %s", name, address)
	} else {
		applicationLogger.Warnf("App proxy %s is already existed", address)
	}

	return nil
}

func (a *AppProxy) OnStop() {
	a.AppConns.OnStop()
	applicationManager.applicationPool.Remove(a.name)
}

func (a *AppProxy) EchoSync(message string) (echoResponse *abcitypes.ResponseEcho, err error) {
	echoResponse, err = a.Query().EchoSync(message)
	if err != nil {
		wrappedErr := fmt.Errorf("error while EchoSync %s: %w", a.name, err)
		applicationLogger.Errorf(wrappedErr.Error())
		return nil, wrappedErr
	}

	return
}

func (a *AppProxy) InfoSync(req abcitypes.RequestInfo) (infoResponse *abcitypes.ResponseInfo, err error) {
	infoResponse, err = a.Query().InfoSync(req)
	if err != nil {
		wrappedErr := fmt.Errorf("error while InfoSync %s: %w", a.name, err)
		applicationLogger.Errorf(wrappedErr.Error())
		return nil, wrappedErr
	}

	return
}

func (a *AppProxy) InitChainSync(req abcitypes.RequestInitChain) (infoResponse *abcitypes.ResponseInitChain, err error) {
	infoResponse, err = a.Consensus().InitChainSync(req)
	if err != nil {
		wrappedErr := fmt.Errorf("error while InfoSync %s: %w", a.name, err)
		applicationLogger.Errorf(wrappedErr.Error())
		return nil, wrappedErr
	}

	return
}

func (a *AppProxy) BeginBlockSync(req abcitypes.RequestBeginBlock) (beginBlockResponse *abcitypes.ResponseBeginBlock, err error) {
	beginBlockResponse, err = a.Consensus().BeginBlockSync(req)
	if err != nil {
		wrappedErr := fmt.Errorf("error while BeginBlockSync %s: %w", a.name, err)
		applicationLogger.Errorf(wrappedErr.Error())
		return nil, wrappedErr
	}

	return
}
