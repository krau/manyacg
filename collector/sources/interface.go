package sources

import (
	"github.com/krau/Picture-collector/core/models"
)

type Source interface {
	GetNewArtworks(limit int) (*models.ArtworkRaw, error)
}
