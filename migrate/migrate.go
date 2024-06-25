package migrate

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/exp/slog"

	"github.com/gravitl/netmaker/database"
	"github.com/gravitl/netmaker/logger"
	"github.com/gravitl/netmaker/logic"
	"github.com/gravitl/netmaker/logic/acls"
	"github.com/gravitl/netmaker/models"
	"github.com/gravitl/netmaker/mq"
	"github.com/gravitl/netmaker/servercfg"
)

// Run - runs all migrations
func Run() {
	updateEnrollmentKeys()
	assignSuperAdmin()
	updateHosts()
	updateNodes()
	updateAcls()
	syncUsers()

}

func assignSuperAdmin() {
	users, err := logic.GetUsers()
	if err != nil || len(users) == 0 {
		return
	}

	if ok, _ := logic.HasSuperAdmin(); ok {
		return
	}
	createdSuperAdmin := false
	owner := servercfg.GetOwnerEmail()
	if owner != "" {
		user, err := logic.GetUser(owner)
		if err != nil {
			log.Fatal("error getting user", "user", owner, "error", err.Error())
		}
		user.IsSuperAdmin = true
		user.IsAdmin = false
		err = logic.UpsertUser(*user)
		if err != nil {
			log.Fatal(
				"error updating user to superadmin",
				"user",
				user.UserName,
				"error",
				err.Error(),
			)
		}
		return
	}
	for _, u := range users {
		if u.IsAdmin {
			user, err := logic.GetUser(u.UserName)
			if err != nil {
				slog.Error("error getting user", "user", u.UserName, "error", err.Error())
				continue
			}
			user.IsSuperAdmin = true
			user.IsAdmin = false
			err = logic.UpsertUser(*user)
			if err != nil {
				slog.Error(
					"error updating user to superadmin",
					"user",
					user.UserName,
					"error",
					err.Error(),
				)
				continue
			} else {
				createdSuperAdmin = true
			}
			break
		}
	}

	if !createdSuperAdmin {
		slog.Error("failed to create superadmin!!")
	}
}

func updateEnrollmentKeys() {
	rows, err := database.FetchRecords(database.ENROLLMENT_KEYS_TABLE_NAME)
	if err != nil {
		return
	}
	for _, row := range rows {
		var key models.EnrollmentKey
		if err = json.Unmarshal([]byte(row), &key); err != nil {
			continue
		}
		if key.Type != models.Undefined {
			logger.Log(2, "migration: enrollment key type already set")
			continue
		} else {
			logger.Log(2, "migration: updating enrollment key type")
			if key.Unlimited {
				key.Type = models.Unlimited
			} else if key.UsesRemaining > 0 {
				key.Type = models.Uses
			} else if !key.Expiration.IsZero() {
				key.Type = models.TimeExpiration
			}
		}
		data, err := json.Marshal(key)
		if err != nil {
			logger.Log(0, "migration: marshalling enrollment key: "+err.Error())
			continue
		}
		if err = database.Insert(key.Value, string(data), database.ENROLLMENT_KEYS_TABLE_NAME); err != nil {
			logger.Log(0, "migration: inserting enrollment key: "+err.Error())
			continue
		}

	}
}

func updateHosts() {
	rows, err := database.FetchRecords(database.HOSTS_TABLE_NAME)
	if err != nil {
		logger.Log(0, "failed to fetch database records for hosts")
	}
	for _, row := range rows {
		var host models.Host
		if err := json.Unmarshal([]byte(row), &host); err != nil {
			logger.Log(0, "failed to unmarshal database row to host", "row", row)
			continue
		}
		if host.PersistentKeepalive == 0 {
			host.PersistentKeepalive = models.DefaultPersistentKeepAlive
			if err := logic.UpsertHost(&host); err != nil {
				logger.Log(0, "failed to upsert host", host.ID.String())
				continue
			}
		}
	}
}

