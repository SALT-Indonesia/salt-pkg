package rbmq

import (
	"encoding/json"
)

type TopUpBalance struct {
	Id     string  `json:"id"`
	Amount float64 `json:"amount"`
}

func (t *TopUpBalance) Serialize() ([]byte, error) {
	return json.Marshal(&t)
}

func (t *TopUpBalance) UnSerialize(data []byte) error {
	err := json.Unmarshal(data, t)
	if err != nil {
		return err
	}
	return nil
}
