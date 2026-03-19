package api

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"LightObjS/internal/Auth"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	jwtSecret  = "super-secret-jwt-key" // In a real app, use an environment variable
	dbFileName = "users.db"
)

var (
	validate = validator.New()
)

func Main() {
	// Ensure the Data directory exists
	dataDir := "./Data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create Data directory: %v", err)
	}

	dbPath := filepath.Join(dataDir, dbFileName)

	// Initialize Auth components for user management
	userRepo, err := auth.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize user repository: %v", err)
	}
	defer userRepo.Close() // Close the database connection when Main exits

	authService := auth.NewAuthService(userRepo, jwtSecret)
	authMiddleware := auth.NewAuthMiddleware(jwtSecret, userRepo)

	// Initialize Auth components for client (API key) management
	clientRepo := auth.NewSQLiteClientRepository(userRepo.DB()) // Access the underlying *sql.DB
	apiKeyService := auth.NewAPIKeyService(clientRepo)
	clientAuthMiddleware := auth.NewClientAuthMiddleware(apiKeyService)

	// Initialize Gin router
	router := gin.Default()

	// Load HTML templates
	router.SetHTMLTemplate(template.Must(template.ParseFiles(
		"web/templates/login.html",
		"web/templates/admin_dashboard.html",
	)))

	// Serve static files
	router.StaticFS("/static", http.Dir("web/static"))

	// Redirect root to login page
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/login")
	})

	// Web interface routes
	router.GET("/web/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	// Public API routes for user authentication (used by web interface via AJAX)
	router.POST("/register", registerHandler(authService))
	router.POST("/login", loginHandler(authService))

	// Protected API routes for dashboard users (used by web interface via AJAX)
	protected := router.Group("/protected")
	protected.Use(authMiddleware.AuthRequired())
	{
		protected.GET("/", protectedHandler)
		protected.POST("/api-keys", generateAPIKeyHandler(apiKeyService))
		protected.GET("/api-keys", listAPIKeysHandler(apiKeyService))
	}

	// Admin Web and API routes
	admin := router.Group("/admin")
	admin.GET("/dashboard", func(c *gin.Context) { // This route serves the HTML page
		c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{})
	})
	// API endpoints for admin dashboard (used by JS via AJAX)
	admin.Use(authMiddleware.AuthRequired(), auth.AdminAuthRequired()) // Apply middleware to API endpoints only
	{
		admin.POST("/users", createUserHandler(authService, apiKeyService))
		admin.GET("/users", listUsersHandler(userRepo))
	}

	// Protected routes for client (API key) access to storage
	storage := router.Group("/api/v1/storage")
	storage.Use(clientAuthMiddleware.ClientAuthRequired())
	{
		storage.POST("/upload", clientProtectedStorageHandler)
		storage.GET("/download/:key", clientProtectedStorageHandler) // Example storage route
	}

	fmt.Println("--- LightObjS API Server Initialized ---")
	fmt.Printf("Listening on :8080\n")
	log.Fatal(router.Run(":8080")) // Listen and serve on 0.0.0.0:8080
}

// registerHandler handles user registration requests.
func registerHandler(service auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := service.Register(req.Username, req.Password, false) // Register regular user
		if err != nil {
			if err == auth.ErrUserAlreadyExists {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				log.Printf("Error registering user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
			}
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": user.ID})
	}
}

// loginHandler handles user login requests.
func loginHandler(service auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req auth.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, err := service.Login(req.Username, req.Password)
		if err != nil {
			if err == auth.ErrInvalidCredentials || err == auth.ErrUserNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			} else {
				log.Printf("Error logging in user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
			}
			return
		}

		c.JSON(http.StatusOK, auth.LoginResponse{Token: token})
	}
}

// protectedHandler is an example of a route that requires user authentication.
func protectedHandler(c *gin.Context) {
	authUser := auth.GetUserFromContext(c)
	if authUser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "You accessed a protected resource!",
		"user_id":  authUser.ID,
		"username": authUser.Username,
	})
}

// generateAPIKeyHandler allows an authenticated user to generate new API keys for storage access.
func generateAPIKeyHandler(apiKeyService auth.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser := auth.GetUserFromContext(c)
		if authUser == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		client, plainTextSecret, err := apiKeyService.GenerateClientCredentials(authUser.ID)
		if err != nil {
			log.Printf("Error generating API key for user %s: %v", authUser.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":    "API Key generated successfully. PLEASE SAVE THIS SECRET, IT WILL NOT BE SHOWN AGAIN.",
			"client_id":  client.ID,
			"api_key":    client.APIKey,
			"api_secret": plainTextSecret, // This is the only time the plain-text secret is returned
		})
	}
}

// listAPIKeysHandler allows an authenticated user to view their generated API keys (excluding secrets).
func listAPIKeysHandler(apiKeyService auth.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authUser := auth.GetUserFromContext(c)
		if authUser == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		clients, err := apiKeyService.GetClientsForUser(authUser.ID)
		if err != nil {
			log.Printf("Error listing API keys for user %s: %v", authUser.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list API keys"})
			return
		}

		// Prepare response, omitting APISecret for security
		responseClients := make([]gin.H, len(clients))
		for i, client := range clients {
			responseClients[i] = gin.H{
				"client_id":  client.ID,
				"api_key":    client.APIKey,
				"created_at": client.CreatedAt,
				"expires_at": client.ExpiresAt,
				"is_active":  client.IsActive,
			}
		}

		c.JSON(http.StatusOK, gin.H{"api_keys": responseClients})
	}
}

// clientProtectedStorageHandler is a placeholder for a storage route protected by client API keys.
func clientProtectedStorageHandler(c *gin.Context) {
	authClient := auth.GetClientFromContext(c)
	if authClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Client not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "You accessed a client-protected storage resource!",
		"client_id":     authClient.ID,
		"api_key":       authClient.APIKey,
		"owned_by_user": authClient.UserID,
		"endpoint":      c.Request.URL.Path,
	})
}

// CreateUserRequest represents the payload for creating a new user by an admin.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	IsAdmin  bool   `json:"is_admin"`
}

// createUserHandler allows an admin to create a new user and automatically generate an API key for them.
func createUserHandler(authService auth.AuthService, apiKeyService auth.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := validate.Struct(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create the user. Pass the IsAdmin flag from the request.
		user, err := authService.Register(req.Username, req.Password, req.IsAdmin)
		if err != nil {
			if err == auth.ErrUserAlreadyExists {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				log.Printf("Error creating user by admin: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			}
			return
		}

		// Generate API key for the new user
		client, plainTextSecret, err := apiKeyService.GenerateClientCredentials(user.ID)
		if err != nil {
			log.Printf("Error generating API key for new user %s: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key for new user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":    "User created and API Key generated successfully",
			"user_id":    user.ID,
			"username":   user.Username,
			"is_admin":   user.IsAdmin,
			"client_id":  client.ID,
			"api_key":    client.APIKey,
			"api_secret": plainTextSecret, // Return plain-text secret once
		})
	}
}

// listUsersHandler allows an admin to list all registered users.
func listUsersHandler(userRepo auth.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := userRepo.GetAllUsers()
		if err != nil {
			log.Printf("Error listing all users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}

		// Prepare response, omitting hashed passwords
		responseUsers := make([]gin.H, len(users))
		for i, user := range users {
			responseUsers[i] = gin.H{
				"id":         user.ID,
				"username":   user.Username,
				"created_at": user.CreatedAt,
				"is_admin":   user.IsAdmin,
			}
		}

		c.JSON(http.StatusOK, gin.H{"users": responseUsers})
	}
}
