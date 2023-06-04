package helpers

import "github.com/google/uuid"

func GetOptionalUUIDParams(value *uuid.UUID) any {
	if value != nil {
		return *value
	} else {
		return nil
	}
}

func GetOptionalStringParams(value *string) any {
	if value != nil {
		return *value
	} else {
		return nil
	}
}

func GetOptionalIntParams(value *int) any {
	if value != nil {
		return *value
	} else {
		return nil
	}
}

type Optional struct {
	Photo string `json:"photo"`
}
