package gormstorage

import (
	pb "github.com/nianticlabs/modron/src/proto/generated"
)

// SeverityScore represents a score for the severity - this number is in the 0.0 - 10.0 range.
type SeverityScore float32

const (
	SeverityLowMax    = 3.9
	SeverityMediumMax = 6.9
	SeverityHighMax   = 8.9

	SeverityMin = 0.0
	SeverityMax = 10.0
)

func (s SeverityScore) ToSeverity() pb.Severity {
	if s < 0 {
		// This should never happen (SeverityScore should be nil if anything)
		return pb.Severity_SEVERITY_UNKNOWN
	}
	if s == SeverityMin {
		return pb.Severity_SEVERITY_INFO
	}
	if s <= SeverityLowMax {
		return pb.Severity_SEVERITY_LOW
	}
	if s <= SeverityMediumMax {
		return pb.Severity_SEVERITY_MEDIUM
	}
	if s <= SeverityHighMax {
		return pb.Severity_SEVERITY_HIGH
	}
	return pb.Severity_SEVERITY_CRITICAL
}

func FromSeverityPb(severity pb.Severity) *SeverityScore {
	switch severity {
	case pb.Severity_SEVERITY_INFO:
		info := SeverityScore(0.0)
		return &info
	case pb.Severity_SEVERITY_LOW:
		low := SeverityScore(SeverityLowMax)
		return &low
	case pb.Severity_SEVERITY_MEDIUM:
		medium := SeverityScore(SeverityMediumMax)
		return &medium
	case pb.Severity_SEVERITY_HIGH:
		high := SeverityScore(SeverityHighMax)
		return &high
	case pb.Severity_SEVERITY_CRITICAL:
		critical := SeverityScore(SeverityMax)
		return &critical
	}
	unknownScore := SeverityScore(-1)
	return &unknownScore
}
