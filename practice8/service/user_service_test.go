package service

import (
	"errors"
	"practice8/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)


func TestGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().GetUserByID(1).Return(user, nil)

	result, err := svc.GetUserByID(1)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}


func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := svc.CreateUser(user)
	assert.NoError(t, err)
}


func TestRegisterUser_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	existing := &repository.User{ID: 2, Name: "Existing", Email: "test@example.com"}
	mockRepo.EXPECT().GetByEmail("test@example.com").Return(existing, nil)

	newUser := &repository.User{Name: "New"}
	err := svc.RegisterUser(newUser, "test@example.com")
	assert.EqualError(t, err, "user with this email already exists")
}

func TestRegisterUser_NewUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{Name: "New User", Email: "new@example.com"}
	mockRepo.EXPECT().GetByEmail("new@example.com").Return(nil, nil)
	mockRepo.EXPECT().CreateUser(user).Return(nil)

	err := svc.RegisterUser(user, "new@example.com")
	assert.NoError(t, err)
}

func TestRegisterUser_RepositoryErrorOnCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{Name: "User"}
	mockRepo.EXPECT().GetByEmail("user@example.com").Return(nil, nil)
	mockRepo.EXPECT().CreateUser(user).Return(errors.New("db connection failed"))

	err := svc.RegisterUser(user, "user@example.com")
	assert.Error(t, err)
}


func TestUpdateUserName_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	err := svc.UpdateUserName(1, "")
	assert.EqualError(t, err, "name cannot be empty")
}

func TestUpdateUserName_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	mockRepo.EXPECT().GetUserByID(99).Return(nil, errors.New("user not found"))

	err := svc.UpdateUserName(99, "New Name")
	assert.Error(t, err)
}

func TestUpdateUserName_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 5, Name: "Old Name"}
	mockRepo.EXPECT().GetUserByID(5).Return(user, nil)
	// Verify that name was actually changed before UpdateUser is called
	mockRepo.EXPECT().UpdateUser(gomock.Any()).DoAndReturn(func(u *repository.User) error {
		assert.Equal(t, "New Name", u.Name)
		return nil
	})

	err := svc.UpdateUserName(5, "New Name")
	assert.NoError(t, err)
}

func TestUpdateUserName_UpdateUserFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	user := &repository.User{ID: 5, Name: "Old Name"}
	mockRepo.EXPECT().GetUserByID(5).Return(user, nil)
	mockRepo.EXPECT().UpdateUser(gomock.Any()).Return(errors.New("db write failed"))

	err := svc.UpdateUserName(5, "New Name")
	assert.Error(t, err)
}


func TestDeleteUser_AttemptToDeleteAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	err := svc.DeleteUser(1)
	assert.EqualError(t, err, "it is not allowed to delete admin user")
}

func TestDeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	mockRepo.EXPECT().DeleteUser(42).Return(nil)

	err := svc.DeleteUser(42)
	assert.NoError(t, err)
}

func TestDeleteUser_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	svc := NewUserService(mockRepo)

	mockRepo.EXPECT().DeleteUser(5).Return(errors.New("db error on delete"))

	err := svc.DeleteUser(5)
	assert.Error(t, err)
}
