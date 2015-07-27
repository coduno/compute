package models

import (
	"golang.org/x/net/context"
	"google.golang.org/cloud/datastore"
)

// ChallangeKind is the kind used to
// store challenges in Datastore.
const ChallengeKind = "challenges"

// Challenge contains the data of a challenge
// with the company that created it.
type Challenge struct {
	Name         string         `json:"name"`
	Instructions string         `json:"instructions"`
	Company      *datastore.Key `json:"company"`
	WebInterface string         `json:"webInterface"`
	Runner       string         `json:"-"`
	Flags        string         `json:"-"`
	Languages    []string       `json:"languages"`
}

// Save a new challange to Datastore.
func (challenge Challenge) Save(ctx context.Context) (*datastore.Key, error) {
	return datastore.Put(ctx, datastore.NewIncompleteKey(ctx, ChallengeKind, nil), &challenge)
}