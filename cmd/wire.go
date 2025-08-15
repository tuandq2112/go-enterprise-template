//go:build wireinject
// +build wireinject

package cmd

import (
	"go-clean-ddd-es-template/internal/application/commands"
	"go-clean-ddd-es-template/internal/application/queries"
	"go-clean-ddd-es-template/internal/application/services"
	"go-clean-ddd-es-template/internal/domain/repositories"
	"go-clean-ddd-es-template/internal/infrastructure/config"
	"go-clean-ddd-es-template/internal/infrastructure/consumers"
	"go-clean-ddd-es-template/internal/infrastructure/database"
	"go-clean-ddd-es-template/internal/infrastructure/grpc"
	"go-clean-ddd-es-template/internal/infrastructure/messagebroker"
	infraRepos "go-clean-ddd-es-template/internal/infrastructure/repositories"
	"go-clean-ddd-es-template/pkg/auth"
	"go-clean-ddd-es-template/pkg/i18n"
	"go-clean-ddd-es-template/pkg/logger"
	"go-clean-ddd-es-template/pkg/middleware"
	"go-clean-ddd-es-template/pkg/tracing"
	"time"

	"github.com/google/wire"
)

// Type aliases to distinguish between different database types
type (
	WriteDatabase database.Database
	ReadDatabase  database.Database
	EventDatabase database.Database
)

// provideConfig provides application configuration
func provideConfig() *config.Config {
	return config.Load()
}

// provideTracer provides tracing service
func provideTracer(cfg *config.Config) (*tracing.Tracer, error) {
	if !cfg.Tracing.Enabled {
		return nil, nil
	}
	return tracing.NewTracer(cfg.Tracing.ServiceName, "1.0.0", cfg.Tracing.Endpoint)
}

// provideLogger provides logger service
func provideLogger(cfg *config.Config) (logger.Logger, error) {
	return logger.NewLoggerFromConfig(cfg.Log.Level, cfg.Log.Format)
}

// provideTranslator provides i18n translator
func provideTranslator(cfg *config.Config) (*i18n.Translator, error) {
	translator := i18n.NewTranslator(cfg.I18n.DefaultLocale)

	// Load translations from the translations directory
	if err := translator.LoadTranslations(cfg.I18n.TranslationsDir); err != nil {
		return nil, err
	}

	// Set as global translator
	i18n.SetGlobalTranslator(translator)

	return translator, nil
}

// provideErrorHandler provides error handler with i18n support
func provideErrorHandler(translator *i18n.Translator, logger logger.Logger) *middleware.ErrorHandler {
	return middleware.NewErrorHandler(translator, logger)
}

// provideDatabaseFactory provides database factory
func provideDatabaseFactory() *database.DatabaseFactory {
	return database.NewDatabaseFactory()
}

// provideWriteDatabase provides write database connection
func provideWriteDatabase(factory *database.DatabaseFactory, cfg *config.Config) (WriteDatabase, error) {
	db, err := factory.CreateDatabase(&cfg.WriteDatabase)
	return WriteDatabase(db), err
}

// provideReadDatabase provides read database connection
func provideReadDatabase(factory *database.DatabaseFactory, cfg *config.Config) (ReadDatabase, error) {
	db, err := factory.CreateDatabase(&cfg.ReadDatabase)
	return ReadDatabase(db), err
}

// provideEventDatabase provides event database connection
func provideEventDatabase(factory *database.DatabaseFactory, cfg *config.Config) (EventDatabase, error) {
	db, err := factory.CreateDatabase(&cfg.EventDatabase)
	return EventDatabase(db), err
}

// provideRepositoryFactory provides repository factory
func provideRepositoryFactory(
	writeDB WriteDatabase,
	readDB ReadDatabase,
	eventDB EventDatabase,
	cfg *config.Config,
) *infraRepos.RepositoryFactory {
	return infraRepos.NewRepositoryFactory(database.Database(writeDB), database.Database(readDB), database.Database(eventDB), cfg)
}

// provideMessageBrokerFactory provides message broker factory
func provideMessageBrokerFactory() *messagebroker.MessageBrokerFactory {
	return messagebroker.NewMessageBrokerFactory()
}

// provideMessageBroker provides message broker using factory
func provideMessageBroker(factory *messagebroker.MessageBrokerFactory, cfg *config.Config) (messagebroker.MessageBroker, error) {
	return factory.CreateMessageBroker(&cfg.MessageBroker)
}

