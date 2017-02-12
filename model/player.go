package model

//Player contains informatiom about a weapon
type Player struct {
	Sov        int64
	Note       string
	ExternalID int
	Name       string
	Sa         string `json:"Sa,omitempty"`
	Da         string `json:"Da,omitempty"`
	Sp         string `json:"Sp,omitempty"`
	Se         string `json:"Se,omitempty"`
	Tff        string `json:"Tff,omitempty"`
}
