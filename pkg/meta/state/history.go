package state

import (
	"github.com/attic-labs/noms/go/datas"
	nt "github.com/attic-labs/noms/go/types"
	"github.com/pkg/errors"
)

// StopIteration can be raised inside the IterHistory callback to stop iteration
// in a way distinguishable from an actual error.
//
// Iterators, python-style.
func StopIteration() error {
	return stopIteration{}
}

type stopIteration struct{}

func (stopIteration) Error() string {
	return "Iteration should stop here"
}

// IsStopIteration returns true if the supplied error is stopIteration
func IsStopIteration(err error) bool {
	if err != nil {
		_, isstopIteration := err.(stopIteration)
		return isstopIteration
	}
	return false
}

// IterHistory iterates backward through history from the current head of the DB.
//
// If the callback function returns a non-nil error, iteration is terminated.
// If the returned error is stopIteration, IterHistory returns nil. Otherwise,
// the error is propagated.
//
// ## Caution
//
// This is fundamentlly a read-only interface. It is not impossible to get a state
// which is not the head, and make edits to it. However, it is a logic error to do so.
// Committing edits made to a state which is not the head is like committing edits
// in git while in detached HEAD mode: the resultant state is orphaned and unreachable.
//
// One may be tempted to manually dig in and issue noms merge commands to make that
// state reachable. Don't do that. Doing so will break an invariant which this function
// depends on.
//
// ## Technical Background
//
// noms is a git-like database. This gives it many properties that we desire, but
// it also means that it has some complications that we don't. As in git, though
// the common case is for a single commit to have a single parent, there exists
// the notion of a merge commit which has more than one parent.
//
// Unlike git, our use of noms can maintain the invariant that any given noms database
// will have at most a single client, and any commit will have at most a single parent.
// We can ensure this because we're using tendermint to establish consensus, instead
// of noms' native merge system. That invariant makes iterators like this possible.
func IterHistory(
	db datas.Database, ds datas.Dataset,
	example State,
	cb func(state State, hash string, height uint64) error,
) error {
	headRef, hasHead := ds.MaybeHeadRef()
	for hasHead {
		// get the data we care about
		metastate, err := metastateAt(db, headRef, example)
		if err != nil {
			return err
		}

		// call the callback
		err = cb(metastate.ChildState, metastate.Blockhash, uint64(metastate.Height))
		if err != nil {
			if IsStopIteration(err) {
				return nil
			}
			return err
		}

		// move back in time
		headRefP, err := parentOf(db, headRef)
		if err != nil {
			return errors.Wrap(err, "IterHistory invariant varied")
		}
		if headRefP == nil {
			hasHead = false
		} else {
			headRef = *headRefP
		}
	}
	return nil
}

// AtHeight retrieves the state as of a given tendermint height and puts it into
// the provided State object.
//
// Runtime is O(n) where n is the difference between the current noms head
// height and the noms head height of the desired TM height.
// n is not visible to external applications, but it will always be
// t * m, where t is the difference between the current tendermint head height
// and the desired TM head height, and m is a float in the range [0,1].
func AtHeight(
	db datas.Database, ds datas.Dataset,
	state State,
	wantHeight uint64,
) error {
	headRef, hasHead := ds.MaybeHeadRef()
	if !hasHead {
		return errors.New("AtHeight: No head in this dataset")
	}
	metastate, err := metastateAt(db, headRef, state)
	if err != nil {
		return err
	}
	if wantHeight > uint64(metastate.Height) {
		return errors.New("Requested height higher than current head")
	} else if wantHeight == 0 || wantHeight == uint64(metastate.Height) {
		state = metastate.ChildState
		return nil
	}

	err = IterHistory(db, ds, state, func(hstate State, hash string, height uint64) error {
		// IterHistory iterates backwards down heights, and doesn't iterate
		// over any heights for which no transactions occurred. Therefore,
		// the correct state is the _first_ in this backwards iteration for
		// which the height <= the desired height
		if height <= wantHeight {
			state = hstate
			return StopIteration()
		}
		return nil
	})
	// errors.Wrap returns nil if err == nil
	return errors.Wrap(err, "AtHeight failed iterating history")
}

func metastateAt(db datas.Database, ref nt.Ref, example State) (Metastate, error) {
	metastate := newMetaState(db, example)
	metastateV := ref.TargetValue(db).(nt.Struct).Get(datas.ValueField).(nt.Struct)
	err := metastate.unmarshal(metastateV, example)
	return metastate, errors.Wrap(err, "Failed to unmarshal metastate")
}

func parentOf(db datas.Database, ref nt.Ref) (*nt.Ref, error) {
	parents := ref.TargetValue(db).(nt.Struct).Get(datas.ParentsField).(nt.Set)
	firstParent := parents.First()
	if firstParent == nil {
		return nil, nil
	}

	if !setSizeEq1(parents) {
		return nil, errors.New("more than 1 commit parent found")
	}

	fpRef := firstParent.(nt.Ref)
	return &fpRef, nil
}

// setSizeEq1 returns true if the set contains exactly one element
//
// This is a dumb function to have to write, but apparently nt.Set doesn't have
// a Size() or Len() function, so we have to do this
func setSizeEq1(set nt.Set) bool {
	size := 0
	set.Iter(func(nt.Value) bool {
		size++
		// return a bool. if true, stop iteration.
		// don't stop when size == 1: if we do, we don't know whether or not
		// the set is larger than 1.
		return size > 1
	})
	return size == 1
}
