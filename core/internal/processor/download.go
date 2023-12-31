package processor

import (
	"strings"
	"sync"

	"github.com/krau/manyacg/core/internal/common"
	"github.com/krau/manyacg/core/internal/common/logger"
	dtoModel "github.com/krau/manyacg/core/internal/model/dto"
	"github.com/krau/manyacg/core/internal/service"
)

func download(artworks []*dtoModel.ArtworkRaw, ch chan *dtoModel.PictureRaw) {
	var wg sync.WaitGroup
	for _, artwork := range artworks {
		for _, picture := range artwork.Pictures {
			wg.Add(1)
			go downloadPicture(picture, ch, &wg)
		}
	}
	wg.Wait()
	close(ch)
}

func downloadPicture(picture *dtoModel.PictureRaw, ch chan *dtoModel.PictureRaw, wg *sync.WaitGroup) {
	defer wg.Done()
	if picture.Binary != nil || picture.Downloaded {
		logger.L.Debugf("Picture already downloaded, pass: %s", picture.DirectURL)
		return
	}
	pictureDB, err := service.GetPictureByDirectURL(picture.DirectURL)
	if err != nil {
		logger.L.Errorf("Failed to get picture by direct url: %s", err)
		return
	}
	if pictureDB != nil {
		if pictureDB.FilePath != "" || pictureDB.Downloaded {
			logger.L.Debugf("Picture already downloaded in database, pass: %s", picture.DirectURL)
			picture.FilePath = pictureDB.FilePath
			picture.Downloaded = true
			return
		}
	}
	logger.L.Debugf("Downloading picture from %s", picture.DirectURL)
	resp, err := common.ReqCilent.R().Get(picture.DirectURL)
	if err != nil {
		logger.L.Errorf("Download failed: %s, error: %s", picture.DirectURL, err)
		return
	}
	defer resp.Body.Close()
	picture.Binary = resp.Bytes()
	picture.Downloaded = true
	picture.Format = strings.Split(picture.DirectURL, ".")[len(strings.Split(picture.DirectURL, "."))-1]
	ch <- picture
}
