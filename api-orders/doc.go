// Package api_orders provides a scalable order management API backend
// built with Go, Gin Framework, MongoDB, and RabbitMQ following
// Hexagonal Architecture (Clean Architecture) principles.
//
// The project structure follows Clean Architecture with clear separation of concerns:
//   - cmd/: Application entry point
//   - internal/core/: Business logic layer (domain, DTOs, use cases, ports)
//   - internal/adapter/: External interface adapters (HTTP handlers, repositories, message producers)
//   - internal/infra/: Infrastructure concerns (database, message broker connections)
//   - wire/: Dependency injection setup using Google Wire
//
// For more information, see README.md
package main
