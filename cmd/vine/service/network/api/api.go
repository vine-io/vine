// Copyright 2020 lack
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package api is the network api
package api

import (
	"context"
	"errors"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lack-io/cli"

	"github.com/lack-io/vine"
	pb "github.com/lack-io/vine/proto/services/network"
	goapi "github.com/lack-io/vine/service/api"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/network"
	"github.com/lack-io/vine/service/network/resolver"
)

var (
	privateBlocks []*net.IPNet
)

func init() {
	for _, b := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10", "fd00::/8"} {
		if _, block, err := net.ParseCIDR(b); err == nil {
			privateBlocks = append(privateBlocks, block)
		}
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, priv := range privateBlocks {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

type node struct {
	id      string
	address string
	network string
	metric  int64
}

type Network struct {
	client pb.NetworkService
	closed chan bool

	mtx sync.RWMutex
	// indexed by address
	nodes map[string]*node
}

func getIP(addr string) (string, error) {
	if strings.HasPrefix(addr, "[::]") {
		return "", errors.New("ip is loopback")
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", errors.New("ip is blank")
	}

	if isPrivateIP(ip) {
		return "", errors.New("private ip")
	}

	return addr, nil
}

func setPeers(peer *pb.Peer, peers map[string]string) {
	if peer == nil || peer.Node == nil {
		return
	}
	ip, err := getIP(peer.Node.Address)
	if err == nil {
		peers[peer.Node.Id] = ip
	} else {
		log.Debugf("Error getting peer IP: %v %+v\n", err, peer.Node)
	}
	for _, p := range peer.Peers {
		setPeers(p, peers)
	}
}

func (n *Network) setCache() {
	rsp, err := n.client.Graph(context.TODO(), &pb.GraphRequest{})
	if err != nil {
		log.Debugf("Failed to get nodes: %v\n", err)
		return
	}

	routesRsp, err := n.client.Routes(context.TODO(), &pb.RoutesRequest{})
	if err != nil {
		log.Debugf("Failed to get routes: %v\n", err)
		return
	}

	// create a map of the peers
	peers := make(map[string]string)

	// set node 0
	setPeers(rsp.Root, peers)

	// don't proceed without peers
	if len(peers) == 0 {
		log.Debugf("Peer node list is 0... not saving")
		return
	}

	// build a route graph
	nodes := make(map[string]*node)

	for _, route := range routesRsp.Routes {
		// find in the peer graph
		ip, ok := peers[route.Router]
		if !ok {
			continue
		}

		// skip routes that don't match in peer ip list
		if ip != route.Gateway && ip != route.Address {
			continue
		}

		// skip already saved routes with better metric
		if n, ok := nodes[ip+route.Network]; ok && n.metric < route.Metric {
			continue
		}

		// add as gateway + network
		nodes[ip+route.Network] = &node{
			id:      route.Router,
			address: ip,
			network: route.Network,
			metric:  route.Metric,
		}
	}

	// set the nodes
	n.mtx.Lock()
	n.nodes = nodes
	n.mtx.Unlock()

	log.Debugf("Set nodes: %+v\n", nodes)
}

func (n *Network) cache() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	// set the cache
	n.setCache()

	for {
		select {
		case <-t.C:
			n.setCache()
		case <-n.closed:
			return
		}
	}
}

func (n *Network) stop() {
	select {
	case <-n.closed:
		return
	default:
		close(n.closed)
	}
}

// TODO: get remote IP and compare to peer list to order by nearest nodes
func (n *Network) Nodes(ctx context.Context, req *map[string]interface{}, rsp *map[string]interface{}) error {
	n.mtx.RLock()
	defer n.mtx.RUnlock()

	network := network.DefaultName

	// check if the option is passed in
	r := *req
	if name, ok := r["name"].(string); ok && len(name) > 0 {
		network = name
	}

	var nodes []*resolver.Record

	// make copy of nodes
	for _, node := range n.nodes {
		// only accept of this network name if exists
		if node.network != network {
			continue
		}
		// append to nodes
		nodes = append(nodes, &resolver.Record{
			Address:  node.address,
			Priority: node.metric,
		})
	}

	// sort by lowest priority
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Priority < nodes[j].Priority })

	// make peer response
	nodeRsp := map[string]interface{}{
		"nodes": nodes,
	}

	// set peer response
	*rsp = nodeRsp
	return nil
}

func Run(ctx *cli.Context) {
	// create the api service
	api := vine.NewService(
		vine.Name("go.vine.api.network"),
	)

	// create the network client
	netClient := pb.NewNetworkService("go.vine.network", api.Client())

	// create new api network handler
	netHandler := &Network{
		client: netClient,
		closed: make(chan bool),
		nodes:  make(map[string]*node),
	}

	// run the cache
	go netHandler.cache()
	defer netHandler.stop()

	// create endpoint
	ep := &goapi.Endpoint{
		Name:    "Network.Nodes",
		Path:    []string{"^/network/?$"},
		Method:  []string{"GET"},
		Handler: "rpc",
	}

	// register the handler
	vine.RegisterHandler(api.Server(), netHandler, goapi.WithEndpoint(ep))

	// run the api
	api.Run()
}
