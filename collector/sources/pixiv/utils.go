package pixiv

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"sync"

	"github.com/krau/manyacg/collector/common"
	"github.com/krau/manyacg/collector/logger"
	sourceModels "github.com/krau/manyacg/collector/sources/models"
	coreModels "github.com/krau/manyacg/core/models"
)

func getArtworkInfo(sourceURL string) (*sourceModels.PixivAjaxResp, error) {
	pid := strings.Split(sourceURL, "/")[len(strings.Split(sourceURL, "/"))-1]
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + pid
	logger.L.Debugf("Fetching artwork info: %s", ajaxURL)
	resp, err := common.Cilent.R().Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivAjaxResp sourceModels.PixivAjaxResp
	err = json.Unmarshal([]byte(resp.String()), &pixivAjaxResp)
	if err != nil {
		return nil, err
	}
	return &pixivAjaxResp, nil
}

func getNewArtworksForURL(url string, limit int, wg *sync.WaitGroup, artworkChan chan *coreModels.ArtworkRaw) {
	defer wg.Done()
	logger.L.Infof("Fetching %s", url)
	resp, err := common.Cilent.R().Get(url)

	if err != nil {
		logger.L.Errorf("Error fetching %s: %v", url, err)
		return
	}

	var pixivRss *sourceModels.PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)

	if err != nil {
		logger.L.Errorf("Error decoding %s: %v", url, err)
		return
	}

	logger.L.Debugf("Got %d items", len(pixivRss.Channel.Items))

	for i, item := range pixivRss.Channel.Items {
		if i >= limit {
			break
		}
		wg.Add(1)
		go func(item sourceModels.Item) {
			defer wg.Done()
			artworkInfo, err := getArtworkInfo(item.Link)
			if err != nil {
				logger.L.Errorf("Error fetching artwork info: %v", err)
				return
			}

			artworkChan <- item.ToArtworkRaw(artworkInfo)
		}(item)
	}
}
