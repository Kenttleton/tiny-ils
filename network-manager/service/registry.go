package service

import (
	"context"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	curiospb "tiny-ils/gen/curiospb"
	internalpb "tiny-ils/gen/internalpb"
	userspb "tiny-ils/gen/userspb"
)

// LocalDirectoryService implements internalpb.LocalDirectoryServer.
// It is the dynamic capability registry — services announce themselves here
// on startup, and callers use Lookup to discover service addresses.
type LocalDirectoryService struct {
	internalpb.UnimplementedLocalDirectoryServer
	mu         sync.RWMutex
	caps       map[string]bool       // set of announced capability names
	allSvcs    map[string][]string   // name → []address (for generic Lookup)
	curiosSvcs []*registeredCurios   // 0..N curios instances (typed, for network.go)
	usersSvcs  []*registeredUsers    // 0..N users instances (typed, for network.go)
}

type registeredCurios struct {
	address string
	conn    *grpc.ClientConn
	client  curiospb.CuriosManagerClient
}

type registeredUsers struct {
	address string
	conn    *grpc.ClientConn
	client  userspb.UsersManagerClient
}

func NewLocalDirectoryService() *LocalDirectoryService {
	return &LocalDirectoryService{
		caps:    make(map[string]bool),
		allSvcs: make(map[string][]string),
	}
}

// Announce registers a local service. Idempotent: duplicate addresses are ignored.
func (r *LocalDirectoryService) Announce(_ context.Context, req *internalpb.LocalServiceInfo) (*internalpb.Empty, error) {
	if req.Address == "" {
		// UI or address-less service — record capability name only.
		r.mu.Lock()
		r.caps[req.Name] = true
		r.mu.Unlock()
		log.Printf("directory: announced capability %q (no address)", req.Name)
		return &internalpb.Empty{}, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Deduplicate by address.
	for _, existing := range r.allSvcs[req.Name] {
		if existing == req.Address {
			return &internalpb.Empty{}, nil
		}
	}

	conn, err := grpc.NewClient(req.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("directory: announce %q@%s: dial error: %v", req.Name, req.Address, err)
		return &internalpb.Empty{}, nil
	}

	r.allSvcs[req.Name] = append(r.allSvcs[req.Name], req.Address)
	r.caps[req.Name] = true

	switch req.Name {
	case "curios":
		r.curiosSvcs = append(r.curiosSvcs, &registeredCurios{
			address: req.Address,
			conn:    conn,
			client:  curiospb.NewCuriosManagerClient(conn),
		})
	case "users":
		r.usersSvcs = append(r.usersSvcs, &registeredUsers{
			address: req.Address,
			conn:    conn,
			client:  userspb.NewUsersManagerClient(conn),
		})
	default:
		// Unknown capability — address is recorded but no typed client created.
		log.Printf("directory: announced unknown capability %q@%s", req.Name, req.Address)
	}

	log.Printf("directory: announced %q @ %s", req.Name, req.Address)
	return &internalpb.Empty{}, nil
}

// Lookup returns all registered services matching the given capability name.
func (r *LocalDirectoryService) Lookup(_ context.Context, req *internalpb.LookupRequest) (*internalpb.LookupResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*internalpb.LocalServiceInfo
	for _, addr := range r.allSvcs[req.Name] {
		result = append(result, &internalpb.LocalServiceInfo{Name: req.Name, Address: addr})
	}
	return &internalpb.LookupResponse{Services: result}, nil
}

// Capabilities returns a snapshot of all registered capability names.
// Called by NetworkService to populate GetNodeInfo / PeerAck responses.
func (r *LocalDirectoryService) Capabilities() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	caps := make([]string, 0, len(r.caps))
	for name := range r.caps {
		caps = append(caps, name)
	}
	return caps
}

// CuriosSvcs returns a snapshot of all registered curios-manager clients.
// The caller may fan out to all of them for search/list operations.
func (r *LocalDirectoryService) CuriosSvcs() []*registeredCurios {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := make([]*registeredCurios, len(r.curiosSvcs))
	copy(snapshot, r.curiosSvcs)
	return snapshot
}

// GetAnyUsersSvc returns the first registered users-manager client.
// Returns an error with codes.Unavailable when none are registered.
func (r *LocalDirectoryService) GetAnyUsersSvc() (*registeredUsers, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.usersSvcs) == 0 {
		return nil, nil
	}
	return r.usersSvcs[0], nil
}
