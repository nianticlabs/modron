package collector

type Type string

const (
	Gcp  Type = "gcp"
	Fake Type = "fake"
)

var validCollectors = []Type{Gcp, Fake}

func ValidCollectors() []string {
	var collectors []string
	for _, c := range validCollectors {
		collectors = append(collectors, string(c))
	}
	return collectors
}
