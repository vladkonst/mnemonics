package content

import (
	"fmt"
	"time"

	"github.com/vladkonst/mnemonics/pkg/apperrors"
)

// MnemonicType defines the content type of a mnemonic.
type MnemonicType string

const (
	MnemonicTypeText  MnemonicType = "text"
	MnemonicTypeImage MnemonicType = "image"
)

// Mnemonic is a learning aid attached to a Theme.
type Mnemonic struct {
	ID          int          `json:"id"`
	ThemeID     int          `json:"theme_id"`
	Type        MnemonicType `json:"type"`
	ContentText *string      `json:"content_text,omitempty"`
	S3ImageKey  *string      `json:"s3_image_key,omitempty"`
	OrderNum    int          `json:"order_num"`
	CreatedAt   time.Time    `json:"created_at"`
}

// Validate checks business rules for a Mnemonic.
func (m *Mnemonic) Validate() error {
	switch m.Type {
	case MnemonicTypeText:
		if m.ContentText == nil || *m.ContentText == "" {
			return fmt.Errorf("text mnemonic requires content_text: %w", apperrors.ErrInvalidInput)
		}
	case MnemonicTypeImage:
		if m.S3ImageKey == nil || *m.S3ImageKey == "" {
			return fmt.Errorf("image mnemonic requires s3_image_key: %w", apperrors.ErrInvalidInput)
		}
	default:
		return fmt.Errorf("unknown mnemonic type %q: %w", m.Type, apperrors.ErrInvalidInput)
	}
	return nil
}
