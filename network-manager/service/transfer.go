package service

import (
	"context"

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

	_, err := s.curiosClient.RequestTransfer(ctx, &curiospb.TransferRequest{
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

	return &pb.RemoteTransferAck{TransferId: req.TransferId, Accepted: true}, nil
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

	var err error
	switch req.NewStatus {
	case "APPROVED":
		_, err = s.curiosClient.ApproveTransfer(ctx, action)
	case "IN_TRANSIT":
		_, err = s.curiosClient.MarkShipped(ctx, action)
	case "RECEIVED":
		_, err = s.curiosClient.ConfirmReceived(ctx, action)
	case "REJECTED":
		_, err = s.curiosClient.RejectTransfer(ctx, action)
	case "CANCELLED":
		_, err = s.curiosClient.CancelTransfer(ctx, action)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown transfer status: %s", req.NewStatus)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update transfer: %v", err)
	}

	return &pb.Empty{}, nil
}
