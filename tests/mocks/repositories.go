package mocks

import (
	"be_uas/app/model/mongodb"
	"be_uas/app/model/postgres"
	"context"

	"github.com/stretchr/testify/mock"
)

// MOCK USER REPO 
type UserRepo struct {
	mock.Mock
}

func (m *UserRepo) GetByUsername(username string) (*postgres.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.User), args.Error(1)
}

// Dummy methods untuk memenuhi interface
func (m *UserRepo) CreateUser(user postgres.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepo) GetAllUsers() ([]postgres.User, error)           { return nil, nil }
func (m *UserRepo) GetUserByID(id string) (*postgres.User, error)   { return nil, nil }
func (m *UserRepo) UpdateUser(user postgres.User) error             { return nil }
func (m *UserRepo) DeleteUser(id string) error                      { return nil }

func (m *UserRepo) GetRoleIDByName(roleName string) (string, error) {
	args := m.Called(roleName)
	return args.String(0), args.Error(1)
}

func (m *UserRepo) UpdateUserRole(userID, roleID string) error      { return nil }

func (m *UserRepo) GetStudentIDByUserID(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// MOCK ACHIEVEMENT REPO (POSTGRES)
type AchievementRepoPG struct {
	mock.Mock
}

func (m *AchievementRepoPG) CreateReference(ref postgres.AchievementReference) error {
	args := m.Called(ref)
	return args.Error(0)
}

func (m *AchievementRepoPG) GetReferenceByID(id string) (*postgres.AchievementReference, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postgres.AchievementReference), args.Error(1)
}

func (m *AchievementRepoPG) UpdateStatus(id, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *AchievementRepoPG) GetAchievementsByAdvisorID(advisorID string) ([]postgres.AchievementReference, error) {
	args := m.Called(advisorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]postgres.AchievementReference), args.Error(1)
}

func (m *AchievementRepoPG) GetAchievementsByStudentID(studentID string) ([]postgres.AchievementReference, error) {
	args := m.Called(studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]postgres.AchievementReference), args.Error(1)
}

func (m *AchievementRepoPG) GetStudentIDByUserID(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *AchievementRepoPG) UpdateVerification(id string, status string, verifiedBy string, rejectionNote *string) error {
	args := m.Called(id, status, verifiedBy, rejectionNote)
	return args.Error(0)
}

func (m *AchievementRepoPG) GetAllAchievements(limit, offset int) ([]postgres.AchievementReference, int, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]postgres.AchievementReference), args.Int(1), args.Error(2)
}

// Dummy methods
func (m *AchievementRepoPG) GetAllReferences(limit, offset int) ([]postgres.AchievementReference, error) {
	return nil, nil
}
func (m *AchievementRepoPG) GetReferencesByStudentID(studentID string) ([]postgres.AchievementReference, error) {
	return nil, nil
}
func (m *AchievementRepoPG) UpdateReferenceVerification(id string, verifiedBy string, verifiedAt interface{}, status string, note string) error {
	args := m.Called(id, verifiedBy, verifiedAt, status, note)
	return args.Error(0)
}
func (m *AchievementRepoPG) GetReferencesByAdvisorID(advisorID string) ([]postgres.AchievementReference, error) {
	return nil, nil
}
func (m *AchievementRepoPG) SoftDeleteReference(id string) error { return nil }

// MOCK ACHIEVEMENT REPO (MONGO) 
type AchievementRepoMongo struct {
	mock.Mock
}

func (m *AchievementRepoMongo) InsertAchievement(ctx context.Context, data mongodb.Achievement) (string, error) {
	args := m.Called(ctx, data)
	return args.String(0), args.Error(1)
}

// Dummy methods
func (m *AchievementRepoMongo) AddAttachment(ctx context.Context, hexID string, attachment mongodb.Attachment) error {
	return nil
}
func (m *AchievementRepoMongo) FindAchievementByID(ctx context.Context, hexID string) (*mongodb.Achievement, error) {
	return nil, nil
}
func (m *AchievementRepoMongo) UpdateAchievement(ctx context.Context, hexID string, data mongodb.Achievement) error {
	return nil
}
func (m *AchievementRepoMongo) SoftDeleteAchievement(ctx context.Context, hexID string) error {
	return nil
}
