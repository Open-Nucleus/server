package gitstore

import (
	"fmt"
	"strings"
	"time"
)

// CommitMessage holds structured commit metadata per spec §3.3.
type CommitMessage struct {
	ResourceType string
	Operation    string
	ResourceID   string
	NodeID       string
	Author       string
	SiteID       string
	Timestamp    time.Time
}

// CommitInfo holds metadata from a parsed git commit.
type CommitInfo struct {
	Hash      string
	Timestamp time.Time
	Message   string
}

// Format produces a structured commit message per spec §3.3.
func (cm CommitMessage) Format() string {
	return fmt.Sprintf("[%s] %s %s\n\nnode: %s\nauthor: %s\nsite: %s\ntimestamp: %s\nfhir_version: R4",
		cm.ResourceType,
		cm.Operation,
		cm.ResourceID,
		cm.NodeID,
		cm.Author,
		cm.SiteID,
		cm.Timestamp.UTC().Format(time.RFC3339),
	)
}

// ParseCommitMessage parses a structured commit message back into a CommitMessage.
func ParseCommitMessage(msg string) (CommitMessage, error) {
	var cm CommitMessage
	lines := strings.Split(msg, "\n")
	if len(lines) < 1 {
		return cm, fmt.Errorf("empty commit message")
	}

	// Parse first line: [ResourceType] OPERATION resource-id
	header := lines[0]
	openBracket := strings.Index(header, "[")
	closeBracket := strings.Index(header, "]")
	if openBracket < 0 || closeBracket < 0 {
		return cm, fmt.Errorf("invalid header format: %s", header)
	}
	cm.ResourceType = header[openBracket+1 : closeBracket]

	rest := strings.TrimSpace(header[closeBracket+1:])
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) >= 1 {
		cm.Operation = parts[0]
	}
	if len(parts) >= 2 {
		cm.ResourceID = parts[1]
	}

	// Parse body key-value pairs
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, ": ", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "node":
			cm.NodeID = kv[1]
		case "author":
			cm.Author = kv[1]
		case "site":
			cm.SiteID = kv[1]
		case "timestamp":
			t, err := time.Parse(time.RFC3339, kv[1])
			if err == nil {
				cm.Timestamp = t
			}
		}
	}

	return cm, nil
}
