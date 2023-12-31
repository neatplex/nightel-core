package story

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/neatplex/nightel-core/internal/database"
	"github.com/neatplex/nightel-core/internal/models"
	"gorm.io/gorm"
	"strings"
)

type Service struct {
	database *database.Database
}

func (s *Service) Index(userId uint64) ([]models.Story, error) {
	var stories []models.Story
	r := s.database.Handler().
		Where("user_id = ?", userId).
		Preload("Audio").Preload("Image").
		Find(&stories)
	if r.Error != nil {
		return nil, errors.New(fmt.Sprintf("cannot index stories: %v", r.Error))
	}
	return stories, nil
}

func (s *Service) Create(story *models.Story) (string, error) {
	if story.Identity == "" {
		story.Identity = strings.ReplaceAll(uuid.NewString(), "-", "0")
	}

	r := s.database.Handler().Create(story)
	if r.Error != nil {
		return "", errors.New(fmt.Sprintf("cannot query to create story: %v", r.Error))
	}
	return story.Identity, nil
}

func (s *Service) FindByIdentity(identity string) (*models.Story, error) {
	var story models.Story
	r := s.database.Handler().Where("identity = ?", identity).First(&story)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.New(fmt.Sprintf("cannot find story by identity `%s`, err: %v", identity, r.Error))
	}
	return &story, nil
}

func New(database *database.Database) *Service {
	return &Service{database: database}
}
