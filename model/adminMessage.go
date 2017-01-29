package model

//AdminMessage contains the status of the current admin user and all other non admin or bold_ally users
type AdminMessage struct {
	TypeID      int
	Type        string
	Logged      bool
	UserEntries []UserEntry
}
