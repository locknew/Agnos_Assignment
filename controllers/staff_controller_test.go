package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"AgnosAssignments/controllers"
	"AgnosAssignments/middlewares"
	dbmodel "AgnosAssignments/model"
	"AgnosAssignments/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultUsername = "admin"
	defaultPassword = "admin1234"
	defaultHospital = "default"
	jwtSecret       = "test-secret"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	_ = godotenv.Load("../.env")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "2208"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "agnos_assignment"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	if err := db.AutoMigrate(&dbmodel.Staff{}, &dbmodel.Patient{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func cleanupStaff(db *gorm.DB, username, hospital string) {
	db.Unscoped().Where("username = ? AND hospital = ?", username, hospital).
		Delete(&dbmodel.Staff{})
}

func cleanupPatient(db *gorm.DB, nationalID, passportID string) {
	db.Unscoped().Where("national_id = ? OR passport_id = ?", nationalID, passportID).
		Delete(&dbmodel.Patient{})
}

func setupLoginRouter(authService *services.AuthService) *gin.Engine {
	r := gin.New()
	r.POST("/staff/login", controllers.NewStaffController(authService).LoginStaff)
	return r
}

func setupCreateRouter(authService *services.AuthService) *gin.Engine {
	r := gin.New()
	r.POST("/staff/create", middlewares.OptionalAuth(authService), controllers.NewStaffController(authService).CreateStaff)
	return r
}

func setupSearchRouter(authService *services.AuthService, patientService *services.PatientService) *gin.Engine {
	r := gin.New()
	protected := r.Group("/")
	protected.Use(middlewares.AuthMiddleware(authService))
	protected.GET("/patient/search", controllers.NewPatientController(patientService).SearchPatient)
	return r
}

func setupPatientCreateRouter(authService *services.AuthService, patientService *services.PatientService) *gin.Engine {
	r := gin.New()
	protected := r.Group("/")
	protected.Use(middlewares.AuthMiddleware(authService))
	protected.POST("/patient/create", controllers.NewPatientController(patientService).CreatePatient)
	return r
}

func getToken(t *testing.T, r *gin.Engine, username, password, hospital string) string {
	t.Helper()
	b, _ := json.Marshal(map[string]string{
		"username": username, "password": password, "hospital": hospital,
	})
	req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp["token"]
}

func post(r *gin.Engine, path string, body map[string]string, token string) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func get(r *gin.Engine, path string, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── /staff/login ──────────────────────────────────────────────────────────────

func TestLoginStaff(t *testing.T) {
	db := setupTestDB(t)
	cleanupStaff(db, defaultUsername, defaultHospital)

	authService := services.NewAuthService(db, jwtSecret)
	authService.CreateStaff(services.CreateStaffInput{
		Username: defaultUsername, Password: defaultPassword, Hospital: defaultHospital,
	})
	defer cleanupStaff(db, defaultUsername, defaultHospital)

	r := setupLoginRouter(authService)

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
		wantToken  bool
	}{
		{
			name:       "success with default staff",
			body:       map[string]string{"username": defaultUsername, "password": defaultPassword, "hospital": defaultHospital},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:       "wrong password",
			body:       map[string]string{"username": defaultUsername, "password": "wrongpassword", "hospital": defaultHospital},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong username",
			body:       map[string]string{"username": "nobody", "password": defaultPassword, "hospital": defaultHospital},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong hospital",
			body:       map[string]string{"username": defaultUsername, "password": defaultPassword, "hospital": "other"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing password",
			body:       map[string]string{"username": defaultUsername, "hospital": defaultHospital},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := post(r, "/staff/login", tc.body, "")
			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d — body: %s", tc.wantStatus, w.Code, w.Body.String())
			}
			if tc.wantToken {
				var resp map[string]string
				json.Unmarshal(w.Body.Bytes(), &resp)
				if resp["token"] == "" {
					t.Errorf("expected non-empty token in response")
				}
			}
		})
	}
}

// ── /staff/create ─────────────────────────────────────────────────────────────

func TestCreateStaff(t *testing.T) {
	db := setupTestDB(t)
	cleanupStaff(db, defaultUsername, defaultHospital)

	authService := services.NewAuthService(db, jwtSecret)
	authService.CreateStaff(services.CreateStaffInput{
		Username: defaultUsername, Password: defaultPassword, Hospital: defaultHospital,
	})
	defer cleanupStaff(db, defaultUsername, defaultHospital)

	token := getToken(t, setupLoginRouter(authService), defaultUsername, defaultPassword, defaultHospital)
	r := setupCreateRouter(authService)

	tests := []struct {
		name       string
		body       map[string]string
		token      string
		wantStatus int
		cleanup    func()
	}{
		{
			name:       "success - create new staff with valid token",
			body:       map[string]string{"username": "newstaff", "password": "password123", "hospital": "hospital_a"},
			token:      token,
			wantStatus: http.StatusCreated,
			cleanup:    func() { cleanupStaff(db, "newstaff", "hospital_a") },
		},
		{
			name:       "success - same username different hospital",
			body:       map[string]string{"username": "newstaff", "password": "password123", "hospital": "hospital_b"},
			token:      token,
			wantStatus: http.StatusCreated,
			cleanup:    func() { cleanupStaff(db, "newstaff", "hospital_b") },
		},
		{
			name:       "fail - duplicate username and hospital",
			body:       map[string]string{"username": defaultUsername, "password": "password123", "hospital": defaultHospital},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "fail - password too short",
			body:       map[string]string{"username": "shortpass", "password": "123", "hospital": "hospital_a"},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "fail - missing username",
			body:       map[string]string{"password": "password123", "hospital": "hospital_a"},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "fail - missing hospital",
			body:       map[string]string{"username": "newstaff", "password": "password123"},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "fail - no token when staff already exist",
			body:       map[string]string{"username": "notoken", "password": "password123", "hospital": "hospital_a"},
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "fail - invalid token",
			body:       map[string]string{"username": "badtoken", "password": "password123", "hospital": "hospital_a"},
			token:      "invalid.token.here",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cleanup != nil {
				defer tc.cleanup()
			}
			w := post(r, "/staff/create", tc.body, tc.token)
			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d — body: %s", tc.wantStatus, w.Code, w.Body.String())
			}
			if tc.wantStatus == http.StatusCreated {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				if resp["username"] == "" {
					t.Errorf("expected username in response, got: %v", resp)
				}
			}
		})
	}
}

