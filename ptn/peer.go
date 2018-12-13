// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ptn

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/consensus/jury"
	//"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"gopkg.in/fatih/set.v0"
)

var (
	errClosed            = errors.New("peer set is closed")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")

	//ID_LENGTH = 32
	//PTNCOIN   = [ID_LENGTH]byte{'p', 't', 'n', 'c', 'o', 'i', 'n'}
)

const (
	maxKnownTxs    = 32768 // Maximum transactions hashes to keep in the known list (prevent DOS)
	maxKnownBlocks = 1024  // Maximum block hashes to keep in the known list (prevent DOS)
	//maxKnownVsss     = 25    // Maximum Vss hashes to keep in the known list (prevent DOS)
	handshakeTimeout = 5 * time.Second

	//transitionStep1  = 1 //All transition mediator each other connected to star vss
	//transitionStep2  = 2 //vss success
	//transitionCancel = 3 //retranstion
)

// PeerInfo represents a short summary of the PalletOne sub-protocol metadata known
// about a connected peer.
type PeerInfo struct {
	Version int `json:"version"` // PalletOne protocol version negotiated
	//Difficulty uint64 `json:"difficulty"` // Total difficulty of the peer's blockchain
	Index uint64 `json:"index"` // Total difficulty of the peer's blockchain
	Head  string `json:"head"`  // SHA3 hash of the peer's best owned block
}

type peerMsg struct {
	head   common.Hash
	number modules.ChainIndex
}

type peer struct {
	id string

	*p2p.Peer
	rw p2p.MsgReadWriter

	version  int         // Protocol version negotiated
	forkDrop *time.Timer // Timed connection dropper if forks aren't validated in time

	peermsg map[modules.IDType16]peerMsg

	lock sync.RWMutex

	knownTxs      *set.Set // Set of transaction hashes known to be known by this peer
	knownBlocks   *set.Set // Set of block hashes known to be known by this peer
	knownGroupSig *set.Set // Set of block hashes known to be known by this peer

	//index modules.ChainIndex
	//mediator bool
	//transitionCh chan int
}

func newPeer(version int, p *p2p.Peer, rw p2p.MsgReadWriter) *peer {

	id := p.ID()
	return &peer{
		Peer:          p,
		rw:            rw,
		version:       version,
		id:            id.TerminalString(),
		knownTxs:      set.New(),
		knownBlocks:   set.New(),
		knownGroupSig: set.New(),
		peermsg:       map[modules.IDType16]peerMsg{},
		//mediator:      false,
		//transitionCh:  make(chan int, 1),
	}
}

/*func (p *peer) ID() int32 {
	p.lock.Lock()
	id := p.id
	p.lock.Unlock()

	return id
}*/
// Info gathers and returns a collection of metadata known about a peer.
func (p *peer) Info( /*assetId modules.IDType16*/ ) *PeerInfo {
	//ptnAssetId, _ := modules.SetIdTypeByHex(dagconfig.DefaultConfig.PtnAssetHex)
	asset := modules.NewPTNAsset()
	hash, number := p.Head(asset.AssetId)

	return &PeerInfo{
		Version: p.version,
		Index:   number.Index,
		Head:    hash.Hex(),
	}
}

// Head retrieves a copy of the current head hash and total difficulty of the
// peer.
//only retain the max index header.will in other mediator,not in ptn mediator.
func (p *peer) Head(assetID modules.IDType16) (hash common.Hash, number modules.ChainIndex) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	msg, ok := p.peermsg[assetID]
	if ok {
		copy(hash[:], msg.head[:])
		number = msg.number
	}
	return hash, number
}

// SetHead updates the head hash and total difficulty of the peer.
//only retain the max index header
func (p *peer) SetHead(hash common.Hash, number modules.ChainIndex) {
	p.lock.Lock()
	defer p.lock.Unlock()

	msg, ok := p.peermsg[number.AssetID]

	if (ok && number.Index > msg.number.Index) || !ok {
		copy(msg.head[:], hash[:])
		msg.number = number
	}
	p.peermsg[number.AssetID] = msg
}

// MarkBlock marks a block as known for the peer, ensuring that the block will
// never be propagated to this particular peer.
func (p *peer) MarkUnit(hash common.Hash) {
	// If we reached the memory allowance, drop a previously known block hash
	for p.knownBlocks.Size() >= maxKnownBlocks {
		p.knownBlocks.Pop()
	}
	p.knownBlocks.Add(hash)
}

