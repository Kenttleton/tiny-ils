package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	networkpb "tiny-ils/gen/networkpb"
	pb "tiny-ils/gen/curiospb"
	"tiny-ils/curios-manager/lcp"
	"tiny-ils/curios-manager/metadata"
	"tiny-ils/curios-manager/store"
	"tiny-ils/shared/models"
)

type CuriosService struct {
	pb.UnimplementedCuriosManagerServer
	curios        *store.CurioStore
	loans         *store.LoanStore
	leases        *store.LeaseStore
	transfers     *store.TransferStore
	lcp           *lcp.Client // nil when LCP is not configured
	networkClient networkpb.NetworkManagerClient // nil when network-manager is not available
	nodeID        string                         // this node's fingerprint
}

func NewCuriosService(c *store.CurioStore, l *store.LoanStore, ls *store.LeaseStore, t *store.TransferStore, lcpClient *lcp.Client, networkClient networkpb.NetworkManagerClient, nodeID string) *CuriosService {
	return &CuriosService{curios: c, loans: l, leases: ls, transfers: t, lcp: lcpClient, networkClient: networkClient, nodeID: nodeID}
}

// ─── Catalog CRUD ────────────────────────────────────────────────────────────

func (s *CuriosService) ListCurios(ctx context.Context, req *pb.ListCuriosRequest) (*pb.CurioList, error) {
	curios, total, err := s.curios.List(ctx, store.ListFilter{
		Query:      req.Query,
		MediaType:  req.MediaType,
		FormatType: req.FormatType,
		Tags:       req.Tags,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list curios: %v", err)
	}
	return &pb.CurioList{Curios: toPBList(curios), Total: total}, nil
}

func (s *CuriosService) GetCurio(ctx context.Context, req *pb.CurioId) (*pb.Curio, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	c, err := s.curios.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "curio not found: %v", err)
	}
	return toPB(c), nil
}

func (s *CuriosService) CreateCurio(ctx context.Context, req *pb.CreateCurioRequest) (*pb.Curio, error) {
	if req.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}
	c := &models.Curio{
		Title:       req.Title,
		Description: req.Description,
		MediaType:   models.MediaType(req.MediaType),
		FormatType:  models.FormatType(req.FormatType),
		Tags:        req.Tags,
		Barcode:     req.Barcode,
		QRCode:      req.QrCode,
	}
	created, err := s.curios.Create(ctx, c)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create curio: %v", err)
	}
	return toPB(created), nil
}

func (s *CuriosService) UpdateCurio(ctx context.Context, req *pb.UpdateCurioRequest) (*pb.Curio, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	existing, err := s.curios.Get(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "curio not found")
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.FormatType != "" {
		existing.FormatType = models.FormatType(req.FormatType)
	}
	if len(req.Tags) > 0 {
		existing.Tags = req.Tags
	}
	if req.Barcode != "" {
		existing.Barcode = req.Barcode
	}
	if req.QrCode != "" {
		existing.QRCode = req.QrCode
	}
	updated, err := s.curios.Update(ctx, existing)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update curio: %v", err)
	}
	return toPB(updated), nil
}

func (s *CuriosService) DeleteCurio(ctx context.Context, req *pb.CurioId) (*pb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	if err := s.curios.Delete(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "delete curio: %v", err)
	}
	return &pb.Empty{}, nil
}

// ─── Metadata enrichment ─────────────────────────────────────────────────────

func (s *CuriosService) EnrichMetadata(ctx context.Context, req *pb.EnrichRequest) (*pb.CurioMetadata, error) {
	if req.MediaType == "" || req.Identifier == "" {
		return nil, status.Errorf(codes.InvalidArgument, "media_type and identifier are required")
	}
	r, err := metadata.Enrich(ctx, models.MediaType(req.MediaType), req.Identifier)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "enrich metadata: %v", err)
	}
	return &pb.CurioMetadata{
		Title:       r.Title,
		Description: r.Description,
		Authors:     r.Authors,
		CoverUrl:    r.CoverURL,
		Tags:        r.Tags,
		Source:      r.Source,
	}, nil
}

