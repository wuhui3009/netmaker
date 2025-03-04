package models

import (
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	// PLACEHOLDER_KEY_TEXT - access key placeholder text if option turned off
	PLACEHOLDER_KEY_TEXT = "ACCESS_KEY"
	// PLACEHOLDER_TOKEN_TEXT - access key token placeholder text if option turned off
	PLACEHOLDER_TOKEN_TEXT = "ACCESS_TOKEN"
)

// AuthParams - struct for auth params
type AuthParams struct {
	MacAddress string `json:"macaddress"`
	ID         string `json:"id"`
	Password   string `json:"password"`
}

// User struct - struct for Users
type User struct {
	UserName string   `json:"username" bson:"username" validate:"min=3,max=40,in_charset|email"`
	Password string   `json:"password" bson:"password" validate:"required,min=5"`
	Networks []string `json:"networks" bson:"networks"`
	IsAdmin  bool     `json:"isadmin" bson:"isadmin"`
	Groups   []string `json:"groups" bson:"groups" yaml:"groups"`
}

// ReturnUser - return user struct
type ReturnUser struct {
	UserName string   `json:"username" bson:"username"`
	Networks []string `json:"networks" bson:"networks"`
	IsAdmin  bool     `json:"isadmin" bson:"isadmin"`
	Groups   []string `json:"groups" bson:"groups"`
}

// UserAuthParams - user auth params struct
type UserAuthParams struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// UserClaims - user claims struct
type UserClaims struct {
	IsAdmin  bool
	UserName string
	Networks []string
	Groups   []string
	jwt.RegisteredClaims
}

// SuccessfulUserLoginResponse - successlogin struct
type SuccessfulUserLoginResponse struct {
	UserName  string
	AuthToken string
}

// Claims is  a struct that will be encoded to a JWT.
// jwt.StandardClaims is an embedded type to provide expiry time
type Claims struct {
	ID         string
	MacAddress string
	Network    string
	jwt.RegisteredClaims
}

// SuccessfulLoginResponse is struct to send the request response
type SuccessfulLoginResponse struct {
	ID        string
	AuthToken string
}

// ErrorResponse is struct for error
type ErrorResponse struct {
	Code    int
	Message string
}

// NodeAuth - struct for node auth
type NodeAuth struct {
	Network    string
	Password   string
	MacAddress string // Depricated
	ID         string
}

// SuccessResponse is struct for sending error message with code.
type SuccessResponse struct {
	Code     int
	Message  string
	Response interface{}
}

// DisplayKey - what is displayed for key
type DisplayKey struct {
	Name string `json:"name" bson:"name"`
	Uses int    `json:"uses" bson:"uses"`
}

// GlobalConfig - global config
type GlobalConfig struct {
	Name string `json:"name" bson:"name"`
}

// CheckInResponse - checkin response
type CheckInResponse struct {
	Success          bool   `json:"success" bson:"success"`
	NeedPeerUpdate   bool   `json:"needpeerupdate" bson:"needpeerupdate"`
	NeedConfigUpdate bool   `json:"needconfigupdate" bson:"needconfigupdate"`
	NeedKeyUpdate    bool   `json:"needkeyupdate" bson:"needkeyupdate"`
	NeedDelete       bool   `json:"needdelete" bson:"needdelete"`
	NodeMessage      string `json:"nodemessage" bson:"nodemessage"`
	IsPending        bool   `json:"ispending" bson:"ispending"`
}

// PeersResponse - peers response
type PeersResponse struct {
	PublicKey           string `json:"publickey" bson:"publickey"`
	Endpoint            string `json:"endpoint" bson:"endpoint"`
	Address             string `json:"address" bson:"address"`
	Address6            string `json:"address6" bson:"address6"`
	LocalAddress        string `json:"localaddress" bson:"localaddress"`
	LocalListenPort     int32  `json:"locallistenport" bson:"locallistenport"`
	IsEgressGateway     string `json:"isegressgateway" bson:"isegressgateway"`
	EgressGatewayRanges string `json:"egressgatewayrange" bson:"egressgatewayrange"`
	ListenPort          int32  `json:"listenport" bson:"listenport"`
	KeepAlive           int32  `json:"persistentkeepalive" bson:"persistentkeepalive"`
}

// ExtPeersResponse - ext peers response
type ExtPeersResponse struct {
	PublicKey       string `json:"publickey" bson:"publickey"`
	Endpoint        string `json:"endpoint" bson:"endpoint"`
	Address         string `json:"address" bson:"address"`
	Address6        string `json:"address6" bson:"address6"`
	LocalAddress    string `json:"localaddress" bson:"localaddress"`
	LocalListenPort int32  `json:"locallistenport" bson:"locallistenport"`
	ListenPort      int32  `json:"listenport" bson:"listenport"`
	KeepAlive       int32  `json:"persistentkeepalive" bson:"persistentkeepalive"`
}

// EgressGatewayRequest - egress gateway request
type EgressGatewayRequest struct {
	NodeID     string   `json:"nodeid" bson:"nodeid"`
	NetID      string   `json:"netid" bson:"netid"`
	NatEnabled string   `json:"natenabled" bson:"natenabled"`
	Ranges     []string `json:"ranges" bson:"ranges"`
}

