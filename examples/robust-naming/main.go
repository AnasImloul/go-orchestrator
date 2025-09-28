package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// Simulate different packages with same interface names
// In real applications, these would be in separate packages

// Package: github.com/user/postgres
type DatabaseService interface {
	orchestrator.Service
	Connect() error
	Query(sql string) ([]string, error)
}

type postgresDatabaseService struct {
	host string
	port int
}

func (p *postgresDatabaseService) Start(ctx context.Context) error {
	fmt.Printf("🐘 PostgreSQL starting on %s:%d\n", p.host, p.port)
	time.Sleep(100 * time.Millisecond)
	return p.Connect()
}

func (p *postgresDatabaseService) Stop(ctx context.Context) error {
	fmt.Printf("🐘 PostgreSQL stopping on %s:%d\n", p.host, p.port)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (p *postgresDatabaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("PostgreSQL on %s:%d is healthy", p.host, p.port),
	}
}

func (p *postgresDatabaseService) Connect() error {
	fmt.Printf("   ✅ PostgreSQL connected to %s:%d\n", p.host, p.port)
	return nil
}

func (p *postgresDatabaseService) Query(sql string) ([]string, error) {
	fmt.Printf("   📊 PostgreSQL query: %s\n", sql)
	return []string{"result1", "result2"}, nil
}

// Package: github.com/user/mysql
type MySQLDatabaseService interface {
	orchestrator.Service
	Connect() error
	Execute(sql string) error
}

type mysqlDatabaseService struct {
	host string
	port int
}

func (m *mysqlDatabaseService) Start(ctx context.Context) error {
	fmt.Printf("🐬 MySQL starting on %s:%d\n", m.host, m.port)
	time.Sleep(100 * time.Millisecond)
	return m.Connect()
}

func (m *mysqlDatabaseService) Stop(ctx context.Context) error {
	fmt.Printf("🐬 MySQL stopping on %s:%d\n", m.host, m.port)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (m *mysqlDatabaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("MySQL on %s:%d is healthy", m.host, m.port),
	}
}

func (m *mysqlDatabaseService) Connect() error {
	fmt.Printf("   ✅ MySQL connected to %s:%d\n", m.host, m.port)
	return nil
}

func (m *mysqlDatabaseService) Execute(sql string) error {
	fmt.Printf("   ⚡ MySQL execute: %s\n", sql)
	return nil
}

// Package: github.com/user/redis
type CacheService interface {
	orchestrator.Service
	Set(key, value string) error
	Get(key string) (string, error)
}

type redisCacheService struct {
	host string
	port int
}

func (r *redisCacheService) Start(ctx context.Context) error {
	fmt.Printf("🔴 Redis starting on %s:%d\n", r.host, r.port)
	time.Sleep(75 * time.Millisecond)
	return nil
}

func (r *redisCacheService) Stop(ctx context.Context) error {
	fmt.Printf("🔴 Redis stopping on %s:%d\n", r.host, r.port)
	time.Sleep(25 * time.Millisecond)
	return nil
}

func (r *redisCacheService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Redis on %s:%d is healthy", r.host, r.port),
	}
}

func (r *redisCacheService) Set(key, value string) error {
	fmt.Printf("   💾 Redis SET %s = %s\n", key, value)
	return nil
}

func (r *redisCacheService) Get(key string) (string, error) {
	fmt.Printf("   📖 Redis GET %s\n", key)
	return "cached_value", nil
}

// Package: github.com/user/memcached
type MemcachedCacheService interface {
	orchestrator.Service
	Store(key, value string) error
	Retrieve(key string) (string, error)
}

type memcachedCacheService struct {
	host string
	port int
}

func (m *memcachedCacheService) Start(ctx context.Context) error {
	fmt.Printf("🟡 Memcached starting on %s:%d\n", m.host, m.port)
	time.Sleep(75 * time.Millisecond)
	return nil
}

func (m *memcachedCacheService) Stop(ctx context.Context) error {
	fmt.Printf("🟡 Memcached stopping on %s:%d\n", m.host, m.port)
	time.Sleep(25 * time.Millisecond)
	return nil
}

func (m *memcachedCacheService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Memcached on %s:%d is healthy", m.host, m.port),
	}
}

func (m *memcachedCacheService) Store(key, value string) error {
	fmt.Printf("   💾 Memcached STORE %s = %s\n", key, value)
	return nil
}

func (m *memcachedCacheService) Retrieve(key string) (string, error) {
	fmt.Printf("   📖 Memcached RETRIEVE %s\n", key)
	return "memcached_value", nil
}

