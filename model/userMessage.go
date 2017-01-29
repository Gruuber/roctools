package model

//UserMessage contains the status of the current user
type UserMessage struct {
	TypeID int
	Type   string
	Logged bool
}
