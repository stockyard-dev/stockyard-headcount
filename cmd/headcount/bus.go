package main

import (
	"context"
	"log"

	"github.com/stockyard-dev/stockyard-headcount/internal/store"
	"github.com/stockyard-dev/stockyard/bus"
)

// subscribeToBus wires headcount as a subscriber so cross-tool events
// become analytics rows automatically. The whole point: other tools
// (dossier, etc) don't have to know headcount exists. They just publish
// to the bus; headcount translates what it recognizes into Track calls.
//
// Design choices:
//   - Store-level Track call, not HTTP self-call, because the HTTP handler
//     adds request-scoped enrichment (IP, UA parsing) we explicitly do NOT
//     want here — these events are server-side, not browser-side.
//   - Unknown topics are dropped silently (SubscribeAll would add noise
//     for every dossier topic we don't map). If a new topic wants to be
//     tracked later, add a Subscribe line below.
//   - Handler errors are logged, not returned as bus retries — a failed
//     Track shouldn't cause the publisher to see an error, and the bus's
//     retry semantics are overkill for analytics.
func subscribeToBus(b *bus.Bus, db *store.DB) {
	b.Subscribe("contacts.created", func(ctx context.Context, e bus.Event) error {
		if err := db.Track(&store.Event{
			Name:  "contact_created",
			Page:  "(bus)",
			Props: `{"source":"dossier"}`,
		}); err != nil {
			log.Printf("headcount: bus -> Track(contact_created) failed: %v", err)
		}
		return nil
	})
	log.Printf("headcount: subscribed to contacts.created")
}
