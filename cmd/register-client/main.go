package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/jyogi-web/jyogi-discord-auth/internal/config"
	"github.com/jyogi-web/jyogi-discord-auth/internal/domain"
	gormRepo "github.com/jyogi-web/jyogi-discord-auth/internal/repository/gorm"
	"github.com/jyogi-web/jyogi-discord-auth/internal/service"
)

func main() {
	clientID := flag.String("id", "", "Client ID")
	clientSecret := flag.String("secret", "", "Client Secret (Plain text)")
	name := flag.String("name", "", "Client Name")
	redirectURIs := flag.String("redirects", "", "Comma separated redirect URIs")
	update := flag.Bool("update", false, "Update existing client instead of creating new one")
	flag.Parse()

	if *clientID == "" {
		flag.Usage()
		log.Fatal("Client ID is required")
	}

	if !*update && (*clientSecret == "" || *name == "" || *redirectURIs == "") {
		flag.Usage()
		log.Fatal("Client Secret, Name, and Redirect URIs are required for new clients")
	}

	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// データベースを初期化
	db, err := gormRepo.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	clientRepo := gormRepo.NewClientRepository(db)
	clientService := service.NewClientService(clientRepo)

	var uris []string
	if *redirectURIs != "" {
		uris = strings.Split(*redirectURIs, ",")
		for i, uri := range uris {
			uris[i] = strings.TrimSpace(uri)
		}
	}

	ctx := context.Background()
	var client *domain.ClientApp

	if *update {
		client, err = clientService.UpdateClient(ctx, *clientID, *clientSecret, *name, uris)
		if err != nil {
			log.Fatalf("Failed to update client: %v", err)
		}
		fmt.Printf("Successfully updated client: %s (ID: %s)\n", client.Name, client.ID)
	} else {
		client, err = clientService.RegisterClient(ctx, *clientID, *clientSecret, *name, uris)
		if err != nil {
			log.Fatalf("Failed to register client: %v", err)
		}
		fmt.Printf("Successfully registered client: %s (ID: %s)\n", client.Name, client.ID)
	}

	fmt.Printf("Client ID: %s\n", client.ClientID)
	if *clientSecret != "" {
		fmt.Println("Client Secret has been hashed and stored securely.")
	}
}
