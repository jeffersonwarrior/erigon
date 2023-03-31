package beacon_changeset

import (
	libcommon "github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/cl/clparams"
	"github.com/ledgerwatch/erigon/cl/cltypes"
)

// This type of changeset is the diff beetwen next state and input state and is used to reverse/forward beacon state.
type ChangeSet struct {
	// Single types.
	slotChange                         *uint64
	forkChange                         *cltypes.Fork
	latestBlockHeaderChange            *cltypes.BeaconBlockHeader
	eth1DataChange                     *cltypes.Eth1Data
	eth1DepositIndexChange             *uint64
	justificationBitsChange            *cltypes.JustificationBits
	previousJustifiedCheckpointChange  *cltypes.Checkpoint
	currentJustifiedCheckpointChange   *cltypes.Checkpoint
	finalizedCheckpointChange          *cltypes.Checkpoint
	currentSyncCommitteeChange         *cltypes.SyncCommittee
	nextSyncCommitteeChange            *cltypes.SyncCommittee
	latestExecutionPayloadHeaderChange *cltypes.Eth1Header
	nextWithdrawalIndexChange          *uint64
	nextWithdrawalValidatorIndexChange *uint64
	versionChange                      *clparams.StateVersion
	// Lists and arrays changesets
	BlockRootsChanges                 *ListChangeSet[libcommon.Hash]
	StateRootsChanges                 *ListChangeSet[libcommon.Hash]
	HistoricalRootsChanges            *ListChangeSet[libcommon.Hash]
	Eth1DataVotesChanges              *ListChangeSet[cltypes.Eth1Data]
	BalancesChanges                   *ListChangeSet[uint64]
	RandaoMixesChanges                *ListChangeSet[libcommon.Hash]
	SlashingsChanges                  *ListChangeSet[uint64]
	PreviousEpochParticipationChanges *ListChangeSet[cltypes.ParticipationFlags]
	CurrentEpochParticipationChanges  *ListChangeSet[cltypes.ParticipationFlags]
	InactivityScoresChanges           *ListChangeSet[uint64]
	historicalSummaryChange           *ListChangeSet[cltypes.HistoricalSummary]
	// Validator fields.
	WithdrawalCredentialsChange      *ListChangeSet[libcommon.Hash]
	EffectiveBalanceChange           *ListChangeSet[uint64]
	SlashedChange                    *ListChangeSet[bool]
	ActivationEligibilityEpochChange *ListChangeSet[uint64]
	ActivationEpochChange            *ListChangeSet[uint64]
	ExitEpochChange                  *ListChangeSet[uint64]
	WithdrawalEpochChange            *ListChangeSet[uint64]
	// Efficient unwinding on reset (only applicable at epoch boundaries)
	previousEpochParticipationAtReset cltypes.ParticipationFlagsList
	currentEpochParticipationAtReset  cltypes.ParticipationFlagsList
	eth1DataVotesAtReset              []*cltypes.Eth1Data
	wasEth1DataVotesReset             bool
	wasEpochParticipationReset        bool
}