// ─── Physical copies and loans ───────────────────────────────────────────────

func (s *CuriosService) ListCopies(ctx context.Context, req *pb.CurioId) (*pb.CopyList, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	copies, err := s.loans.ListCopies(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list copies: %v", err)
	}
	var pbCopies []*pb.PhysicalCopy
	for _, c := range copies {
		pbCopies = append(pbCopies, &pb.PhysicalCopy{
			Id:        c.ID.String(),
			CurioId:   c.CurioID.String(),
			Condition: string(c.Condition),
			Location:  c.Location,
			NodeId:    c.NodeID,
			Status:    string(c.Status),
			CreatedAt: c.CreatedAt.Unix(),
		})
	}
	return &pb.CopyList{Copies: pbCopies}, nil
}

func (s *CuriosService) ListLoans(ctx context.Context, req *pb.ListLoansRequest) (*pb.LoanList, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int(req.Offset)
	var userID *uuid.UUID
	if req.UserId != "" {
		id, err := uuid.Parse(req.UserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
		}
		userID = &id
	}
	loans, total, err := s.loans.ListLoans(ctx, req.ActiveOnly, userID, req.UserNodeId, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list loans: %v", err)
	}
	var pbLoans []*pb.PhysicalLoan
	for _, l := range loans {
		pl := loanToPB(&l.PhysicalLoan)
		pl.CurioId = l.CurioID.String()
		pl.CurioTitle = l.CurioTitle
		pbLoans = append(pbLoans, pl)
	}
	return &pb.LoanList{Loans: pbLoans, Total: int32(total)}, nil
}

func (s *CuriosService) CheckoutCopy(ctx context.Context, req *pb.CheckoutRequest) (*pb.PhysicalLoan, error) {
	copyID, err := uuid.Parse(req.CopyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid copy id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	dueDate := time.Now().Add(14 * 24 * time.Hour) // default 2 weeks
	if req.DueDate > 0 {
		dueDate = time.Unix(req.DueDate, 0)
	}
	loan, err := s.loans.Checkout(ctx, copyID, userID, req.UserNodeId, dueDate, req.UserNodeId)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "checkout: %v", err)
	}
	return loanToPB(loan), nil
}

func (s *CuriosService) ReturnCopy(ctx context.Context, req *pb.ReturnRequest) (*pb.PhysicalLoan, error) {
	copyID, err := uuid.Parse(req.CopyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid copy id")
	}
	loan, err := s.loans.Return(ctx, copyID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "return copy: %v", err)
	}
	return loanToPB(loan), nil
}

func (s *CuriosService) PlaceHold(ctx context.Context, req *pb.HoldRequest) (*pb.Hold, error) {
	curioID, err := uuid.Parse(req.CurioId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	hold, err := s.loans.PlaceHold(ctx, curioID, userID, req.UserNodeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "place hold: %v", err)
	}
	return &pb.Hold{
		Id:       hold.ID.String(),
		CurioId:  hold.CurioID.String(),
		UserId:   hold.UserID.String(),
		PlacedAt: hold.PlacedAt.Unix(),
	}, nil
}

func (s *CuriosService) CancelHold(ctx context.Context, req *pb.HoldId) (*pb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid hold id")
	}
	if err := s.loans.CancelHold(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "cancel hold: %v", err)
	}
	return &pb.Empty{}, nil
}

// ─── Digital assets and leasing ──────────────────────────────────────────────

func (s *CuriosService) GetDigitalAsset(ctx context.Context, req *pb.CurioId) (*pb.DigitalAsset, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	a, err := s.leases.GetAsset(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "digital asset not found: %v", err)
	}
	return assetToPB(a), nil
}