func updateNodes() {
	nodes, err := logic.GetAllNodes()
	if err != nil {
		slog.Error("migration failed for nodes", "error", err)
		return
	}
	for _, node := range nodes {
		if node.IsEgressGateway {
			egressRanges, update := removeInterGw(node.EgressGatewayRanges)
			if update {
				node.EgressGatewayRequest.Ranges = egressRanges
				node.EgressGatewayRanges = egressRanges
				logic.UpsertNode(&node)
			}
		}
	}
}

func removeInterGw(egressRanges []string) ([]string, bool) {
	update := false
	for i := len(egressRanges) - 1; i >= 0; i-- {
		if egressRanges[i] == "0.0.0.0/0" || egressRanges[i] == "::/0" {
			update = true
			egressRanges = append(egressRanges[:i], egressRanges[i+1:]...)
		}
	}
	return egressRanges, update
}

func updateAcls() {
	// get all networks
	networks, err := logic.GetNetworks()
	if err != nil && !database.IsEmptyRecord(err) {
		slog.Error("acls migration failed. error getting networks", "error", err)
		return
	}

	// get current acls per network
	for _, network := range networks {
		var networkAcl acls.ACLContainer
		networkAcl, err := networkAcl.Get(acls.ContainerID(network.NetID))
		if err != nil {
			if database.IsEmptyRecord(err) {
				continue
			}
			slog.Error(fmt.Sprintf("error during acls migration. error getting acls for network: %s", network.NetID), "error", err)
			continue
		}
		// convert old acls to new acls with clients
		// TODO: optimise O(n^2) operation
		clients, err := logic.GetNetworkExtClients(network.NetID)
		if err != nil {
			slog.Error(fmt.Sprintf("error during acls migration. error getting clients for network: %s", network.NetID), "error", err)
			continue
		}
		clientsIdMap := make(map[string]struct{})
		for _, client := range clients {
			clientsIdMap[client.ClientID] = struct{}{}
		}
		nodeIdsMap := make(map[string]struct{})
		for nodeId := range networkAcl {
			nodeIdsMap[string(nodeId)] = struct{}{}
		}
		/*
			initially, networkACL has only node acls so we add client acls to it
			final shape:
			{
				"node1": {
					"node2": 2,
					"client1": 2,
					"client2": 1,
				},
				"node2": {
					"node1": 2,
					"client1": 2,
					"client2": 1,
				},
				"client1": {
					"node1": 2,
					"node2": 2,
					"client2": 1,
				},
				"client2": {
					"node1": 1,
					"node2": 1,
					"client1": 1,
				},
			}
		*/
		for _, client := range clients {
			networkAcl[acls.AclID(client.ClientID)] = acls.ACL{}
			// add client values to node acls and create client acls with node values
			for id, nodeAcl := range networkAcl {
				// skip if not a node
				if _, ok := nodeIdsMap[string(id)]; !ok {
					continue
				}
				if nodeAcl == nil {
					slog.Warn("acls migration bad data: nil node acl", "node", id, "network", network.NetID)
					continue
				}
				nodeAcl[acls.AclID(client.ClientID)] = acls.Allowed
				networkAcl[acls.AclID(client.ClientID)][id] = acls.Allowed
				if client.DeniedACLs == nil {
					continue
				} else if _, ok := client.DeniedACLs[string(id)]; ok {
					nodeAcl[acls.AclID(client.ClientID)] = acls.NotAllowed
					networkAcl[acls.AclID(client.ClientID)][id] = acls.NotAllowed
				}
			}
			// add clients to client acls response
			for _, c := range clients {
				if c.ClientID == client.ClientID {
					continue
				}
				networkAcl[acls.AclID(client.ClientID)][acls.AclID(c.ClientID)] = acls.Allowed
				if client.DeniedACLs == nil {
					continue
				} else if _, ok := client.DeniedACLs[c.ClientID]; ok {
					networkAcl[acls.AclID(client.ClientID)][acls.AclID(c.ClientID)] = acls.NotAllowed
				}
			}
			// delete oneself from its own acl
			delete(networkAcl[acls.AclID(client.ClientID)], acls.AclID(client.ClientID))
		}

		// remove non-existent client and node acls
		for objId := range networkAcl {
			if _, ok := nodeIdsMap[string(objId)]; ok {
				continue
			}
			if _, ok := clientsIdMap[string(objId)]; ok {
				continue
			}
			// remove all occurances of objId from all acls
			for objId2 := range networkAcl {
				delete(networkAcl[objId2], objId)
			}
			delete(networkAcl, objId)
		}

		// save new acls
		slog.Debug(fmt.Sprintf("(migration) saving new acls for network: %s", network.NetID), "networkAcl", networkAcl)
		if _, err := networkAcl.Save(acls.ContainerID(network.NetID)); err != nil {
			slog.Error(fmt.Sprintf("error during acls migration. error saving new acls for network: %s", network.NetID), "error", err)
			continue
		}
		slog.Info(fmt.Sprintf("(migration) successfully saved new acls for network: %s", network.NetID))
	}
}