// provideUserEventHandler provides user event handler
func provideUserEventHandler(readRepository repositories.UserReadRepository) *consumers.UserEventHandler {
	return consumers.NewUserEventHandler(readRepository)
}

// provideProductEventHandler provides product event handler
func provideProductEventHandler() *consumers.ProductEventHandler {
	return consumers.NewProductEventHandler()
}

// provideEventConsumer provides generic event consumer with multiple handlers
func provideEventConsumer(
	broker messagebroker.MessageBroker,
	userEventHandler *consumers.UserEventHandler,
	productEventHandler *consumers.ProductEventHandler,
	cfg *config.Config,
) *consumers.EventConsumer {
	consumer := broker.GetConsumer()
	topics := []string{"user.created", "user.updated", "user.deleted", "product.created", "product.updated", "product.deleted"}

	eventConsumer := consumers.NewEventConsumer(consumer, cfg.MessageBroker.GroupID, topics)

	// Register user event handlers
	eventConsumer.RegisterEventHandler("user.created", userEventHandler)
	eventConsumer.RegisterEventHandler("user.updated", userEventHandler)
	eventConsumer.RegisterEventHandler("user.deleted", userEventHandler)

	// Register product event handlers
	eventConsumer.RegisterEventHandler("product.created", productEventHandler)
	eventConsumer.RegisterEventHandler("product.updated", productEventHandler)
	eventConsumer.RegisterEventHandler("product.deleted", productEventHandler)

	return eventConsumer
}

// provideUserWriteRepository provides user write repository
func provideUserWriteRepository(factory *infraRepos.RepositoryFactory) (repositories.UserWriteRepository, error) {
	return factory.CreateUserWriteRepository()
}

// provideUserReadRepository provides user read repository
func provideUserReadRepository(factory *infraRepos.RepositoryFactory) (repositories.UserReadRepository, error) {
	return factory.CreateUserReadRepository()
}

// provideUserRepository provides user repository (combines write and read)
func provideUserRepository(writeRepo repositories.UserWriteRepository, readRepo repositories.UserReadRepository) repositories.UserRepository {
	// For now, we'll use writeRepo as the main repository since it has all the methods
	// In a real implementation, you might want to create a composite repository
	return writeRepo.(repositories.UserRepository)
}

// provideEventStore provides event store
func provideEventStore(factory *infraRepos.RepositoryFactory) (repositories.EventStore, error) {
	return factory.CreateEventStore()
}

// provideEventPublisher provides event publisher
func provideEventPublisher(broker messagebroker.MessageBroker) repositories.EventPublisher {
	return infraRepos.NewMessageBrokerEventPublisher(broker)
}

// Command Handlers (Write Operations)
func provideUserCreateCommandHandler(
	userWriteRepo repositories.UserWriteRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
) *commands.UserCreateCommandHandler {
	return commands.NewUserCreateCommandHandler(userWriteRepo, eventStore, eventPublisher)
}

func provideUserUpdateCommandHandler(
	userWriteRepo repositories.UserWriteRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
) *commands.UserUpdateCommandHandler {
	return commands.NewUserUpdateCommandHandler(userWriteRepo, eventStore, eventPublisher)
}

func provideUserDeleteCommandHandler(
	userWriteRepo repositories.UserWriteRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
) *commands.UserDeleteCommandHandler {
	return commands.NewUserDeleteCommandHandler(userWriteRepo, eventStore, eventPublisher)
}

// Query Handlers (Read Operations)
func provideUserGetQueryHandler(userReadRepository repositories.UserReadRepository) *queries.UserGetQueryHandler {
	return queries.NewUserGetQueryHandler(userReadRepository)
}

func provideUserListQueryHandler(userReadRepository repositories.UserReadRepository) *queries.UserListQueryHandler {
	return queries.NewUserListQueryHandler(userReadRepository)
}

func provideUserGetByEmailQueryHandler(userReadRepository repositories.UserReadRepository) *queries.UserGetByEmailQueryHandler {
	return queries.NewUserGetByEmailQueryHandler(userReadRepository)
}

func provideUserEventsQueryHandler(userReadRepository repositories.UserReadRepository) *queries.UserEventsQueryHandler {
	return queries.NewUserEventsQueryHandler(userReadRepository)
}

