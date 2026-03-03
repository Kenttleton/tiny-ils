package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	curiospb "tiny-ils/gen/curiospb"
)

// StartHTTPServer starts the HTTP interop server on addr.
// This server is the external HTTP face of the node, complementing the gRPC listener.
// It handles:
//   - Digital asset passthrough (content delivery / lease redemption)
//   - ISO 18626 ILL stub (future: external library interop)
//   - SRU catalog search stub (future: external catalog search interop)
func StartHTTPServer(addr string, svc *NetworkService) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/digital/lease/", svc.handleDigitalLease)
	mux.HandleFunc("/ill/v1/request", svc.handleILLRequest)
	mux.HandleFunc("/sru", svc.handleSRU)
	return http.ListenAndServe(addr, mux)
}

// handleDigitalLease redeems a digital lease and delivers content access.
//
// GET /digital/lease/{lease_id}
// Authorization: Bearer <user-jwt>
//
// The caller presents the JWT that was issued by their home node.
// The endpoint verifies the JWT, checks lease ownership and validity,
// then redirects to the LCP license URL or returns the access token.
func (s *NetworkService) handleDigitalLease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	leaseID := strings.TrimPrefix(r.URL.Path, "/digital/lease/")
	if leaseID == "" {
		http.Error(w, "lease_id required", http.StatusBadRequest)
		return
	}

	// Parse Bearer token from Authorization header.
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization: Bearer <user-jwt> required", http.StatusUnauthorized)
		return
	}
	jwtStr := strings.TrimPrefix(authHeader, "Bearer ")

	ctx := r.Context()

	// Verify the JWT against the peer's registered public key.
	claims, err := s.verifyForeignJWT(ctx, jwtStr, "")
	if err != nil {
		http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Fetch lease from curios-manager.
	lease, err := s.curiosClient.GetLease(ctx, &curiospb.LeaseId{Id: leaseID})
	if err != nil {
		http.Error(w, "lease not found", http.StatusNotFound)
		return
	}

	// Verify the lease belongs to the authenticated user.
	if lease.UserId != claims.UserID || lease.UserNodeId != claims.Issuer {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Verify the lease is still active.
	if lease.Revoked {
		http.Error(w, "lease has been revoked", http.StatusGone)
		return
	}
	if lease.ExpiresAt > 0 && time.Unix(lease.ExpiresAt, 0).Before(time.Now()) {
		http.Error(w, "lease has expired", http.StatusGone)
		return
	}

	// Deliver content access.
	if lease.LicenseUrl != "" {
		// LCP-backed: redirect to LSD status/license URL.
		http.Redirect(w, r, lease.LicenseUrl, http.StatusFound)
		return
	}

	if lease.AccessToken != "" {
		// Provider-backed: return the access token for the client to use.
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:exhaustruct
			"access_token": lease.AccessToken,
			"expires_at":   lease.ExpiresAt,
		})
		return
	}

	// TODO: non-LCP/non-provider asset delivery — communities configure their own
	// file server or DRM system. The lease record holds access_token as the hook point.
	http.Error(w, "content delivery not configured on this node", http.StatusNotImplemented)
}

// handleILLRequest is a stub for the ISO 18626 International ILL Protocol.
//
// POST /ill/v1/request
//
// TODO: ISO 18626 ILL Protocol (ISO 18626:2014/Amd.1:2019)
// When implemented:
//   - Validate auth (mTLS cert match or shared secret — design TBD)
//   - Decode ISO 18626 RequestMessage (JSON)
//   - Delegate to RequestBorrow gRPC handler using requesting agency ID + bibliographic info
//   - Return ISO 18626 Answer message
//
// Ref: https://www.niso.org/standards-committees/iso-ill
func (s *NetworkService) handleILLRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]any{ //nolint:exhaustruct
		"confirmationHeader": map[string]any{
			"supplyingAgencyId": map[string]string{
				"agencyIdType":  "ISIL",
				"agencyIdValue": s.nodeID,
			},
			"messageStatus": "ERROR",
			"errorData": []map[string]string{
				{
					"errorType":  "NotImplemented",
					"errorValue": "ISO 18626 HTTP endpoint not yet active; use gRPC RequestBorrow",
				},
			},
		},
	})
}

// handleSRU is a stub for the SRU 2.0 (Search/Retrieve via URL) catalog search endpoint.
//
// GET /sru?operation=searchRetrieve&query=<CQL>&maximumRecords=<n>
//
// TODO: SRU 2.0 (Search/Retrieve via URL — https://www.loc.gov/standards/sru/)
// When implemented:
//   - Parse CQL query from ?query= parameter
//   - Delegate to SearchCatalog on this node (local catalog only; no fan-out)
//   - Return SRU XML response with Dublin Core or MARC records
//   - Auth: public read or IP-restricted — design TBD
func (s *NetworkService) handleSRU(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<searchRetrieveResponse xmlns="http://docs.oasis-open.org/ns/search-ws/sruResponse">
  <version>2.0</version>
  <numberOfRecords>0</numberOfRecords>
  <!-- TODO: SRU endpoint not yet implemented; CQL query ignored.
       Will delegate to SearchCatalog (local catalog, no fan-out).
       See https://www.loc.gov/standards/sru/ -->
</searchRetrieveResponse>
`)
}
