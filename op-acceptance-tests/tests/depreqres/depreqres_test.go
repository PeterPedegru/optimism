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

func TestDeprecateReqResCLSync_ELSequencer(gt *testing.T) {
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

	l.Info("anteva stopping L2 batcher")
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

	l.Info("anteva L2CL status", "unsafeL2", ssA.UnsafeL2.ID(), "safeL2", ssA.SafeL2.ID())
	l.Info("anteva L2CLB status", "unsafeL2", ssB.UnsafeL2.ID(), "safeL2", ssB.SafeL2.ID())

	sys.L2Batcher.Start()

	sys.L2CLB.ReachedNode(types.LocalUnsafe, sys.L2CL, 100)

	ssA = sys.L2CL.SyncStatus()
	ssB = sys.L2CLB.SyncStatus()

	l.Info("anteva L2CL status", "unsafeL2", ssA.UnsafeL2.ID(), "safeL2", ssA.SafeL2.ID())
	l.Info("anteva L2CLB status", "unsafeL2", ssB.UnsafeL2.ID(), "safeL2", ssB.SafeL2.ID())

	l.Info("test completed")
}

func TestDeprecateReqResCLSync_StopOpNode(gt *testing.T) {
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

	l.Info("anteva stopping L2 batcher")
	sys.L2Batcher.Stop()

	l.Info("anteva stopping L2CL2")
	sys.L2CLB.Stop()
	sys.L2ELB.Stop()

	unsafeAfterStopped := sys.L2CL.SyncStatus().UnsafeL2.ID()

	time.Sleep(180 * time.Second)

	unsafeBeforeStarted := sys.L2CL.SyncStatus().UnsafeL2.ID()

	l.Info("anteva blocks to sync", "after_stopped", unsafeAfterStopped, "before_started", unsafeBeforeStarted)

	l.Info("anteva starting L2CLB")
	sys.L2ELB.Start()
	l.Info("anteva started L2ELB")
	sys.L2CLB.Start()
	l.Info("anteva started L2CLB")
	sys.L2ELB.PeerWith(sys.L2EL)
	l.Info("anteva L2ELB peered with L2EL")

	unsafeTargetId := sys.L2CL.SyncStatus().UnsafeL2.ID()
	safeTargetId := sys.L2CL.SyncStatus().SafeL2.ID()

	l.Info("anteva target for L2CLB", "unsafeL2", unsafeTargetId, "safeL2", safeTargetId)

	dsl.CheckAll(t,
		sys.L2CLB.ReachedRefFn(types.LocalUnsafe, unsafeTargetId, 50),
		sys.L2CLB.ReachedRefFn(types.LocalSafe, safeTargetId, 50),
	)

	syncedUnsafe := sys.L2CLB.SyncStatus().UnsafeL2.ID()
	syncedSafe := sys.L2CLB.SyncStatus().SafeL2.ID()
	l.Info("anteva L2CLB synced", "syncedUnsafe", syncedUnsafe, "syncedSafe", syncedSafe)

	l.Info("test completed")
}
