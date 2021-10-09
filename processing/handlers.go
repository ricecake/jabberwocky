package processing

import (
	"context"

	"github.com/apex/log"

	"jabberwocky/cluster"
	"jabberwocky/storage"
	"jabberwocky/transport"
)

func HandleEvent(ctx context.Context) {
	log.Info("Starting event loop") // Should this be "request"?  It's more of a request loop... kinda.  Reactive something or other
	for msg := range cluster.Router.GetEventOutbound() {
		// TODO: make this some form of bounded work queue, so that it won't block, but also wont overload the system
		// TODO: should we make this loop responsible for filling in defaults for new objects, or leave that in the storage loop?  kinda like leaving it in storage.
		switch msg.Type {
		case "agent":
			switch msg.SubType {
			case "connect":
				agent, ok := msg.Content.(storage.Agent)
				if !ok {
					log.Errorf("Bad agent in connection message %+v", msg)
					break
				}

				servers, err := storage.ListLiveServers(ctx)
				if err != nil {
					log.Error(err.Error())
					break
				}

				srvList := transport.NewMessage("server", "list", servers)
				cluster.Router.Send(agent.Uuid, cluster.LOCAL_AGENT, srvList)
			}
		}
	}
	log.Info("Leaving event loop")
}

func HandleStorage(ctx context.Context) {
	log.Info("Starting storage loop")
	for msg := range cluster.Router.GetStorageOutbound() {
		switch msg.Type {
		case "server":
			err := storage.SaveServer(ctx, msg.Content.(storage.Server))
			if err != nil {
				log.Error(err.Error())
			}
		case "agent":
			agent, ok := msg.Content.(storage.Agent)
			if !ok {
				log.Errorf("Bad agent in connection message %+v", msg)
				break
			}

			err := storage.SaveAgent(ctx, agent)
			if err != nil {
				log.Error(err.Error())
			}

			if msg.SubType != "sync" { // TODO should make something that will more easily emit a sync message if we've made a change
				msg.SubType = "sync"
				cluster.Router.Emit(cluster.LOCAL_SERVER, msg)
			}
		case "script":
			switch msg.SubType {
			case "create":
				log.Info("Createing a script")
			}
		default:
			log.WithFields(log.Fields{
				"system":  "storage",
				"type":    msg.Type,
				"subtype": msg.SubType,
			}).Info("Unknown message type")
		}
	}
	log.Info("Leaving storage loop")
}

func HandleOutput(ctx context.Context) {
	log.Info("Starting processing loop")
	for msg := range cluster.Router.GetProcessingOutbound() {
		log.WithFields(log.Fields{
			"system":  "output",
			"type":    msg.Type,
			"subtype": msg.SubType,
		}).Info("Unknown message type")
	}
	log.Info("Leaving processing loop")
}
