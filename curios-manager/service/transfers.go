package service

import (
	"context"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	networkpb "tiny-ils/gen/networkpb"
	pb "tiny-ils/gen/curiospb"
	"tiny-ils/shared/models"
)

// ─── Transfer RPCs ────────────────────────────────────────────────────────────

// RequestTransfer creates a local ledger entry and, for cross-node transfers,
// forwards the request to the source node via the network-manager.
func (s *CuriosService) RequestTransfer(ctx context.Context, req *pb.TransferRequest) (*pb.CopyTransfer, error) {
	initiatedBy, err := uuid.Parse(req.InitiatedBy)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiated_by")
	}
	if req.SourceNode == "" || req.DestNode == "" {
		return nil, status.Errorf(codes.InvalidArgument, "source_node and dest_node are required")
	}

	crossNode := req.SourceNode != s.nodeID

	// Derive global_copy_id if not provided.
	globalCopyID := req.GlobalCopyId
	if globalCopyID == "" {
		if req.CopyId == "" {
			return nil, status.Errorf(codes.InvalidArgument, "copy_id or global_copy_id is required")
		}
		globalCopyID = req.SourceNode + "/" + req.CopyId
	}

	// Generate transfer ID: "{source_node}/{dest_node}/{uuid_v7}".
	transferID := req.Id
	if transferID == "" {
		v7, err := uuid.NewV7()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "generate transfer id: %v", err)
		}
		transferID = req.SourceNode + "/" + req.DestNode + "/" + v7.String()
	}

	// Create the local ledger entry.
	// For cross-node (dest initiating): no copy FK step (copy lives on source node).
	// For local: copy FK step marks the copy REQUESTED.
	t, err := s.transfers.Create(ctx, transferID, globalCopyID, initiatedBy,
		models.TransferType(req.TransferType), req.SourceNode, req.DestNode, req.Notes)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "request transfer: %v", err)
	}

	// For cross-node transfers, forward to source node so it creates its ledger entry.
	if crossNode && s.networkClient != nil {
		if _, err := s.networkClient.ForwardTransfer(ctx, &networkpb.ForwardTransferRequest{
			TargetNodeId: req.SourceNode,
			Request: &networkpb.RemoteTransferRequest{
				TransferId:   transferID,
				CopyId:       req.CopyId,
				TransferType: req.TransferType,
				SourceNode:   req.SourceNode,
				DestNode:     req.DestNode,
				InitiatedBy:  req.InitiatedBy,
				Notes:        req.Notes,
			},
		}); err != nil {
			// Roll back local entry if the source node rejects.
			return nil, status.Errorf(codes.Internal, "forward transfer to source node: %v", err)
		}
	}

	return transferToPB(t), nil
}

func (s *CuriosService) ApproveTransfer(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.Approve(ctx, req.TransferId, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "approve transfer: %v", err)
	}
	s.relayTransferUpdate(req.TransferId, "APPROVED", req.ActorId, t)
	return transferToPB(t), nil
}

func (s *CuriosService) RejectTransfer(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.Reject(ctx, req.TransferId, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "reject transfer: %v", err)
	}
	s.relayTransferUpdate(req.TransferId, "REJECTED", req.ActorId, t)
	return transferToPB(t), nil
}

func (s *CuriosService) MarkShipped(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.MarkShipped(ctx, req.TransferId, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "mark shipped: %v", err)
	}
	s.relayTransferUpdate(req.TransferId, "IN_TRANSIT", req.ActorId, t)
	return transferToPB(t), nil
}

func (s *CuriosService) ConfirmReceived(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.ConfirmReceived(ctx, req.TransferId, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "confirm received: %v", err)
	}
	s.relayTransferUpdate(req.TransferId, "RECEIVED", req.ActorId, t)
	return transferToPB(t), nil
}

func (s *CuriosService) CancelTransfer(ctx context.Context, req *pb.TransferAction) (*pb.CopyTransfer, error) {
	actorID, err := uuid.Parse(req.ActorId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid actor_id")
	}
	t, err := s.transfers.Cancel(ctx, req.TransferId, actorID)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "cancel transfer: %v", err)
	}
	s.relayTransferUpdate(req.TransferId, "CANCELLED", req.ActorId, t)
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
	t, err := s.transfers.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "transfer not found: %v", err)
	}
	return transferToPB(t), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// relayTransferUpdate fires off a best-effort status notification to the peer node.
// It determines the peer from the transfer record's source_node/dest_node fields.
func (s *CuriosService) relayTransferUpdate(transferID, newStatus, actorID string, t *models.CopyTransfer) {
	if s.networkClient == nil {
		return
	}
	peerNodeID := peerNode(t.SourceNode, t.DestNode, s.nodeID)
	if peerNodeID == "" {
		return // local transfer; no relay needed
	}
	go func() {
		ctx := context.Background()
		if _, err := s.networkClient.RelayTransferUpdate(ctx, &networkpb.RelayTransferUpdateRequest{
			TargetNodeId: peerNodeID,
			Update: &networkpb.TransferUpdate{
				TransferId: transferID,
				NewStatus:  newStatus,
				ActorId:    actorID,
			},
		}); err != nil {
			log.Printf("relay transfer update %s → %s: %v", transferID, peerNodeID, err)
		}
	}()
}

// peerNode returns the node that is NOT thisNodeID among source and dest.
// Returns "" if both are the same node (local transfer).
func peerNode(sourceNode, destNode, thisNodeID string) string {
	if sourceNode != thisNodeID {
		return sourceNode
	}
	if destNode != thisNodeID {
		return destNode
	}
	return ""
}

// transferToPB converts a models.CopyTransfer to the protobuf representation.
func transferToPB(t *models.CopyTransfer) *pb.CopyTransfer {
	p := &pb.CopyTransfer{
		Id:           t.ID,
		GlobalCopyId: t.GlobalCopyID,
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

