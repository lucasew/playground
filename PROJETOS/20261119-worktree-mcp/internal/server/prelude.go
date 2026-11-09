package server

import (
	// Register the native exec driver that x/git runs git through.
	_ "github.com/lewtec/lewkit/x/driver/exec/prelude"
)