func New(validatorSetSize, blockRootsLength, stateRootsLength, slashingsLength, historicalSummariesLength, historicalRootsLength, votesLength, randaoMixesLength int) *ChangeSet {
	return &ChangeSet{
		BlockRootsChanges:                 NewListChangeSet[libcommon.Hash](blockRootsLength),
		StateRootsChanges:                 NewListChangeSet[libcommon.Hash](stateRootsLength),
		HistoricalRootsChanges:            NewListChangeSet[libcommon.Hash](historicalRootsLength),
		Eth1DataVotesChanges:              NewListChangeSet[cltypes.Eth1Data](votesLength),
		BalancesChanges:                   NewListChangeSet[uint64](validatorSetSize),
		RandaoMixesChanges:                NewListChangeSet[libcommon.Hash](randaoMixesLength),
		SlashingsChanges:                  NewListChangeSet[uint64](slashingsLength),
		PreviousEpochParticipationChanges: NewListChangeSet[cltypes.ParticipationFlags](validatorSetSize),
		CurrentEpochParticipationChanges:  NewListChangeSet[cltypes.ParticipationFlags](validatorSetSize),
		InactivityScoresChanges:           NewListChangeSet[uint64](validatorSetSize),
		historicalSummaryChange:           NewListChangeSet[cltypes.HistoricalSummary](historicalSummariesLength),
		// Validators section
		WithdrawalCredentialsChange:      NewListChangeSet[libcommon.Hash](validatorSetSize),
		EffectiveBalanceChange:           NewListChangeSet[uint64](validatorSetSize),
		ActivationEligibilityEpochChange: NewListChangeSet[uint64](validatorSetSize),
		ActivationEpochChange:            NewListChangeSet[uint64](validatorSetSize),
		ExitEpochChange:                  NewListChangeSet[uint64](validatorSetSize),
		WithdrawalEpochChange:            NewListChangeSet[uint64](validatorSetSize),
		SlashedChange:                    NewListChangeSet[bool](validatorSetSize),
	}
}

func (r *ChangeSet) OnSlotChange(prevSlot uint64, replaceExisting bool) {
	if !replaceExisting && r.slotChange != nil {
		return
	}
	r.slotChange = new(uint64)
	*r.slotChange = prevSlot
}

func (r *ChangeSet) OnForkChange(fork *cltypes.Fork, replaceExisting bool) {
	if !replaceExisting && r.forkChange != nil {
		return
	}
	r.forkChange = new(cltypes.Fork)
	*r.forkChange = *fork
}

func (r *ChangeSet) OnLatestHeaderChange(h *cltypes.BeaconBlockHeader, replaceExisting bool) {
	if !replaceExisting && r.latestBlockHeaderChange != nil {
		return
	}
	r.latestBlockHeaderChange = new(cltypes.BeaconBlockHeader)
	*r.latestBlockHeaderChange = *h
}

func (r *ChangeSet) OnEth1DataChange(e *cltypes.Eth1Data, replaceExisting bool) {
	if !replaceExisting && r.latestBlockHeaderChange != nil {
		return
	}
	r.eth1DataChange = new(cltypes.Eth1Data)
	*r.eth1DataChange = *e
}

func (r *ChangeSet) OnJustificationBitsChange(j cltypes.JustificationBits, replaceExisting bool) {
	if !replaceExisting && r.justificationBitsChange != nil {
		return
	}
	r.justificationBitsChange = new(cltypes.JustificationBits)
	*r.justificationBitsChange = j.Copy()
}

func (r *ChangeSet) OnEth1DepositIndexChange(e uint64, replaceExisting bool) {
	if !replaceExisting && r.eth1DepositIndexChange != nil {
		return
	}
	r.eth1DepositIndexChange = new(uint64)
	*r.eth1DepositIndexChange = e
}

func (r *ChangeSet) OnPreviousJustifiedCheckpointChange(c *cltypes.Checkpoint, replaceExisting bool) {
	if !replaceExisting && r.previousJustifiedCheckpointChange != nil {
		return
	}
	r.previousJustifiedCheckpointChange = c.Copy()
}

func (r *ChangeSet) OnCurrentJustifiedCheckpointChange(c *cltypes.Checkpoint, replaceExisting bool) {
	if !replaceExisting && r.currentJustifiedCheckpointChange != nil {
		return
	}
	r.currentJustifiedCheckpointChange = c.Copy()
}

func (r *ChangeSet) OnFinalizedCheckpointChange(c *cltypes.Checkpoint, replaceExisting bool) {
	if !replaceExisting && r.finalizedCheckpointChange != nil {
		return
	}
	r.finalizedCheckpointChange = c.Copy()
}

