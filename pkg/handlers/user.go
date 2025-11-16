package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"goland_api/pkg/database"
	"goland_api/pkg/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	jwt "github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Документация для метода GetUsers
// @Summary Возвращает список всех пользователей
// @Description Получение списка всех пользователей
// @Tags Пользователи
// @Accept  application/json
// @Produce  application/json
// @Success 200 {object} []models.User
// @Failure 400 Bad Request
// @Failure 500 Internal Server Error
// @Router /api/users [get]
func GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := database.DB.Query("SELECT id, name, email, phone, city, logo, media, status, created_at FROM users")
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()

		users := []models.UserView{}
		for rows.Next() {
			var user models.UserView
			if err := rows.Scan(
				&user.ID,
				&user.Name,
				&user.Email,
				&user.Phone,
				&user.City,
				&user.Logo,
				&user.Media,
				&user.Status,
				&user.CreatedAt); err != nil {
				log.Println(err)
			}
			users = append(users, user)
		}
		if err := rows.Err(); err != nil {
			log.Println(err)
		}

		json.NewEncoder(w).Encode(users)
	}
}

func getUserFromToken(token *jwt.Token) (error, *models.UserView) {
	// Извлекаем claims
	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		// Получаем значение jti
		userEmail := claims.Id // jti хранится в поле Id структуры Claims
		errorResponse, userView := getUserViewByEmail(userEmail)
		if errorResponse != nil {
			return errorResponse, nil
		}

		return nil, &userView
	}

	return fmt.Errorf("Не смог прочитать токен"), nil
}

func getUserViewById(paramId int64) (error, models.UserView) {
	var userView models.UserView
	var role models.Role

	err := database.DB.QueryRow(
		"SELECT "+
			"u.id, u.name, u.email, u.phone, u.city, u.logo, u.media, u.status, u.created_at, "+
			"r.id, r.name "+
			"FROM users u "+
			"join roles r on r.id = u.role_id "+
			"WHERE u.id = $1", paramId).Scan(
		&userView.ID,
		&userView.Name,
		&userView.Email,
		&userView.Phone,
		&userView.City,
		&userView.Logo,
		&userView.Media,
		&userView.Status,
		&userView.CreatedAt,
		&role.ID,
		&role.Name,
	)
	userView.Role = role

	return err, userView
}

func getUserViewByEmail(paramEmail string) (error, models.UserView) {
	var userView models.UserView
	var role models.Role

	err := database.DB.QueryRow(
		"SELECT "+
			"u.id, u.name, u.email, u.phone, u.city, u.logo, u.media, u.status, u.created_at, "+
			"r.id, r.name "+
			"FROM users u "+
			"join roles r on r.id = u.role_id "+
			"WHERE email = $1", paramEmail).Scan(
		&userView.ID,
		&userView.Name,
		&userView.Email,
		&userView.Phone,
		&userView.City,
		&userView.Logo,
		&userView.Media,
		&userView.Status,
		&userView.CreatedAt,
		&role.ID,
		&role.Name,
	)
	userView.Role = role

	return err, userView
}

func getUserViewByIdByEmail(paramEmail string) (error, models.CreateUserRequest) {
	var user models.CreateUserRequest
	err := database.DB.QueryRow("SELECT id, name, email, phone, password FROM users WHERE email = $1", paramEmail).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Password,
	)

	return err, user
}

// Документация для метода GetUser
// @Summary Возвращает информацию о пользователе по ID
// @Description Получение информации о пользователе по идентификатору
// @Tags Пользователи
// @Param id path int true "ID пользователя"
// @Success 200 {object} models.User
// @Failure 400 Bad Request
// @Failure 404 Not Found
// @Router /api/users/{id} [get]
func GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramId, _ := strconv.Atoi(vars["id"])

		errorResponse, userView := getUserViewById(int64(paramId))
		if errorResponse != nil {
			SendJSONError(w, http.StatusBadRequest, errorResponse.Error())
			return
		}

		json.NewEncoder(w).Encode(userView)
	}
}

// Документация для метода InfoUser
// @Summary Возвращает информацию о текущем пользователе
// @Description Получение информации о текущем аутентифицированном пользователе
// @Tags Аутентификация
// @Accept  application/json
// @Produce  application/json
// @Success 200 {object} models.UserView
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/auth/info [get]
func InfoUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		json.NewEncoder(w).Encode(AUTH)
	}
}