func MigrateEmqx() {

	err := mq.SendPullSYN()
	if err != nil {
		logger.Log(0, "failed to send pull syn to clients", "error", err.Error())

	}
	time.Sleep(time.Second * 3)
	slog.Info("proceeding to kicking out clients from emqx")
	err = mq.KickOutClients()
	if err != nil {
		logger.Log(2, "failed to migrate emqx: ", "kickout-error", err.Error())
	}

}

func syncUsers() {
	// create default network user roles for existing networks
	networks, _ := logic.GetNetworks()
	nodes, err := logic.GetAllNodes()
	if err == nil {
		for _, netI := range networks {
			networkNodes := logic.GetNetworkNodesMemory(nodes, netI.NetID)
			for _, networkNodeI := range networkNodes {
				if networkNodeI.IsIngressGateway {
					h, err := logic.GetHost(networkNodeI.HostID.String())
					if err == nil {
						logic.CreateRole(models.UserRolePermissionTemplate{
							ID:                  models.UserRole(fmt.Sprintf("net-%s-user-gw-%s", netI.NetID, h.Name)),
							DenyDashboardAccess: true,
							NetworkID:           netI.NetID,
							NetworkLevelAccess: map[models.RsrcType]map[models.RsrcID]models.RsrcPermissionScope{
								models.RemoteAccessGwRsrc: {
									models.RsrcID(networkNodeI.ID.String()): models.RsrcPermissionScope{
										VPNaccess: true,
									},
								},
							},
						})
					}

				}
			}
		}
	}

	users, err := logic.GetUsersDB()
	if err == nil {
		for _, user := range users {
			if user.IsSuperAdmin {
				user.PlatformRoleID = models.SuperAdminRole
				logic.UpsertUser(user)
			} else if user.IsAdmin {
				user.PlatformRoleID = models.AdminRole
				logic.UpsertUser(user)
			} else {
				user.PlatformRoleID = models.ServiceUser
				logic.UpsertUser(user)
			}
			if len(user.RemoteGwIDs) > 0 {
				// define user roles for network
				// assign relevant network role to user
				for remoteGwID := range user.RemoteGwIDs {
					gwNode, err := logic.GetNodeByID(remoteGwID)
					if err != nil {
						continue
					}
					h, err := logic.GetHost(gwNode.HostID.String())
					if err != nil {
						continue
					}
					r, err := logic.GetRole(models.UserRole(fmt.Sprintf("net-%s-user-gw-%s", gwNode.Network, h.Name)))
					if err != nil {
						continue
					}
					if netRoles, ok := user.NetworkRoles[models.NetworkID(gwNode.Network)]; ok {
						netRoles[r.ID] = struct{}{}
					} else {
						user.NetworkRoles[models.NetworkID(gwNode.Network)] = map[models.UserRole]struct{}{
							r.ID: {},
						}
					}
				}
				logic.UpsertUser(user)
			}
		}
	}
}
