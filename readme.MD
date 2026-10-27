# Before You Leave
aplikasi untuk pengingat barang ketika ingin kemanapun

## Problem
biasanya orang orang lupa akan membuat barang penting ketika berpegian nah ini aplikasi untuk menyelesaikan masalah itu sendiri

## Solution
yaitu membuat aplikasi pengigat dan bukan hanya pengingat namun bisa menjadi pembiasaan untuk mengurangi kejadian tersebut

## Target Development
1. Perencanaan (done)
2. membuat struktur file (done)
3. connect ke database (done)
4. membuat model (done)
5. auth + middleware (done)
6. CRUD destination (done)
7. CRUD checklist item (done)
8. CRUD jadwal mingguan (done)
9. worker reminder + centangan per reminder (done)

## Core Features
1. User membuat tempat
2. User membuat checklist
3. User bisa melihat checklist
4. User bisa mencentang item
5. App memberikan reminder

## Konsep Penting
- `destinations` = tempat tujuan (misal Sekolah). `is_active=false` = libur total.
- `destination_schedules` = 1 baris = 1 hari aktif. Sekolah Senin-Jumat = 5 baris, tidak ada baris Minggu = tidak ada reminder Minggu.
- `checklist_items` = TEMPLATE barang, tidak berubah per tanggal.
- `reminders` = instance per tanggal, dibangkitkan worker dari jadwal. Status: `pending` -> `sent` -> `read`, plus `skipped` untuk libur.
- `reminder_item_checks` = snapshot centangan per tanggal (histori pembiasaan).

## Endpoint
Semua endpoint di bawah butuh header `Authorization: Bearer <token>`, kecuali auth.

```
# Auth
POST   /api/auth/register
POST   /api/auth/login
GET    /api/auth/me

# Destinasi
POST   /api/destination
GET    /api/destination?search=&is_active=&start_date=&end_date=&page=&limit=
GET    /api/destination/:id
PATCH  /api/destination/:id
DELETE /api/destination/:id

# Checklist item (nested di bawah destination)
POST   /api/destination/:destinationId/items
GET    /api/destination/:destinationId/items?search=&is_required=&page=&limit=
GET    /api/destination/:destinationId/items/:itemId
PATCH  /api/destination/:destinationId/items/:itemId
DELETE /api/destination/:destinationId/items/:itemId

# Jadwal mingguan
POST   /api/destination/:destinationId/schedules
GET    /api/destination/:destinationId/schedules?is_active=&page=&limit=
PUT    /api/destination/:destinationId/schedules
GET    /api/destination/:destinationId/schedules/:scheduleId
PATCH  /api/destination/:destinationId/schedules/:scheduleId
DELETE /api/destination/:destinationId/schedules/:scheduleId

# Reminder
GET    /api/reminder?status=&destination_id=&unread_only=&start_date=&end_date=&page=&limit=
GET    /api/reminder/:reminderId
PATCH  /api/reminder/:reminderId/read
PATCH  /api/reminder/:reminderId/skip
PATCH  /api/reminder/:reminderId/items/:checklistItemId
```

`PUT /api/destination/:destinationId/schedules` menerima array dan replace-all dalam satu
transaksi, dipakai form FE yang mencentang hari aktif sekaligus (misal Senin-Jumat).

`day_of_week`: 0=Minggu, 1=Senin, ... 6=Sabtu.
`remind_d_minus`: 1=H-1 malam (default), 0=hari-H. Untuk `0`, `remind_time` harus lebih
awal dari `departure_time`, kalau tidak reminder bunyi setelah user berangkat.

## Worker Reminder
Dijalankan otomatis dari `main.go`, dua loop terpisah:
- **generator** (default tiap 60 menit): membuat reminder ke-N hari ke depan dari jadwal aktif.
- **dispatcher** (default tiap 5 menit): `pending` -> `sent` untuk yang sudah lewat waktunya.

Semua bisa diubah lewat `.env`, lihat `.env.example`:
`APP_TIMEZONE`, `REMINDER_HORIZON_DAYS`, `REMINDER_GENERATE_INTERVAL_MIN`, `REMINDER_DISPATCH_INTERVAL_MIN`.

GenerateReminders aman dipanggil berulang karena `uq_reminders_schedule_for_date` mencegah
reminder dobel untuk jadwal + tanggal yang sama.

## Future Features
1. GPS











================================================================================================================= To Install Package Migration : go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

================================================================================================================= To Create Migration : migrate create -ext sql -dir databases/migrations -seq create_examples_table

================================================================================================================= To Running (UP) Migration (Example Mysql) :  

================================================================================================================== To Running (DOWN) Migration : migrate -database "mysql://db_user:db_password@tcp(127.0.0.1:3306)/db_name" -path database/migration down

To Init Swag : swag init To Swag Generate : swag init --parseDependency --parseInternal
