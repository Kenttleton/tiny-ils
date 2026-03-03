package models

import (
	"time"

	"github.com/google/uuid"
)

type MediaType string

const (
	MediaTypeThing MediaType = "THING"
	MediaTypeBook  MediaType = "BOOK"
	MediaTypeVideo MediaType = "VIDEO"
	MediaTypeAudio MediaType = "AUDIO"
	MediaTypeGame  MediaType = "GAME"
)

type FormatType string

const (
	FormatTypeDigital  FormatType = "DIGITAL"
	FormatTypePhysical FormatType = "PHYSICAL"
	FormatTypeBoth     FormatType = "BOTH"
)

type CopyCondition string

const (
	ConditionNew  CopyCondition = "NEW"
	ConditionGood CopyCondition = "GOOD"
	ConditionFair CopyCondition = "FAIR"
	ConditionPoor CopyCondition = "POOR"
	ConditionLost CopyCondition = "LOST"
)

type Curio struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	MediaType   MediaType  `json:"mediaType"`
	FormatType  FormatType `json:"formatType"`
	Tags        []string   `json:"tags"`
	Barcode     string     `json:"barcode,omitempty"`
	QRCode      string     `json:"qrCode,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type CopyStatus string

const (
	CopyStatusAvailable CopyStatus = "AVAILABLE"
	CopyStatusOnLoan    CopyStatus = "ON_LOAN"
	CopyStatusRequested CopyStatus = "REQUESTED"
	CopyStatusInTransit CopyStatus = "IN_TRANSIT"
)

type TransferType string

const (
	TransferTypeILL       TransferType = "ILL"
	TransferTypeReturn    TransferType = "RETURN"
	TransferTypePermanent TransferType = "PERMANENT"
)

type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "PENDING"
	TransferStatusApproved  TransferStatus = "APPROVED"
	TransferStatusInTransit TransferStatus = "IN_TRANSIT"
	TransferStatusReceived  TransferStatus = "RECEIVED"
	TransferStatusRejected  TransferStatus = "REJECTED"
	TransferStatusCancelled TransferStatus = "CANCELLED"
)

type PhysicalCopy struct {
	ID         uuid.UUID     `json:"id"`
	CurioID    uuid.UUID     `json:"curioId"`
	Condition  CopyCondition `json:"condition"`
	Location   string        `json:"location"`
	NodeID     string        `json:"nodeId"`
	HomeNodeID string        `json:"homeNodeId"`
	Status     CopyStatus    `json:"status"`
	CreatedAt  time.Time     `json:"createdAt"`
}

type CopyTransfer struct {
	ID           string         `json:"id"`           // "{source_node}/{dest_node}/{uuid_v7}"
	GlobalCopyID string         `json:"globalCopyId"` // "{home_node}/{copy_uuid}"
	TransferType TransferType   `json:"transferType"`
	SourceNode   string         `json:"sourceNode"`
	DestNode     string         `json:"destNode"`
	InitiatedBy  uuid.UUID      `json:"initiatedBy"`
	ApprovedBy   *uuid.UUID     `json:"approvedBy,omitempty"`
	Status       TransferStatus `json:"status"`
	Notes        string         `json:"notes,omitempty"`
	RequestedAt  time.Time      `json:"requestedAt"`
	ApprovedAt   *time.Time     `json:"approvedAt,omitempty"`
	ShippedAt    *time.Time     `json:"shippedAt,omitempty"`
	ReceivedAt   *time.Time     `json:"receivedAt,omitempty"`
}

type PhysicalLoan struct {
	ID             uuid.UUID  `json:"id"`
	CopyID         uuid.UUID  `json:"copyId"`
	UserID         uuid.UUID  `json:"userId"`
	UserNodeID     string     `json:"userNodeId"`
	CheckedOut     time.Time  `json:"checkedOut"`
	DueDate        time.Time  `json:"dueDate"`
	ReturnedAt     *time.Time `json:"returnedAt,omitempty"`
	RequestingNode string     `json:"requestingNode,omitempty"`
}

type Hold struct {
	ID        uuid.UUID  `json:"id"`
	CurioID   uuid.UUID  `json:"curioId"`
	UserID    uuid.UUID  `json:"userId"`
	PlacedAt  time.Time  `json:"placedAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	Fulfilled bool       `json:"fulfilled"`
}

type DigitalAsset struct {
	ID             uuid.UUID `json:"id"`
	CurioID        uuid.UUID `json:"curioId"`
	Format         string    `json:"format"`
	FileRef        string    `json:"fileRef"`
	Checksum       string    `json:"checksum"`
	MaxConcurrent  int       `json:"maxConcurrent"`
	LCPContentID   string    `json:"lcpContentId,omitempty"`
	StorageBackend string    `json:"storageBackend"` // "local" | "provider"
	Encrypted      bool      `json:"encrypted"`
}

// DigitalLease is stubbed — access token delivery mechanism is pluggable (TODO).
type DigitalLease struct {
	ID          uuid.UUID  `json:"id"`
	AssetID     uuid.UUID  `json:"assetId"`
	UserID      uuid.UUID  `json:"userId"`
	UserNodeID  string     `json:"userNodeId"`
	AccessToken string     `json:"accessToken"` // TODO: pluggable delivery
	IssuedAt    time.Time  `json:"issuedAt"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	Revoked     bool       `json:"revoked"`
}
