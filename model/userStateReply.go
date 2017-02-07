package model

//UserStateReply reply to user get state, contains info about player and user
type UserStateReply struct {
	Type   int
	Player Player
}
