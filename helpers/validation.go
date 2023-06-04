package helpers

import "github.com/google/uuid"

func IsValidUUID(value uuid.UUID) any {
	var emptyUUID uuid.UUID

	if emptyUUID.String() != value.String() {
		return true
	} else {
		return false
	}
}