// @Summary Аутентификация пользователя
// @Description Аутентификация пользователя по email и паролю с возвратом JWT токена
// @Tags Аутентификация
// @Param credentials body models.LoginUserRequest true "Учетные данные пользователя"
// @Accept  application/json
// @Produce  application/json
// @Success 200 {string} string "JWT токен"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /api/auth/login [post]
func Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			errorValidation, userRequest := validateLoginUserRequest(r)
			if errorValidation != nil {
				SendJSONError(w, http.StatusBadRequest, errorValidation.Error())
				return
			}

			errorQuery, user := getUserViewByIdByEmail(userRequest.Email)
			if errorQuery != nil {
				SendJSONError(w, http.StatusBadRequest, errorQuery.Error())
				return
			}

			checkPassword := checkPasswordHash(userRequest.Password, user.Password)

			if checkPassword != true {
				SendJSONError(w, http.StatusBadRequest, "Invalid password")
				return
			}

			tokenString, errorToken := getNewToken(user.Name, user.Email)
			if errorToken != nil {
				SendJSONError(w, http.StatusBadRequest, errorToken.Error())
			}

			json.NewEncoder(w).Encode(tokenString)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "POST, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// @Summary Обновление JWT токена
// @Description Обновление JWT токена для аутентифицированного пользователя
// @Tags Аутентификация
// @Produce  application/json
// @Success 200 {string} string "Новый JWT токен"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /api/auth/refresh [post]
func Refresh() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			tokenString, errorToken := getNewToken(AUTH.Name, AUTH.Email)
			if errorToken != nil {
				SendJSONError(w, http.StatusBadRequest, errorToken.Error())
			}

			json.NewEncoder(w).Encode(tokenString)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "POST, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// Документация для метода CreateUser
// @Summary Создание нового пользователя
// @Description Создание нового пользователя
// @Tags Пользователи
// @Param createUser body models.CreateUserRequest true "Данные для создания пользователя"
// @Consumes application/json
// @Produces application/json
// @Success 201 {object} string
// @Failure 422 Unprocessable Entity
// @Router /api/auth [post]
func CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPost {
			errValidate, userRequest := validateCreateUserRequest(r)

			if errValidate != nil {
				// Преобразуем ошибки валидации в JSON
				var validationErrors []models.ValidationErrorResponse
				for _, errValidate := range errValidate.(validator.ValidationErrors) {
					validationErrors = append(validationErrors, models.ValidationErrorResponse{
						Field:   errValidate.Field(),
						Message: fmt.Sprintf("Ошибка в поле '%s': '%s'", errValidate.Field(), errValidate.Tag()),
					})
				}
				// Создаем структуру для ответа с ошибкой
				errorData := models.ErrorResponse{
					StatusCode: http.StatusBadRequest,
					Message:    "Возникла ошибка при регистрации",
					Errors:     validationErrors,
				}
				// Сериализуем ошибки в JSON
				jsonResponse, err := json.Marshal(errorData)
				if err != nil {
					SendJSONError(w, http.StatusInternalServerError, "Ошибка при формировании ответа")
					return
				}
				// Устанавливаем заголовок Content-Type
				w.Header().Set("Content-Type", "application/json")
				// Устанавливаем код состояния HTTP
				w.WriteHeader(http.StatusBadRequest)
				// Отправляем JSON-ответ
				w.Write(jsonResponse)
				return
			}

			var user models.CreateUserRequest
			user.Name = userRequest.Name
			user.Email = userRequest.Email
			user.Phone = userRequest.Phone
			user.Password = getHashPassword(userRequest.Password)

			err := database.DB.QueryRow("INSERT INTO users (name, email, phone, password) VALUES ($1, $2, $3, $4) RETURNING id", user.Name, user.Email, user.Phone, user.Password).Scan(&user.ID)
			if err != nil {
				SendJSONError(w, http.StatusBadRequest, "Возникла ошибка при регистрации")
				return
			}

			tokenString, errorToken := getNewToken(user.Name, user.Email)
			if errorToken != nil {
				SendJSONError(w, http.StatusBadRequest, "Возникла ошибка при регистрации")
				return
			}

			json.NewEncoder(w).Encode(tokenString)
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "POST, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// Документация для метода UpdateUser
// @Summary Обновление существующего пользователя
// @Description Обновление существующего пользователя
// @Tags Пользователи
// @Param updateUser body models.UpdateUserRequest true "Данные для обновления пользователя"
// @Consumes application/json
// @Produces application/json
// @Param id path int true "ID пользователя"
// @Success 204 No Content
// @Failure 422 Unprocessable Entity
// @Failure 404 Not Found
// @Router /api/users [put]
func UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Устанавливаем заголовки для CORS
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Отправляем успешный ответ
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodPut {
			errValidate, userRequest := validateUpdatedAtUserRequest(r)
			if errValidate != nil {
				// Преобразуем ошибки валидации в JSON
				var validationErrors []models.ValidationErrorResponse
				for _, errValidate := range errValidate.(validator.ValidationErrors) {
					validationErrors = append(validationErrors, models.ValidationErrorResponse{
						Field:   errValidate.Field(),
						Message: fmt.Sprintf("Ошибка в поле '%s': '%s'", errValidate.Field(), errValidate.Tag()),
					})
				}
				// Создаем структуру для ответа с ошибкой
				errorData := models.ErrorResponse{
					StatusCode: http.StatusBadRequest,
					Message:    "Возникла ошибка при регистрации",
					Errors:     validationErrors,
				}
				// Сериализуем ошибки в JSON
				jsonResponse, err := json.Marshal(errorData)
				if err != nil {
					SendJSONError(w, http.StatusInternalServerError, "Ошибка при формировании ответа")
					return
				}
				// Устанавливаем заголовок Content-Type
				w.Header().Set("Content-Type", "application/json")
				// Устанавливаем код состояния HTTP
				w.WriteHeader(http.StatusBadRequest)
				// Отправляем JSON-ответ
				w.Write(jsonResponse)
				return
			}

			if userRequest != nil {
				AUTH.Name = userRequest.Name
				AUTH.Email = userRequest.Email
				AUTH.Phone = userRequest.Phone
				userPassword := getHashPassword(userRequest.Password)

				_, err := database.DB.Exec("UPDATE users SET name = $1, email = $2, phone = $3, password = $4 WHERE id = $5",
					AUTH.Name,
					AUTH.Email,
					AUTH.Phone,
					userPassword,
					userRequest.ID)

				if err != nil {
					log.Println(err)
				}

				json.NewEncoder(w).Encode(AUTH)
				return
			}

			SendJSONError(w, http.StatusInternalServerError, "Ошибка при формировании ответа")
			return
		}

		// Если метод не поддерживается
		w.Header().Set("Allow", "PUT, OPTIONS")
		SendJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
	}
}

