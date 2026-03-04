package store

import "github.com/google/uuid"

// saltedID generates a UUID that blends serviceID with fresh randomness.
// The result is a valid UUID v5 that cannot collide with IDs produced by a
// different service instance, while revealing nothing about the service's
// identity or count to an external observer.
func saltedID(serviceID uuid.UUID) uuid.UUID {
	name := uuid.New()
	return uuid.NewSHA1(serviceID, name[:])
}
