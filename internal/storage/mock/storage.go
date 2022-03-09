// Code generated by MockGen. DO NOT EDIT.
// Source: storage.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	types "github.com/Decentr-net/decentr/x/community/types"
	storage "github.com/Decentr-net/theseus/internal/storage"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockStorage is a mock of Storage interface
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *MockStorageMockRecorder
}

// MockStorageMockRecorder is the mock recorder for MockStorage
type MockStorageMockRecorder struct {
	mock *MockStorage
}

// NewMockStorage creates a new mock instance
func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &MockStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStorage) EXPECT() *MockStorageMockRecorder {
	return m.recorder
}

// InTx mocks base method
func (m *MockStorage) InTx(ctx context.Context, f func(storage.Storage) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InTx", ctx, f)
	ret0, _ := ret[0].(error)
	return ret0
}

// InTx indicates an expected call of InTx
func (mr *MockStorageMockRecorder) InTx(ctx, f interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InTx", reflect.TypeOf((*MockStorage)(nil).InTx), ctx, f)
}

// SetHeight mocks base method
func (m *MockStorage) SetHeight(ctx context.Context, height uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeight", ctx, height)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeight indicates an expected call of SetHeight
func (mr *MockStorageMockRecorder) SetHeight(ctx, height interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeight", reflect.TypeOf((*MockStorage)(nil).SetHeight), ctx, height)
}

// GetHeight mocks base method
func (m *MockStorage) GetHeight(ctx context.Context) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHeight", ctx)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHeight indicates an expected call of GetHeight
func (mr *MockStorageMockRecorder) GetHeight(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHeight", reflect.TypeOf((*MockStorage)(nil).GetHeight), ctx)
}

// RefreshViews mocks base method
func (m *MockStorage) RefreshViews(ctx context.Context, postView, statsView bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshViews", ctx, postView, statsView)
	ret0, _ := ret[0].(error)
	return ret0
}

// RefreshViews indicates an expected call of RefreshViews
func (mr *MockStorageMockRecorder) RefreshViews(ctx, postView, statsView interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshViews", reflect.TypeOf((*MockStorage)(nil).RefreshViews), ctx, postView, statsView)
}

// Follow mocks base method
func (m *MockStorage) Follow(ctx context.Context, follower, followee string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Follow", ctx, follower, followee)
	ret0, _ := ret[0].(error)
	return ret0
}

// Follow indicates an expected call of Follow
func (mr *MockStorageMockRecorder) Follow(ctx, follower, followee interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Follow", reflect.TypeOf((*MockStorage)(nil).Follow), ctx, follower, followee)
}

// Unfollow mocks base method
func (m *MockStorage) Unfollow(ctx context.Context, follower, followee string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unfollow", ctx, follower, followee)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unfollow indicates an expected call of Unfollow
func (mr *MockStorageMockRecorder) Unfollow(ctx, follower, followee interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unfollow", reflect.TypeOf((*MockStorage)(nil).Unfollow), ctx, follower, followee)
}

// ListPosts mocks base method
func (m *MockStorage) ListPosts(ctx context.Context, p *storage.ListPostsParams) ([]*storage.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPosts", ctx, p)
	ret0, _ := ret[0].([]*storage.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPosts indicates an expected call of ListPosts
func (mr *MockStorageMockRecorder) ListPosts(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPosts", reflect.TypeOf((*MockStorage)(nil).ListPosts), ctx, p)
}

// CreatePost mocks base method
func (m *MockStorage) CreatePost(ctx context.Context, p *storage.CreatePostParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePost", ctx, p)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePost indicates an expected call of CreatePost
func (mr *MockStorageMockRecorder) CreatePost(ctx, p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePost", reflect.TypeOf((*MockStorage)(nil).CreatePost), ctx, p)
}

// GetPost mocks base method
func (m *MockStorage) GetPost(ctx context.Context, id storage.PostID) (*storage.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPost", ctx, id)
	ret0, _ := ret[0].(*storage.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPost indicates an expected call of GetPost
func (mr *MockStorageMockRecorder) GetPost(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPost", reflect.TypeOf((*MockStorage)(nil).GetPost), ctx, id)
}

// GetPostBySlug mocks base method
func (m *MockStorage) GetPostBySlug(ctx context.Context, slug string) (*storage.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPostBySlug", ctx, slug)
	ret0, _ := ret[0].(*storage.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPostBySlug indicates an expected call of GetPostBySlug
func (mr *MockStorageMockRecorder) GetPostBySlug(ctx, slug interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPostBySlug", reflect.TypeOf((*MockStorage)(nil).GetPostBySlug), ctx, slug)
}

// DeletePost mocks base method
func (m *MockStorage) DeletePost(ctx context.Context, id storage.PostID, timestamp time.Time, deletedBy string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePost", ctx, id, timestamp, deletedBy)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePost indicates an expected call of DeletePost
func (mr *MockStorageMockRecorder) DeletePost(ctx, id, timestamp, deletedBy interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePost", reflect.TypeOf((*MockStorage)(nil).DeletePost), ctx, id, timestamp, deletedBy)
}

// GetLikes mocks base method
func (m *MockStorage) GetLikes(ctx context.Context, likedBy string, id ...storage.PostID) (map[storage.PostID]types.LikeWeight, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, likedBy}
	for _, a := range id {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetLikes", varargs...)
	ret0, _ := ret[0].(map[storage.PostID]types.LikeWeight)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLikes indicates an expected call of GetLikes
func (mr *MockStorageMockRecorder) GetLikes(ctx, likedBy interface{}, id ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, likedBy}, id...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLikes", reflect.TypeOf((*MockStorage)(nil).GetLikes), varargs...)
}

// SetLike mocks base method
func (m *MockStorage) SetLike(ctx context.Context, id storage.PostID, weight types.LikeWeight, timestamp time.Time, likeOwner string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetLike", ctx, id, weight, timestamp, likeOwner)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetLike indicates an expected call of SetLike
func (mr *MockStorageMockRecorder) SetLike(ctx, id, weight, timestamp, likeOwner interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLike", reflect.TypeOf((*MockStorage)(nil).SetLike), ctx, id, weight, timestamp, likeOwner)
}

// AddPDV mocks base method
func (m *MockStorage) AddPDV(ctx context.Context, address string, amount int64, timestamp time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddPDV", ctx, address, amount, timestamp)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddPDV indicates an expected call of AddPDV
func (mr *MockStorageMockRecorder) AddPDV(ctx, address, amount, timestamp interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPDV", reflect.TypeOf((*MockStorage)(nil).AddPDV), ctx, address, amount, timestamp)
}

// GetProfileStats mocks base method
func (m *MockStorage) GetProfileStats(ctx context.Context, addr ...string) ([]*storage.ProfileStats, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range addr {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetProfileStats", varargs...)
	ret0, _ := ret[0].([]*storage.ProfileStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProfileStats indicates an expected call of GetProfileStats
func (mr *MockStorageMockRecorder) GetProfileStats(ctx interface{}, addr ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, addr...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfileStats", reflect.TypeOf((*MockStorage)(nil).GetProfileStats), varargs...)
}

// GetPostStats mocks base method
func (m *MockStorage) GetPostStats(ctx context.Context, id ...storage.PostID) (map[storage.PostID]storage.Stats, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx}
	for _, a := range id {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetPostStats", varargs...)
	ret0, _ := ret[0].(map[storage.PostID]storage.Stats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPostStats indicates an expected call of GetPostStats
func (mr *MockStorageMockRecorder) GetPostStats(ctx interface{}, id ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx}, id...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPostStats", reflect.TypeOf((*MockStorage)(nil).GetPostStats), varargs...)
}

// GetDecentrStats mocks base method
func (m *MockStorage) GetDecentrStats(ctx context.Context) (*storage.DecentrStats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDecentrStats", ctx)
	ret0, _ := ret[0].(*storage.DecentrStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDecentrStats indicates an expected call of GetDecentrStats
func (mr *MockStorageMockRecorder) GetDecentrStats(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDecentrStats", reflect.TypeOf((*MockStorage)(nil).GetDecentrStats), ctx)
}

// ResetAccount mocks base method
func (m *MockStorage) ResetAccount(ctx context.Context, owner string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResetAccount", ctx, owner)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResetAccount indicates an expected call of ResetAccount
func (mr *MockStorageMockRecorder) ResetAccount(ctx, owner interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResetAccount", reflect.TypeOf((*MockStorage)(nil).ResetAccount), ctx, owner)
}
