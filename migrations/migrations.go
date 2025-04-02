package migrations

import (
	"log"
	"os/exec"
)

func RunMigrations(dsn string) error {
	cmd := exec.Command("goose", "-dir", "migrations", "postgres", dsn, "up")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Migration failed: %s\n", output)
		return err
	}
	log.Println("Migrations applied successfully")
	return nil
}