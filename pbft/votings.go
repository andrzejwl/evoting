package pbft

type Voting struct {
	blockId  int
	yesVotes []VoteRequest // those who vote to append the block
	noVotes  []VoteRequest // those who vote to reject the block
	client   Node          // requesting party
}

func (v Voting) HasVoted(node Node) bool {
	for _, v := range v.yesVotes {
		if node.Identifier == v.VoterId {
			return true
		}
	}

	for _, v := range v.noVotes {
		if node.Identifier == v.VoterId {
			return true
		}
	}

	return false
}

func (v Voting) Results() (yes int, no int) {
	// Returns two integers - the first one is YES votes, the second one is NO votes
	yes = len(v.yesVotes)
	no = len(v.noVotes)
	return
}

type VoteRequest struct {
	BlockId int    `json:"block-id"`
	Vote    string `json:"vote"` // this should be digitally signed by the voter
	VoterId string `json:"voter-id"`
}
