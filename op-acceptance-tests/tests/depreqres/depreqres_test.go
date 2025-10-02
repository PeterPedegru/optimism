package depreqres

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

func XTestDeprecateReqResCLSync_ELSequencer(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewSingleChainMultiNode(t)
	require := t.Require()
	l := t.Logger()

	// Test that we can get chain IDs from both L2CL nodes
	l2CLChainID := sys.L2CL.ID().ChainID()
	require.Equal(eth.ChainIDFromUInt64(901), l2CLChainID, "first L2CL should be on chain 901")

	l2CL2ChainID := sys.L2CLB.ID().ChainID()
	require.Equal(eth.ChainIDFromUInt64(901), l2CL2ChainID, "second L2CL should be on chain 901")

	target := uint64(10)
	dsl.CheckAll(t,
		sys.L2CL.AdvancedFn(types.LocalUnsafe, target, 30),
		sys.L2CLB.AdvancedFn(types.LocalUnsafe, target, 30),
	)

	l.Info("test completed")
}
func TestDeprecateReqResCLSync_DisableGossip(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewSingleChainMultiNode(t)
	l := t.Logger()

	target := uint64(10)
	dsl.CheckAll(t,
		sys.L2CL.AdvancedFn(types.LocalUnsafe, target, 30),
		sys.L2CLB.AdvancedFn(types.LocalUnsafe, target, 30),
	)

	sys.L2Batcher.Stop()

	l.Info("anteva disconnect L2CL from L2CLB")
	sys.L2CLB.DisconnectPeer(sys.L2CL)

	unsafeAfterStopped := sys.L2CL.SyncStatus().UnsafeL2.ID()
	time.Sleep(20 * time.Second)
	unsafeBeforeStarted := sys.L2CL.SyncStatus().UnsafeL2.ID()

	l.Info("anteva connect L2CL to L2CLB")
	sys.L2CLB.ConnectPeer(sys.L2CL)

	l.Info("anteva blocks to sync", "after_stopped", unsafeAfterStopped, "before_started", unsafeBeforeStarted)

	ssA := sys.L2CL.SyncStatus()
	ssB := sys.L2CLB.SyncStatus()
	elUnsafe, elSafe, elFinalized := sys.L2ELB.Status()

	l.Info("anteva L2CL status", "unsafeL2", ssA.UnsafeL2.ID(), "safeL2", ssA.SafeL2.ID())
	l.Info("anteva L2ELB status", "unsafeL2", elUnsafe.ID(), "safeL2", elSafe.ID(), "finalizedL2", elFinalized.ID())
	l.Info("anteva L2CLB status", "unsafeL2", ssB.UnsafeL2.ID(), "safeL2", ssB.SafeL2.ID())

	time.Sleep(30 * time.Second)

	ssA = sys.L2CL.SyncStatus()
	ssB = sys.L2CLB.SyncStatus()
	elUnsafe, elSafe, elFinalized = sys.L2ELB.Status()

	l.Info("anteva L2CL status - after sleep", "unsafeL2", ssA.UnsafeL2.ID(), "safeL2", ssA.SafeL2.ID())
	l.Info("anteva L2ELB status - after sleep", "unsafeL2", elUnsafe.ID(), "safeL2", elSafe.ID(), "finalizedL2", elFinalized.ID())
	l.Info("anteva L2CLB status - after sleep", "unsafeL2", ssB.UnsafeL2.ID(), "safeL2", ssB.SafeL2.ID())

	sys.L2CLB.ReachedNode(types.LocalUnsafe, sys.L2CL, 100)
	dsl.CheckAll(t,
		sys.L2CLB.NotAdvancedFn(types.LocalSafe, 5),
		sys.L2CLB.NotAdvancedFn(types.LocalSafe, 5),
	)
}
