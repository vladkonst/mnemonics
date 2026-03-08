package content

import (
	"fmt"
	"time"
)

// MnemonicType defines the content type of a mnemonic.
type MnemonicType string

const (
	MnemonicTypeText  MnemonicType = "text"
	MnemonicTypeImage MnemonicType = "image"
)

// Mnemonic is a learning aid attached to a Theme.
type Mnemonic struct {
	ID          int
	ThemeID     int
	Type        MnemonicType
	ContentText *string
	S3ImageKey  *string
	OrderNum    int
	CreatedAt   time.Time
}

// Validate checks business rules for a Mnemonic.
func (m *Mnemonic) Validate() error {
	switch m.Type {
	case MnemonicTypeText:
		if m.ContentText == nil || *m.ContentText == "" {
			return fmt.Errorf("text mnemonic requires content_text")
		}
	case MnemonicTypeImage:
		if m.S3ImageKey == nil || *m.S3ImageKey == "" {
			return fmt.Errorf("image mnemonic requires s3_image_key")
		}
	default:
		return fmt.Errorf("unknown mnemonic type %q", m.Type)
	}
	return nil
}
