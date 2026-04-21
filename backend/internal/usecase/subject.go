package usecase

import (
    "context"

    "stud_hub/internal/models"

    "github.com/google/uuid"
)

type SubjectUseCase struct {
    subjectRepo SubjectRepository
}

func NewSubjectUseCase(subjectRepo SubjectRepository) *SubjectUseCase {
    return &SubjectUseCase{subjectRepo: subjectRepo}
}

func (u *SubjectUseCase) GetByID(ctx context.Context, id uuid.UUID) (models.Subject, error) {
    return u.subjectRepo.GetByID(ctx, id)
}
