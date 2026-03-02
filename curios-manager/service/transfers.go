package service

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "tiny-ils/gen/curiospb"
	"tiny-ils/shared/models"
)

// ─── Transfer RPCs ────────────────────────────────────────────────────────────

func (s *CuriosService) RequestTransfer(ctx context.Context, req *pb.TransferRequest) (*pb.CopyTransfer, error) {
	copyID, err := uuid.Parse(req.CopyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid copy_id")
	}
	initiatedBy, err := uuid.Parse(req.InitiatedBy)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiated_by")
	}
	if req.SourceNode == "" || req.DestNode == "" {
		return nil, status.Errorf(codes.InvalidArgument, "source_node and dest_node are required")
	}
	t, err := s.transfers.Create(ctx, copyID, initiatedBy, models.TransferType(req.TransferType), req.SourceNode, req.DestNode, req.Notes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "request transfer: %v", err)
	}
	return transferToPB(t), nil
}

func (s *CuriosService) ApproveTransfer(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	id, err := uuid.Parse(req.TransferId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transfer_id")
	}
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.Approve(ctx, id, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "approve transfer: %v", err)
	}
	return transferToPB(t), nil
}

func (s *CuriosService) RejectTransfer(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	id, err := uuid.Parse(req.TransferId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transfer_id")
	}
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.Reject(ctx, id, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "reject transfer: %v", err)
	}
	return transferToPB(t), nil
}

func (s *CuriosService) MarkShipped(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	id, err := uuid.Parse(req.TransferId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transfer_id")
	}
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.MarkShipped(ctx, id, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "mark shipped: %v", err)
	}
	return transferToPB(t), nil
}

func (s *CuriosService) ConfirmReceived(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	id, err := uuid.Parse(req.TransferId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transfer_id")
	}
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.ConfirmReceived(ctx, id, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "confirm received: %v", err)
	}
	return transferToPB(t), nil
}

func (s *CuriosService) CancelTransfer(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	id, err := uuid.Parse(req.TransferId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transfer_id")
	}
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.Cancel(ctx, id, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "cancel transfer: %v", err)
	}
	return transferToPB(t), nil
}

func (s *CuriosService) ListTransfers(ctx context.Context, req *pb.ListTransfersRequest) (*pb.TransferList, error) {
	transfers, err := s.transfers.List(ctx, req.Status, req.NodeId, req.TransferType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list transfers: %v", err)
	}
	out := make([]*pb.CopyTransfer, len(transfers))
	for i, t := range transfers {
		out[i] = transferToPB(t)
	}
	return &pb.TransferList{Transfers: out}, nil
}

func (s *CuriosService) GetTransfer(ctx context.Context, req *pb.TransferId) (*pb.CopyTransfer, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transfer_id")
	}
	t, err := s.transfers.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "transfer not found: %v", err)
	}
	return transferToPB(t), nil
}

// ─── Conversion helper ────────────────────────────────────────────────────────

func transferToPB(t *models.CopyTransfer) *pb.CopyTransfer {
	p := &pb.CopyTransfer{
		Id:           t.ID.String(),
		CopyId:       t.CopyID.String(),
		TransferType: string(t.TransferType),
		SourceNode:   t.SourceNode,
		DestNode:     t.DestNode,
		InitiatedBy:  t.InitiatedBy.String(),
		Status:       string(t.Status),
		Notes:        t.Notes,
		RequestedAt:  t.RequestedAt.Unix(),
	}
	if t.ApprovedBy != nil {
		p.ApprovedBy = t.ApprovedBy.String()
	}
	if t.ApprovedAt != nil {
		p.ApprovedAt = t.ApprovedAt.Unix()
	}
	if t.ShippedAt != nil {
		p.ShippedAt = t.ShippedAt.Unix()
	}
	if t.ReceivedAt != nil {
		p.ReceivedAt = t.ReceivedAt.Unix()
	}
	return p
}
