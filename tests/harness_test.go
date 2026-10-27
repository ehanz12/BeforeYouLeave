package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/ehanz12/BeforeYouLeave/configs"
	"github.com/ehanz12/BeforeYouLeave/databases"
	"github.com/ehanz12/BeforeYouLeave/routes"
	"github.com/ehanz12/BeforeYouLeave/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// testTables urutan tidak penting karena FOREIGN_KEY_CHECKS dimatikan.
var testTables = []string{
	"reminder_item_checks",
	"reminders",
	"destination_schedules",
	"checklist_items",
	"destinations",
	"users",
}

var app *fiber.App

// TestMain memakai database terpisah (byl_test) supaya data dev tidak tersentuh.
// Env di-set sebelum LoadEnv karena godotenv.Load() tidak menimpa variabel yang sudah ada.
func TestMain(m *testing.M) {
	setOrFail("DB_NAME", "byl_test")
	setOrFail("DB_USER", "root")
	setOrFail("DB_PASSWORD", "678999")
	setOrFail("DB_HOST", "127.0.0.1")
	setOrFail("DB_PORT", "3306")
	setOrFail("APP_TIMEZONE", "Asia/Jakarta")
	setOrFail("REMINDER_HORIZON_DAYS", "14")
	setOrFail("REMINDER_GENERATE_INTERVAL_MIN", "60")
	setOrFail("REMINDER_DISPATCH_INTERVAL_MIN", "5")
	setOrFail("JWT_SECRET", "TEST_SECRET_ONLY")

	configs.LoadEnv()

	if err := databases.ConnectDB(); err != nil {
		panic("gagal connect DB test: " + err.Error())
	}
	// Matikan log SQL GORM supaya output test bisa dibaca.
	databases.DB = databases.DB.Session(&gorm.Session{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	// Tabel harus sudah ada dari hasil `migrate up` (lihat tests/README_TEST.md).
	// Di sini hanya dicek, tidak di-drop, supaya AUTO_INCREMENT tetap konsisten.
	for _, tbl := range testTables {
		if !databases.DB.Migrator().HasTable(tbl) {
			panic("tabel " + tbl + " belum ada, jalankan migrate up dulu")
		}
	}

	app = fiber.New()
	app.Use(cors.New(cors.Config{
		// Origin harus spesifik, bukan "*", kalau AllowCredentials=true.
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	routes.SetupRoute(app)

	code := m.Run()

	os.Exit(code)
}

func setOrFail(k, v string) {
	if err := os.Setenv(k, v); err != nil {
		panic(err)
	}
}

// resetDB mengosongkan semua tabel + reset auto_increment di antara test.
func resetDB(t *testing.T) {
	t.Helper()
	db := databases.DB
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		t.Fatal(err)
	}
	for _, tbl := range testTables {
		if err := db.Exec("TRUNCATE TABLE " + tbl).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------- request helper

type apiResp struct {
	Status int
	Body   map[string]interface{}
	Raw    []byte
}

func (r apiResp) data() map[string]interface{} {
	d, _ := r.Body["data"].(map[string]interface{})
	return d
}

func (r apiResp) list() []interface{} {
	d, _ := r.Body["data"].([]interface{})
	return d
}

func (r apiResp) errMsg() string {
	s, _ := r.Body["error"].(string)
	return s
}

func call(t *testing.T, method, path, token string, body interface{}) apiResp {
	t.Helper()
	var headers map[string]string
	if token != "" {
		headers = map[string]string{"Authorization": "Bearer " + token}
	}
	return callWithHeaders(t, method, path, headers, body)
}

// callWithHeaders dipakai untuk menguji kasus token yang tidak lolos format
// "Bearer <token>", yang tidak bisa dilakukan lewat call().
func callWithHeaders(t *testing.T, method, path string, headers map[string]string, body interface{}) apiResp {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	out := apiResp{Status: resp.StatusCode, Raw: raw}
	_ = json.Unmarshal(raw, &out.Body)
	return out
}

// wantStatus gagal cepat dengan pesan body supaya mudah di-debug.
func wantStatus(t *testing.T, r apiResp, want int, ctx string) {
	t.Helper()
	if r.Status != want {
		t.Fatalf("%s: status = %d, want %d | body: %s", ctx, r.Status, want, string(r.Raw))
	}
}

func idOf(t *testing.T, r apiResp, ctx string) uint64 {
	t.Helper()
	v, ok := r.data()["id"].(float64)
	if !ok {
		t.Fatalf("%s: tidak ada id di response | body: %s", ctx, string(r.Raw))
	}
	return uint64(v)
}

// ---------------------------------------------------------------- fixtures

const testPassword = "password123"

func makeUser(t *testing.T, name, email string) uint64 {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	user := map[string]interface{}{
		"name": name, "email": email, "password": string(hash),
	}
	if err := databases.DB.Table("users").Create(user).Error; err != nil {
		t.Fatal(err)
	}
	var id uint64
	if err := databases.DB.Table("users").Where("email = ?", email).Select("id").Scan(&id).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

func tokenFor(t *testing.T, email string) string {
	t.Helper()
	r := call(t, "POST", "/api/auth/login", "", map[string]string{
		"email": email, "password": testPassword,
	})
	wantStatus(t, r, 200, "login "+email)
	tok, ok := r.Body["access_token"].(string)
	if !ok {
		t.Fatalf("tidak ada access_token | body: %s", string(r.Raw))
	}
	return tok
}

// makeDestination lewat API supaya alur bisnis ikut teruji.
func makeDestination(t *testing.T, token, name string) uint64 {
	t.Helper()
	r := call(t, "POST", "/api/destination", token, map[string]interface{}{
		"name": name, "description": "dari test", "latitude": -6.2, "longitude": 106.8, "radius": 5,
	})
	wantStatus(t, r, 200, "create destination "+name)
	return idOf(t, r, "create destination "+name)
}

func makeItem(t *testing.T, token string, desID uint64, name string) uint64 {
	t.Helper()
	r := call(t, "POST", nested("/api/destination", desID, "/items"), token, map[string]interface{}{
		"name": name, "is_required": true,
	})
	wantStatus(t, r, 200, "create item "+name)
	return idOf(t, r, "create item "+name)
}

// nested = bangun path "/api/destination/5/items".
func nested(prefix string, id uint64, suffix string) string {
	return prefix + "/" + strconv.FormatUint(id, 10) + suffix
}

// countRows dipakai untuk assert langsung ke DB.
func countRows(t *testing.T, table string, where string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	q := databases.DB.Table(table)
	if where != "" {
		q = q.Where(where, args...)
	}
	if err := q.Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func db() *gorm.DB { return databases.DB }

func u64s(v uint64) string { return strconv.FormatUint(v, 10) }

// generateReminders memanggil worker generator langsung (bukan lewat goroutine),
// supaya test deterministik.
func generateReminders(t *testing.T) {
	t.Helper()
	if _, err := services.GenerateReminders(); err != nil {
		t.Fatalf("GenerateReminders: %v", err)
	}
}

func dispatchReminders(t *testing.T) int64 {
	t.Helper()
	n, err := services.DispatchDueReminders()
	if err != nil {
		t.Fatalf("DispatchDueReminders: %v", err)
	}
	return n
}

// firstReminderID = ambil satu reminder milik destination untuk diuji.
func firstReminderID(t *testing.T, destinationID uint64) uint64 {
	t.Helper()
	var id uint64
	err := db().Table("reminders").
		Where("destination_id = ?", destinationID).
		Order("remind_at ASC").Limit(1).Select("id").Scan(&id).Error
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// makeAllDaysSchedule = jadwal aktif untuk SETIAP hari (0..6) dengan H-1 23:00.
//
// Dipakai supaya jumlah reminder yang dihasilkan bisa diprediksi tanpa bergantung
// tanggal hari ini: offset 0 (hari ini) selalu ter-skip karena remind_at-nya
// sudah lewat, sedangkan hari-hari berikutnya selalu di masa depan.
func makeAllDaysSchedule(t *testing.T, token string, desID uint64) {
	t.Helper()
	r := call(t, "PUT", nested("/api/destination", desID, "/schedules"), token, map[string]interface{}{
		"schedules": allDaysPayload(1, "23:00"),
	})
	wantStatus(t, r, 200, "bulk replace 7 hari")
	if n := len(r.list()); n != 7 {
		t.Fatalf("jadwal tersimpan = %d, want 7", n)
	}
}

func allDaysPayload(remindDMinus int, remindTime string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, 7)
	for day := 0; day <= 6; day++ {
		out = append(out, map[string]interface{}{
			"day_of_week": day, "departure_time": "07:00",
			"remind_d_minus": remindDMinus, "remind_time": remindTime, "is_active": true,
		})
	}
	return out
}
