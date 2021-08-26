package pbft

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Blockchain struct {
	Chain            []Block           `json:"chain"`
	Peers            []Node            `json:"-"` // right now not shared but could be used to propagate new peers
	Votings          map[string]Voting `json:"-"` // key - block ID, block ID stored as string so that the map can be encoded as JSON
	Identifier       string            `json:"node-id"`
	BlockBuffer      map[int]Block     `json:"-"`
	DiscoveryAddress string            `json:"-"`
	Self             Node              `json:"-"`
}

type VotingInfo struct {
	VotingData Voting `json:"voting-data"`
	BlockData  Block  `json:"block-data"`
}

func NewBlockchain(port int) *Blockchain {
	var bc Blockchain
	// uuid in production, for debug id=hostname
	// bc.Identifier = uuid.NewString()
	bc.DiscoveryAddress = os.Getenv("DISCOVERY_ADDR")
	bc.Votings = make(map[string]Voting)
	bc.BlockBuffer = make(map[int]Block)

	hostname := os.Getenv("HOSTNAME")
	// DEBUG mode
	bc.Identifier = hostname

	// generate priv/pub signing key pair
	priv, pub := GenerateSigningKeyPair()

	self := Node{hostname, port, bc.Identifier, "blockchain", pub, &priv}
	bc.Self = self

	bc.RegisterNode()

	bc.RefreshPeers()
	if len(bc.Peers) < 1 {
		fmt.Println("[INFO] too few peers, creating genesis block (peers:", bc.Peers, ")")
		// create genesis block
		initBlock := Block{Identifier: 0, Timestamp: int(time.Now().Unix()), Transactions: []Transaction{}, PreviousBlockHash: ""}
		bc.Chain = append(bc.Chain, initBlock)
	} else {
		fmt.Println("[INFO] fetching blockchain from a peer")
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

func (bc *Blockchain) PeerById(id string) Node {
	for _, peer := range bc.Peers {
		if peer.Identifier == id {
			return peer
		}
	}
	return Node{}
}

func (bc *Blockchain) InsertVote(vote VoteRequest, selfVote bool) {
	voting, exists := bc.Votings[strconv.Itoa(vote.BlockId)]

	voter := bc.PeerById(vote.VoterId)
	if voter == (Node{}) && !selfVote {
		fmt.Println("[ERROR] voter of given id not found", voter, "vote:", vote)
		return
	}

	if selfVote {
		voter = bc.Self
	}

	yesError := VerifySignature(voter.PublicKey, []byte(vote.Vote), "yes")
	noError := VerifySignature(voter.PublicKey, []byte(vote.Vote), "no")

	voteMsg := ""

	if yesError == nil {
		voteMsg = "yes"
	} else if noError == nil {
		voteMsg = "no"
	} else {
		fmt.Println("voter signature doesnt match", voter, "yes", yesError, ", no", noError)
		return
	}

	if !exists {
		// create new voting
		bc.Votings[strconv.Itoa(vote.BlockId)] = Voting{vote.BlockId, []VoteRequest{}, []VoteRequest{}, vote.Client}
		voting = bc.Votings[strconv.Itoa(vote.BlockId)]
	}

	if voteMsg == "yes" {
		voting.YesVotes = append(voting.YesVotes, vote)
	} else if voteMsg == "no" {
		voting.NoVotes = append(voting.NoVotes, vote)
	}

	bc.Votings[strconv.Itoa(voting.BlockId)] = voting
	bc.CheckVotingResults(voting.BlockId)
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

	// remove self from the peer list
	bc.Peers = nil
	for _, peer := range newPeers {
		if peer.Identifier != bc.Self.Identifier {
			bc.Peers = append(bc.Peers, peer)
		}
	}
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

func (bc Blockchain) SignMessage(message string) string {
	/*
		Returns hex encoded string (signed message).
	*/
	signed, _ := SignData([]byte(message), bc.Self.privateKey)

	signedHex := make([]byte, hex.EncodedLen(len(signed)))
	hex.Encode(signedHex, signed)

	return string(signedHex)
}

func (bc *Blockchain) HttpGetChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("GET /chain Request from:", r.RemoteAddr)
	fmt.Fprint(w, json.NewEncoder(w).Encode(bc))
}

func (bc *Blockchain) HttpRequest(w http.ResponseWriter, r *http.Request) {
	// PBFT: Request Phase
	// Node receives transaction data from a client (and thus becomes the "primary replica").

	var req Request
	error := json.NewDecoder(r.Body).Decode(&req)

	w.Header().Set("Content-Type", "application/json")

	if error != nil {
		http.Error(w, HttpJsonBodyPadding(error.Error()), http.StatusBadRequest)
		return
	}
	_, encodingErr := json.Marshal(req.Transactions)

	if encodingErr != nil {
		http.Error(w, HttpJsonBodyPadding(error.Error()), http.StatusBadRequest)
		return
	}

	var newBlock Block
	newBlock.Transactions = append(newBlock.Transactions, req.Transactions...)
	newBlock.Identifier = bc.LastBlock().Identifier + 1
	newBlock.Timestamp = int(time.Now().Unix())
	newBlock.PreviousBlockHash = calculateHash(bc.LastBlock())

	bc.BlockBuffer[newBlock.Identifier] = newBlock

	newVoting := Voting{BlockId: newBlock.Identifier, YesVotes: []VoteRequest{}, NoVotes: []VoteRequest{}, Client: req.Client}
	jsonvoting, jsonerr := json.Marshal(newVoting)

	if jsonerr != nil {
		fmt.Println("json:", string(jsonvoting), "error", jsonerr)
	}

	votingData := VotingInfo{VotingData: newVoting, BlockData: newBlock}

	fmt.Println("[PBFT] Request, new block:", newBlock)

	bc.Votings[strconv.Itoa(newBlock.Identifier)] = newVoting
	// validate
	vote := VoteRequest{BlockId: newVoting.BlockId, Vote: bc.SignMessage("yes"), VoterId: bc.Self.Identifier, Client: req.Client}
	bc.InsertVote(vote, true)

	success := bc.PropagateMessage("pre-prepare", votingData)
	if success {
		fmt.Fprint(w, json.NewEncoder(w).Encode(votingData))
	} else {
		http.Error(w, HttpJsonBodyPadding("node connection error"), http.StatusBadRequest)
	}
	bc.PropagateMessage("prepare", vote)
}

func (bc Blockchain) HttpPrePrepare(w http.ResponseWriter, r *http.Request) {
	// PBFT: Pre-prepare Phase
	// Node receives a block to validate from the primary node.

	w.Header().Set("Content-Type", "application/json")

	var votingInfo VotingInfo
	decodingErr := json.NewDecoder(r.Body).Decode(&votingInfo)
	if decodingErr != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
	}

	block := votingInfo.BlockData
	voting := votingInfo.VotingData

	fmt.Println("[PBFT] Pre-Prepare, block to validate:", block)
	fmt.Fprint(w, json.NewEncoder(w).Encode(HttpJsonBodyPadding("ok")))

	// temporarily we assume all transactions are valid

	vote := VoteRequest{block.Identifier, bc.SignMessage("yes"), bc.Identifier, voting.Client}

	if block.Identifier < bc.LastBlock().Identifier+1 {
		vote.Vote = bc.SignMessage("no")
		bc.InsertVote(vote, false)
		bc.PropagateMessage("prepare", vote)
		return
	}
	// further checks
	bc.BlockBuffer[block.Identifier] = block
	bc.Votings[strconv.Itoa(block.Identifier)] = voting
	bc.InsertVote(vote, true)
	bc.PropagateMessage("prepare", vote)
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
	fmt.Println("[PBFT] Prepare, received vote:", vote)

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

	bc.InsertVote(vote, false)
	fmt.Fprint(w, json.NewEncoder(w).Encode(HttpJsonBodyPadding("ok")))
	// check if their result is the same. If so, check if f+1 votes already received. If so, proceed to commit phase.
}

func (bc Blockchain) HttpGetPending(w http.ResponseWriter, r *http.Request) {
	// Debug endpoint
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(bc.Votings)
	fmt.Fprint(w, json.NewEncoder(w).Encode(bc.Votings))
}

func (bc *Blockchain) CheckVotingResults(blockId int) {
	// Checks if sufficient number of votes has been casted. If so, proceed to commit.
	voting, exists := bc.Votings[strconv.Itoa(blockId)]
	if !exists {
		return
	}

	yesVotes, noVotes := voting.Results()
	minVotes := len(bc.Peers) + 1 - bc.MaximumFaultyNodes()

	if yesVotes >= minVotes || noVotes >= minVotes {
		fmt.Println("[INFO] committing block", blockId)
		bc.Commit(blockId)
	}
}

func (bc *Blockchain) Commit(blockId int) {
	// PBFT: Commit Phase
	// Client node receives replies (votes) from other nodes with the same transactions executed.
	// Before sending the result to client, nodes await f+1 votes where f is the maximum number of faulty nodes.

	block, bExists := bc.BlockBuffer[blockId]
	voting, vExists := bc.Votings[strconv.Itoa(blockId)]

	if !bExists || !vExists {
		fmt.Println("[DEBUG commit] block / voting dont exist, block:", block, ", voting:", voting)
		return
	}

	bc.RefreshPeers() // so that we get accurate minVotes

	yesVotes, noVotes := voting.Results()
	minVotes := len(bc.Peers) + 1 - bc.MaximumFaultyNodes()

	if yesVotes >= minVotes {
		delete(bc.Votings, strconv.Itoa(blockId))
		block.PreviousBlockHash = calculateHash(bc.LastBlock())
		bc.Chain = append(bc.Chain, block)
		message := fmt.Sprintf("{\"node-id\": %v}", blockId)
		messageBuffer, _ := json.Marshal(message)
		_, err := http.Post(fmt.Sprintf("http://%v/commit", voting.Client.String()), "application/json", bytes.NewBuffer(messageBuffer))
		if err != nil {
			// retry?
			return
		}
	} else if noVotes >= minVotes {
		// ??? notify the client their block was rejected?
		return
	}
}