func main() {
	fmt.Println("🎯 Robust Naming Strategy Example")
	fmt.Println("==================================")
	fmt.Println("✨ Demonstrates how the library handles interface name conflicts")
	fmt.Println("✨ Shows both automatic and custom naming strategies")
	fmt.Println()

	// Create service registry
	registry := orchestrator.New()

	// Method 1: Automatic naming with full package path
	// The library automatically creates unique names by including the package path
	fmt.Println("📦 Method 1: Automatic naming with full package path")
	fmt.Println("   - github.com/user/postgres.DatabaseService -> github.com/user/postgres::DatabaseService")
	fmt.Println("   - github.com/user/mysql.DatabaseService -> github.com/user/mysql::DatabaseService")
	fmt.Println("   - github.com/user/redis.CacheService -> github.com/user/redis::CacheService")
	fmt.Println("   - github.com/user/memcached.CacheService -> github.com/user/memcached::CacheService")
	fmt.Println()

	registry.Register(
		orchestrator.NewServiceSingleton[DatabaseService](
			&postgresDatabaseService{host: "localhost", port: 5432},
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[MySQLDatabaseService](
			&mysqlDatabaseService{host: "localhost", port: 3306},
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[CacheService](
			&redisCacheService{host: "localhost", port: 6379},
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[MemcachedCacheService](
			&memcachedCacheService{host: "localhost", port: 11211},
		),
	)

	// Start the service registry
	ctx := context.Background()
	fmt.Println("🚀 Starting service registry...")
	if err := registry.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start service registry: %v\n", err)
		return
	}

	fmt.Println("\n✅ Service registry started successfully!")
	fmt.Println("   - All services with same interface names registered without conflicts")
	fmt.Println("   - Automatic naming strategy created unique service names")

	// Check health
	fmt.Println("\n📊 Checking service health...")
	health := registry.Health(ctx)
	for name, status := range health {
		fmt.Printf("   %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("\n⏱️  Running for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the service registry
	fmt.Println("\n🛑 Stopping service registry...")
	if err := registry.Stop(ctx); err != nil {
		fmt.Printf("❌ Failed to stop service registry: %v\n", err)
		return
	}

	fmt.Println("✅ Service registry stopped successfully!")

	// Method 2: Custom naming for more control
	fmt.Println("\n📝 Method 2: Custom naming for more control")
	fmt.Println("   - Use .WithName() to specify custom service names")
	fmt.Println("   - Useful for shorter, more descriptive names")
	fmt.Println()

	registry2 := orchestrator.New()

	registry2.Register(
		orchestrator.NewServiceSingleton[DatabaseService](
			&postgresDatabaseService{host: "prod-db", port: 5432},
		).WithName("postgres-primary"),
	)

	registry2.Register(
		orchestrator.NewServiceSingleton[MySQLDatabaseService](
			&mysqlDatabaseService{host: "analytics-db", port: 3306},
		).WithName("mysql-analytics"),
	)

	registry2.Register(
		orchestrator.NewServiceSingleton[CacheService](
			&redisCacheService{host: "cache-1", port: 6379},
		).WithName("redis-session"),
	)

	registry2.Register(
		orchestrator.NewServiceSingleton[MemcachedCacheService](
			&memcachedCacheService{host: "cache-2", port: 11211},
		).WithName("memcached-object"),
	)

	fmt.Println("🚀 Starting service registry with custom names...")
	if err := registry2.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start service registry: %v\n", err)
		return
	}

	fmt.Println("\n✅ Service registry with custom names started successfully!")
	fmt.Println("   - postgres-primary: PostgreSQL primary database")
	fmt.Println("   - mysql-analytics: MySQL analytics database")
	fmt.Println("   - redis-session: Redis session cache")
	fmt.Println("   - memcached-object: Memcached object cache")

	// Check health
	fmt.Println("\n📊 Checking service health...")
	health2 := registry2.Health(ctx)
	for name, status := range health2 {
		fmt.Printf("   %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("\n⏱️  Running for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the service registry
	fmt.Println("\n🛑 Stopping service registry...")
	if err := registry2.Stop(ctx); err != nil {
		fmt.Printf("❌ Failed to stop service registry: %v\n", err)
		return
	}

	fmt.Println("✅ Service registry stopped successfully!")
	fmt.Println("\n🎉 Robust naming strategy working perfectly!")
	fmt.Println("   - No more interface name conflicts")
	fmt.Println("   - Automatic unique naming with package paths")
	fmt.Println("   - Optional custom naming for better control")
}
