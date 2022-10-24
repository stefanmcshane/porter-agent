package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/logger"
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
	mockRepository "github.com/porter-dev/porter-agent/internal/repository/mocks"
)

func TestArgoCDResourceHookConsumer_Consume(t *testing.T) {
	testTime := time.Now()

	tests := []struct {
		name        string
		argoEvent   types.ArgoCDResourceHook
		mockSetup   func(mer *mockRepository.MockEventRepository)
		expectedErr error
	}{
		{
			name: "ConsumeEvent_OutOfSync_happy_path",
			argoEvent: types.ArgoCDResourceHook{
				Application:          "test porter application",
				ApplicationNamespace: "test namespace",
				Status:               "OutOfSync",
				Author:               "porter developer",
				Timestamp:            testTime.String(),
			},
			mockSetup: func(mer *mockRepository.MockEventRepository) {
				mer.EXPECT().CreateEvent(gomock.Any()).Return(&models.Event{
					ReleaseName:      "test porter application",
					ReleaseNamespace: "test namespace",
					Type:             types.EventTypeDeploymentFinished,
					Timestamp:        &testTime,
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "ConsumeEvent_Synced_happy_path",
			argoEvent: types.ArgoCDResourceHook{
				Application:          "test porter application",
				ApplicationNamespace: "test namespace",
				Status:               "Synced",
				Author:               "porter developer",
				Timestamp:            testTime.String(),
			},
			mockSetup: func(mer *mockRepository.MockEventRepository) {
				mer.EXPECT().CreateEvent(gomock.Any()).Return(&models.Event{
					ReleaseName:      "test porter application",
					ReleaseNamespace: "test namespace",
					Type:             types.EventTypeDeploymentFinished,
					Timestamp:        &testTime,
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "ConsumeEvent_unsupported_type",
			argoEvent: types.ArgoCDResourceHook{
				Application:          "test porter application",
				ApplicationNamespace: "test namespace",
				Status:               "ThisIsNotASupporteredType",
				Author:               "porter developer",
				Timestamp:            testTime.String(),
			},
			mockSetup:   func(mer *mockRepository.MockEventRepository) {},
			expectedErr: errors.New("unsupported type ThisIsNotASupporteredType"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			lo := logger.NewConsole(false)

			ctrl := gomock.NewController(t)
			mer := mockRepository.NewMockEventRepository(ctrl)
			tt.mockSetup(mer)

			consumer := ArgoCDResourceHookConsumer{
				logger: lo,
				Repository: &repository.Repository{
					Event: mer,
				},
			}
			err := consumer.Consume(ctx, tt.argoEvent)
			errorConditions(t, tt.expectedErr, err)
		})
	}
}

func errorConditions(t *testing.T, expectedError error, actualError error) {
	if expectedError == actualError {
		return
	}
	if expectedError != nil && actualError == nil {
		t.Fatalf("expected error [%s] to occur, but no error did", expectedError)
	}
	if expectedError == nil && actualError != nil {
		t.Fatalf("unexpected error occurred. actual error: %s", actualError.Error())
	}
}