func (s *CuriosService) CreateDigitalAsset(ctx context.Context, req *pb.CreateDigitalAssetRequest) (*pb.DigitalAsset, error) {
	curioID, err := uuid.Parse(req.CurioId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	a := &models.DigitalAsset{
		CurioID:        curioID,
		Format:         req.Format,
		FileRef:        req.FileRef,
		Checksum:       req.Checksum,
		MaxConcurrent:  int(req.MaxConcurrent),
		LCPContentID:   req.LcpContentId,
		StorageBackend: req.StorageBackend,
		Encrypted:      req.Encrypted,
	}
	if a.StorageBackend == "" {
		a.StorageBackend = "local"
	}
	created, err := s.leases.CreateAsset(ctx, a)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create digital asset: %v", err)
	}
	return assetToPB(created), nil
}

func (s *CuriosService) IssueLease(ctx context.Context, req *pb.LeaseRequest) (*pb.DigitalLease, error) {
	curioID, err := uuid.Parse(req.CurioId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid curio id")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id")
	}
	asset, err := s.leases.GetAsset(ctx, curioID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "digital asset not found")
	}

	expiresAt := time.Now().Add(14 * 24 * time.Hour)
	if req.ExpiresAt > 0 {
		expiresAt = time.Unix(req.ExpiresAt, 0)
	}

	// Concurrent seat check
	if asset.MaxConcurrent > 0 {
		active, err := s.leases.CountActiveLeases(ctx, asset.ID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "seat check: %v", err)
		}
		if active >= asset.MaxConcurrent {
			return nil, status.Errorf(codes.ResourceExhausted,
				"no seats available (%d/%d active)", active, asset.MaxConcurrent)
		}
	}

	// Determine access token: use LCP license if content is registered with LCP
	accessToken := ""
	licenseURL := ""

	if s.lcp != nil && asset.LCPContentID != "" {
		providerURL := os.Getenv("LCP_PROVIDER_URL")
		if providerURL == "" {
			providerURL = os.Getenv("BFF_PUBLIC_URL")
		}
		end := expiresAt
		licenseReq := lcp.LicenseRequest{
			Provider:  providerURL,
			User:      lcp.LicenseUser{ID: userID.String()},
			Rights:    lcp.LicenseRights{End: &end},
			ContentID: asset.LCPContentID,
		}
		doc, err := s.lcp.IssueLicense(ctx, licenseReq)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "lcp issue license: %v", err)
		}
		accessToken = doc.ID
		if doc.Rights.End != nil {
			expiresAt = *doc.Rights.End
		}
		lsdPublicURL := os.Getenv("LSD_PUBLIC_URL")
		if lsdPublicURL != "" {
			licenseURL = fmt.Sprintf("%s/licenses/%s/status", lsdPublicURL, doc.ID)
		}
	}

	lease, err := s.leases.IssueLease(ctx, asset.ID, userID, req.UserNodeId, accessToken, expiresAt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "issue lease: %v", err)
	}
	return &pb.DigitalLease{
		Id:          lease.ID.String(),
		AssetId:     lease.AssetID.String(),
		UserId:      lease.UserID.String(),
		UserNodeId:  lease.UserNodeID,
		AccessToken: lease.AccessToken,
		IssuedAt:    lease.IssuedAt.Unix(),
		ExpiresAt:   lease.ExpiresAt.Unix(),
		Revoked:     lease.Revoked,
		LicenseUrl:  licenseURL,
	}, nil
}

func (s *CuriosService) RevokeLease(ctx context.Context, req *pb.LeaseId) (*pb.Empty, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid lease id")
	}
	// If LCP-backed, also call lsdserver to return the license
	if s.lcp != nil {
		lease, err := s.leases.GetLease(ctx, id)
		if err == nil && lease.AccessToken != "" {
			// Best-effort: don't fail revoke if LSD call fails
			_ = s.lcp.ReturnLicense(ctx, lease.AccessToken)
		}
	}
	if err := s.leases.RevokeLease(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "revoke lease: %v", err)
	}
	return &pb.Empty{}, nil
}

