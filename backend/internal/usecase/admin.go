// filepath: /home/nok1o/university_projects/bmstu_student_app/backend/internal/usecase/admin.go
package usecase

import (
    "context"

    "stud_hub/internal/models"

    "github.com/google/uuid"
)

type AdminUserRepository interface {
    GetAllUsers(ctx context.Context, page, perPage int) ([]models.User, int, error)
    GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
    UpdateUser(ctx context.Context, user models.User) (models.User, error)
    DeleteUser(ctx context.Context, id uuid.UUID) error
    ChangeUserRole(ctx context.Context, id uuid.UUID, role models.RoleType) error
    BlockUser(ctx context.Context, id uuid.UUID, block bool) error
}

type AdminUseCase struct {
    userRepo AdminUserRepository
}

func NewAdminUseCase(userRepo AdminUserRepository) *AdminUseCase {
    return &AdminUseCase{
        userRepo: userRepo,
    }
}

func (a *AdminUseCase) GetAllUsers(ctx context.Context, page, perPage int) ([]models.User, int, error) {
    return a.userRepo.GetAllUsers(ctx, page, perPage)
}

func (a *AdminUseCase) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
    return a.userRepo.GetUserByID(ctx, id)
}

func (a *AdminUseCase) UpdateUser(ctx context.Context, user models.User) (models.User, error) {
    return a.userRepo.UpdateUser(ctx, user)
}

func (a *AdminUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
    return a.userRepo.DeleteUser(ctx, id)
}

func (a *AdminUseCase) BlockUser(ctx context.Context, id uuid.UUID, block bool) error {
    return a.userRepo.BlockUser(ctx, id, block)
}

func (a *AdminUseCase) ChangeUserRole(ctx context.Context, id uuid.UUID, role models.RoleType) error {
    return a.userRepo.ChangeUserRole(ctx, id, role)
}