// Документация для метода DeleteUser
// @Summary Удаляет пользователя по ID
// @Description Удаление пользователя по идентификатору
// @Tags Пользователи
// @Param id path int true "ID пользователя"
// @Success 204 No Content
// @Failure 404 Not Found
// @Router /api/users/{id} [delete]
func DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramId, _ := strconv.Atoi(vars["id"])

		var user models.User
		err := database.DB.QueryRow("SELECT * FROM users WHERE id = $1", paramId).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := database.DB.Exec("DELETE FROM users WHERE id = $1", paramId)
			if err != nil {
				//todo : fix error handling
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("User deleted")
		}
	}
}

func isUniqueEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	var checkUser models.CreateUserRequest
	err := database.DB.QueryRow("SELECT id, name, email, phone, password FROM users WHERE email = $1", email).Scan(
		&checkUser.ID,
		&checkUser.Name,
		&checkUser.Email,
		&checkUser.Phone,
		&checkUser.Password,
	)
	if err == sql.ErrNoRows {
		return true
	}
	return false
}

func isUniquePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	var checkUser models.CreateUserRequest
	err := database.DB.QueryRow("SELECT id, name, email, phone, password FROM users WHERE phone = $1", phone).Scan(
		&checkUser.ID,
		&checkUser.Name,
		&checkUser.Email,
		&checkUser.Phone,
		&checkUser.Password,
	)
	if err == sql.ErrNoRows {
		return true
	}
	return false
}

