package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrInvalidSignature = errors.New("Invalid webhook signature")
	ErrMissingSignature = errors.New("Missing webhook signature")
	ErrInvalidPayload   = errors.New("Invalid webhook payload")
)

// Represents a GitHub push webhook payload
type PushEvent struct {
	Ref        string     `json:"ref"`    // "refs/heads/main"
	Before     string     `json:"before"` // Previous commit SHA
	After      string     `json:"after"`  // New commit SHA
	Created    bool       `json:"created"`
	Deleted    bool       `json:"deleted"`
	Forced     bool       `json:"forced"`
	Repository Repository `json:"repository"`
	HeadCommit *Commit    `json:"head_commit"`
	Pusher     Pusher     `json:"pusher"`
	Sender     Sender     `json:"sender"`
}

type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	CloneURL string `json:"clone_url"`
	SSHURL   string `json:"ssh_url"`
}

type Commit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Author    Author `json:"author"`
	Committer Author `json:"committer"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type Pusher struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Sender struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatar_url"`
}

// Verify webhook payload came from GitHub using HMAC SHA256
// Signature header is in format: sha256=<hex>
func ValidateSignature(payload []byte, signatureHeader,
	secret string) error {
	if signatureHeader == "" {
		return ErrMissingSignature
	}

	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return ErrInvalidSignature
	}

	signatureHex := strings.TrimPrefix(signatureHeader, "sha256=")
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return ErrInvalidSignature
	}

	// Compute HMAC SHA256
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedSignature := mac.Sum(nil)

	// Compare signatures (contant-time to avoid timing attacks)
	if subtle.ConstantTimeCompare(signature, expectedSignature) != 1 {
		return ErrInvalidSignature
	}

	return nil
}

// Parse a GitHub push webhook payload
func ParsePushEvent(payload []byte) (*PushEvent, error) {
	var event PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, ErrInvalidPayload
	}

	return &event, nil
}

// Extract branch from the ref ("refs/heads/main" -> "main")
func (e *PushEvent) GetBranch() string {
	return strings.TrimPrefix(e.Ref, "refs/heads/")
}

// Returns true if the push event should trigger a deployment
func (e *PushEvent) ShouldDeploy() bool {
	// Don't deploy deleted branch
	if e.Deleted {
		return false
	}
	// Don't deploy if no commit
	if e.HeadCommit == nil {
		return false
	}
	// Don't deploy if commit SHA is all zeros (GitHub quirk)
	if e.After == "0000000000000000000000000000000000000000" {
		return false
	}

	return true
}

// Return commit details for the deployment record
func (e *PushEvent) GetCommitInfo() (sha, message, author string) {
	sha = e.After
	if e.HeadCommit != nil {
		message = e.HeadCommit.Message
		author = e.HeadCommit.Author.Name
		if author == "" {
			author = e.HeadCommit.Author.Username
		}
	}
	if author == "" {
		author = e.Pusher.Name
	}
	return
}
