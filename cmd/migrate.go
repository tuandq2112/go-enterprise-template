package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/internal/infrastructure/database"
	"go-clean-ddd-es-template/pkg/logger"
	"go-clean-ddd-es-template/pkg/migrations"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run database migrations for write and event databases`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("up")
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback all migrations",
	Run: func(cmd *cobra.Command, args []string) {
		runMigrations("down")
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show current migration version",
	Run: func(cmd *cobra.Command, args []string) {
		showMigrationVersion()
	},
}

var migrateForceCmd = &cobra.Command{
	Use:   "force [version]",
	Short: "Force migration version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		forceMigrationVersion(version)
	},
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createMigration(args[0])
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateVersionCmd)
	migrateCmd.AddCommand(migrateForceCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runMigrations(action string) {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger, err := logger.NewLoggerFromConfig(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Create database connections
	writeDB, err := database.NewPostgresConnection(cfg.WriteDatabase)
	if err != nil {
		logger.Fatal("Failed to connect to write database", zap.Error(err))
	}
	defer writeDB.Close()

	eventDB, err := database.NewPostgresConnection(cfg.EventDatabase)
	if err != nil {
		logger.Fatal("Failed to connect to event database", zap.Error(err))
	}
	defer eventDB.Close()

	// Create migration manager
	migrationManager, err := migrations.NewMigrationManager(
		writeDB,
		eventDB,
		"./migrations/write",
		"./migrations/event",
	)
	if err != nil {
		logger.Fatal("Failed to create migration manager", zap.Error(err))
	}
	defer migrationManager.Close()

	ctx := context.Background()

	// Initialize migration systems
	if err := migrationManager.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize migrations", zap.Error(err))
	}

	switch action {
	case "up":
		logger.Info("Running write database migrations...")
		if err := migrationManager.RunWriteDBMigrations(ctx); err != nil {
			logger.Fatal("Failed to run write database migrations", zap.Error(err))
		}
		logger.Info("Write database migrations completed")

		logger.Info("Running event database migrations...")
		if err := migrationManager.RunEventDBMigrations(ctx); err != nil {
			logger.Fatal("Failed to run event database migrations", zap.Error(err))
		}
		logger.Info("Event database migrations completed")

	case "down":
		logger.Info("Rolling back event database migrations...")
		if err := migrationManager.EventDBMigrator.Down(ctx); err != nil {
			logger.Fatal("Failed to rollback event database migrations", zap.Error(err))
		}
		logger.Info("Event database migrations rolled back")

		logger.Info("Rolling back write database migrations...")
		if err := migrationManager.WriteDBMigrator.Down(ctx); err != nil {
			logger.Fatal("Failed to rollback write database migrations", zap.Error(err))
		}
		logger.Info("Write database migrations rolled back")
	}
}

func showMigrationVersion() {
	// Load configuration
	cfg := config.Load()

	// Create database connections
	writeDB, err := database.NewPostgresConnection(cfg.WriteDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to write database: %v", err)
	}
	defer writeDB.Close()

	eventDB, err := database.NewPostgresConnection(cfg.EventDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to event database: %v", err)
	}
	defer eventDB.Close()

	// Create migration manager
	migrationManager, err := migrations.NewMigrationManager(
		writeDB,
		eventDB,
		"./migrations/write",
		"./migrations/event",
	)
	if err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	defer migrationManager.Close()

	ctx := context.Background()

	// Get versions
	writeVersion, writeDirty, err := migrationManager.GetWriteDBVersion(ctx)
	if err != nil {
		log.Printf("Failed to get write database version: %v", err)
	} else {
		fmt.Printf("Write Database Version: %d (dirty: %t)\n", writeVersion, writeDirty)
	}

	eventVersion, eventDirty, err := migrationManager.GetEventDBVersion(ctx)
	if err != nil {
		log.Printf("Failed to get event database version: %v", err)
	} else {
		fmt.Printf("Event Database Version: %d (dirty: %t)\n", eventVersion, eventDirty)
	}
}

func forceMigrationVersion(version int) {
	// Load configuration
	cfg := config.Load()

	// Create database connections
	writeDB, err := database.NewPostgresConnection(cfg.WriteDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to write database: %v", err)
	}
	defer writeDB.Close()

	eventDB, err := database.NewPostgresConnection(cfg.EventDatabase)
	if err != nil {
		log.Fatalf("Failed to connect to event database: %v", err)
	}
	defer eventDB.Close()

	// Create migration manager
	migrationManager, err := migrations.NewMigrationManager(
		writeDB,
		eventDB,
		"./migrations/write",
		"./migrations/event",
	)
	if err != nil {
		log.Fatalf("Failed to create migration manager: %v", err)
	}
	defer migrationManager.Close()

	ctx := context.Background()

	// Force version for both databases
	if err := migrationManager.WriteDBMigrator.Force(ctx, version); err != nil {
		log.Fatalf("Failed to force write database version: %v", err)
	}
	fmt.Printf("Write database version forced to: %d\n", version)

	if err := migrationManager.EventDBMigrator.Force(ctx, version); err != nil {
		log.Fatalf("Failed to force event database version: %v", err)
	}
	fmt.Printf("Event database version forced to: %d\n", version)
}

func createMigration(name string) {
	// Create write database migration
	if err := migrations.CreateMigrationFile("./migrations/write", name); err != nil {
		log.Fatalf("Failed to create write database migration: %v", err)
	}

	// Create event database migration
	if err := migrations.CreateMigrationFile("./migrations/event", name); err != nil {
		log.Fatalf("Failed to create event database migration: %v", err)
	}

	fmt.Printf("Created migration files for: %s\n", name)
}
