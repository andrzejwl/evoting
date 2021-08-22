package pbft

type Voting struct {
	BlockId  int           `json:"block-id"`
	YesVotes []VoteRequest `json:"yes-votes"` // those who vote to append the block
	NoVotes  []VoteRequest `json:"no-votes"`  // those who vote to reject the block
	Client   Node          `json:"client"`    // requesting party
}

func (v Voting) HasVoted(node Node) bool {
	for _, v := range v.YesVotes {
		if node.Identifier == v.VoterId {
			return true
		}
	}

	for _, v := range v.NoVotes {
		if node.Identifier == v.VoterId {
			return true
		}
	}

	return false
}

func (v Voting) Results() (yes int, no int) {
	// Returns two integers - the first one is YES votes, the second one is NO votes
	yes = len(v.YesVotes)
	no = len(v.NoVotes)
	return
}

type VoteRequest struct {
	BlockId int    `json:"block-id"`
	Vote    string `json:"vote"` // this should be digitally signed by the voter
	VoterId string `json:"voter-id"`
	Client  Node   `json:"client"`
}
