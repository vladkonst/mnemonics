package content

import (
	"testing"
)

func TestTest_Validate(t *testing.T) {
	validQuestion := Question{ID: 1, Text: "Q?", Type: QuestionTypeMultipleChoice, CorrectAnswer: "A"}

	cases := []struct {
		name    string
		test    Test
		wantErr bool
	}{
		{
			name:    "valid test",
			test:    Test{Questions: []Question{validQuestion}, PassingScore: 70, Difficulty: 3},
			wantErr: false,
		},
		{
			name:    "no questions",
			test:    Test{Questions: nil, PassingScore: 70, Difficulty: 3},
			wantErr: true,
		},
		{
			name:    "empty questions slice",
			test:    Test{Questions: []Question{}, PassingScore: 70, Difficulty: 3},
			wantErr: true,
		},
		{
			name:    "passing score too low",
			test:    Test{Questions: []Question{validQuestion}, PassingScore: -1, Difficulty: 3},
			wantErr: true,
		},
		{
			name:    "passing score too high",
			test:    Test{Questions: []Question{validQuestion}, PassingScore: 101, Difficulty: 3},
			wantErr: true,
		},
		{
			name:    "difficulty too low",
			test:    Test{Questions: []Question{validQuestion}, PassingScore: 70, Difficulty: 0},
			wantErr: true,
		},
		{
			name:    "difficulty too high",
			test:    Test{Questions: []Question{validQuestion}, PassingScore: 70, Difficulty: 6},
			wantErr: true,
		},
		{
			name:    "boundary passing score 0",
			test:    Test{Questions: []Question{validQuestion}, PassingScore: 0, Difficulty: 1},
			wantErr: false,
		},
		{
			name:    "boundary passing score 100",
			test:    Test{Questions: []Question{validQuestion}, PassingScore: 100, Difficulty: 5},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.test.Validate()
			if c.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMnemonic_Validate(t *testing.T) {
	text := "some text"
	key := "s3/key.png"

	cases := []struct {
		name    string
		m       Mnemonic
		wantErr bool
	}{
		{
			name:    "valid text mnemonic",
			m:       Mnemonic{Type: MnemonicTypeText, ContentText: &text},
			wantErr: false,
		},
		{
			name:    "text mnemonic missing content",
			m:       Mnemonic{Type: MnemonicTypeText},
			wantErr: true,
		},
		{
			name:    "text mnemonic empty content",
			m:       Mnemonic{Type: MnemonicTypeText, ContentText: strPtr("")},
			wantErr: true,
		},
		{
			name:    "valid image mnemonic",
			m:       Mnemonic{Type: MnemonicTypeImage, S3ImageKey: &key},
			wantErr: false,
		},
		{
			name:    "image mnemonic missing key",
			m:       Mnemonic{Type: MnemonicTypeImage},
			wantErr: true,
		},
		{
			name:    "image mnemonic empty key",
			m:       Mnemonic{Type: MnemonicTypeImage, S3ImageKey: strPtr("")},
			wantErr: true,
		},
		{
			name:    "unknown type",
			m:       Mnemonic{Type: "video"},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.m.Validate()
			if c.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
