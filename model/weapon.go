package model

//Weapon contains informatiom about a weapon
type Weapon struct {
	Name     string
	Type     string
	Quantity int
	Damage   string
	Strength string
}

//ToString cretes string representation of struct
func (w *Weapon) ToString() string {
	var strVal = "Name : " + w.Name + " , "
	strVal += "Type : " + w.Type + " , "
	strVal += "Quantity : " + string(w.Quantity) + " , "
	strVal += "Damage : " + w.Damage + " , "
	strVal += "Strength : " + w.Strength + " , "
	return strVal
}
