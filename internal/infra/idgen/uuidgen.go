package idgen

import "github.com/google/uuid"

type UUIDGen struct{}

func (UUIDGen) New() string {
	return uuid.New().String()
}
