package pbft

type Transaction struct {
	/*
		not storing issuer ID for privacy reasons (tokenId instead)
		not storing amount because it is always a single vote
	*/
	TokenId string `json:"Token"`
	ToId    string `json:"ToId"`
}

func validateTransaction(ta Transaction) (valid bool, err string) {
	/*
		Returns possible errors as string for more verbose output/log.
	*/
	valid = true

	return
}
