package depreqres

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

func TestELSyncFillingUnsafeChainGap(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewSingleChainMultiNode(t)
	require := t.Require()
	l := t.Logger()

	l.Info("Confirm that the CL nodes are progressing the unsafe chain")
	target := uint64(10)
	dsl.CheckAll(t,
		sys.L2CL.AdvancedFn(types.LocalUnsafe, target, 30),
		sys.L2CLB.AdvancedFn(types.LocalUnsafe, target, 30),
	)

	l.Info("Stop the L2 batcher")
	sys.L2Batcher.Stop()

	l.Info("Disconnect L2CL from L2CLB")
	sys.L2CLB.DisconnectPeer(sys.L2CL)

	ssA_before := sys.L2CL.SyncStatus()
	ssB_before := sys.L2CLB.SyncStatus()

	time.Sleep(20 * time.Second)

	ssA_after := sys.L2CL.SyncStatus()
	ssB_after := sys.L2CLB.SyncStatus()

	require.Greater(ssA_after.UnsafeL2.Number, ssA_before.UnsafeL2.Number, "unsafe chain for L2CL should have advanced")
	require.Equal(ssB_after.UnsafeL2.Number, ssB_before.UnsafeL2.Number, "unsafe chain for L2CLB should have stalled")

	l.Info("Re-connect L2CL to L2CLB")
	sys.L2CLB.ConnectPeer(sys.L2CL)

	l.Info("Expect L2CLB to catch up with L2CL for the unsafe chain")
	sys.L2CLB.ReachedNode(types.LocalUnsafe, sys.L2CL, 100)

	sys.L2ELB.Reached(types.LocalUnsafe, ssA_after.UnsafeL2.Number, 100)

	l.Info("Confirm that the safe chain for both CL nodes is stalled, since L2 batcher is stopped")
	dsl.CheckAll(t,
		sys.L2CLB.NotAdvancedFn(types.LocalSafe, 5),
		sys.L2CLB.NotAdvancedFn(types.LocalSafe, 5),
	)
}
