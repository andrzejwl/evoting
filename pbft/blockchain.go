package pbft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Blockchain struct {
	Chain            []Block        `json:"chain"`
	Peers            []Node         `json:"-"` // right now not shared but could be used to propagate new peers
	Votings          map[int]Voting `json:"-"` // key - block ID
	Identifier       string         `json:"node-id"`
	BlockBuffer      map[int]Block  `json:"-"`
	DiscoveryAddress string         `json:"-"`
	Self             Node           `json:"-"`
}

func NewBlockchain(port int) *Blockchain {
	var bc Blockchain
	bc.Identifier = uuid.NewString()
	bc.DiscoveryAddress = "127.0.0.1:9999"
	bc.Votings = make(map[int]Voting)
	bc.BlockBuffer = make(map[int]Block)

	self := Node{"127.0.0.1", port, bc.Identifier, "blockchain"}
	bc.Self = self

	bc.RegisterNode()

	bc.RefreshPeers()
	if len(bc.Peers) < 2 {
		// create genesis block
		initBlock := Block{Identifier: 0, Timestamp: int(time.Now().Unix()), Transactions: []Transaction{}, PreviousBlockHash: ""}
		bc.Chain = append(bc.Chain, initBlock)
	} else {
		// fetch blockchain from a random peer
		peer := RandomNode(bc.Peers)

		// prevent asking self for the current chain
		for peer.Address == bc.Self.Address && peer.Port == bc.Self.Port {
			peer = RandomNode(bc.Peers)
		}

		resp, error := http.Get(fmt.Sprintf("http://%v/chain", peer))
		if error != nil {
			fmt.Printf("[ERROR] failed to fetch blockchain data from %v", peer)
		} else {
			newChain, decodingErr := ReconstructBlockchain(resp.Body)
			if decodingErr != "" {
				fmt.Println("[ERROR] erroring parsing peer blockchain:", decodingErr)
			}
			bc.Chain = newChain.Chain
		}
	}

	return &bc
}

func ReconstructBlockchain(r io.ReadCloser) (Blockchain, string) {
	var bc Blockchain
	decodingErr := json.NewDecoder(r).Decode(&bc)

	if decodingErr != nil {
		return Blockchain{}, decodingErr.Error()
	}

	return bc, ""
}

func (bc Blockchain) MaximumFaultyNodes() int {
	// Max faulty nodes for PBFT is floor((n-1)/3).

	return int(len(bc.Peers) / 3) // peers + self = n
}

func (bc Blockchain) RegisterNode() {
	messageBuffer, _ := json.Marshal(bc.Self)
	resp, err := http.Post(fmt.Sprintf("http://%v/register", bc.DiscoveryAddress), "application/json", bytes.NewBuffer(messageBuffer))

	if err != nil || resp.StatusCode != 200 {
		fmt.Println("[ERROR] failed to register node at node discovery service")
		if err != nil {
			fmt.Println("details:", err.Error())
		} else {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("status code:", resp.StatusCode, ", discovery response:", string(bodyBytes))
		}
	} else {
		fmt.Println("[INFO] Node registered")
	}
}

func (bc Blockchain) LastBlock() Block {
	if len(bc.Chain) < 1 { // chain empty
		return Block{}
	}
	return bc.Chain[len(bc.Chain)-1]
}

func (bc Blockchain) PeerById(id string) Node {
	for _, peer := range bc.Peers {
		if peer.Identifier == id {
			return peer
		}
	}
	return Node{}
}

func (bc *Blockchain) InsertVote(vote VoteRequest) {
	voting, exists := bc.Votings[vote.BlockId]

	if !exists {
		// create new voting
	}

	if vote.Vote == "yes" {
		voting.yesVotes = append(voting.yesVotes, vote)
	} else if vote.Vote == "no" {
		voting.noVotes = append(voting.noVotes, vote)
	}

	bc.Votings[voting.blockId] = voting
}

func (bc *Blockchain) RefreshPeers() {
	var newPeers []Node
	resp, err := http.Get(fmt.Sprintf("http://%v/get-blockchain", bc.DiscoveryAddress))

	if err != nil {
		fmt.Println("[CLIENT] failed to refresh nodes")
		return
	}

	decodingErr := json.NewDecoder(resp.Body).Decode(&newPeers)

	if decodingErr != nil {
		fmt.Println("[ERROR] node discovery's response is ambiguous")
		return
	}

	bc.Peers = newPeers
}

func (bc *Blockchain) PropagateMessage(endpoint string, message interface{}) bool {
	// Makes an HTTP post request to all discovered peers.
	// Returns true if propagation was successful, else false.

	messageBuffer, bufferErr := json.Marshal(message)
	// TODO: might want to use digital signatures here for node authentication
	if bufferErr != nil {
		return false
	}

	failedCtr := 0

	bc.RefreshPeers()

	for _, peer := range bc.Peers {
		_, err := http.Post(fmt.Sprintf("http://%v/%v", peer.String(), endpoint), "application/json", bytes.NewBuffer(messageBuffer))
		if err != nil {
			failedCtr++
			if failedCtr > bc.MaximumFaultyNodes() {
				return false
			}
		}
	}
	return true
}

