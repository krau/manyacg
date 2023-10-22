package azurebus

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/krau/manyacg/core/config"
	"github.com/krau/manyacg/core/logger"
	"github.com/krau/manyacg/core/models"
)

type MessengerAzureBus struct{}

func (a *MessengerAzureBus) SubscribeArtworks(count int, ch chan []*models.ArtworkRaw) {
	if azureClient == nil {
		logger.L.Errorf("Azure client is nil")
		return
	}
	for {
		logger.L.Infof("Receiving messages")
		messages, err := azureSubscriber.ReceiveMessages(context.Background(), count, nil)
		if err != nil {
			logger.L.Errorf("Error receiving messages: %s", err.Error())
			return
		}
		logger.L.Debugf("Got %d messages", len(messages))
		artworks := make([]*models.ArtworkRaw, 0)
		for _, message := range messages {
			artwork := &models.ArtworkRaw{}
			err := json.Unmarshal(message.Body, artwork)
			if err != nil {
				logger.L.Errorf("Error unmarshalling message: %s", err.Error())
				continue
			}
			artworks = append(artworks, artwork)
			if config.Cfg.App.Debug != true {
				azureSubscriber.CompleteMessage(context.Background(), message, nil)
			}
		}
		ch <- artworks
	}
}

func (a *MessengerAzureBus) SendProcessedArtworks(artworks []*models.ArtworkRaw) error {
	if azureSender == nil {
		return errors.New("Azure sender is nil")
	}
	batch, err := azureSender.NewMessageBatch(context.Background(), nil)
	if err != nil {
		return err
	}
	succeeded := 0
	for _, artwork := range artworks {
		message := &models.MessageProcessedArtwork{
			ArtworkID:   artwork.ID,
			Title:       artwork.Title,
			Author:      artwork.Author,
			Description: artwork.Description,
			SourceURL:   artwork.SourceURL,
			Source:      string(artwork.Source),
			Tags:        artwork.Tags,
			R18:         artwork.R18,
			Pictures: make([]*struct {
				DirectURL string `json:"direct_url"`
			}, len(artwork.Pictures)),
		}
		for i, pic := range artwork.Pictures {
			message.Pictures[i] = &struct {
				DirectURL string `json:"direct_url"`
			}{
				DirectURL: pic.DirectURL,
			}
		}
		messageBytes, err := json.Marshal(message)
		if err != nil {
			logger.L.Errorf("Error marshalling message: %s", err.Error())
			continue
		}
		err = batch.AddMessage(&azservicebus.Message{
			Body:      messageBytes,
			MessageID: &artwork.SourceURL,
		}, nil)
		if err != nil {
			logger.L.Errorf("Error adding message to batch: %s", err.Error())
			continue
		}
		succeeded++
	}
	if err := azureSender.SendMessageBatch(context.Background(), batch, nil); err != nil {
		return err
	}
	logger.L.Infof("Sent %d processed artwork", succeeded)
	return nil
}
