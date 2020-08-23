package rtp

//SeqMod rtp seq mod
const SeqMod = 1 << 16

//SourceStateInfo ...
type SourceStateInfo struct {
	MaxSeq        uint16 /* highest seq. number seen */
	Cycles        uint32 /* shifted count of seq. number cycles */
	BaseSeq       uint32 /* base seq number */
	BadSseq       uint32 /* last ’bad’ seq number + 1 */
	Probation     uint32 /* sequ. packets till source is valid */
	Received      uint32 /* packets received */
	ExpectedPrior uint32 /* packet expected at last interval */
	ReceivedPrior uint32 /* packet received at last interval */
	Transit       uint32 /* relative trans time for prev pkt */
	Jitter        uint32 /* estimated jitter */
}

const maxDropout = 3000
const maxMisOrder = 100
const minSequential = 2

func (state *SourceStateInfo) resetSeq(seq uint16) {
	state.BaseSeq = uint32(seq)
	state.MaxSeq = seq
	state.BadSseq = SeqMod
	state.Cycles = 0
	state.Received = 0
	state.ReceivedPrior = 0
	state.ExpectedPrior = 0
}

//InitSeq ...
func (state *SourceStateInfo) InitSeq(seq uint16) {
	state.resetSeq(seq)
	state.MaxSeq = seq - 1
	state.Probation = minSequential
}

//UpdateSeq ...
func (state *SourceStateInfo) UpdateSeq(seq uint16) bool {
	delta := seq - state.MaxSeq
	//直到收到>minSequential个数的seq后，该源才有效
	if state.Probation != 0 {
		if seq == state.MaxSeq+1 {
			state.Probation--
			state.MaxSeq = seq
			if state.Probation == 0 {
				state.resetSeq(seq)
				//in rfc
				state.Received++
				//my logic
				//state.Received = minSequential
				return true
			}
		} else {
			state.Probation = minSequential - 1
			state.MaxSeq = seq
		}
		return false
	} else if delta < maxDropout {
		if seq < state.MaxSeq {
			state.Cycles += SeqMod
		}
		state.MaxSeq = seq
	} else if delta <= SeqMod-maxMisOrder {
		if uint32(seq) == state.BadSseq {
			state.resetSeq(seq)
		} else {
			state.BadSseq = (uint32(seq) + 1) & (SeqMod - 1)
			return false
		}
	} else {

	}
	state.Received++
	return true
}

//GetLostValues ...
func (state *SourceStateInfo) GetLostValues() (fraction uint8, lost uint32, extendedHighestSequenceNumber uint32) {
	extendMax := state.Cycles + uint32(state.MaxSeq)
	expected := extendMax - state.BadSseq + 1
	lost = expected - state.Received
	expectedInterval := expected - state.ExpectedPrior
	state.ExpectedPrior = expected
	receivedInterval := state.Received - state.ReceivedPrior
	state.ReceivedPrior = state.Received
	lostInterval := int64(expectedInterval) - int64(receivedInterval)
	if expectedInterval == 0 || lostInterval <= 0 {
		fraction = 0
	} else {
		fraction = uint8((uint32(lostInterval) << 8) / expectedInterval)
	}

	extendedHighestSequenceNumber = state.Cycles | uint32(state.MaxSeq)
	return
}