// MarkBlock marks a block as known for the peer, ensuring that the block will
// never be propagated to this particular peer.
func (p *peer) MarkGroupSig(hash common.Hash) {
	// If we reached the memory allowance, drop a previously known block hash
	for p.knownGroupSig.Size() >= maxKnownBlocks {
		p.knownGroupSig.Pop()
	}
	p.knownGroupSig.Add(hash)
}

// MarkTransaction marks a transaction as known for the peer, ensuring that it
// will never be propagated to this particular peer.
func (p *peer) MarkTransaction(hash common.Hash) {
	// If we reached the memory allowance, drop a previously known transaction hash
	for p.knownTxs.Size() >= maxKnownTxs {
		p.knownTxs.Pop()
	}
	p.knownTxs.Add(hash)
}

// SendTransactions sends transactions to the peer and includes the hashes
// in its transaction hash set for future reference.
func (p *peer) SendTransactions(txs modules.Transactions) error {
	for _, tx := range txs {
		p.knownTxs.Add(tx.Hash())
	}
	return p2p.Send(p.rw, TxMsg, txs)
}

func (p *peer) SendContractExeTransaction(event jury.ContractExeEvent) error {
	return p2p.Send(p.rw, ContractExecMsg, event)
}

func (p *peer) SendContractSigTransaction(event jury.ContractSigEvent) error {
	return p2p.Send(p.rw, ContractSigMsg, event)
}

//SendConsensus sends consensus msg to the peer
func (p *peer) SendConsensus(msgs string) error {
	return p2p.Send(p.rw, ConsensusMsg, msgs)
}

// SendNewBlockHashes announces the availability of a number of blocks through
// a hash notification.
func (p *peer) SendNewUnitHashes(hashes []common.Hash, numbers []modules.ChainIndex) error {
	for _, hash := range hashes {
		p.knownBlocks.Add(hash)
	}
	request := make(newBlockHashesData, len(hashes))
	for i := 0; i < len(hashes); i++ {
		request[i].Hash = hashes[i]
		request[i].Number = numbers[i]
	}
	return p2p.Send(p.rw, NewBlockHashesMsg, request)
}

// SendNewBlock propagates an entire block to a remote peer.
func (p *peer) SendNewUnit(unit *modules.Unit) error {
	p.knownBlocks.Add(unit.UnitHash)
	return p2p.Send(p.rw, NewBlockMsg, unit)
}

// SendBlockHeaders sends a batch of block headers to the remote peer.
func (p *peer) SendUnitHeaders(headers []*modules.Header) error {
	return p2p.Send(p.rw, BlockHeadersMsg, headers)
}

// SendBlockBodies sends a batch of block contents to the remote peer.
func (p *peer) SendBlockBodies(bodies []blockBody) error {
	return p2p.Send(p.rw, BlockBodiesMsg, blockBodiesData(bodies))
}

// SendBlockBodiesRLP sends a batch of block contents to the remote peer from
// an already RLP encoded format.
func (p *peer) SendBlockBodiesRLP(bodies [][]byte /*[]rlp.RawValue*/) error {
	return p2p.Send(p.rw, BlockBodiesMsg, bodies)
}

// SendNodeDataRLP sends a batch of arbitrary internal data, corresponding to the
// hashes requested.
func (p *peer) SendNodeData(data [][]byte) error {
	return p2p.Send(p.rw, NodeDataMsg, data)
}

// SendReceiptsRLP sends a batch of transaction receipts, corresponding to the
// ones requested from an already RLP encoded format.
func (p *peer) SendReceiptsRLP(receipts []rlp.RawValue) error {
	return p2p.Send(p.rw, ReceiptsMsg, receipts)
}

// RequestOneHeader is a wrapper around the header query functions to fetch a
// single header. It is used solely by the fetcher.
func (p *peer) RequestOneHeader(hash common.Hash) error {
	log.Debug("Fetching single header", "hash", hash)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Hash: hash}, Amount: uint64(1), Skip: uint64(0), Reverse: false})
}

// RequestHeadersByHash fetches a batch of blocks' headers corresponding to the
// specified header query, based on the hash of an origin block.
func (p *peer) RequestHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	log.Debug("Fetching batch of headers", "count", amount, "fromhash", origin, "skip", skip, "reverse", reverse)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Hash: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

// RequestDagHeadersByHash fetches a batch of blocks' headers corresponding to the
// specified header query, based on the hash of an origin block.
func (p *peer) RequestDagHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	//log.Debug("Fetching batch of headers", "count", amount, "fromhash", origin, "skip", skip, "reverse", reverse)
	return nil
}

