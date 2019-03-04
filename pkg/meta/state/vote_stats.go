package state

import (
	"encoding/base64"

	log "github.com/sirupsen/logrus"
	abci "github.com/tendermint/tendermint/abci/types"
)

// HistorySize is how much history we keep for node performance analysis
//
// We may want this to be a variable stored in metastate in the future,
// but for now, a const is good enough.
const HistorySize = 20

// generate noms marshalers
//nomsify NodeRoundStats RoundStats VoteStats

// NodeRoundStats contains information about the votes of a particular node in a particular round
type NodeRoundStats struct {
	Power            int64
	Voted            bool
	AgainstConsensus bool
}

// RoundStats contains information about the validator set votes in a particular round
type RoundStats struct {
	Height     uint64
	Validators map[string]NodeRoundStats
}

// MakeRoundStats collects data from a RequestBeginBlock and converts it to a RoundStats,
// logging as it goes.
func MakeRoundStats(logger log.FieldLogger, req abci.RequestBeginBlock) RoundStats {
	// note: LastRoundStats never indicates what round it actually was; it just
	// asserts that it was the last round. I choose to interpret this as the
	// round whose height is 1 less than the current.
	rs := RoundStats{
		Height:     uint64(req.Header.Height) - 1,
		Validators: make(map[string]NodeRoundStats),
	}

	var (
		voted     map[string]struct{}
		abstained map[string]struct{}
	)

	// fill in the validators
	for _, voteInfo := range req.LastCommitInfo.Votes {
		addr := base64.StdEncoding.EncodeToString(voteInfo.Validator.Address)
		if voteInfo.SignedLastBlock {
			voted[addr] = struct{}{}
		} else {
			abstained[addr] = struct{}{}
		}
		nrs := NodeRoundStats{
			Power: voteInfo.Validator.Power,
			Voted: voteInfo.SignedLastBlock,
		}
		rs.Validators[addr] = nrs
	}
	logger.WithFields(log.Fields{
		"lastCommit.round":     req.LastCommitInfo.Round,
		"validators.voted":     voted,
		"validators.abstained": abstained,
	}).Info("validator vote report")

	// handle the byzantine evidence, which is in a different struct for some reason
	for _, ev := range req.ByzantineValidators {
		addr := base64.StdEncoding.EncodeToString(ev.Validator.Address)
		logger = logger.WithFields(log.Fields{
			"evidence":        ev,
			"evidence.type":   ev.Type,
			"evidence.height": ev.Height,
			"validator":       addr,
		})
		if uint64(ev.Height) != rs.Height {
			logger.WithField("lastCommit.height", rs.Height).Warn("presented with historical evidence of byzantine validation")
			// but we can't do anything about it now
			continue
		}

		// note that it voted against the consensus
		nrs, ok := rs.Validators[addr]
		if !ok {
			logger.Warn("evidence of byzantine validation but validator not in validator set")
			// we're not going to inject a bogus validator into the validator set
			continue
		}
		nrs.AgainstConsensus = true
		rs.Validators[addr] = nrs

		logger.Warn("evidence of byzantine validation")
	}

	return rs
}

// VoteStats is a rolling window of the N most recent rounds of statistics
//
// It's a struct mainly due to nomsify limitations, but this also lets us
// internalize certain functions into methods.
type VoteStats struct {
	History []RoundStats
}

// Append the provided RoundStats to the history
//
// Retain no more than HistorySize items.
func (vs *VoteStats) Append(rs RoundStats) {
	idx0 := len(vs.History) - HistorySize + 1
	if idx0 < 0 {
		idx0 = 0
	}
	vs.History = append(vs.History[idx0:], rs)
}

// AppendRoundStats appends round statistics of the current round to the metastate
func (m *Metastate) AppendRoundStats(logger log.FieldLogger, req abci.RequestBeginBlock) {
	m.Stats.Append(MakeRoundStats(logger, req))
}
