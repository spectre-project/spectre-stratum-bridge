package spectrestratum

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/spectre-project/spectre-stratum-bridge/src/gostratum"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const version = "v0.3.18"
const minBlockWaitTime = 250 * time.Millisecond

type BridgeConfig struct {
	StratumPort     string        `yaml:"stratum_port"`
	RPCServer       string        `yaml:"spectred_address"`
	PromPort        string        `yaml:"prom_port"`
	PrintStats      bool          `yaml:"print_stats"`
	UseLogFile      bool          `yaml:"log_to_file"`
	HealthCheckPort string        `yaml:"health_check_port"`
	SoloMining      bool          `yaml:"solo_mining"`
	BlockWaitTime   time.Duration `yaml:"block_wait_time"`
	MinShareDiff    uint          `yaml:"min_share_diff"`
	VarDiff         bool          `yaml:"var_diff"`
	SharesPerMin    uint          `yaml:"shares_per_min"`
	VarDiffStats    bool          `yaml:"var_diff_stats"`
	ExtranonceSize  uint          `yaml:"extranonce_size"`
}

func configureZap(cfg BridgeConfig) (*zap.SugaredLogger, func()) {
	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.RFC3339TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(pe)
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	if !cfg.UseLogFile {
		return zap.New(zapcore.NewCore(consoleEncoder,
			zapcore.AddSync(colorable.NewColorableStdout()), zap.InfoLevel)).Sugar(), func() {}
	}

	// log file fun
	logFile, err := os.OpenFile("bridge.log", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), zap.InfoLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(colorable.NewColorableStdout()), zap.InfoLevel),
	)
	return zap.New(core).Sugar(), func() { logFile.Close() }
}

func ListenAndServe(cfg BridgeConfig) error {
	logger, logCleanup := configureZap(cfg)
	defer logCleanup()

	if cfg.PromPort != "" {
		StartPromServer(logger, cfg.PromPort)
	}

	blockWaitTime := cfg.BlockWaitTime
	if blockWaitTime < minBlockWaitTime {
		blockWaitTime = minBlockWaitTime
	}
	sprApi, err := NewSpectreAPI(cfg.RPCServer, blockWaitTime, logger)
	if err != nil {
		return err
	}

	if cfg.HealthCheckPort != "" {
		logger.Info("enabling health check on port " + cfg.HealthCheckPort)
		http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		go http.ListenAndServe(cfg.HealthCheckPort, nil)
	}

	shareHandler := newShareHandler(sprApi.spectred)
	minDiff := cfg.MinShareDiff
	if minDiff == 0 {
		minDiff = 4
	}
	extranonceSize := cfg.ExtranonceSize
	if extranonceSize > 3 {
		extranonceSize = 3
	}
	clientHandler := newClientListener(logger, shareHandler, float64(minDiff), int8(extranonceSize))
	handlers := gostratum.DefaultHandlers()
	// override the submit handler with an actual useful handler
	handlers[string(gostratum.StratumMethodSubmit)] =
		func(ctx *gostratum.StratumContext, event gostratum.JsonRpcEvent) error {
			if err := shareHandler.HandleSubmit(ctx, event, cfg.SoloMining); err != nil {
				ctx.Logger.Sugar().Error(err) // sink error
			}
			return nil
		}

	stratumConfig := gostratum.StratumListenerConfig{
		Port:           cfg.StratumPort,
		HandlerMap:     handlers,
		StateGenerator: MiningStateGenerator,
		ClientListener: clientHandler,
		Logger:         logger.Desugar(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sprApi.Start(ctx, func() {
		clientHandler.NewBlockAvailable(sprApi, cfg.SoloMining)
	})

	if cfg.VarDiff || cfg.SoloMining {
		go shareHandler.startVardiffThread(cfg.SharesPerMin, cfg.VarDiffStats)
	}

	if cfg.PrintStats {
		go shareHandler.startStatsThread()
	}

	return gostratum.NewListener(stratumConfig).Listen(context.Background())
}
