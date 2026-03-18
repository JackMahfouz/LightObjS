package main

import (
	"context"
	"fmt"
	"time"

	"LightObjS/internal/domain"
	"LightObjS/internal/storage"

	"github.com/google/uuid"
)

func main() {
	// 1. Initialize Infrastructure (Adapter)
	s3Store := storage.NewS3Provider("production-assets", "us-east-1")

	// 2. Mock a RESTful Upload request
	ctx := context.Background()
	newObj := &domain.Object{
		ID:          uuid.New().String(),
		Key:         "uploads/2024/report.pdf",
		Size:        2048,
		ContentType: "application/pdf",
		Metadata: map[string]string{
			"Author":  "Architect",
			"Project": "Go-Object-Store",
		},
		CreatedAt: time.Now(),
	}

	fmt.Println("--- Object Storage Service Initialized ---")

	// 3. Execute Operations via the Storage Interface
	if err := s3Store.Upload(ctx, newObj, []byte("binary-content")); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Successfully processed object: %s (UUID: %s)\n", newObj.Key, newObj.ID)

	// Demonstrate Metadata retrieval
	fmt.Printf("Metadata: %v\n", newObj.Metadata)
}
