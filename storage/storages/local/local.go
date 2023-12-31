package local

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/krau/manyacg/core/api/rpc/proto"
	coreErrors "github.com/krau/manyacg/core/pkg/common/errors"
	"github.com/krau/manyacg/storage/client"
	"github.com/krau/manyacg/storage/config"
	"github.com/krau/manyacg/storage/logger"
)

var dir string = config.Cfg.Storages.Local.Dir

func (s *StorageLocal) SaveArtworks(artworks []*proto.ProcessedArtworkInfo) {
	for _, artwork := range artworks {
		if artwork == nil {
			logger.L.Fatalf("Artwork is nil")
			continue
		}
		go s.saveArtwork(artwork)
	}
}

func (s *StorageLocal) saveArtwork(artwork *proto.ProcessedArtworkInfo) {
	ctx := context.Background()
	pictures := artwork.Pictures
	// 以来源名/作者名/标题为目录名
	artworkDir := dir + "/" + strings.ReplaceAll(artwork.Source.String(), "/", "_") + "/" + strings.ReplaceAll(artwork.Author, "/", "_") + "/" + strings.ReplaceAll(artwork.Title, "/", "_")
	if _, err := os.Stat(artworkDir); os.IsNotExist(err) {
		err := os.MkdirAll(artworkDir, os.ModePerm)
		if err != nil {
			logger.L.Errorf("Error creating dir: %v", err)
			return
		}
	}
	for _, picture := range pictures {
		// 以直链尾部为文件名
		fileName := artworkDir + "/" + strings.Split(picture.DirectURL, "/")[len(strings.Split(picture.DirectURL, "/"))-1]

		stream, err := client.ArtworkClient.GetPictureData(ctx, &proto.GetPictureDataRequest{PictureID: picture.PictureID})
		if err != nil {
			logger.L.Errorf("Error getting picture data: %v", err)
			if errors.Is(err, coreErrors.ErrPictureNotFound) {
				return
			}
			return
		}

		var file *os.File

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.L.Errorf("Error getting picture data: %v", err)
				// 删除文件
				err := os.Remove(fileName)
				if err != nil {
					logger.L.Errorf("Error removing file: %v", err)
				}
				return
			}
			if file == nil {
				file, err = os.Create(fileName)
				if err != nil {
					logger.L.Errorf("Error creating file: %v", err)
					return
				}
			}
			_, err = file.Write(resp.Binary)
			if err != nil {
				logger.L.Errorf("Error writing file: %v", err)
				return
			}
		}
		file.Close()
		logger.L.Infof("Saved picture %s", fileName)
	}
	dirFiles, err := os.ReadDir(artworkDir)
	if err != nil {
		logger.L.Errorf("Error reading dir: %v", err)
		return
	}
	if len(dirFiles) == 0 {
		logger.L.Debugf("Removing empty dir %s", artworkDir)
		err := os.Remove(artworkDir)
		if err != nil {
			logger.L.Errorf("Error removing empty dir: %v", err)
			return
		}
	}
}
