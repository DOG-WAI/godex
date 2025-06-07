package entity

type PhishingSite struct {
	Domain string `json:"domain"`
	Source string `json:"source"`
}

type PhishingSiteCheckRet struct {
	Query  string `json:"query"`
	Domain string `json:"domain"`
	Source string `json:"source"`
}