// ── /patient/create ───────────────────────────────────────────────────────────

func TestCreatePatient(t *testing.T) {
	db := setupTestDB(t)
	cleanupStaff(db, defaultUsername, "hospital_a")

	authService := services.NewAuthService(db, jwtSecret)
	authService.CreateStaff(services.CreateStaffInput{
		Username: defaultUsername, Password: defaultPassword, Hospital: "hospital_a",
	})
	defer cleanupStaff(db, defaultUsername, "hospital_a")

	token := getToken(t, setupLoginRouter(authService), defaultUsername, defaultPassword, "hospital_a")
	patientService := services.NewPatientService(db)
	r := setupPatientCreateRouter(authService, patientService)

	tests := []struct {
		name       string
		body       map[string]string
		token      string
		wantStatus int
		cleanup    func()
	}{
		{
			name: "success - all required fields",
			body: map[string]string{
				"first_name_en": "John", "last_name_en": "Doe",
				"date_of_birth": "1990-05-15", "gender": "male",
				"national_id": "1111111111111", "passport_id": "PP111111",
			},
			token:      token,
			wantStatus: http.StatusCreated,
			cleanup:    func() { cleanupPatient(db, "1111111111111", "PP111111") },
		},
		{
			name: "fail - missing first_name_en",
			body: map[string]string{
				"last_name_en": "Doe", "date_of_birth": "1990-05-15",
				"gender": "male", "national_id": "2222222222222", "passport_id": "PP222222",
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "fail - missing last_name_en",
			body: map[string]string{
				"first_name_en": "John", "date_of_birth": "1990-05-15",
				"gender": "male", "national_id": "3333333333333", "passport_id": "PP333333",
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "fail - missing date_of_birth",
			body: map[string]string{
				"first_name_en": "John", "last_name_en": "Doe",
				"gender": "male", "national_id": "4444444444444", "passport_id": "PP444444",
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "fail - missing gender",
			body: map[string]string{
				"first_name_en": "John", "last_name_en": "Doe",
				"date_of_birth": "1990-05-15", "national_id": "5555555555555", "passport_id": "PP555555",
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "fail - missing national_id",
			body: map[string]string{
				"first_name_en": "John", "last_name_en": "Doe",
				"date_of_birth": "1990-05-15", "gender": "male", "passport_id": "PP666666",
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "fail - missing passport_id",
			body: map[string]string{
				"first_name_en": "John", "last_name_en": "Doe",
				"date_of_birth": "1990-05-15", "gender": "male", "national_id": "7777777777777",
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "fail - no token",
			body:       map[string]string{},
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cleanup != nil {
				defer tc.cleanup()
			}
			w := post(r, "/patient/create", tc.body, tc.token)
			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d — body: %s", tc.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

// ── /patient/search ───────────────────────────────────────────────────────────

func TestSearchPatient(t *testing.T) {
	db := setupTestDB(t)

	cleanupStaff(db, defaultUsername, "hospital_a")
	cleanupStaff(db, "staff_b", "hospital_b")

	authService := services.NewAuthService(db, jwtSecret)
	authService.CreateStaff(services.CreateStaffInput{
		Username: defaultUsername, Password: defaultPassword, Hospital: "hospital_a",
	})
	authService.CreateStaff(services.CreateStaffInput{
		Username: "staff_b", Password: "password123", Hospital: "hospital_b",
	})
	defer cleanupStaff(db, defaultUsername, "hospital_a")
	defer cleanupStaff(db, "staff_b", "hospital_b")

	cleanupPatient(db, "1234567890123", "PP123456")
	patientService := services.NewPatientService(db)
	patientService.CreatePatient(services.CreatePatientInput{
		Hospital:    "hospital_a",
		FirstNameEn: "John",
		LastNameEn:  "Doe",
		DateOfBirth: "1990-05-15",
		Gender:      "male",
		NationalID:  "1234567890123",
		PassportID:  "PP123456",
	})
	defer cleanupPatient(db, "1234567890123", "PP123456")

	tokenA := getToken(t, setupLoginRouter(authService), defaultUsername, defaultPassword, "hospital_a")
	tokenB := getToken(t, setupLoginRouter(authService), "staff_b", "password123", "hospital_b")

	r := setupSearchRouter(authService, patientService)

	tests := []struct {
		name        string
		query       string
		token       string
		wantStatus  int
		wantResults int
	}{
		{
			name:        "success - search by national_id same hospital",
			query:       "/patient/search?national_id=1234567890123",
			token:       tokenA,
			wantStatus:  http.StatusOK,
			wantResults: 1,
		},
		{
			name:        "success - search by passport_id same hospital",
			query:       "/patient/search?passport_id=PP123456",
			token:       tokenA,
			wantStatus:  http.StatusOK,
			wantResults: 1,
		},
		{
			name:        "success - search by national_id different hospital returns empty",
			query:       "/patient/search?national_id=1234567890123",
			token:       tokenB,
			wantStatus:  http.StatusOK,
			wantResults: 0,
		},
		{
			name:        "success - search by passport_id different hospital returns empty",
			query:       "/patient/search?passport_id=PP123456",
			token:       tokenB,
			wantStatus:  http.StatusOK,
			wantResults: 0,
		},
		{
			name:        "success - national_id not found returns empty",
			query:       "/patient/search?national_id=0000000000000",
			token:       tokenA,
			wantStatus:  http.StatusOK,
			wantResults: 0,
		},
		{
			name:       "fail - no auth token",
			query:      "/patient/search?national_id=1234567890123",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "fail - invalid token",
			query:      "/patient/search?national_id=1234567890123",
			token:      "invalid.token.here",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "fail - search by patient_hn not allowed",
			query:      "/patient/search?patient_hn=HN001",
			token:      tokenA,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "fail - neither national_id nor passport_id provided",
			query:      "/patient/search",
			token:      tokenA,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "fail - both national_id and passport_id provided",
			query:      "/patient/search?national_id=1234567890123&passport_id=PP123456",
			token:      tokenA,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := get(r, tc.query, tc.token)
			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d — body: %s", tc.wantStatus, w.Code, w.Body.String())
			}

			if tc.wantResults >= 0 && tc.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				data, ok := resp["data"].([]interface{})
				if !ok {
					t.Errorf("expected data array in response, got: %v", resp)
					return
				}
				if len(data) != tc.wantResults {
					t.Errorf("expected %d results, got %d", tc.wantResults, len(data))
				}
			}
		})
	}
}
