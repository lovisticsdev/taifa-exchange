package ids

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func New(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "ID"
	}

	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf(
			"%s-%s-fallback",
			strings.ToUpper(prefix),
			time.Now().UTC().Format("20060102150405"),
		)
	}

	return fmt.Sprintf(
		"%s-%s-%s",
		strings.ToUpper(prefix),
		time.Now().UTC().Format("20060102150405"),
		hex.EncodeToString(random),
	)
}

func NewCorrelationID() string {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf(
			"corr-%s-fallback",
			time.Now().UTC().Format("20060102150405"),
		)
	}

	return fmt.Sprintf(
		"corr-%s-%s",
		time.Now().UTC().Format("20060102150405"),
		hex.EncodeToString(random),
	)
}

func NewDecisionID() string {
	return New("DEC")
}

func NewPolicyID() string {
	return New("POL")
}

func NewAuditEventID() string {
	return New("EVT")
}