func (s *CuriosService) GetLease(ctx context.Context, req *pb.LeaseId) (*pb.DigitalLease, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid lease id")
	}
	lease, err := s.leases.GetLease(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "lease not found: %v", err)
	}
	licenseURL := ""
	if lease.AccessToken != "" {
		if lsdURL := os.Getenv("LSD_PUBLIC_URL"); lsdURL != "" {
			licenseURL = fmt.Sprintf("%s/licenses/%s/status", lsdURL, lease.AccessToken)
		}
	}
	return &pb.DigitalLease{
		Id:          lease.ID.String(),
		AssetId:     lease.AssetID.String(),
		UserId:      lease.UserID.String(),
		UserNodeId:  lease.UserNodeID,
		AccessToken: lease.AccessToken,
		IssuedAt:    lease.IssuedAt.Unix(),
		ExpiresAt:   lease.ExpiresAt.Unix(),
		Revoked:     lease.Revoked,
		LicenseUrl:  licenseURL,
	}, nil
}

func (s *CuriosService) ListLeases(ctx context.Context, req *pb.ListLeasesRequest) (*pb.DigitalLeaseList, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id")
	}
	leases, err := s.leases.ListLeases(ctx, userID, req.UserNodeId, req.ActiveOnly)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list leases: %v", err)
	}
	var pbLeases []*pb.DigitalLease
	for _, l := range leases {
		pbLeases = append(pbLeases, &pb.DigitalLease{
			Id:          l.ID.String(),
			AssetId:     l.AssetID.String(),
			UserId:      l.UserID.String(),
			UserNodeId:  l.UserNodeID,
			AccessToken: l.AccessToken,
			IssuedAt:    l.IssuedAt.Unix(),
			ExpiresAt:   l.ExpiresAt.Unix(),
			Revoked:     l.Revoked,
		})
	}
	return &pb.DigitalLeaseList{Leases: pbLeases}, nil
}

// ─── Conversion helpers ──────────────────────────────────────────────────────

func assetToPB(a *models.DigitalAsset) *pb.DigitalAsset {
	return &pb.DigitalAsset{
		Id:             a.ID.String(),
		CurioId:        a.CurioID.String(),
		Format:         a.Format,
		FileRef:        a.FileRef,
		Checksum:       a.Checksum,
		MaxConcurrent:  int32(a.MaxConcurrent),
		LcpContentId:   a.LCPContentID,
		StorageBackend: a.StorageBackend,
		Encrypted:      a.Encrypted,
	}
}

func toPB(c *models.Curio) *pb.Curio {
	return &pb.Curio{
		Id:          c.ID.String(),
		Title:       c.Title,
		Description: c.Description,
		MediaType:   string(c.MediaType),
		FormatType:  string(c.FormatType),
		Tags:        c.Tags,
		Barcode:     c.Barcode,
		QrCode:      c.QRCode,
		CreatedAt:   c.CreatedAt.Unix(),
		UpdatedAt:   c.UpdatedAt.Unix(),
	}
}

func toPBList(cs []*models.Curio) []*pb.Curio {
	out := make([]*pb.Curio, len(cs))
	for i, c := range cs {
		out[i] = toPB(c)
	}
	return out
}

func loanToPB(l *models.PhysicalLoan) *pb.PhysicalLoan {
	pb := &pb.PhysicalLoan{
		Id:             l.ID.String(),
		CopyId:         l.CopyID.String(),
		UserId:         l.UserID.String(),
		UserNodeId:     l.UserNodeID,
		CheckedOut:     l.CheckedOut.Unix(),
		DueDate:        l.DueDate.Unix(),
		RequestingNode: l.RequestingNode,
	}
	if l.ReturnedAt != nil {
		pb.ReturnedAt = l.ReturnedAt.Unix()
	}
	return pb
}
