package hashgraph

import (
	"math"
	"time"
)

//Node :
type Node struct {
	Address   string             // ip:port of the peer
	Hashgraph map[string][]Event // local copy of hashgraph, map to peer address -> peer events
	Events    map[string]Event   // events as a map of signature -> event
}

//SyncEventsDTO : Data transfer object for 2nd call in Gossip: SyncAllEvents
type SyncEventsDTO struct {
	SenderAddress string
	MissingEvents map[string][]Event // all missing events, map to peer address -> all events I don't know about this peer
}

//GetNumberOfMissingEvents : Node A calls Node B to learn which events B does not know and A knows.
func (n *Node) GetNumberOfMissingEvents(numEventsAlreadyKnown map[string]int, numEventsToSend *map[string]int) error {
	for addr := range n.Hashgraph {
		(*numEventsToSend)[addr] = numEventsAlreadyKnown[addr] - len(n.Hashgraph[addr])
	}
	return nil
}

//SyncAllEvents : Node A first calls GetNumberOfMissingEvents on B, and then sends the missing events in this function
func (n *Node) SyncAllEvents(events SyncEventsDTO, success *bool) error {
	for addr := range events.MissingEvents {
		for _, missingEvent := range events.MissingEvents[addr] {
			n.Hashgraph[addr] = append(n.Hashgraph[addr], missingEvent)
			n.Events[missingEvent.Signature] = missingEvent
		}
	}

	// TODO: create new event
	newEvent := Event{
		Owner:           n.Address,
		Signature:       time.Now().String(), // todo: use RSA
		SelfParentHash:  n.Hashgraph[n.Address][len(n.Hashgraph[n.Address])-1].Signature,
		OtherParentHash: n.Hashgraph[events.SenderAddress][len(n.Hashgraph[events.SenderAddress])-1].Signature,
		Timestamp:       0,   // todo: use date time
		Transactions:    nil, // todo: use the transaction buffer which grows with user input
		Round:           0,
		IsWitness:       false,
	}
	n.Events[newEvent.Signature] = newEvent
	n.Hashgraph[n.Address] = append(n.Hashgraph[n.Address], newEvent)

	// big boi todo's here :)
	n.DivideRounds(newEvent) // IMPORTANT TODO: PASS A POINTER HERE INSTEAD OF COPY?
	n.DecideFame()
	n.FindOrder()

	return nil
}

//DivideRounds : Calculates the round of a new event
func (n Node) DivideRounds(e Event) {
	numNodes := len(n.Hashgraph) // number of nodes, required for strong seeing and supermajority
	selfParent := n.Events[e.SelfParentHash]
	otherParent := n.Events[e.OtherParentHash]
	r := max(selfParent.Round, otherParent.Round)
	// Find round r witnesses
	witnesses := n.findWitnessesOfARound(r)
	// Count strongly seen witnesses for this round
	stronglySeenWitnessCount := 0
	for _, w := range witnesses {
		if n.stronglySee( /* ... */ ) {
			stronglySeenWitnessCount++
		}
	}
	if stronglySeenWitnessCount > int(math.Ceil(2.0*float64(numNodes)/3.0)) {
		e.Round = r + 1
	} else {
		e.Round = r
	}
	e.IsWitness = e.Round > selfParent.Round // we do not check if there is no self parent, because we never create the initial event here
}

//DecideFame : Decides if a witness is famous or not
func (n Node) DecideFame() {

}

//FindOrder : Arrive at a consensus on the order of events
func (n Node) FindOrder() {

}

// If we can reach to target using downward edges only, we can see it. Downward in this case means that we reach through either parent. This function is used for voting
// todo: if required we can optimize  with a global variable to indicate early exit if target is reached
func (n Node) see(current Event, target Event) bool {
	if current.Signature == target.Signature {
		return true
	}
	if current.Round == 1 && current.IsWitness {
		return false
	}
	// Go has short-circuit evaluation, which we utilize here
	return n.see(n.Events[current.SelfParentHash], target) || n.see(n.Events[current.OtherParentHash], target)
}

// If we see the target, and we go through 2n/3 different nodes as we do that, we say we strongly see that target. This function is used for choosing the famous witness
func (n Node) stronglySee(current Event, target Event, seenEvents []Event) (bool, int) {
	// todo
}

// Find witnesses of round r, which is the first event with round r in every node
func (n Node) findWitnessesOfARound(r uint32) map[string]Event {
	var witnesses map[string]Event
	witnesses = make(map[string]Event)
	for addr := range n.Hashgraph {
		for _, event := range n.Hashgraph[addr] {
			if event.Round == r && event.IsWitness {
				witnesses[addr] = event
				break
			}
			if event.Round > r {
				// we do not need to check further events because deeper events always have greater r
				break
			}
		}
	}
	return witnesses // it is possible that a round does not have a witness on each node sometimes
}

// There is no built-in max function for uint32...
func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
