package repository_test

import (
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/crisbusta/proindustrial-backend-public/internal/repository"
)

// Regresión: una solicitud anterior a la migración 011, con una descripción
// más larga que el tope canónico, quedaba imposible de resolver. El CHECK que
// 011 agregó como NOT VALID se evalúa igual sobre la fila completa en cada
// UPDATE, así que tanto el rechazo como la aprobación chocaban con
// provider_registrations_description_len sin tocar description.
//
// Requiere una base desechable:
//
//	TEST_DATABASE_URL=postgres://... go test ./internal/repository/ -run LegacyLongDescription
func TestLegacyLongDescriptionRegistrations(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no definida; se omite el test de integración")
	}

	m, err := migrate.New("file://../../migrations", url)
	if err != nil {
		t.Fatalf("migrator: %v", err)
	}
	// Estado histórico: hasta 010, antes de que existieran los CHECK de longitud.
	if err := m.Migrate(10); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate a 010: %v", err)
	}

	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	longDesc := strings.Repeat("á", 700) // 700 caracteres, sobre el tope de 600
	insert := func(email string) string {
		var id string
		err := db.Get(&id, `
			INSERT INTO provider_registrations (company_name, email, phone, region, services, description, status)
			VALUES ('Legacy SpA', $1, '+56900000000', 'Arica y Parinacota', '{termofusion}', $2, 'pending')
			RETURNING id`, email, longDesc)
		if err != nil {
			t.Fatalf("insert fila histórica %s: %v", email, err)
		}
		return id
	}
	toReject := insert("legacy-reject@example.com")
	toApprove := insert("legacy-approve@example.com")

	// Resto de las migraciones, incluidas las de endurecimiento y el backfill.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	repo := repository.NewAdminRepo(db)

	t.Run("rechazo", func(t *testing.T) {
		reg, err := repo.RejectRegistration(toReject, "descripción muy larga, sugerimos no más de 600 caracteres")
		if err != nil {
			t.Fatalf("RejectRegistration devolvió error: %v", err)
		}
		if reg.Status != "rejected" {
			t.Fatalf("status = %q, se esperaba \"rejected\"", reg.Status)
		}
	})

	t.Run("aprobación", func(t *testing.T) {
		res, err := repo.ApproveRegistration(toApprove, "$2a$10$abcdefghijklmnopqrstuv", "clave-inicial")
		if err != nil {
			t.Fatalf("ApproveRegistration devolvió error: %v", err)
		}
		if res.Registration.Status != "approved" {
			t.Fatalf("status = %q, se esperaba \"approved\"", res.Registration.Status)
		}
	})

	t.Run("sin filas por sobre el tope", func(t *testing.T) {
		var n int
		if err := db.Get(&n, `
			SELECT (SELECT count(*) FROM provider_registrations WHERE char_length(description) > 600)
			     + (SELECT count(*) FROM companies              WHERE char_length(description) > 600)`); err != nil {
			t.Fatalf("consulta de longitudes: %v", err)
		}
		if n != 0 {
			t.Fatalf("%d filas siguen por sobre el tope de 600 caracteres", n)
		}
	})
}
