package tests

import (
	"testing"
)

func TestRegisterAndLogin(t *testing.T) {
	resetDB(t)

	body := map[string]string{
		"name": "Reihan", "email": "reihan@test.com",
		"password": testPassword, "phone": "0812345",
	}

	r := call(t, "POST", "/api/auth/register", "", body)
	wantStatus(t, r, 201, "register")
	if r.Body["success"] != true {
		t.Fatalf("register success != true: %s", r.Raw)
	}
	if countRows(t, "users", "email = ?", body["email"]) != 1 {
		t.Fatal("user tidak tersimpan")
	}

	// Password harus di-hash, jangan pernah plaintext.
	var password string
	if err := db().Table("users").Where("email = ?", body["email"]).Select("password").Scan(&password).Error; err != nil {
		t.Fatal(err)
	}
	if password == body["password"] {
		t.Fatal("password tersimpan sebagai plaintext !")
	}

	// Register ulang email yang sama harus ditolak.
	r = call(t, "POST", "/api/auth/register", "", body)
	wantStatus(t, r, 400, "register duplikat")

	// Field wajib kosong harus ditolak.
	r = call(t, "POST", "/api/auth/register", "", map[string]string{"name": "X"})
	wantStatus(t, r, 400, "register tanpa field wajib")
}

func TestLogin(t *testing.T) {
	resetDB(t)
	makeUser(t, "Reihan", "reihan@test.com")

	t.Run("berhasil", func(t *testing.T) {
		r := call(t, "POST", "/api/auth/login", "", map[string]string{
			"email": "reihan@test.com", "password": testPassword,
		})
		wantStatus(t, r, 200, "login")
		if tok, ok := r.Body["access_token"].(string); !ok || tok == "" {
			t.Fatalf("tidak dapat token: %s", r.Raw)
		}
	})

	t.Run("password salah", func(t *testing.T) {
		r := call(t, "POST", "/api/auth/login", "", map[string]string{
			"email": "reihan@test.com", "password": "salahpassword",
		})
		wantStatus(t, r, 400, "login password salah")
	})

	t.Run("email tidak terdaftar", func(t *testing.T) {
		r := call(t, "POST", "/api/auth/login", "", map[string]string{
			"email": "nope@test.com", "password": testPassword,
		})
		wantStatus(t, r, 400, "email tidak terdaftar")
	})
}

func TestProtectedRouteRejectsBadToken(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	token := tokenFor(t, "a@test.com")

	cases := []struct {
		name   string
		header map[string]string
		want   int
	}{
		{"tanpa header", nil, 401},
		{"tanpa prefix Bearer", map[string]string{"Authorization": token}, 401},
		{"token ngawur", map[string]string{"Authorization": "Bearer abc.def.ghi"}, 401},
		{"token valid", map[string]string{"Authorization": "Bearer " + token}, 200},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := callWithHeaders(t, "GET", "/api/destination", c.header, nil)
			if r.Status != c.want {
				t.Errorf("status = %d, want %d | body: %s", r.Status, c.want, r.Raw)
			}
		})
	}
}

func TestMe(t *testing.T) {
	resetDB(t)
	userID := makeUser(t, "Reihan", "reihan@test.com")
	token := tokenFor(t, "reihan@test.com")

	r := call(t, "GET", "/api/auth/me", token, nil)
	wantStatus(t, r, 200, "me")

	data, _ := r.Body["data"].(map[string]interface{})
	if data == nil {
		t.Fatalf("tidak ada data: %s", r.Raw)
	}
	if got := uint64(data["id"].(float64)); got != userID {
		t.Errorf("id = %d, want %d", got, userID)
	}
	// Password tidak boleh bocor lewat API.
	if _, leaked := data["password"]; leaked {
		t.Error("password bocor di response /auth/me !")
	}
}

