// filePath: kai/kai.go

package kai

import (
	"go.uber.org/zap"
)

// json is an alias for the jsoniter library configured to be compatible with the standard library's JSON handling.
// var json = jsoniter.ConfigCompatibleWithStandardLibrary
var logger *zap.Logger

func init() {

	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger.Info("kai init")
}

const NO_INSCRIPTION_BLOCK_HASH = "0x0000000000000000000000000000000000000000000000000000000000000000"