func validateCreateUserRequest(r *http.Request) (error, models.CreateUserRequest) {
	var userRequest models.CreateUserRequest
	// Парсим JSON из тела запроса
	if errJson := json.NewDecoder(r.Body).Decode(&userRequest); errJson != nil {
		fmt.Println("Неверный формат запроса JSON")
		return nil, userRequest
	}
	vars := mux.Vars(r)
	paramId, _ := strconv.Atoi(vars["id"])
	userRequest.ID = int64(paramId)

	validate := validator.New()
	validate.RegisterValidation("email", isUniqueEmail)
	validate.RegisterValidation("phone", isUniquePhone)
	errValidate := validate.Struct(userRequest)
	if errValidate != nil {
		// Если есть ошибки валидации, выводим их
		//for _, errValidate := range errValidate.(validator.ValidationErrors) {
		//	fmt.Println("Ошибка в поле '" + errValidate.Field() + "': '" + errValidate.Tag() + "'")
		//}
		return errValidate, userRequest
	} else {
		fmt.Println("Валидация прошла успешно!")
		return nil, userRequest
	}
}

// isUniqueEmailFactory создает функцию isUniqueEmail с захваченной переменной
func isUniqueEmailFactory(userRequest models.UpdateUserRequest) validator.Func {
	return func(fl validator.FieldLevel) bool {
		email := fl.Field().String()
		var checkUser models.CreateUserRequest
		err := database.DB.QueryRow("SELECT id, name, email, phone, password FROM users WHERE email = $1 AND id <> $2", email, userRequest.ID).Scan(
			&checkUser.ID,
			&checkUser.Name,
			&checkUser.Email,
			&checkUser.Phone,
			&checkUser.Password,
		)
		if err == sql.ErrNoRows {
			return true
		}
		return false
	}
}

// isUniquePhoneFactory создает функцию isUniqueEmail с захваченной переменной
func isUniquePhoneFactory(userRequest models.UpdateUserRequest) validator.Func {
	return func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		var checkUser models.CreateUserRequest
		err := database.DB.QueryRow("SELECT id, name, email, phone, password FROM users WHERE phone = $1 AND id <> $2", phone, userRequest.ID).Scan(
			&checkUser.ID,
			&checkUser.Name,
			&checkUser.Email,
			&checkUser.Phone,
			&checkUser.Password,
		)
		if err == sql.ErrNoRows {
			return true
		}
		return false
	}
}

func validateUpdatedAtUserRequest(r *http.Request) (error, *models.UpdateUserRequest) {
	var userRequest models.UpdateUserRequest
	// Парсим JSON из тела запроса
	if errJson := json.NewDecoder(r.Body).Decode(&userRequest); errJson != nil {
		fmt.Println("Неверный формат запроса JSON")
		return errJson, nil
	}

	userRequest.ID = AUTH.ID

	validate := validator.New()
	validate.RegisterValidation("email", isUniqueEmailFactory(userRequest))
	validate.RegisterValidation("phone", isUniquePhoneFactory(userRequest))

	errValidate := validate.Struct(userRequest)
	if errValidate != nil {
		return errValidate, nil
	}

	return nil, &userRequest
}

func validateLoginUserRequest(r *http.Request) (error, models.LoginUserRequest) {
	var req models.LoginUserRequest
	if validation := json.NewDecoder(r.Body).Decode(&req); validation != nil {
		return validation, req
	}
	validate := validator.New()
	if validation := validate.Struct(req); validation != nil {
		return validation, req
	}

	return nil, req
}

// getNewToken создает новый JWT-токен
func getNewToken(name string, email string) (string, error) {
	// ExpiresAt в миллисекундах от Unix epoch
	expiresAt := time.Now().Add(time.Hour).UnixMilli()
	claims := models.Claims{
		Username: name,
		StandardClaims: jwt.StandardClaims{
			Id:        email,
			Subject:   name,
			ExpiresAt: expiresAt,
		},
	}

	jwt_secret := os.Getenv("JWT_SECRET")
	var secretKey = []byte(jwt_secret)
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken проверяет и парсит JWT-токен
func ParseToken(tokenString string) (*jwt.Token, error) {
	jwt_secret := os.Getenv("JWT_SECRET")
	var secretKey = []byte(jwt_secret)

	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Возвращаем ключ для проверки подписи
		return []byte(secretKey), nil
	})
	if err != nil {
		log.Fatal("Ошибка при парсинге токена:", err)
	}

	// Извлекаем claims
	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		// Получаем значение jti
		// jti хранится в поле Id структуры StandardClaims
		_ = claims // явно указываем, что переменная используется
	} else {
		log.Fatal("Неверный токен или claims")
	}

	return token, err
}

func getHashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}
	return string(bytes)
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