func (r *ChangeSet) OnCurrentSyncCommitteeChange(c *cltypes.SyncCommittee, replaceExisting bool) {
	if !replaceExisting && r.currentSyncCommitteeChange != nil {
		return
	}
	r.currentSyncCommitteeChange = new(cltypes.SyncCommittee)
	*r.currentSyncCommitteeChange = *c
	r.currentSyncCommitteeChange.PubKeys = make([][48]byte, len(c.PubKeys))
	copy(r.currentSyncCommitteeChange.PubKeys, c.PubKeys)
}

func (r *ChangeSet) OnNextSyncCommitteeChange(c *cltypes.SyncCommittee, replaceExisting bool) {
	if !replaceExisting && r.nextSyncCommitteeChange != nil {
		return
	}
	r.nextSyncCommitteeChange = new(cltypes.SyncCommittee)
	*r.nextSyncCommitteeChange = *c
	r.nextSyncCommitteeChange.PubKeys = make([][48]byte, len(c.PubKeys))
	copy(r.nextSyncCommitteeChange.PubKeys, c.PubKeys)
}

func (r *ChangeSet) OnEth1Header(e *cltypes.Eth1Header, replaceExisting bool) {
	if !replaceExisting && r.latestExecutionPayloadHeaderChange != nil {
		return
	}
	r.latestExecutionPayloadHeaderChange = new(cltypes.Eth1Header)
	*r.latestExecutionPayloadHeaderChange = *e
	r.latestExecutionPayloadHeaderChange.Extra = libcommon.Copy(e.Extra)
}

func (r *ChangeSet) OnNextWithdrawalIndexChange(index uint64, replaceExisting bool) {
	if !replaceExisting && r.nextWithdrawalIndexChange != nil {
		return
	}
	r.nextWithdrawalIndexChange = new(uint64)
	*r.nextWithdrawalIndexChange = index
}

func (r *ChangeSet) OnNextWithdrawalValidatorIndexChange(index uint64, replaceExisting bool) {
	if !replaceExisting && r.nextWithdrawalValidatorIndexChange != nil {
		return
	}
	r.nextWithdrawalValidatorIndexChange = new(uint64)
	*r.nextWithdrawalValidatorIndexChange = index
}

func (r *ChangeSet) OnVersionChange(v clparams.StateVersion, replaceExisting bool) {
	if !replaceExisting && r.versionChange != nil {
		return
	}
	r.versionChange = new(clparams.StateVersion)
	*r.versionChange = v
}

func (r *ChangeSet) HasValidatorSetNotChanged(validatorSetLength int) bool {
	return validatorSetLength == r.WithdrawalCredentialsChange.ListLength() && r.WithdrawalCredentialsChange.Empty() && r.ActivationEligibilityEpochChange.Empty() && r.ActivationEpochChange.Empty() &&
		r.EffectiveBalanceChange.Empty() && r.SlashedChange.Empty() && r.ExitEpochChange.Empty() && r.WithdrawalEpochChange.Empty()
}

func (r *ChangeSet) ApplyHistoricalSummaryChanges(input []*cltypes.HistoricalSummary) (output []*cltypes.HistoricalSummary, changed bool) {
	output = input
	if r.historicalSummaryChange.Empty() && r.historicalSummaryChange.ListLength() == len(output) {
		return
	}
	changed = true
	historicalSummarryLength := r.historicalSummaryChange.ListLength()
	if historicalSummarryLength != len(output) {
		output = make([]*cltypes.HistoricalSummary, historicalSummarryLength)
		copy(output, input)
	}
	r.historicalSummaryChange.ChangesWithHandler(func(value cltypes.HistoricalSummary, index int) {
		*output[index] = value
	})
	return
}

