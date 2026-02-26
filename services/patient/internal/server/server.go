package server

import (
	"encoding/json"
	"time"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/config"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/pipeline"
	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the PatientService gRPC server.
type Server struct {
	patientv1.UnimplementedPatientServiceServer
	pipeline *pipeline.Writer
	idx      sqliteindex.Index
	git      gitstore.Store
	cfg      *config.Config
}

// NewServer creates a new gRPC server for the patient service.
func NewServer(cfg *config.Config, pw *pipeline.Writer, idx sqliteindex.Index, git gitstore.Store) *Server {
	return &Server{
		pipeline: pw,
		idx:      idx,
		git:      git,
		cfg:      cfg,
	}
}

// Helper: convert MutationContext from proto to pipeline type
func mutCtxFromProto(ctx *patientv1.MutationContext) pipeline.MutationContext {
	mc := pipeline.MutationContext{}
	if ctx != nil {
		mc.PractitionerID = ctx.PractitionerId
		mc.NodeID = ctx.NodeId
		mc.SiteID = ctx.SiteId
		if ctx.Timestamp != nil {
			mc.Timestamp = ctx.Timestamp.AsTime()
		}
	}
	if mc.Timestamp.IsZero() {
		mc.Timestamp = time.Now().UTC()
	}
	return mc
}

// Helper: map pipeline errors to gRPC status codes per spec §11.1
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if ve, ok := err.(*pipeline.ValidationError); ok {
		return status.Errorf(codes.InvalidArgument, "VALIDATION_ERROR: %s", ve.Message)
	}
	msg := err.Error()
	if contains(msg, "not found") {
		return status.Errorf(codes.NotFound, "%s", msg)
	}
	if contains(msg, "lock timeout") {
		return status.Errorf(codes.Aborted, "write lock timeout")
	}
	return status.Errorf(codes.Internal, "%s", msg)
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsImpl(s, sub))
}

func containsImpl(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// Helper: convert WriteResult to GitCommitInfo proto
func toGitCommitInfo(result *pipeline.WriteResult) *patientv1.GitCommitInfo {
	return &patientv1.GitCommitInfo{
		CommitHash: result.CommitHash,
		Message:    result.CommitMsg,
		Timestamp:  timestamppb.New(result.Timestamp),
	}
}

// Helper: convert FHIR JSON to FHIRResource proto
func toFHIRResource(resourceType, id string, fhirJSON []byte) *commonv1.FHIRResource {
	return &commonv1.FHIRResource{
		ResourceType: resourceType,
		Id:           id,
		JsonPayload:  fhirJSON,
	}
}

// Helper: convert PatientRow to FHIRResource proto
func patientRowToProto(row *fhir.PatientRow) *commonv1.FHIRResource {
	return &commonv1.FHIRResource{
		ResourceType: fhir.ResourcePatient,
		Id:           row.ID,
		JsonPayload:  []byte(row.FHIRJson),
	}
}

// Helper: convert PaginationOpts from proto
func paginationFromProto(pg *commonv1.PaginationRequest) fhir.PaginationOpts {
	opts := fhir.PaginationOpts{Page: 1, PerPage: 25}
	if pg != nil {
		if pg.Page > 0 {
			opts.Page = int(pg.Page)
		}
		if pg.PerPage > 0 {
			opts.PerPage = int(pg.PerPage)
		}
		opts.Sort = pg.Sort
	}
	return opts
}

// Helper: convert Pagination to proto
func paginationToProto(pg *fhir.Pagination) *commonv1.PaginationResponse {
	if pg == nil {
		return nil
	}
	return &commonv1.PaginationResponse{
		Page:       int32(pg.Page),
		PerPage:    int32(pg.PerPage),
		Total:      int32(pg.Total),
		TotalPages: int32(pg.TotalPages),
	}
}

// Helper: convert row fhir_json to bytes
func rowFHIRBytes(fhirJSON string) []byte {
	return []byte(fhirJSON)
}

// levenshtein computes the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

// soundex computes a basic Soundex code for phonetic matching.
func soundex(s string) string {
	if len(s) == 0 {
		return ""
	}

	result := make([]byte, 0, 4)
	s = toLower(s)
	result = append(result, toUpper(s[0]))

	lastCode := soundexCode(s[0])
	for i := 1; i < len(s) && len(result) < 4; i++ {
		code := soundexCode(s[i])
		if code != '0' && code != lastCode {
			result = append(result, code)
		}
		lastCode = code
	}

	for len(result) < 4 {
		result = append(result, '0')
	}
	return string(result)
}

func soundexCode(c byte) byte {
	switch c {
	case 'b', 'f', 'p', 'v':
		return '1'
	case 'c', 'g', 'j', 'k', 'q', 's', 'x', 'z':
		return '2'
	case 'd', 't':
		return '3'
	case 'l':
		return '4'
	case 'm', 'n':
		return '5'
	case 'r':
		return '6'
	default:
		return '0'
	}
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

func toUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

// parseGivenNames parses a JSON array string of given names.
func parseGivenNames(givenNamesJSON string) []string {
	var names []string
	json.Unmarshal([]byte(givenNamesJSON), &names)
	return names
}
