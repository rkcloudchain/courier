/*
Copyright Rockontrol Corp. All Rights Reserved.
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gossip

import (
	"bytes"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/rkcloudchain/rksync/channel"
	"github.com/rkcloudchain/rksync/common"
	"github.com/rkcloudchain/rksync/config"
	"github.com/rkcloudchain/rksync/discovery"
	"github.com/rkcloudchain/rksync/filter"
	"github.com/rkcloudchain/rksync/identity"
	"github.com/rkcloudchain/rksync/lib"
	"github.com/rkcloudchain/rksync/logging"
	"github.com/rkcloudchain/rksync/protos"
	"github.com/rkcloudchain/rksync/rpc"
	"github.com/rkcloudchain/rksync/util"
	"google.golang.org/grpc"
)

const (
	presumedDeadChanSize = 100
	acceptChanSize       = 100
)

// NewGossipService creates a gossip instance attached to a gRPC server
func NewGossipService(gConf *config.GossipConfig, idConf *config.IdentityConfig, s *grpc.Server,
	selfIdentity common.PeerIdentityType, secureDialOpts func() []grpc.DialOption) (Gossip, error) {

	g := &gossipService{
		selfIdentity:          selfIdentity,
		conf:                  gConf,
		id:                    idConf.ID,
		presumedDead:          make(chan common.PKIidType, presumedDeadChanSize),
		toDieChan:             make(chan struct{}, 1),
		stopFlag:              int32(0),
		includeIdentityPeriod: time.Now().Add(gConf.PublishCertPeriod),
		ChannelDeMultiplexer:  rpc.NewChannelDemultiplexer(),
	}
	g.chainStateMsgStore = g.newChainStateMsgStore()

	var err error
	g.idMapper, err = identity.NewIdentity(idConf, selfIdentity, func(pkiID common.PKIidType) {
		g.srv.CloseConn(&common.NetworkMember{PKIID: pkiID})
	})
	if err != nil {
		return nil, err
	}

	g.selfPKIid = g.idMapper.GetPKIidOfCert(selfIdentity)
	g.chanState = newChannelState(g)
	g.srv = rpc.NewServer(s, g.idMapper, selfIdentity, secureDialOpts)
	g.emitter = newBatchingEmitter(gConf.PropagateIterations, gConf.MaxPropagationBurstSize,
		gConf.MaxPropagationBurstLatency, g.sendGossipBatch)

	g.discAdapter = g.newDiscoveryAdapter()
	g.disc = discovery.NewDiscoveryService(g.selfNetworkMember(), g.discAdapter, g.newDiscoverySecurityAdapter(), g.disclosurePolicy)
	logging.Infof("Creating gossip service with self membership of %s", g.selfNetworkMember())

	g.stopSignal.Add(2)
	go g.start()
	go g.connect2BootstrapPeers()

	return g, nil
}

type gossipService struct {
	id                    string
	selfIdentity          common.PeerIdentityType
	selfPKIid             common.PKIidType
	includeIdentityPeriod time.Time
	idMapper              identity.Identity
	srv                   *rpc.Server
	conf                  *config.GossipConfig
	emitter               batchingEmitter
	disc                  discovery.Discovery
	stopSignal            sync.WaitGroup
	stopFlag              int32
	toDieChan             chan struct{}
	presumedDead          chan common.PKIidType
	discAdapter           *discoveryAdapter
	chanState             *channelState
	chainStateMsgStore    lib.MessageStore
	*rpc.ChannelDeMultiplexer
}

func (g *gossipService) SelfChannelInfo(chainID string) *protos.ChainState {
	ch := g.chanState.getChannelByChainID(chainID)
	if ch == nil {
		return nil
	}
	return ch.Self()
}

func (g *gossipService) SelfPKIid() common.PKIidType {
	return g.selfPKIid
}

func (g *gossipService) Peers() []common.NetworkMember {
	if g.toDie() {
		return []common.NetworkMember{}
	}
	return g.disc.GetMembership()
}

func (g *gossipService) Accept(acceptor common.MessageAcceptor, passThrough bool) (<-chan *protos.RKSyncMessage, <-chan protos.ReceivedMessage) {
	if passThrough {
		return nil, g.srv.Accept(acceptor)
	}
	acceptByType := func(o interface{}) bool {
		if o, isRKSyncMsg := o.(*protos.RKSyncMessage); isRKSyncMsg {
			return acceptor(o)
		}
		if o, isSignedMsg := o.(*protos.SignedRKSyncMessage); isSignedMsg {
			return acceptor(o.RKSyncMessage)
		}
		logging.Warning("Message type: ", reflect.TypeOf(o), "cannot be evaluated")
		return false
	}

	inCh := g.AddChannel(acceptByType)
	outCh := make(chan *protos.RKSyncMessage, acceptChanSize)
	go func() {
		for {
			select {
			case s := <-g.toDieChan:
				g.toDieChan <- s
				return
			case m := <-inCh:
				if m == nil {
					return
				}
				outCh <- m.(*protos.SignedRKSyncMessage).RKSyncMessage
			}
		}
	}()
	return outCh, nil
}

func (g *gossipService) InitializeChannel(chainID string, chainState *protos.ChainState) error {
	if chainID == "" {
		return errors.New("Channel ID must be provided")
	}
	if g.toDie() {
		return errors.New("RKSync service is stopping")
	}

	mac := channel.GenerateMAC(g.selfPKIid, chainID)
	if !bytes.Equal(chainState.ChainMac, mac) {
		return errors.Errorf("Channel MAC doesn't match, expected: %s, actual: %s", hex.EncodeToString(mac), hex.EncodeToString(chainState.ChainMac))
	}

	signedMsg, err := chainState.Envelope.ToRKSyncMessage()
	if err != nil {
		return errors.Wrapf(err, "Failed to parse channel %s state information", chainID)
	}

	err = signedMsg.Verify(g.selfPKIid, func(peerIdentity []byte, signature, message []byte) error {
		return g.idMapper.Verify(peerIdentity, signature, message)
	})
	if err != nil {
		return errors.Wrapf(err, "Failed verifying %s chain state information signature: %s", chainID, err)
	}

	stateInfo, err := chainState.GetChainStateInfo()
	if err != nil {
		return errors.Errorf("Channel %s: state information format error: %s", chainID, err)
	}
	if !bytes.Equal(common.PKIidType(stateInfo.Leader), g.selfPKIid) {
		return errors.Errorf("Channel %s: current peer's PKI-ID (%s) doesn't match the leader PKI-ID (%s)", chainID, g.selfPKIid, common.PKIidType(stateInfo.Leader))
	}

	gc := g.chanState.joinChannel(chainID, true)
	return gc.InitializeWithChainState(chainState)
}

func (g *gossipService) AddMemberToChan(chainID string, member common.PKIidType) (*protos.ChainState, error) {
	gc := g.chanState.getChannelByChainID(chainID)
	if gc == nil {
		return nil, errors.Errorf("Channel %s not yet created", chainID)
	}

	return gc.AddMember(member)
}

func (g *gossipService) AddFileToChan(chainID string, file common.FileSyncInfo) (*protos.ChainState, error) {
	gc := g.chanState.getChannelByChainID(chainID)
	if gc == nil {
		return nil, errors.Errorf("Channel %s not yet created", chainID)
	}

	return gc.AddFile(file)
}

func (g *gossipService) GetPKIidOfCert(nodeID string, cert *x509.Certificate) (common.PKIidType, error) {
	nodeIDRaw := []byte(nodeID)
	pb := &pem.Block{Bytes: cert.Raw, Type: "CERTIFICATE"}
	pemBytes := pem.EncodeToMemory(pb)
	if pemBytes == nil {
		return nil, errors.New("Encoding of certificate failed")
	}

	raw := append(nodeIDRaw, pemBytes...)
	digest := util.ComputeSHA256(raw)
	return digest, nil
}

func (g *gossipService) CreateChannel(chainID string, files []common.FileSyncInfo) (*protos.ChainState, error) {
	if chainID == "" {
		return nil, errors.New("Channel ID must be provided")
	}
	if g.toDie() {
		return nil, errors.New("RKSync service is stopping")
	}

	gc := g.chanState.joinChannel(chainID, true)
	return gc.Initialize([]common.PKIidType{g.selfPKIid}, files)
}

func (g *gossipService) CloseChannel(chainID string) {
	if chainID == "" {
		return
	}

	g.chanState.removeChannel(chainID)
}

func (g *gossipService) Stop() {
	if g.toDie() {
		return
	}

	atomic.StoreInt32(&g.stopFlag, int32(1))
	logging.Info("Stopping gossip")
	defer logging.Info("Stopped gossip")
	g.discAdapter.close()
	g.disc.Stop()
	g.chanState.stop()
	g.toDieChan <- struct{}{}
	g.emitter.Stop()
	g.ChannelDeMultiplexer.Close()
	g.stopSignal.Wait()
	g.srv.Stop()
}

func (g *gossipService) selfNetworkMember() common.NetworkMember {
	return common.NetworkMember{
		Endpoint: g.conf.Endpoint,
		PKIID:    g.srv.GetPKIid(),
	}
}

func (g *gossipService) sendGossipBatch(a []interface{}) {
	msgs2Gossip := make([]*emittedRKSyncMessage, len(a))
	for i, e := range a {
		msgs2Gossip[i] = e.(*emittedRKSyncMessage)
	}
	g.gossipBatch(msgs2Gossip)
}

func (g *gossipService) gossipBatch(msgs []*emittedRKSyncMessage) {
	if g.disc == nil {
		logging.Error("Discovery has not been initialized yet, aborting")
		return
	}

	var chainStateMsgs []*emittedRKSyncMessage

	isAChainStateMsg := func(o interface{}) bool {
		return o.(*emittedRKSyncMessage).IsChainStateMsg()
	}

	chainStateMsgs, msgs = partitionMessages(isAChainStateMsg, msgs)
	for _, chainStateMsg := range chainStateMsgs {
		peerSelector := func(member common.NetworkMember) bool {
			return chainStateMsg.filter(member.PKIID)
		}
		gc := g.chanState.getChannelByChainID(string(chainStateMsg.Channel))
		if gc != nil {
			peerSelector = filter.CombineRoutingFilters(peerSelector, gc.IsMemberInChan)
		}

		peers2Send := filter.SelectPeers(g.conf.PropagatePeerNum, g.disc.GetMembership(), peerSelector)
		g.srv.Send(chainStateMsg.SignedRKSyncMessage, peers2Send...)
	}

	for _, msg := range msgs {
		if !msg.IsAliveMsg() {
			logging.Error("Unknow message type", msg)
			continue
		}

		selector := filter.CombineRoutingFilters(filter.SelectAllPolicy, func(member common.NetworkMember) bool {
			return msg.filter(member.PKIID)
		})
		peers2Send := filter.SelectPeers(g.conf.PropagatePeerNum, g.disc.GetMembership(), selector)
		g.srv.Send(msg.SignedRKSyncMessage, peers2Send...)
	}
}

func (g *gossipService) start() {
	go g.syncDiscovery()
	go g.handlePresumedDead()

	msgSelector := func(msg interface{}) bool {
		gMsg, isRKSyncMsg := msg.(protos.ReceivedMessage)
		if !isRKSyncMsg {
			return false
		}

		isConn := gMsg.GetRKSyncMessage().GetConn() != nil
		isEmpty := gMsg.GetRKSyncMessage().GetEmpty() != nil

		return !(isConn || isEmpty)
	}

	incMsgs := g.srv.Accept(msgSelector)

	go g.acceptMessages(incMsgs)

	logging.Info("RKSync gossip instance", g.id, "started")
}

func (g *gossipService) acceptMessages(incMsgs <-chan protos.ReceivedMessage) {
	defer logging.Debug("Exiting")
	defer g.stopSignal.Done()
	for {
		select {
		case s := <-g.toDieChan:
			g.toDieChan <- s
			return
		case msg := <-incMsgs:
			g.handleMessage(msg)
		}
	}
}

func (g *gossipService) handleMessage(m protos.ReceivedMessage) {
	if g.toDie() {
		return
	}

	if m == nil || m.GetRKSyncMessage() == nil {
		return
	}

	msg := m.GetRKSyncMessage()

	logging.Debug("Entering,", m.GetConnectionInfo(), "sent us", msg)
	defer logging.Debug("Exiting")

	if !g.validateMsg(m) {
		logging.Warning("Message", msg, "isn't valid")
		return
	}

	if msg.IsChainStateMsg() {
		g.emitter.Add(&emittedRKSyncMessage{
			SignedRKSyncMessage: msg,
			filter:              m.GetConnectionInfo().ID.IsNotSameFilter,
		})

		added := g.chainStateMsgStore.Add(msg)
		if added {
			gc := g.chanState.lookupChannelForMsg(m)
			if gc == nil && g.isInChannel(m) {
				gc = g.chanState.joinChannel(string(msg.Channel), false)
			}

			if gc != nil {
				gc.HandleMessage(m)
			}
		}
		return
	}

	if msg.IsChannelRestricted() {
		gc := g.chanState.lookupChannelForMsg(m)
		if gc != nil {
			gc.HandleMessage(m)
		}
		return
	}

	if selectOnlyDiscoveryMessages(m) {
		if m.GetRKSyncMessage().GetMemReq() != nil {
			sMsg, err := m.GetRKSyncMessage().GetMemReq().SelfInformation.ToRKSyncMessage()
			if err != nil {
				logging.Warningf("Got membership request with invalid selfInfo: %+v", errors.WithStack(err))
				return
			}
			if !sMsg.IsAliveMsg() {
				logging.Warning("Got membership request with selfInfo that isn't an AliveMessage")
				return
			}
			if !bytes.Equal(sMsg.GetAliveMsg().Membership.PkiId, m.GetConnectionInfo().ID) {
				logging.Warning("Got membership request with selfInfo that doesn't match the handshake")
				return
			}
		}
		g.forwardDiscoveryMsg(m)
	}
}

func (g *gossipService) isInChannel(m protos.ReceivedMessage) bool {
	msg := m.GetRKSyncMessage()
	chainStateInfo, err := msg.GetState().GetChainStateInfo()
	if err != nil {
		logging.Errorf("Failed unmarshalling ChainStateInfo message: %v", err)
		return false
	}

	for _, member := range chainStateInfo.Properties.Members {
		if bytes.Equal(member, g.selfPKIid) {
			return true
		}
	}

	return false
}

func (g *gossipService) forwardDiscoveryMsg(msg protos.ReceivedMessage) {
	g.discAdapter.RLock()
	defer g.discAdapter.RUnlock()
	if g.discAdapter.toDie() {
		return
	}

	g.discAdapter.incChan <- msg
}

func (g *gossipService) handlePresumedDead() {
	defer logging.Debug("Exiting")
	defer g.stopSignal.Done()
	for {
		select {
		case s := <-g.toDieChan:
			g.toDieChan <- s
			return
		case deadEndpoint := <-g.srv.PresumedDead():
			g.presumedDead <- deadEndpoint
		}
	}
}

// validateMsg checks the signature of the message if exists.
func (g *gossipService) validateMsg(msg protos.ReceivedMessage) bool {
	if err := msg.GetRKSyncMessage().IsTagLegal(); err != nil {
		logging.Warningf("Tag of %v isn't legal: %v", msg.GetRKSyncMessage(), errors.WithStack(err))
		return false
	}

	return true
}

func (g *gossipService) syncDiscovery() {
	logging.Debug("Entering discovery sync with interval", g.conf.PullInterval)
	defer logging.Debug("Exiting discovery sync loop")

	for !g.toDie() {
		g.disc.InitiateSync(g.conf.PullPeerNum)
		time.Sleep(g.conf.PullInterval)
	}
}

func (g *gossipService) connect2BootstrapPeers() {
	for _, endpoint := range g.conf.BootstrapPeers {
		identifier := func() (common.PKIidType, error) {
			remotePeerIdentity, err := g.srv.Handshake(&common.NetworkMember{Endpoint: endpoint})
			if err != nil {
				return nil, errors.WithStack(err)
			}
			pkiID := g.idMapper.GetPKIidOfCert(remotePeerIdentity)
			if len(pkiID) == 0 {
				return nil, errors.Errorf("Wasn't able to extract PKI-ID of remote peer with identity of %v", remotePeerIdentity)
			}
			return pkiID, nil
		}
		g.disc.Connect(common.NetworkMember{Endpoint: endpoint}, identifier)
	}
}

func (g *gossipService) toDie() bool {
	return atomic.LoadInt32(&g.stopFlag) == int32(1)
}

func (g *gossipService) newChainStateMsgStore() lib.MessageStore {
	pol := protos.NewRKSyncMessageComparator()
	return lib.NewMessageStoreExpirable(pol,
		lib.Noop,
		g.conf.PublishStateInfoInterval*100,
		nil,
		nil,
		lib.Noop)
}

func selectOnlyDiscoveryMessages(m interface{}) bool {
	msg, isRKSyncMsg := m.(protos.ReceivedMessage)
	if !isRKSyncMsg {
		return false
	}
	alive := msg.GetRKSyncMessage().GetAliveMsg()
	memRes := msg.GetRKSyncMessage().GetMemRes()
	memReq := msg.GetRKSyncMessage().GetMemReq()

	selected := alive != nil || memRes != nil || memReq != nil
	return selected
}

func (g *gossipService) newDiscoveryAdapter() *discoveryAdapter {
	return &discoveryAdapter{
		srv:      g.srv,
		stopping: int32(0),
		gossipFunc: func(msg *protos.SignedRKSyncMessage) {
			if g.conf.PropagateIterations == 0 {
				return
			}
			g.emitter.Add(&emittedRKSyncMessage{
				SignedRKSyncMessage: msg,
				filter:              func(_ common.PKIidType) bool { return true },
			})
		},
		forwardFunc: func(msg protos.ReceivedMessage) {
			if g.conf.PropagateIterations == 0 {
				return
			}
			g.emitter.Add(&emittedRKSyncMessage{
				SignedRKSyncMessage: msg.GetRKSyncMessage(),
				filter:              msg.GetConnectionInfo().ID.IsNotSameFilter,
			})
		},
		incChan:          make(chan protos.ReceivedMessage),
		presumedDead:     g.presumedDead,
		disclosurePolicy: g.disclosurePolicy,
	}
}

func (g *gossipService) disclosurePolicy(remotePeer *common.NetworkMember) (discovery.Sieve, discovery.EnvelopeFilter) {
	return func(msg *protos.SignedRKSyncMessage) bool {
			if !msg.IsAliveMsg() {
				logging.Fatal("Programing error, this should be used only on alive message")
			}

			return msg.GetAliveMsg().Membership.Endpoint != "" && remotePeer.Endpoint != ""

		}, func(msg *protos.SignedRKSyncMessage) *protos.Envelope {
			envelope := proto.Clone(msg.Envelope).(*protos.Envelope)
			return envelope
		}
}

// discoveryAdapter is used to supply the discovery module with needed abilities
type discoveryAdapter struct {
	stopping         int32
	srv              *rpc.Server
	presumedDead     chan common.PKIidType
	incChan          chan protos.ReceivedMessage
	gossipFunc       func(message *protos.SignedRKSyncMessage)
	forwardFunc      func(message protos.ReceivedMessage)
	disclosurePolicy discovery.DisclosurePolicy
	sync.RWMutex
}

func (da *discoveryAdapter) close() {
	atomic.StoreInt32(&da.stopping, int32(1))

	da.Lock()
	defer da.Unlock()
	close(da.incChan)
}

func (da *discoveryAdapter) toDie() bool {
	return atomic.LoadInt32(&da.stopping) == int32(1)
}

func (da *discoveryAdapter) Gossip(msg *protos.SignedRKSyncMessage) {
	da.RLock()
	defer da.RUnlock()
	if da.toDie() {
		return
	}

	da.gossipFunc(msg)
}

func (da *discoveryAdapter) Forward(msg protos.ReceivedMessage) {
	da.RLock()
	defer da.RUnlock()
	if da.toDie() {
		return
	}

	da.forwardFunc(msg)
}

func (da *discoveryAdapter) SendToPeer(peer *common.NetworkMember, msg *protos.SignedRKSyncMessage) {
	da.RLock()
	defer da.RUnlock()
	if da.toDie() {
		return
	}

	if memReq := msg.GetMemReq(); memReq != nil && len(peer.PKIID) != 0 {
		selfMsg, err := memReq.SelfInformation.ToRKSyncMessage()
		if err != nil {
			panic(errors.Wrap(err, "Tried to send a membership request with a malformed AliveMessage"))
		}

		_, omitConcealedFields := da.disclosurePolicy(peer)
		selfMsg.Envelope = omitConcealedFields(selfMsg)
		oldKnown := memReq.Known
		memReq = &protos.MembershipRequest{
			SelfInformation: selfMsg.Envelope,
			Known:           oldKnown,
		}
		msgCopy := proto.Clone(msg.RKSyncMessage).(*protos.RKSyncMessage)

		msgCopy.Content = &protos.RKSyncMessage_MemReq{
			MemReq: memReq,
		}
		msg, err := (&protos.SignedRKSyncMessage{
			RKSyncMessage: msgCopy,
		}).NoopSign()

		if err != nil {
			return
		}
		da.srv.Send(msg, peer)
		return
	}
	da.srv.Send(msg, peer)
}

func (da *discoveryAdapter) Ping(peer *common.NetworkMember) bool {
	err := da.srv.Probe(peer)
	return err == nil
}

func (da *discoveryAdapter) Accept() <-chan protos.ReceivedMessage {
	return da.incChan
}

func (da *discoveryAdapter) PresumedDead() <-chan common.PKIidType {
	return da.presumedDead
}

func (da *discoveryAdapter) CloseConn(peer *common.NetworkMember) {
	da.srv.CloseConn(peer)
}

func (g *gossipService) newDiscoverySecurityAdapter() *discoverySecurityAdapter {
	return &discoverySecurityAdapter{
		idMapper:              g.idMapper,
		includeIdentityPeriod: g.includeIdentityPeriod,
		identity:              g.selfIdentity,
	}
}

type discoverySecurityAdapter struct {
	identity              common.PeerIdentityType
	includeIdentityPeriod time.Time
	idMapper              identity.Identity
}

func (sa *discoverySecurityAdapter) ValidateAliveMsg(m *protos.SignedRKSyncMessage) bool {
	am := m.GetAliveMsg()
	if am == nil || am.Membership == nil || am.Membership.PkiId == nil || !m.IsSigned() {
		logging.Warning("Invalid alive message:", m)
		return false
	}

	if am.Identity != nil {
		identity := common.PeerIdentityType(am.Identity)
		claimedPKIID := am.Membership.PkiId
		err := sa.idMapper.Put(claimedPKIID, identity)
		if err != nil {
			logging.Debug("Falied validating identity of %v reason %+v", am, errors.WithStack(err))
			return false
		}
	} else {
		cert, _ := sa.idMapper.Get(am.Membership.PkiId)
		if cert == nil {
			logging.Debug("Don't have certificate for", am)
			return false
		}
	}

	logging.Debug("Fetched identity of", am.Membership.PkiId, "from identity store")
	return sa.validateAliveMsgSignature(m, am.Membership.PkiId)
}

func (sa *discoverySecurityAdapter) SignMessage(m *protos.RKSyncMessage) *protos.Envelope {
	signer := func(msg []byte) ([]byte, error) {
		return sa.idMapper.Sign(msg)
	}
	if m.IsAliveMsg() && time.Now().Before(sa.includeIdentityPeriod) {
		m.GetAliveMsg().Identity = sa.identity
	}

	signedMsg := &protos.SignedRKSyncMessage{RKSyncMessage: m}
	e, err := signedMsg.Sign(signer)
	if err != nil {
		logging.Warningf("Failed signing message: %+v", errors.WithStack(err))
		return nil
	}

	return e
}

func (sa *discoverySecurityAdapter) validateAliveMsgSignature(m *protos.SignedRKSyncMessage, id common.PKIidType) bool {
	am := m.GetAliveMsg()
	verifier := func(pkiID []byte, signature, message []byte) error {
		return sa.idMapper.Verify(common.PKIidType(pkiID), signature, message)
	}

	err := m.Verify(id, verifier)
	if err != nil {
		logging.Warningf("Failed verifying: %v: %+v", am, errors.WithStack(err))
		return false
	}
	return true
}

// partitionMessages receives a predicate and a slice of rksync messages
// and returns a tuple of two slices: the messages that hold for the predicate
// and the rest
func partitionMessages(pred common.MessageAcceptor, a []*emittedRKSyncMessage) ([]*emittedRKSyncMessage, []*emittedRKSyncMessage) {
	s1 := []*emittedRKSyncMessage{}
	s2 := []*emittedRKSyncMessage{}
	for _, m := range a {
		if pred(m) {
			s1 = append(s1, m)
		} else {
			s2 = append(s2, m)
		}
	}
	return s1, s2
}
