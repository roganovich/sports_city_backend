package handlers

import (
	"encoding/json"
	"fmt"
	"goland_api/pkg/models"
	"net/http"
	"net/url"
	"os"
	"strconv"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// AUTH - глобальная переменная с авторизацией
var AUTH *models.UserView

// Массив допустимых ролей
var UserRoles = map[int]bool{
	1: true,
	2: true,
}
var AdminRoles = map[int]bool{
	10: true,
	11: true,
}

// Middleware для проверки авторизации
func getAuth(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenString := authHeader[len("Bearer "):]
	token, errToken := ParseToken(tokenString)
	if errToken != nil {
		SendJSONError(w, http.StatusBadRequest, "Неверный токен")
		return
	}

	if !token.Valid {
		SendJSONError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	errorResponse, userView := getUserFromToken(token)
	if errorResponse != nil {
		SendJSONError(w, http.StatusBadRequest, "Неверный токен")
		return
	}
	AUTH = userView
	return
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getAuth(w, r)
		// Если токен валиден, передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	})
}

func AuthUserMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getAuth(w, r)
		if _, exists := UserRoles[AUTH.Role.ID]; !exists {
			SendJSONError(w, http.StatusBadRequest, "Роль пользователя недопустима")
			return
		}
		// Если токен валиден, передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	})
}

func AuthAdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getAuth(w, r)
		if _, exists := AdminRoles[AUTH.Role.ID]; !exists {
			SendJSONError(w, http.StatusBadRequest, "Роль пользователя недопустима")
			return
		}
		// Если токен валиден, передаем запрос следующему обработчику
		next.ServeHTTP(w, r)
	})
}

/**
* Функция проверяет, что пользователь является клиентов
 */
func IsUser(auth models.UserView) bool {
	if _, exists := UserRoles[auth.Role.ID]; !exists {
		return false
	}
	return true
}

/**
* Функция проверяет, что пользователь является админом
 */
func IsAdmin(auth models.UserView) bool {
	if _, exists := AdminRoles[auth.Role.ID]; !exists {
		return false
	}
	return true
}

// Middleware для обработки CORS
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого источника (или укажите конкретный домен, например, "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Разрешаем методы, которые могут быть использованы
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		// Разрешаем заголовки, которые могут быть отправлены
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Если это OPTIONS запрос, просто завершаем его
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}

func JsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// SendJSONError отправляет JSON ответ об ошибке
func SendJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResponse := models.SimpleErrorResponse{
		Code:    code,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		// Если не удалось закодировать JSON, отправляем простой текст
		http.Error(w, message, code)
	}
}

var (
	validate *validator.Validate
	trans    ut.Translator
)

func getIntParam(params url.Values, key string, defaultValue int) int {
	value := params.Get(key) // Получаем значение параметра
	if value == "" {
		return defaultValue // Возвращаем значение по умолчанию, если параметр отсутствует
	}

	result, err := strconv.Atoi(value) // Преобразуем строку в int
	if err != nil {
		return defaultValue // Возвращаем значение по умолчанию в случае ошибки
	}

	return result
}

// varDump will print out any number of variables given to it
// e.g. varDump("test", 1234)
func varDump(myVar ...interface{}) {
	fmt.Printf("%v\n", myVar)
}

// dd will print out variables given to it (like varDump()) but
// will also stop execution from continuing.
func dd(myVar ...interface{}) {
	varDump(myVar...)
	os.Exit(1)
}
