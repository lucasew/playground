//go:build unix

package server

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessAliveSeesChild(t *testing.T) {
	cmd := exec.Command("sleep", "5")
	require.NoError(t, cmd.Start())
	defer cmd.Process.Kill()
	require.True(t, processAlive(cmd.Process.Pid))
	require.False(t, processAlive(0))
	require.False(t, processAlive(1<<30))
}
