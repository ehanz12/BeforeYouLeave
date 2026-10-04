# Before You Leave
aplikasi untuk pengingat barang ketika ingin kemanapun

## Problem
biasanya orang orang lupa akan membuat barang penting ketika berpegian nah ini aplikasi untuk menyelesaikan masalah itu sendiri

## Solution
yaitu membuat aplikasi pengigat dan bukan hanya pengingat namun bisa menjadi pembiasaan untuk mengurangi kejadian tersebut

## Target Development
1. Perencanaan (done)
2. membuat struktur file (done)
4. connect ke database (done)
5. membuat model
6. membuat handlers
5. route awal
6. auth
7. CRUD lainnya...

## Core Features
1. User membuat tempat
2. User membuat checklist
3. User bisa melihat checklist
4. User bisa mencentang item
5. App menyimpan data secara lokal
6. App memberikan reminder

## Future Features
1. GPS











================================================================================================================= To Install Package Migration : go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

================================================================================================================= To Create Migration : migrate create -ext sql -dir databases/migrations -seq create_examples_table

================================================================================================================= To Running (UP) Migration (Example Mysql) : migrate -database "mysql://db_user:db_password@tcp(127.0.0.1:3306)/db_name" -path database/migration up

================================================================================================================== To Running (DOWN) Migration : migrate -database "mysql://db_user:db_password@tcp(127.0.0.1:3306)/db_name" -path database/migration down

To Init Swag : swag init To Swag Generate : swag init --parseDependency --parseInternal