func (r *ChangeSet) CompactChanges() {

	r.BlockRootsChanges.CompactChangesReverse()
	r.StateRootsChanges.CompactChangesReverse()
	r.HistoricalRootsChanges.CompactChangesReverse()
	r.SlashingsChanges.CompactChangesReverse()
	r.RandaoMixesChanges.CompactChangesReverse()
	r.BalancesChanges.CompactChangesReverse()
	if len(r.eth1DataVotesAtReset) > 0 {
		r.Eth1DataVotesChanges = nil
	} else {
		r.Eth1DataVotesChanges.CompactChangesReverse()
	}
	if len(r.previousEpochParticipationAtReset) > 0 {
		r.PreviousEpochParticipationChanges = nil
		r.CurrentEpochParticipationChanges = nil
	} else {
		r.PreviousEpochParticipationChanges.CompactChangesReverse()
		r.CurrentEpochParticipationChanges.CompactChangesReverse()
	}
	r.InactivityScoresChanges.CompactChangesReverse()
	r.HistoricalRootsChanges.CompactChangesReverse()
	r.WithdrawalCredentialsChange.CompactChangesReverse()
	r.EffectiveBalanceChange.CompactChangesReverse()
	r.ExitEpochChange.CompactChangesReverse()
	r.ActivationEligibilityEpochChange.CompactChangesReverse()
	r.ActivationEpochChange.CompactChangesReverse()
	r.SlashedChange.CompactChangesReverse()
	r.WithdrawalEpochChange.CompactChangesReverse()
}

func (r *ChangeSet) ReportVotesReset(previousVotes []*cltypes.Eth1Data) {
	if r.wasEth1DataVotesReset {
		return
	}
	// Copy the slice over
	for _, vote := range previousVotes {
		copyVote := *vote
		r.eth1DataVotesAtReset = append(r.eth1DataVotesAtReset, &copyVote)
	}
	r.wasEth1DataVotesReset = true
}

func (r *ChangeSet) ReportEpochParticipationReset(prevParticipation, currParticpation cltypes.ParticipationFlagsList) {
	if r.wasEpochParticipationReset {
		return
	}
	r.previousEpochParticipationAtReset = prevParticipation.Copy()
	r.currentEpochParticipationAtReset = currParticpation.Copy()
	r.wasEpochParticipationReset = true
}

func (r *ChangeSet) ApplyEth1DataVotesChanges(initialVotes []*cltypes.Eth1Data) (output []*cltypes.Eth1Data, changed bool) {
	if r.wasEth1DataVotesReset {
		return r.eth1DataVotesAtReset, true
	}
	output = initialVotes
	if r.Eth1DataVotesChanges.Empty() && r.Eth1DataVotesChanges.ListLength() == len(output) {
		return
	}
	changed = true
	if r.Eth1DataVotesChanges.ListLength() != len(output) {
		output = make([]*cltypes.Eth1Data, r.Eth1DataVotesChanges.ListLength())
		copy(output, initialVotes)
	}
	r.Eth1DataVotesChanges.ChangesWithHandler(func(value cltypes.Eth1Data, index int) {
		*output[index] = value
	})
	return
}

func (r *ChangeSet) ApplyEpochParticipationChanges(
	previousEpochParticipation cltypes.ParticipationFlagsList,
	currentEpochParticipation cltypes.ParticipationFlagsList) (newPreviousEpochParticipation cltypes.ParticipationFlagsList, newCurrentEpochParticipation cltypes.ParticipationFlagsList,
	previousParticipationChanged bool, currentParticipationChanged bool) {
	if r.wasEpochParticipationReset {
		return r.previousEpochParticipationAtReset, r.currentEpochParticipationAtReset, true, true
	}
	newPreviousEpochParticipation, previousParticipationChanged = r.PreviousEpochParticipationChanges.ApplyChanges(previousEpochParticipation)
	newCurrentEpochParticipation, currentParticipationChanged = r.CurrentEpochParticipationChanges.ApplyChanges(currentEpochParticipation)
	return
}

func (r *ChangeSet) ApplySlotChange(prevSlot uint64) (uint64, bool) {
	if r.slotChange == nil {
		return prevSlot, false
	}
	return *r.slotChange, true
}

