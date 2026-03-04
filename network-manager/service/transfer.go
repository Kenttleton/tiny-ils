package service

import (
	"context"
	"slices"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	curiospb "tiny-ils/gen/curiospb"
	pb "tiny-ils/gen/networkpb"
)

// InitiateRemoteTransfer is called by a remote node that wants to transfer a
// copy owned by this node. It verifies the requesting node's JWT and delegates
// to the local curios-manager to create the transfer record.
func (s *NetworkService) InitiateRemoteTransfer(ctx context.Context, req *pb.RemoteTransferRequest) (*pb.RemoteTransferAck, error) {
	if req.UserJwt != "" {
		if _, err := s.verifyForeignJWT(ctx, req.UserJwt, req.SourceNode); err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}
	}

	curios := s.firstCuriosSvc()
	if curios == nil {
		return nil, status.Errorf(codes.Unavailable, "no curios service registered")
	}
	xfer, err := curios.RequestTransfer(ctx, &curiospb.TransferRequest{
		Id:           req.TransferId,
		CopyId:       req.CopyId,
		TransferType: req.TransferType,
		SourceNode:   req.SourceNode,
		DestNode:     req.DestNode,
		InitiatedBy:  req.InitiatedBy,
		Notes:        req.Notes,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create remote transfer: %v", err)
	}

	return &pb.RemoteTransferAck{TransferId: xfer.Id, Accepted: true}, nil
}

// NotifyTransferUpdate is called by a remote node when a transfer's status
// changes. It delegates the appropriate action to the local curios-manager.
func (s *NetworkService) NotifyTransferUpdate(ctx context.Context, req *pb.TransferUpdate) (*pb.Empty, error) {
	if req.UserJwt != "" {
		if _, err := s.verifyForeignJWT(ctx, req.UserJwt, ""); err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}
	}

	action := &curiospb.TransferAction{
		TransferId: req.TransferId,
		ActorId:    req.ActorId,
	}

	curios := s.firstCuriosSvc()
	if curios == nil {
		return nil, status.Errorf(codes.Unavailable, "no curios service registered")
	}
	var err error
	switch req.NewStatus {
	case "APPROVED":
		_, err = curios.ApproveTransfer(ctx, action)
	case "IN_TRANSIT":
		_, err = curios.MarkShipped(ctx, action)
	case "RECEIVED":
		_, err = curios.ConfirmReceived(ctx, action)
	case "REJECTED":
		_, err = curios.RejectTransfer(ctx, action)
	case "CANCELLED":
		_, err = curios.CancelTransfer(ctx, action)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown transfer status: %s", req.NewStatus)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update transfer: %v", err)
	}

	return &pb.Empty{}, nil
}

// ForwardTransfer is called by the LOCAL curios-manager to ask this network-manager
// to call InitiateRemoteTransfer on the specified peer node.
func (s *NetworkService) ForwardTransfer(ctx context.Context, req *pb.ForwardTransferRequest) (*pb.ForwardTransferAck, error) {
	peer, err := s.peers.Get(ctx, req.TargetNodeId)
	if err != nil || peer == nil {
		return nil, status.Errorf(codes.NotFound, "peer %q not found", req.TargetNodeId)
	}
	if len(peer.Capabilities) > 0 && !slices.Contains(peer.Capabilities, "curios") {
		return nil, status.Errorf(codes.FailedPrecondition, "peer %q lacks curios capability", req.TargetNodeId)
	}

	conn, err := grpc.NewClient(peer.Address, PeerDialOptions(s.nodeCert)...)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "dial peer %q: %v", req.TargetNodeId, err)
	}
	defer conn.Close()

	ack, err := pb.NewNetworkManagerClient(conn).InitiateRemoteTransfer(ctx, req.Request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "InitiateRemoteTransfer on %q: %v", req.TargetNodeId, err)
	}
	return &pb.ForwardTransferAck{TransferId: ack.TransferId, Accepted: ack.Accepted}, nil
}

// RelayTransferUpdate is called by the LOCAL curios-manager to ask this network-manager
// to call NotifyTransferUpdate on the specified peer node.
func (s *NetworkService) RelayTransferUpdate(ctx context.Context, req *pb.RelayTransferUpdateRequest) (*pb.Empty, error) {
	peer, err := s.peers.Get(ctx, req.TargetNodeId)
	if err != nil || peer == nil {
		return nil, status.Errorf(codes.NotFound, "peer %q not found", req.TargetNodeId)
	}
	if len(peer.Capabilities) > 0 && !slices.Contains(peer.Capabilities, "curios") {
		return nil, status.Errorf(codes.FailedPrecondition, "peer %q lacks curios capability", req.TargetNodeId)
	}

	conn, err := grpc.NewClient(peer.Address, PeerDialOptions(s.nodeCert)...)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "dial peer %q: %v", req.TargetNodeId, err)
	}
	defer conn.Close()

	if _, err := pb.NewNetworkManagerClient(conn).NotifyTransferUpdate(ctx, req.Update); err != nil {
		return nil, status.Errorf(codes.Internal, "NotifyTransferUpdate on %q: %v", req.TargetNodeId, err)
	}
	return &pb.Empty{}, nil
}
