package api

import (
	"context"
	"io"
	"testing"
)

func TestService_GenerateDescriptionStreaming_requiresStepSettings(t *testing.T) {
	svc := NewService()
	_, err := svc.GenerateDescriptionStreaming(context.Background(), GenerateDescriptionRequest{
		SourceBranch: "a",
		TargetBranch: "b",
		Files:        nil,
		Prompt:       "p",
	}, io.Discard)
	if err == nil {
		t.Fatalf("expected error when step settings are nil, got nil")
	}
}
