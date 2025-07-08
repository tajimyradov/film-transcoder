package repository

import (
	models "github.com/tajimyradov/transcode/models"

	"github.com/jmoiron/sqlx"
)

type VideosRepository struct {
	db *sqlx.DB
}

func NewVideosRepository(db *sqlx.DB) *VideosRepository {
	return &VideosRepository{db: db}
}

func (f *VideosRepository) GetStudioByID(abbr string) (models.Studio, error) {
	var studio models.Studio
	query := `select id, name, abbreviated from studios where abbreviated = $1`
	err := f.db.QueryRow(query, abbr).Scan(&studio.ID, &studio.Name, &studio.Abbreviated)
	if err != nil {
		return models.Studio{}, err
	}
	return studio, nil
}

func (f *VideosRepository) GetOriginalFileOfVideo(id int) (string, error) {
	var res string
	query := `select filepath from files where video_id=$1`
	err := f.db.QueryRow(query, id).Scan(&res)
	if err != nil {
		return "", err
	}

	return res, nil
}

func (f *VideosRepository) UpdateTranscodeFiles(id int, hls string) error {
	return nil
}
