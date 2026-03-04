package service

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"tiny-ils/network-manager/store"
)

// TrustLevel represents the degree of trust granted to a caller.
type TrustLevel int

const (
	TrustNone      TrustLevel = iota // no mTLS cert presented
	TrustCert                        // valid cert, peer not yet CONNECTED
	TrustConnected                   // cert + CONNECTED in peers table (or own-node cert)
)

// rpcTrust maps RPC method names to the minimum TrustLevel required.
var rpcTrust = map[string]TrustLevel{
	// Fully public — no cert required.
	"GetNodeInfo":   TrustNone,
	"GetNodeConfig": TrustNone, // frontend reads before setup/auth; grpc_address is not sensitive

	// Any peer with a valid cert may register or search (PENDING is sufficient).
	"RegisterPeer":  TrustCert,
	"SearchNetwork": TrustCert,
	"ShareCatalog":  TrustCert,

	// CONNECTED peers only — all borrow, transfer, auth, and admin operations.
	"ListPeers":              TrustConnected,
	"ConnectPeer":            TrustConnected,
	"ApprovePeer":            TrustConnected,
	"VerifyUser":             TrustConnected,
	"IssueGuestToken":        TrustConnected,
	"AuthenticateGuest":      TrustConnected,
	"GetUserLoans":           TrustConnected,
	"RequestBorrow":          TrustConnected,
	"ReturnCurio":            TrustConnected,
	"RequestDigitalLease":    TrustConnected,
	"RevokeDigitalLease":     TrustConnected,
	"InitiateRemoteTransfer": TrustConnected,
	"NotifyTransferUpdate":   TrustConnected,
	"SetNodeAddress":         TrustConnected, // internal-only; external port blocks external peers
	// ForwardTransfer and RelayTransferUpdate are internal-only (called by local
	// curios-manager on the internal port which has no interceptor). They are NOT
	// listed here so they default to TrustConnected on the external server —
	// remote peers calling them would be rejected unless CONNECTED, and they
	// would receive a routing error since these RPCs have no cross-node purpose.
}

// ServerTLSConfig returns a tls.Config for the network-manager gRPC server.
// Client certificates are requested but not required for backward compatibility
// with callers that have not yet loaded their key pair.
func ServerTLSConfig(cert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequestClientCert,
		// Skip normal chain verification — trust is established via public-key
		// pinning (own key = TrustConnected, peers table = CONNECTED peers).
		VerifyPeerCertificate: func([][]byte, [][]*x509.Certificate) error { return nil },
	}
}

// PeerDialOptions returns gRPC dial options for outgoing peer connections using mTLS.
func PeerDialOptions(cert tls.Certificate) []grpc.DialOption {
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// Skip chain verification — server identity is pinned via peers table.
		InsecureSkipVerify: true, //nolint:gosec // intentional: public-key pinned
	}
	return []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
	}
}

// peerLookup is the interface the trust interceptor needs from the peer store.
// *store.PeerStore satisfies this interface via duck typing.
type peerLookup interface {
	GetByPublicKey(ctx context.Context, pubKeyB64 string) (*store.Peer, error)
}

// TrustInterceptor enforces per-RPC trust based on the caller's mTLS cert.
// selfPubKeyB64 is this node's own base64-encoded public key; a caller
// presenting the own-node cert (the local BFF) is granted TrustConnected.
func TrustInterceptor(peers peerLookup, selfPubKeyB64 string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		name := rpcName(info.FullMethod)
		required, ok := rpcTrust[name]
		if !ok {
			required = TrustConnected
		}
		if required == TrustNone {
			return handler(ctx, req)
		}
		cert := clientCert(ctx)
		level := resolveTrust(ctx, peers, cert, selfPubKeyB64)
		if level < required {
			return nil, status.Errorf(codes.PermissionDenied, "insufficient trust for %s", name)
		}
		return handler(ctx, req)
	}
}

// TrustStreamInterceptor is the streaming equivalent of TrustInterceptor.
func TrustStreamInterceptor(peers peerLookup, selfPubKeyB64 string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		name := rpcName(info.FullMethod)
		required, ok := rpcTrust[name]
		if !ok {
			required = TrustConnected
		}
		if required == TrustNone {
			return handler(srv, ss)
		}
		cert := clientCert(ss.Context())
		level := resolveTrust(ss.Context(), peers, cert, selfPubKeyB64)
		if level < required {
			return status.Errorf(codes.PermissionDenied, "insufficient trust for %s", name)
		}
		return handler(srv, ss)
	}
}

// clientCert extracts the mTLS client certificate from an incoming gRPC context.
func clientCert(ctx context.Context) *x509.Certificate {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil
	}
	if len(tlsInfo.State.PeerCertificates) == 0 {
		return nil
	}
	return tlsInfo.State.PeerCertificates[0]
}

// resolveTrust determines the trust level of the caller based on their cert.
// A caller presenting this node's own public key (i.e. the local BFF) is
// granted TrustConnected without a peers table lookup.
func resolveTrust(ctx context.Context, peers peerLookup, cert *x509.Certificate, selfPubKeyB64 string) TrustLevel {
	if cert == nil {
		return TrustNone
	}
	pub, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return TrustNone
	}
	pubKeyB64 := base64.StdEncoding.EncodeToString(pub)

	// Own-node cert (the local BFF authenticating as this node).
	if pubKeyB64 == selfPubKeyB64 {
		return TrustConnected
	}

	p, err := peers.GetByPublicKey(ctx, pubKeyB64)
	if err != nil || p == nil {
		return TrustCert // cert is valid but not a registered peer
	}
	if p.Status == "CONNECTED" {
		return TrustConnected
	}
	return TrustCert // PENDING
}

// rpcName extracts the method name from a full gRPC method path.
// e.g. "/network.NetworkManager/RegisterPeer" → "RegisterPeer"
func rpcName(fullMethod string) string {
	if i := strings.LastIndex(fullMethod, "/"); i >= 0 {
		return fullMethod[i+1:]
	}
	return fullMethod
}
