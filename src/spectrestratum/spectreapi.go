package spectrestratum

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spectre-project/spectre-stratum-bridge/src/gostratum"
	"github.com/spectre-project/spectred/app/appmessage"
	"github.com/spectre-project/spectred/infrastructure/network/rpcclient"
	"go.uber.org/zap"
)

type SpectreApi struct {
	address       string
	blockWaitTime time.Duration
	logger        *zap.SugaredLogger
	spectred      *rpcclient.RPCClient
	connected     bool
}

func NewSpectreAPI(address string, blockWaitTime time.Duration, logger *zap.SugaredLogger) (*SpectreApi, error) {
	client, err := rpcclient.NewRPCClient(address)
	if err != nil {
		return nil, err
	}

	return &SpectreApi{
		address:       address,
		blockWaitTime: blockWaitTime,
		logger:        logger.With(zap.String("component", "spectreapi:"+address)),
		spectred:      client,
		connected:     true,
	}, nil
}

func (sprApi *SpectreApi) Start(ctx context.Context, blockCb func()) {
	sprApi.waitForSync(true)
	go sprApi.startBlockTemplateListener(ctx, blockCb)
	go sprApi.startStatsThread(ctx)
}

func (sprApi *SpectreApi) startStatsThread(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			sprApi.logger.Warn("context cancelled, stopping stats thread")
			return
		case <-ticker.C:
			dagResponse, err := sprApi.spectred.GetBlockDAGInfo()
			if err != nil {
				sprApi.logger.Warn("failed to get network hashrate from spectre, prom stats will be out of date", zap.Error(err))
				continue
			}
			response, err := sprApi.spectred.EstimateNetworkHashesPerSecond(dagResponse.TipHashes[0], 1000)
			if err != nil {
				sprApi.logger.Warn("failed to get network hashrate from spectre, prom stats will be out of date", zap.Error(err))
				continue
			}
			RecordNetworkStats(response.NetworkHashesPerSecond, dagResponse.BlockCount, dagResponse.Difficulty)
		}
	}
}

func (sprApi *SpectreApi) reconnect() error {
	if sprApi.spectred != nil {
		return sprApi.spectred.Reconnect()
	}

	client, err := rpcclient.NewRPCClient(sprApi.address)
	if err != nil {
		return err
	}
	sprApi.spectred = client
	return nil
}

func (sprApi *SpectreApi) waitForSync(verbose bool) error {
	if verbose {
		sprApi.logger.Info("checking spectred sync state")
	}
	for {
		clientInfo, err := sprApi.spectred.GetInfo()
		if err != nil {
			return errors.Wrapf(err, "error fetching server info from spectred @ %s", sprApi.address)
		}
		if clientInfo.IsSynced {
			break
		}
		sprApi.logger.Warn("Spectre is not synced, waiting for sync before starting bridge")
		time.Sleep(5 * time.Second)
	}
	if verbose {
		sprApi.logger.Info("spectred synced, starting server")
	}
	return nil
}

func (sprApi *SpectreApi) startBlockTemplateListener(ctx context.Context, blockReadyCb func()) {
	blockReadyChan := make(chan bool)
	err := sprApi.spectred.RegisterForNewBlockTemplateNotifications(func(_ *appmessage.NewBlockTemplateNotificationMessage) {
		blockReadyChan <- true
	})
	if err != nil {
		sprApi.logger.Error("fatal: failed to register for block notifications from spectre")
	}

	ticker := time.NewTicker(sprApi.blockWaitTime)
	for {
		if err := sprApi.waitForSync(false); err != nil {
			sprApi.logger.Error("error checking spectred sync state, attempting reconnect: ", err)
			if err := sprApi.reconnect(); err != nil {
				sprApi.logger.Error("error reconnecting to spectred, waiting before retry: ", err)
				time.Sleep(5 * time.Second)
			}
		}
		select {
		case <-ctx.Done():
			sprApi.logger.Warn("context cancelled, stopping block update listener")
			return
		case <-blockReadyChan:
			blockReadyCb()
			ticker.Reset(sprApi.blockWaitTime)
		case <-ticker.C: // timeout, manually check for new blocks
			blockReadyCb()
		}
	}
}

func (sprApi *SpectreApi) GetBlockTemplate(
	client *gostratum.StratumContext) (*appmessage.GetBlockTemplateResponseMessage, error) {
	template, err := sprApi.spectred.GetBlockTemplate(client.WalletAddr,
		fmt.Sprintf(`'%s' via spectre-project/spectre-stratum-bridge_%s`, client.RemoteApp, version))
	if err != nil {
		return nil, errors.Wrap(err, "failed fetching new block template from spectre")
	}
	return template, nil
}
