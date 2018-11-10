package overlaynetwork

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/libp2p/go-libp2p-record"
)

// TODO: we should take our time a define a number of validators that we
//  think will be useful for DHT data. Right now a basic Sha256 validator
// is probably sufficient.

var (
	// ErrInvalidNamespace represents an error that is returned when the passed
	// key does not match the required namespace
	ErrInvalidNamespace = errors.New("invalid namespace")

	// ErrInvalidSha256Record represents an error that is returned when the
	// key is not the hex encoded sha256 hash of the value.
	ErrInvalidSha256Record = errors.New("value does not hash to the key")
)

// Sha256Validator is a basic validator used by the DHT to validate that
// the key for any given record is the hex encoded sha256 hash of the value.
type Sha256Validator struct {}

// Validate validates the given record, returning an error if it's
// invalid (e.g., expired, signed by the wrong key, etc.).
func (v *Sha256Validator) Validate(key string, value []byte) error {
	ns, key, err := record.SplitKey(key)
	if err != nil {
		return err
	}
	if ns != "sha256" {
		return errors.New("namespace not 'sha256'")
	}
	h := sha256.Sum256(value)
	if key != hex.EncodeToString(h[:]) {
		return ErrInvalidSha256Record
	}
	return nil
}

// Select selects the best record from the set of records (e.g., the
// newest).
//
// Decisions made by select should be stable.
func (v *Sha256Validator) Select(key string, values [][]byte) (int, error) {
	return 0, nil
}