// TestCrossUserIsolation = token user B tidak boleh menyentuh data user A
// lewat endpoint mana pun, termasuk yang nested dan reminder.
func TestCrossUserIsolation(t *testing.T) {
	resetDB(t)
	makeUser(t, "A", "a@test.com")
	makeUser(t, "B", "b@test.com")
	tokenA := tokenFor(t, "a@test.com")
	tokenB := tokenFor(t, "b@test.com")

	desA := makeDestination(t, tokenA, "Sekolah A")
	itemA := makeItem(t, tokenA, desA, "Buku A")

	r := call(t, "POST", nested("/api/destination", desA, "/schedules"), tokenA, map[string]interface{}{
		"day_of_week": 1, "departure_time": "07:00", "remind_time": "20:00", "remind_d_minus": 1,
	})
	wantStatus(t, r, 200, "create schedule A")
	schA := idOf(t, r, "schedule A")

	generateReminders(t)
	reminderA := firstReminderID(t, desA)
	if reminderA == 0 {
		t.Fatal("tidak ada reminder yang ter-generate, test tidak bermakna")
	}

	desPath := nested("/api/destination", desA, "")
	attempts := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"baca destinasi", "GET", desPath, nil},
		{"ubah destinasi", "PATCH", desPath, map[string]string{"name": "Dibajak"}},
		{"hapus destinasi", "DELETE", desPath, nil},
		{"list item", "GET", nested("/api/destination", desA, "/items"), nil},
		{"baca item", "GET", nested("/api/destination", desA, "/items") + "/" + u64s(itemA), nil},
		{"ubah item", "PATCH", nested("/api/destination", desA, "/items") + "/" + u64s(itemA), map[string]string{"name": "Dibajak"}},
		{"hapus item", "DELETE", nested("/api/destination", desA, "/items") + "/" + u64s(itemA), nil},
		{"list jadwal", "GET", nested("/api/destination", desA, "/schedules"), nil},
		{"baca jadwal", "GET", nested("/api/destination", desA, "/schedules") + "/" + u64s(schA), nil},
		{"ubah jadwal", "PATCH", nested("/api/destination", desA, "/schedules") + "/" + u64s(schA), map[string]bool{"is_active": false}},
		{"hapus jadwal", "DELETE", nested("/api/destination", desA, "/schedules") + "/" + u64s(schA), nil},
		{"baca reminder", "GET", "/api/reminder/" + u64s(reminderA), nil},
		{"tandai read", "PATCH", "/api/reminder/" + u64s(reminderA) + "/read", nil},
		{"skip reminder", "PATCH", "/api/reminder/" + u64s(reminderA) + "/skip", nil},
		{"centang item", "PATCH", "/api/reminder/" + u64s(reminderA) + "/items/" + u64s(itemA), map[string]bool{"is_checked": true}},
	}
	for _, a := range attempts {
		t.Run(a.name, func(t *testing.T) {
			r := call(t, a.method, a.path, tokenB, a.body)
			if r.Status == 200 {
				t.Errorf("token B=%s %s -> 200, harusnya ditolak | %s", a.method, a.path, r.Raw)
			}
		})
	}

	// Data A harus utuh setelah semua percobaan.
	if countRows(t, "destinations", "id = ?", desA) != 1 {
		t.Error("destinasi A terhapus oleh user B !")
	}
	if countRows(t, "checklist_items", "id = ?", itemA) != 1 {
		t.Error("item A terhapus oleh user B !")
	}
	if countRows(t, "destination_schedules", "id = ?", schA) != 1 {
		t.Error("jadwal A terhapus oleh user B !")
	}

	// List user B harus benar-benar kosong (tidak boleh ada kebocoran).
	r = call(t, "GET", "/api/destination", tokenB, nil)
	wantStatus(t, r, 200, "list destination B")
	if n := len(r.list()); n != 0 {
		t.Errorf("user B melihat %d destinasi, want 0", n)
	}
	r = call(t, "GET", "/api/reminder", tokenB, nil)
	wantStatus(t, r, 200, "list reminder B")
	if n := len(r.list()); n != 0 {
		t.Errorf("user B melihat %d reminder, want 0", n)
	}
}
