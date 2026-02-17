package ask

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

type stubProvider struct {
	resp string
	err  error
}

func (s *stubProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	return s.resp, s.err
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		resp    string
		err     error
		want    string
		wantErr bool
	}{
		{
			name:    "successfully asks a question and outputs response",
			args:    []string{"what", "is", "Go?"},
			resp:    "Go is a programming language.",
			want:    "Go is a programming language.\n",
			wantErr: false,
		},
		{
			name:    "returns error when no arguments provided",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "returns error when provider fails",
			args:    []string{"hello"},
			err:     errors.New("network error"),
			wantErr: true,
		},
		{
			name:    "handles single word question",
			args:    []string{"help"},
			resp:    "How can I help you?",
			want:    "How can I help you?\n",
			wantErr: false,
		},
		{
			name:    "handles long question with many arguments",
			args:    []string{"explain", "the", "difference", "between", "interfaces", "and", "structs", "in", "Go"},
			resp:    "Interfaces define behavior, structs define data.",
			want:    "Interfaces define behavior, structs define data.\n",
			wantErr: false,
		},
		{
			name:    "handles empty response from provider",
			args:    []string{"test"},
			resp:    "",
			want:    "\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			provider := &stubProvider{resp: tt.resp, err: tt.err}

			err := Run(context.Background(), provider, &output, nil, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotOutput := output.String()
			if gotOutput != tt.want {
				t.Errorf("output = %q, want %q", gotOutput, tt.want)
			}
		})
	}
}

func TestRun_NilOutputDefaultsToDiscard(t *testing.T) {
	provider := &stubProvider{resp: "test response"}
	err := Run(context.Background(), provider, nil, nil, []string{"test"})
	if err != nil {
		t.Errorf("Run() with nil output error = %v, want nil", err)
	}
}