func (bc *Blockchain) HttpGetChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("GET /chain Request from:", r.RemoteAddr)
	fmt.Fprint(w, json.NewEncoder(w).Encode(bc))
}

func (bc Blockchain) HttpRequest(w http.ResponseWriter, r *http.Request) {
	// PBFT: Request Phase
	// Node receives transaction data from a client (and thus becomes the "primary replica").

	var transactions []Transaction
	error := json.NewDecoder(r.Body).Decode(&transactions)

	w.Header().Set("Content-Type", "application/json")

	if error != nil {
		http.Error(w, HttpJsonBodyPadding(error.Error()), http.StatusBadRequest)
		return
	}
	_, encodingErr := json.Marshal(transactions)

	if encodingErr != nil {
		http.Error(w, HttpJsonBodyPadding(error.Error()), http.StatusBadRequest)
		return
	}

	var newBlock Block
	newBlock.Transactions = append(newBlock.Transactions, transactions...)
	newBlock.Identifier = bc.LastBlock().Identifier + 1
	newBlock.Timestamp = int(time.Now().Unix())

	success := bc.PropagateMessage("pre-prepare", newBlock)
	if success {
		fmt.Fprint(w, json.NewEncoder(w).Encode(newBlock))
	} else {
		http.Error(w, HttpJsonBodyPadding("node connection error"), http.StatusBadRequest)
	}
}

func (bc Blockchain) HttpPrePrepare(w http.ResponseWriter, r *http.Request) {
	// PBFT: Pre-prepare Phase
	// Node receives a block to validate from the primary node.

	w.Header().Set("Content-Type", "application/json")

	var block Block
	decodingErr := json.NewDecoder(r.Body).Decode(&block)
	if decodingErr != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
	}
	fmt.Fprint(w, json.NewEncoder(w).Encode(HttpJsonBodyPadding("ok")))

	// temporarily we assume all transactions are valid

	if block.Identifier < bc.LastBlock().Identifier+1 {
		bc.PropagateMessage("prepare", VoteRequest{block.Identifier, "no", bc.Identifier})
		return
	}
	// further checks
	bc.BlockBuffer[block.Identifier] = block
	bc.PropagateMessage("prepare", VoteRequest{block.Identifier, "yes", bc.Identifier})
}

func (bc *Blockchain) HttpPrepare(w http.ResponseWriter, r *http.Request) {
	// PBFT: Prepare Phase
	// Accept other nodes' votes.

	w.Header().Set("Content-Type", "application/json")
	var vote VoteRequest
	decodingErr := json.NewDecoder(r.Body).Decode(&vote)

	if decodingErr != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
	}
	fmt.Fprint(w, json.NewEncoder(w).Encode(HttpJsonBodyPadding("ok")))

	voter := bc.PeerById(vote.VoterId)
	if voter == (Node{}) {
		// TODO: verify signature
		http.Error(w, HttpJsonBodyPadding("peer not found"), http.StatusForbidden)
		return
	}

	if vote == (VoteRequest{}) {
		http.Error(w, HttpJsonBodyPadding("vote invalid"), http.StatusBadRequest)
		return
	}

	bc.InsertVote(vote)
	fmt.Fprint(w, json.NewEncoder(w).Encode(HttpJsonBodyPadding("ok")))

	// check if their result is the same. If so, check if f+1 votes already received. If so, proceed to commit phase.
}

func (bc *Blockchain) CheckVotingResults(blockId int) {
	// Checks if sufficient number of votes has been casted. If so, proceed to commit.
	voting, exists := bc.Votings[blockId]
	if !exists {
		return
	}

	yesVotes, noVotes := voting.Results()
	minVotes := len(bc.Peers) + 1 - bc.MaximumFaultyNodes()

	if yesVotes > minVotes || noVotes > minVotes {
		bc.Commit(blockId)
	}
}

func (bc *Blockchain) Commit(blockId int) {
	// PBFT: Commit Phase
	// Client node receives replies (votes) from other nodes with the same transactions executed.
	// Before sending the result to client, nodes await f+1 votes where f is the maximum number of faulty nodes.

	block, bExists := bc.BlockBuffer[blockId]
	voting, vExists := bc.Votings[blockId]

	if !bExists || !vExists {
		return
	}

	yesVotes, noVotes := voting.Results()
	minVotes := len(bc.Peers) + 1 - bc.MaximumFaultyNodes()

	if yesVotes > minVotes {
		delete(bc.Votings, blockId)
		block.PreviousBlockHash = calculateHash(bc.LastBlock())
		bc.Chain = append(bc.Chain, block)
		message := fmt.Sprintf("{\"node-id\": %v}", blockId)
		messageBuffer, _ := json.Marshal(message)
		_, err := http.Post(fmt.Sprintf("http://%v/commit", voting.client.String()), "application/json", bytes.NewBuffer(messageBuffer))
		if err != nil {
			// retry?
			return
		}
	} else if noVotes > minVotes {
		// ??? notify the client their block was rejected?
		return
	}
}
