package models

type Cinema struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	IMG         string `db:"img" json:"img"`
	TicketPrice int    `db:"ticket_price" json:"ticket_price"`
}
