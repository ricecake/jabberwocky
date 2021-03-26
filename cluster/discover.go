package cluster

import (
	"context"
	"strconv"
	"time"

	"github.com/apex/log"
	"github.com/koron/go-ssdp"
	"github.com/spf13/viper"
)

/*
this should be ssdp stuff.

For sanity, we should only broadcast search messages when unclustered, and we should only listen for search messages of our type.
That way, we can hopefully reduce broadcast noise.

Should still accept a seed node in lieu of discovery, but for now assume discovery.
*/

const ssdpService = "urn:swim:jabberwocky:gossip"

func startDiscovery(ctx context.Context) error {
	ad, err := ssdp.Advertise(
		ssdpService,
		"uuid:"+dconf.Name,
		viper.GetString("server.advertise_gossip_interface")+":"+strconv.Itoa(dconf.BindPort),
		"Jabberwocky",
		60,
	)
	if err != nil {
		return err
	}

	searchTimer := time.Tick(time.Duration(5) * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				m.Close()
				ad.Close()
				break
			case <-searchTimer:
				if mlist.NumMembers() > 1 {
					continue
				}

				list, err := ssdp.Search(ssdpService, 1, "")
				if err != nil {
					log.Error(err.Error())
					continue
				}

				var joinList []string
				seenList := make(map[string]bool)
				for _, res := range list {
					if res.Server != "Jabberwocky" || res.USN == "uuid:"+dconf.Name {
						continue
					}

					if !seenList[res.Location] {
						joinList = append(joinList, res.Location)
						seenList[res.Location] = true
					}
				}

				if len(joinList) > 0 {
					joined, err := mlist.Join(joinList)
					if err != nil {
						log.Error(err.Error())
					}
					log.Infof("Joined %d new nodes", joined)
				}
			}
		}
	}()

	return nil
}
