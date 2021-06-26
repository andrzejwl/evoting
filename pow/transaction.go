package pow

type Transaction struct {
	/*
		not storing issuer ID for privacy reasons (tokenId instead)
		not storing amount because it is always a single vote
	*/
	TokenId string `json:"Token"`
	ToId    string `json:"ToId"`
}

func validateTransaction(ta Transaction) (valid bool, err string) {
	valid = true

	return
}
