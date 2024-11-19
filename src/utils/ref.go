package utils

func RefOrNull(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