func (p *peer) RequestLeafNodes() error {
	return nil
}

// RequestHeadersByNumber fetches a batch of blocks' headers corresponding to the
// specified header query, based on the number of an origin block.
func (p *peer) RequestHeadersByNumber(origin modules.ChainIndex, amount int, skip int, reverse bool) error {
	log.Debug("Fetching batch of headers", "count", amount, "index", origin.Index, "skip", skip, "reverse", reverse)
	return p2p.Send(p.rw, GetBlockHeadersMsg, &getBlockHeadersData{Origin: hashOrNumber{Number: origin}, Amount: uint64(amount), Skip: uint64(skip), Reverse: reverse})
}

// RequestBodies fetches a batch of blocks' bodies corresponding to the hashes
// specified.
func (p *peer) RequestBodies(hashes []common.Hash) error {
	log.Debug("Fetching batch of block bodies", "peer id:", p.id, "count", len(hashes))
	return p2p.Send(p.rw, GetBlockBodiesMsg, hashes)
}

// RequestNodeData fetches a batch of arbitrary data from a node's known state
// data, corresponding to the specified hashes.
func (p *peer) RequestNodeData(hashes []common.Hash) error {
	log.Debug("Fetching batch of state data", "count", len(hashes))
	return p2p.Send(p.rw, GetNodeDataMsg, hashes)
}

// RequestReceipts fetches a batch of transaction receipts from a remote node.
func (p *peer) RequestReceipts(hashes []common.Hash) error {
	log.Debug("Fetching batch of receipts", "count", len(hashes))
	return p2p.Send(p.rw, GetReceiptsMsg, hashes)
}

// Handshake executes the ptn protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis blocks.
func (p *peer) Handshake(network uint64, index modules.ChainIndex, genesis common.Hash,
	/*mediator bool,*/ headHash common.Hash) error {
	// Send out own handshake in a new thread
	errc := make(chan error, 2)
	var status statusData // safe to read after two values have been received from errc

	go func() {
		errc <- p2p.Send(p.rw, StatusMsg, &statusData{
			ProtocolVersion: uint32(p.version),
			NetworkId:       network,
			Index:           index,
			GenesisUnit:     genesis,
			//Mediator:        mediator,
			CurrentHeader: headHash,
		})
	}()
	go func() {
		errc <- p.readStatus(network, &status, genesis)
	}()
	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}
	//p.mediator = status.Mediator
	p.SetHead(status.CurrentHeader, status.Index)
	return nil
}

func (p *peer) readStatus(network uint64, status *statusData, genesis common.Hash) (err error) {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != StatusMsg {
		return errResp(ErrNoStatusMsg, "first msg has code %x (!= %x)", msg.Code, StatusMsg)
	}
	if msg.Size > ProtocolMaxMsgSize {
		return errResp(ErrMsgTooLarge, "%v > %v", msg.Size, ProtocolMaxMsgSize)
	}
	// Decode the handshake and make sure everything matches
	if err := msg.Decode(&status); err != nil {
		return errResp(ErrDecode, "msg %v: %v", msg, err)
	}
	if status.GenesisUnit != genesis {
		return errResp(ErrGenesisBlockMismatch, "%x (!= %x)", status.GenesisUnit[:8], genesis[:8])
	}
	if status.NetworkId != network {
		return errResp(ErrNetworkIdMismatch, "%d (!= %d)", status.NetworkId, network)
	}
	if int(status.ProtocolVersion) != p.version {
		return errResp(ErrProtocolVersionMismatch, "%d (!= %d)", status.ProtocolVersion, p.version)
	}

	return nil
}

// String implements fmt.Stringer.
func (p *peer) String() string {
	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("ptn/%2d", p.version),
	)
}

// peerSet represents the collection of active peers currently participating in
// the PalletOne sub-protocol.
type peerSet struct {
	peers map[string]*peer
	//knownVss     *set.Set
	//knownVssResp *set.Set
	//mediators    *set.Set
	lock   sync.RWMutex
	closed bool
}

// newPeerSet creates a new peer set to track the active participants.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
		//knownVss:     set.New(),
		//knownVssResp: set.New(),
		//mediators:    set.New(),
	}
}

//func (ps *peerSet) MediatorsAllConnected() int {
//	return 0
//}

//func (ps *peerSet) MediatorsSize() int {
//	ps.lock.Lock()
//	defer ps.lock.Unlock()
//	return ps.mediators.Size()
//}

