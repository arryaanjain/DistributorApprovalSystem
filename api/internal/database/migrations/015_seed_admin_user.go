package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(Up015, Down015)
}

func Up015(ctx context.Context, tx *sql.Tx) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		return fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD environment variables are required to seed admin user")
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO users (name, email, password_hash, role, is_active)
		VALUES ('Kresconet Super Admin', $1, crypt($2, gen_salt('bf')), 'super_admin', TRUE)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			role = 'super_admin',
			is_active = TRUE;
	`, adminEmail, adminPassword)
	if err != nil {
		return fmt.Errorf("seeding admin user: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO products (sku, name, description, category, price_paise, moq, is_active, is_sample, is_regular)
		VALUES
			('SMP-OIL-01', 'Sample Pack: Edible Oils (3x500ml)', 'Sample kit of Sunflower, Mustard, and Groundnut oils', 'Edible Oils', 49900, 1, TRUE, TRUE, FALSE),
			('SMP-STP-01', 'Sample Pack: Premium Staples & Grains', 'Trial sample bundle of Chakki Atta, Sugar, and Basmati Rice', 'Staples', 29900, 1, TRUE, TRUE, FALSE)
		ON CONFLICT DO NOTHING;

		UPDATE products SET is_regular = TRUE WHERE is_sample = FALSE;
	`)
	return err
}

func Down015(ctx context.Context, tx *sql.Tx) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail != "" {
		_, err := tx.ExecContext(ctx, "DELETE FROM users WHERE email = $1", adminEmail)
		if err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, "DELETE FROM products WHERE sku IN ('SMP-OIL-01', 'SMP-STP-01')")
	return err
}