// provideUserService provides user service
func provideUserService(
	createCommandHandler *commands.UserCreateCommandHandler,
	updateCommandHandler *commands.UserUpdateCommandHandler,
	deleteCommandHandler *commands.UserDeleteCommandHandler,
	getQueryHandler *queries.UserGetQueryHandler,
	listQueryHandler *queries.UserListQueryHandler,
	getByEmailQueryHandler *queries.UserGetByEmailQueryHandler,
	eventsQueryHandler *queries.UserEventsQueryHandler,
) *services.UserService {
	return services.NewUserService(
		createCommandHandler,
		updateCommandHandler,
		deleteCommandHandler,
		getQueryHandler,
		listQueryHandler,
		getByEmailQueryHandler,
		eventsQueryHandler,
	)
}

// provideJWTService provides JWT service
func provideJWTService(cfg *config.Config) (*auth.JWTService, error) {
	return auth.NewJWTService(cfg.Auth.PrivateKeyPath, cfg.Auth.PublicKeyPath, time.Duration(cfg.Auth.TokenExpiry)*time.Hour)
}

// providePasswordService provides password service
func providePasswordService() *auth.PasswordService {
	return auth.NewPasswordService(12) // bcrypt cost of 12
}

// provideAuthRegisterCommandHandler provides auth register command handler
func provideAuthRegisterCommandHandler(
	userRepo repositories.UserRepository,
	eventStore repositories.EventStore,
	eventPublisher repositories.EventPublisher,
	passwordService *auth.PasswordService,
	jwtService *auth.JWTService,
) *commands.AuthRegisterCommandHandler {
	return commands.NewAuthRegisterCommandHandler(userRepo, eventStore, eventPublisher, passwordService, jwtService)
}

// provideAuthLoginCommandHandler provides auth login command handler
func provideAuthLoginCommandHandler(
	userRepo repositories.UserRepository,
	passwordService *auth.PasswordService,
	jwtService *auth.JWTService,
) *commands.AuthLoginCommandHandler {
	return commands.NewAuthLoginCommandHandler(userRepo, passwordService, jwtService)
}

// provideAuthService provides auth service
func provideAuthService(
	registerHandler *commands.AuthRegisterCommandHandler,
	loginHandler *commands.AuthLoginCommandHandler,
	jwtService *auth.JWTService,
) *services.AuthService {
	return services.NewAuthService(registerHandler, loginHandler, jwtService)
}

// provideGRPCServer provides gRPC server
func provideGRPCServer(
	userService *services.UserService,
	authService *services.AuthService,
	tracer *tracing.Tracer,
	logger logger.Logger,
) *grpc.GRPCServer {
	return grpc.NewGRPCServer(userService, authService, tracer, logger)
}

// InitializeGRPCServer initializes gRPC server with all dependencies
func InitializeGRPCServer() (*grpc.GRPCServer, error) {
	wire.Build(
		provideConfig,
		provideTracer,
		provideLogger,
		provideDatabaseFactory,
		provideWriteDatabase,
		provideReadDatabase,
		provideEventDatabase,
		provideMessageBrokerFactory,
		provideMessageBroker,
		provideRepositoryFactory,
		provideUserWriteRepository,
		provideUserReadRepository,
		provideUserRepository,
		provideEventStore,
		provideEventPublisher,
		// Command Handlers (Write Operations)
		provideUserCreateCommandHandler,
		provideUserUpdateCommandHandler,
		provideUserDeleteCommandHandler,
		// Query Handlers (Read Operations)
		provideUserGetQueryHandler,
		provideUserListQueryHandler,
		provideUserGetByEmailQueryHandler,
		provideUserEventsQueryHandler,
		// Services
		provideUserService,
		provideJWTService,
		providePasswordService,
		provideAuthRegisterCommandHandler,
		provideAuthLoginCommandHandler,
		provideAuthService,
		provideGRPCServer,
	)
	return &grpc.GRPCServer{}, nil
}

// InitializeEventConsumer initializes event consumer with all dependencies
func InitializeEventConsumer() (*consumers.EventConsumer, error) {
	wire.Build(
		provideConfig,
		provideDatabaseFactory,
		provideWriteDatabase,
		provideReadDatabase,
		provideEventDatabase,
		provideMessageBrokerFactory,
		provideMessageBroker,
		provideRepositoryFactory,
		provideUserReadRepository,
		provideUserEventHandler,
		provideProductEventHandler,
		provideEventConsumer,
	)
	return &consumers.EventConsumer{}, nil
}
