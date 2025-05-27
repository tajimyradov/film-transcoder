package repository

import (
	"github.com/jmoiron/sqlx"
	"transcode/models"
)

type FilmsRepository struct {
	db *sqlx.DB
}

func NewFilmsRepository(db *sqlx.DB) *FilmsRepository {
	return &FilmsRepository{db: db}
}

func (f *FilmsRepository) GetStudioByID(abbr string) (models.Studio, error) {
	var studio models.Studio
	query := `select id, name, type, abbreviated from studios where abbreviated = $1 and type=2`
	err := f.db.QueryRow(query, abbr).Scan(&studio.ID, &studio.Name, &studio.Type, &studio.Abbreviated)
	if err != nil {
		return models.Studio{}, err
	}
	return studio, nil
}
