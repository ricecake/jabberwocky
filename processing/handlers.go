package processing

import (
	"context"

	"github.com/apex/log"

	"jabberwocky/cluster"
	"jabberwocky/storage"
	"jabberwocky/transport"
)

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

			switch msg.SubType {
			case "connect":
				servers, err := storage.ListLiveServers(ctx)
				if err != nil {
					log.Error(err.Error())
					break
				}

				srvList := transport.NewMessage("server", "list", servers)
				cluster.Router.Send(agent.Uuid, cluster.LOCAL_AGENT, srvList)
			}

			if msg.SubType != "sync" {
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
			"type":    msg.Type,
			"subtype": msg.SubType,
		}).Info("Unknown message type")
	}
	log.Info("Leaving processing loop")
}
