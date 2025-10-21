// main.go - Backend en Go para AetherPlatform
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
	"github.com/gin-gonic/gin"
	 _ "aetherplatform/docs" // Importa los docs generados por swag
    ginSwagger "github.com/swaggo/gin-swagger"
    swaggerFiles "github.com/swaggo/files"
)

// @title AetherPlatform API
// @version 1.0
// @description API de autenticación para AetherPlatform
// @host localhost:8000
// @basePath /AetherPlatform/api


// Estructuras
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"-"`
	Role     string `json:"role,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	    User  struct {
        ID    string `json:"id"`
        Email string `json:"email"`
        Name  string `json:"name,omitempty"`
    } `json:"user"`}


type ErrorResponse struct {
	Message string `json:"message"`
}

// Clave secreta para JWT (en producción usar variable de entorno)
var jwtSecret = []byte("tu-clave-secreta-super-segura-cambiar-en-produccion")

// Base de datos temporal en memoria (en producción usar una BD real)
var users = map[string]User{}

// Middleware para verificar JWT
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Token no proporcionado")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Formato de token inválido")
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inesperado")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "Token inválido")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Claims inválidos")
			return
		}

		// Pasar el user ID al contexto
		r.Header.Set("User-ID", claims["user_id"].(string))
		next(w, r)
	}
}

// Handlers
// @Summary Login de usuario
// @Description Valida credenciales y devuelve token JWT
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   loginRequest body LoginRequest true "Credenciales"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		respondWithError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Buscar usuario
	user, exists := users[loginReq.Email]
	if !exists {
		respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Generar JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expira en 24 horas
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generando token")
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{
		Token: tokenString,
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name,omitempty"`
		}{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		},
	})
}

// Gin handler wrapper for handleLogin
func ginHandleLogin(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Datos inválidos"})
		return
	}

	user, exists := users[loginReq.Email]
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Credenciales inválidas"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Message: "Credenciales inválidas"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Message: "Error generando token"})
		return
	}

	resp := LoginResponse{
		Token: tokenString,
	}
	resp.User.ID = user.ID
	resp.User.Email = user.Email
	resp.User.Name = user.Name

	c.JSON(http.StatusOK, resp)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	// Verificar si el usuario ya existe
	if _, exists := users[user.Email]; exists {
		respondWithError(w, http.StatusConflict, "El usuario ya existe")
		return
	}

	// Hashear contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error procesando contraseña")
		return
	}

	// Crear usuario
	user.ID = fmt.Sprintf("user_%d", time.Now().Unix())
	user.Password = string(hashedPassword)
	user.Role = "user"
	users[user.Email] = user

	// Limpiar contraseña antes de responder
	user.Password = ""
	respondWithJSON(w, http.StatusCreated, user)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// En una implementación real, podrías invalidar el token aquí
	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Sesión cerrada exitosamente",
	})
}

func handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	
	// Buscar usuario por ID
	var foundUser *User
	for _, user := range users {
		if user.ID == userID {
			foundUser = &user
			break
		}
	}
	
	if foundUser == nil {
		respondWithError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}
	// Aquí podrías devolver información del usuario o simplemente confirmar que el token es válido
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"user": foundUser,
	})
}
	


func handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("User-ID")
	
	// Generar nuevo token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generando token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": tokenString,
	})
}

// Helpers
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ErrorResponse{Message: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func main() {
	r := gin.Default()
	api := r.Group("/AetherPlatform/api")
	{
		api.POST("/auth/login/", ginHandleLogin)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run((":8000"))
	// Crear usuario de prueba
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	users["test@example.com"] = User{
		ID:       "user_1",
		Email:    "test@example.com",
		Name:     "Usuario de Prueba",
		Password: string(hashedPassword),
		Role:     "admin",
	}

	// Configurar rutas
	router := mux.NewRouter()
	
	// Rutas públicas
	router.HandleFunc("/api/auth/login", handleLogin).Methods("POST")
	router.HandleFunc("/api/auth/register", handleRegister).Methods("POST")
	
	// Rutas protegidas
	router.HandleFunc("/api/auth/logout", authMiddleware(handleLogout)).Methods("POST")
	router.HandleFunc("/api/auth/verify", authMiddleware(handleVerifyToken)).Methods("GET")
	router.HandleFunc("/api/auth/refresh", authMiddleware(handleRefreshToken)).Methods("POST")

	// Configurar CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	fmt.Println("🚀 Servidor AetherPlatform corriendo en http://localhost:8080")
	fmt.Println("📧 Usuario de prueba: test@example.com / password123")
	log.Fatal(http.ListenAndServe(":8080", handler))
}