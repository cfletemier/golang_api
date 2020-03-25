package models

//db model
type Person struct {
	Id			string	`json:"id"`
	FirstName	string	`json:"firstName"`
	LastName	string	`json:"lastName"`
	Age			int		`json:"age"`
}