// RelayRequest - relay request struct
type RelayRequest struct {
	NodeID     string   `json:"nodeid" bson:"nodeid"`
	NetID      string   `json:"netid" bson:"netid"`
	RelayAddrs []string `json:"relayaddrs" bson:"relayaddrs"`
}

// HostRelayRequest - struct for host relay creation
type HostRelayRequest struct {
	HostID       string   `json:"host_id"`
	RelayedHosts []string `json:"relayed_hosts"`
}

// IngressRequest - ingress request struct
type IngressRequest struct {
	ExtclientDNS string `json:"extclientdns"`
	Failover     bool   `json:"failover"`
}

// ServerUpdateData - contains data to configure server
// and if it should set peers
type ServerUpdateData struct {
	UpdatePeers bool       `json:"updatepeers" bson:"updatepeers"`
	Node        LegacyNode `json:"servernode" bson:"servernode"`
}

// Telemetry - contains UUID of the server and timestamp of last send to posthog
// also contains assymetrical encryption pub/priv keys for any server traffic
type Telemetry struct {
	UUID           string `json:"uuid" bson:"uuid"`
	LastSend       int64  `json:"lastsend" bson:"lastsend"`
	TrafficKeyPriv []byte `json:"traffickeypriv" bson:"traffickeypriv"`
	TrafficKeyPub  []byte `json:"traffickeypub" bson:"traffickeypub"`
}

// ServerAddr - to pass to clients to tell server addresses and if it's the leader or not
type ServerAddr struct {
	IsLeader bool   `json:"isleader" bson:"isleader" yaml:"isleader"`
	Address  string `json:"address" bson:"address" yaml:"address"`
}

// TrafficKeys - struct to hold public keys
type TrafficKeys struct {
	Mine   []byte `json:"mine" bson:"mine" yaml:"mine"`
	Server []byte `json:"server" bson:"server" yaml:"server"`
}

// HostPull - response of a host's pull
type HostPull struct {
	Host         Host                 `json:"host" yaml:"host"`
	Peers        []wgtypes.PeerConfig `json:"peers" yaml:"peers"`
	ServerConfig ServerConfig         `json:"server_config" yaml:"server_config"`
	PeerIDs      PeerMap              `json:"peer_ids,omitempty" yaml:"peer_ids,omitempty"`
}

// NodeGet - struct for a single node get response
type NodeGet struct {
	Node         Node                 `json:"node" bson:"node" yaml:"node"`
	Host         Host                 `json:"host" yaml:"host"`
	Peers        []wgtypes.PeerConfig `json:"peers" bson:"peers" yaml:"peers"`
	HostPeers    []wgtypes.PeerConfig `json:"host_peers" bson:"host_peers" yaml:"host_peers"`
	ServerConfig ServerConfig         `json:"serverconfig" bson:"serverconfig" yaml:"serverconfig"`
	PeerIDs      PeerMap              `json:"peerids,omitempty" bson:"peerids,omitempty" yaml:"peerids,omitempty"`
}

// NodeJoinResponse data returned to node in response to join
type NodeJoinResponse struct {
	Node         Node                 `json:"node" bson:"node" yaml:"node"`
	Host         Host                 `json:"host" yaml:"host"`
	ServerConfig ServerConfig         `json:"serverconfig" bson:"serverconfig" yaml:"serverconfig"`
	Peers        []wgtypes.PeerConfig `json:"peers" bson:"peers" yaml:"peers"`
}

// ServerConfig - struct for dealing with the server information for a netclient
type ServerConfig struct {
	CoreDNSAddr string       `yaml:"corednsaddr"`
	API         string       `yaml:"api"`
	APIPort     string       `yaml:"apiport"`
	DNSMode     string       `yaml:"dnsmode"`
	Version     string       `yaml:"version"`
	MQPort      string       `yaml:"mqport"`
	MQUserName  string       `yaml:"mq_username"`
	MQPassword  string       `yaml:"mq_password"`
	Server      string       `yaml:"server"`
	Broker      string       `yaml:"broker"`
	Is_EE       bool         `yaml:"isee"`
	StunPort    int          `yaml:"stun_port"`
	StunList    []StunServer `yaml:"stun_list"`
	TrafficKey  []byte       `yaml:"traffickey"`
	TurnDomain  string       `yaml:"turn_domain"`
	TurnPort    int          `yaml:"turn_port"`
	UseTurn     bool         `yaml:"use_turn"`
}

// User.NameInCharset - returns if name is in charset below or not
func (user *User) NameInCharSet() bool {
	charset := "abcdefghijklmnopqrstuvwxyz1234567890-."
	for _, char := range user.UserName {
		if !strings.Contains(charset, strings.ToLower(string(char))) {
			return false
		}
	}
	return true
}

// ServerIDs - struct to hold server ids.
type ServerIDs struct {
	ServerIDs []string `json:"server_ids"`
}

// JoinData - struct to hold data required for node to join a network on server
type JoinData struct {
	Host Host   `json:"host" yaml:"host"`
	Node Node   `json:"node" yaml:"node"`
	Key  string `json:"key" yaml:"key"`
}

// StunServer - struct to hold data required for using stun server
type StunServer struct {
	Domain string `json:"domain" yaml:"domain"`
	Port   int    `json:"port" yaml:"port"`
}
