package model

//UserEntry contains information about non admin or bold_ally users
type UserEntry struct {
	ExternalID int
	Username   string
	Challenge  string
	TypeID     int
}
