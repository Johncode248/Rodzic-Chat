package databasepg

import "fmt"

type Order struct {
	Key       string    `json:"key,omitempty"`
	Direction Direction `json:"direction,omitempty"`
}

func (o *Order) ToString() string {
	return fmt.Sprintf("%s %s", o.Key, o.Direction.String())
}
