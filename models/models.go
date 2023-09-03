package models

type RealmInfo struct {
	Realms      []Realm `json:"realms"`
	OnlineCount int     `json:"online_count"`
}

type Realm struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsOnline bool   `json:"isOnline"`
	Online   int    `json:"online"`
}