func (r *ChangeSet) ApplyForkChange(fork *cltypes.Fork) (*cltypes.Fork, bool) {
	if r.forkChange == nil {
		return fork, false
	}
	return r.forkChange.Copy(), true
}

func (r *ChangeSet) ApplyLatestBlockHeader(header *cltypes.BeaconBlockHeader) (*cltypes.BeaconBlockHeader, bool) {
	if r.latestBlockHeaderChange == nil {
		return header, false
	}
	return r.latestBlockHeaderChange.Copy(), true
}

func (r *ChangeSet) ApplyEth1DataChange(data *cltypes.Eth1Data) (*cltypes.Eth1Data, bool) {
	if r.eth1DataChange == nil {
		return data, false
	}
	return r.eth1DataChange.Copy(), true
}

func (r *ChangeSet) ApplyEth1DepositIndexChange(depositIndex uint64) (uint64, bool) {
	if r.eth1DepositIndexChange == nil {
		return depositIndex, false
	}
	return *r.eth1DepositIndexChange, true
}

func (r *ChangeSet) ApplyJustificationBitsChange(bits cltypes.JustificationBits) (cltypes.JustificationBits, bool) {
	if r.justificationBitsChange == nil {
		return bits, false
	}
	return r.justificationBitsChange.Copy(), true
}

func (r *ChangeSet) ApplyCurrentJustifiedCheckpointChange(c *cltypes.Checkpoint) (*cltypes.Checkpoint, bool) {
	if r.currentJustifiedCheckpointChange == nil {
		return c, false
	}
	return r.currentJustifiedCheckpointChange.Copy(), true
}

func (r *ChangeSet) ApplyPreviousJustifiedCheckpointChange(c *cltypes.Checkpoint) (*cltypes.Checkpoint, bool) {
	if r.previousJustifiedCheckpointChange == nil {
		return c, false
	}
	return r.previousJustifiedCheckpointChange.Copy(), true
}

func (r *ChangeSet) ApplyFinalizedCheckpointChange(c *cltypes.Checkpoint) (*cltypes.Checkpoint, bool) {
	if r.finalizedCheckpointChange == nil {
		return c, false
	}
	return r.finalizedCheckpointChange.Copy(), true
}

func (r *ChangeSet) ApplyCurrentSyncCommitteeChange(committee *cltypes.SyncCommittee) (*cltypes.SyncCommittee, bool) {
	if r.currentSyncCommitteeChange == nil {
		return committee, false
	}
	return r.currentSyncCommitteeChange.Copy(), true
}

func (r *ChangeSet) ApplyNextSyncCommitteeChange(committee *cltypes.SyncCommittee) (*cltypes.SyncCommittee, bool) {
	if r.nextSyncCommitteeChange == nil {
		return committee, false
	}
	return r.nextSyncCommitteeChange.Copy(), true
}

func (r *ChangeSet) ApplyLatestExecutionPayloadHeaderChange(eth1Header *cltypes.Eth1Header) (*cltypes.Eth1Header, bool) {
	if r.latestExecutionPayloadHeaderChange == nil {
		return eth1Header, false
	}
	return r.latestExecutionPayloadHeaderChange.Copy(), true
}

func (r *ChangeSet) ApplyNextWithdrawalIndexChange(index uint64) (uint64, bool) {
	if r.nextWithdrawalIndexChange == nil {
		return index, false
	}
	return *r.nextWithdrawalIndexChange, true
}

func (r *ChangeSet) ApplyNextValidatorWithdrawalIndexChange(index uint64) (uint64, bool) {
	if r.nextWithdrawalValidatorIndexChange == nil {
		return index, false
	}
	return *r.nextWithdrawalValidatorIndexChange, true
}

func (r *ChangeSet) ApplyVersionChange(version clparams.StateVersion) (clparams.StateVersion, bool) {
	if r.versionChange == nil {
		return version, false
	}
	return *r.versionChange, true
}