//func (ps *peerSet) MediatorsReset(nodes []string) {
//	ps.lock.Lock()
//	defer ps.lock.Unlock()
//	ps.mediators.Clear()
//	for _, node := range nodes {
//		ps.mediators.Add(node)
//	}
//}

//func (ps *peerSet) MediatorsClean() {
//	ps.lock.Lock()
//	defer ps.lock.Unlock()
//	ps.mediators.Clear()
//}

//Make sure there is plenty of connection for Mediator
//func (ps *peerSet) noMediatorCheck(maxPeers int, mediators int) bool {
//	ps.lock.RLock()
//	defer ps.lock.RUnlock()
//
//	size := 0
//	for _, p := range ps.peers {
//		if !p.mediator {
//			size++
//		}
//	}
//	if size > maxPeers-mediators {
//		return false
//	}
//	return true
//}

//Make sure there is plenty of connection for Mediator
//func (ps *peerSet) MediatorCheck() bool {
//	ps.lock.RLock()
//	defer ps.lock.RUnlock()
//	ps.mediators.Size()
//	return true
//}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
func (ps *peerSet) Register(p *peer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if ps.closed {
		return errClosed
	}
	if _, ok := ps.peers[p.id]; ok {
		return errAlreadyRegistered
	}
	ps.peers[p.id] = p
	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if _, ok := ps.peers[id]; !ok {
		return errNotRegistered
	}
	delete(ps.peers, id)
	return nil
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id string) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

// Len returns if the current number of peers in the set.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// PeersWithoutBlock retrieves a list of peers that do not have a given block in
// their set of known hashes.
func (ps *peerSet) PeersWithoutUnit(hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownBlocks.Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

//GroupSig
func (ps *peerSet) PeersWithoutGroupSig(hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownGroupSig.Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

// PeersWithoutTx retrieves a list of peers that do not have a given transaction
// in their set of known hashes.
func (ps *peerSet) PeersWithoutTx(hash common.Hash) []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		if !p.knownTxs.Has(hash) {
			list = append(list, p)
		}
	}
	return list
}

// PeersWithoutVss retrieves a list of peers that do not have a given transaction
// in their set of known hashes.
//func (ps *peerSet) PeersWithoutVss(nodeId string) bool {
//	ps.lock.RLock()
//	defer ps.lock.RUnlock()
//
//	return ps.knownVss.Has(nodeId)
//}

// MarkVss marks a block as known for the peer, ensuring that the block will
// never be propagated to this particular peer.
//func (ps *peerSet) MarkVss(nodeId string) {
//	ps.lock.RLock()
//	defer ps.lock.RUnlock()
//	// If we reached the memory allowance, drop a previously known block hash
//	for ps.knownVss.Size() >= maxKnownVsss {
//		ps.knownVss.Pop()
//	}
//	ps.knownVss.Add(nodeId)
//}

// PeersWithoutVssResp retrieves a list of peers that do not have a given transaction
// in their set of known hashes.
//func (ps *peerSet) PeersWithoutVssResp(nodeId string) bool {
//	ps.lock.RLock()
//	defer ps.lock.RUnlock()
//
//	return ps.knownVssResp.Has(nodeId)
//}

// MarkVssResp marks a block as known for the peer, ensuring that the block will
// never be propagated to this particular peer.
//func (ps *peerSet) MarkVssResp(nodeId string) {
//	// If we reached the memory allowance, drop a previously known block hash
//	for ps.knownVssResp.Size() >= maxKnownVsss {
//		ps.knownVssResp.Pop()
//	}
//	ps.knownVssResp.Add(nodeId)
//}

// BestPeer retrieves the known peer with the currently highest total difficulty.
func (ps *peerSet) BestPeer(assetId modules.IDType16) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	var (
		bestPeer *peer
		bestTd   uint64 = 0 //*big.Int
	)
	for _, p := range ps.peers {
		if _, number := p.Head(assetId); bestPeer == nil || number.Index > bestTd /*td.Cmp(bestTd) > 0*/ {
			bestPeer, bestTd = p, number.Index
		}
	}
	if bestPeer != nil {
		log.Debug("peerSet", "BestPeer:", bestPeer.id)
	}

	return bestPeer
}

// Close disconnects all peers.
// No new peers can be registered after Close has returned.
func (ps *peerSet) Close() {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, p := range ps.peers {
		p.Disconnect(p2p.DiscQuitting)
	}
	for id, _ := range ps.peers {
		delete(ps.peers, id)
	}
	ps.peers = nil
	ps.closed = true
}

func (ps *peerSet) GetPeers() []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		list = append(list, p)
	}
	return list
}
