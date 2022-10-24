// Code generated by MockGen. DO NOT EDIT.
// Source: event.go

// Package mock_repository is a generated GoMock package.
package mock_repository

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/porter-dev/porter-agent/internal/models"
	utils "github.com/porter-dev/porter-agent/internal/utils"
)

// MockEventRepository is a mock of EventRepository interface.
type MockEventRepository struct {
	ctrl     *gomock.Controller
	recorder *MockEventRepositoryMockRecorder
}

// MockEventRepositoryMockRecorder is the mock recorder for MockEventRepository.
type MockEventRepositoryMockRecorder struct {
	mock *MockEventRepository
}

// NewMockEventRepository creates a new mock instance.
func NewMockEventRepository(ctrl *gomock.Controller) *MockEventRepository {
	mock := &MockEventRepository{ctrl: ctrl}
	mock.recorder = &MockEventRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventRepository) EXPECT() *MockEventRepositoryMockRecorder {
	return m.recorder
}

// CreateEvent mocks base method.
func (m *MockEventRepository) CreateEvent(event *models.Event) (*models.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEvent", event)
	ret0, _ := ret[0].(*models.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEvent indicates an expected call of CreateEvent.
func (mr *MockEventRepositoryMockRecorder) CreateEvent(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEvent", reflect.TypeOf((*MockEventRepository)(nil).CreateEvent), event)
}

// DeleteEvent mocks base method.
func (m *MockEventRepository) DeleteEvent(uid string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEvent", uid)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEvent indicates an expected call of DeleteEvent.
func (mr *MockEventRepositoryMockRecorder) DeleteEvent(uid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEvent", reflect.TypeOf((*MockEventRepository)(nil).DeleteEvent), uid)
}

// ListEvents mocks base method.
func (m *MockEventRepository) ListEvents(filter *utils.ListEventsFilter, opts ...utils.QueryOption) ([]*models.Event, *utils.PaginatedResult, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{filter}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListEvents", varargs...)
	ret0, _ := ret[0].([]*models.Event)
	ret1, _ := ret[1].(*utils.PaginatedResult)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListEvents indicates an expected call of ListEvents.
func (mr *MockEventRepositoryMockRecorder) ListEvents(filter interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{filter}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEvents", reflect.TypeOf((*MockEventRepository)(nil).ListEvents), varargs...)
}

// ReadEvent mocks base method.
func (m *MockEventRepository) ReadEvent(id uint) (*models.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadEvent", id)
	ret0, _ := ret[0].(*models.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadEvent indicates an expected call of ReadEvent.
func (mr *MockEventRepositoryMockRecorder) ReadEvent(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadEvent", reflect.TypeOf((*MockEventRepository)(nil).ReadEvent), id)
}

// UpdateEvent mocks base method.
func (m *MockEventRepository) UpdateEvent(event *models.Event) (*models.Event, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateEvent", event)
	ret0, _ := ret[0].(*models.Event)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateEvent indicates an expected call of UpdateEvent.
func (mr *MockEventRepositoryMockRecorder) UpdateEvent(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEvent", reflect.TypeOf((*MockEventRepository)(nil).UpdateEvent), event)
}
