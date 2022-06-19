package connection

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"

	gormLogger "gorm.io/gorm/logger"

	"gorm.io/driver/postgres"

	zlogger "github.com/rs/zerolog/log"

	_ "github.com/lib/pq" // Used in test too.
	"gorm.io/gorm"

	gormmysql "gorm.io/driver/mysql"
)

// GormLogLevel represents Gorm log level. Default is silent.
var GormLogLevel = gormLogger.Silent

var (
	psqlOnce       sync.Once
	psqlDBInstance *gorm.DB
)

// NewPostgreSQL instantiates postgresql connection once. Any further calls
// returns cached instance.
func NewPostgreSQL() *gorm.DB {
	psqlOnce.Do(func() {
		host := os.Getenv("POSTGRES_HOST")
		port := os.Getenv("POSTGRES_PORT")
		username := os.Getenv("POSTGRES_USER")
		dbName := os.Getenv("POSTGRES_DATABASE")
		pwd := os.Getenv("POSTGRES_PASSWORD")
		sslMode := os.Getenv("POSTGRES_SSL_MODE")
		sslCert := os.Getenv("POSTGRES_SSL_CERT")

		buf := bytes.Buffer{}
		for i := 0; i < len(pwd); i++ {
			buf.WriteString("*")
		}

		zlogger.Info().
			Str("host", fmt.Sprintf("%s:%s", host, port)).
			Str("username", username).
			Str("db", dbName).
			Str("password", buf.String()).
			Str("ssl_mode", sslMode).
			Str("ssl_cert", sslCert).
			Msg("Connect PostgreSQL database")

		dsn := fmt.Sprintf(`host=%s
			port=%s
			user=%s
			dbname=%s
			password=%s
			sslmode=%s
			sslrootcert=%s
			timezone=utc`,
			host,
			port,
			username,
			dbName,
			pwd,
			sslMode,
			sslCert)

		customLogger := gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  true,
			IgnoreRecordNotFoundError: false,
			LogLevel:                  GormLogLevel,
		})

		postgreSQL, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: customLogger,
		})

		psqlDBInstance = postgreSQL
	})

	return psqlDBInstance
}

var (
	sqlDBOnce     sync.Once
	sqlDBInstance *gorm.DB
)

// NewMySQL instantiates sql connection once. Any further calls
// returns cached instance.
func NewMySQL() *gorm.DB {
	sqlDBOnce.Do(func() {
		username := os.Getenv("MYSQL_USERNAME")
		pwd := os.Getenv("MYSQL_PASSWORD")
		host := os.Getenv("MYSQL_HOST")
		dbName := os.Getenv("MYSQL_DATABASE")

		buf := bytes.Buffer{}
		for i := 0; i < len(pwd); i++ {
			buf.WriteString("*")
		}

		zlogger.Info().
			Str("username", username).
			Str("password", buf.String()).
			Str("host", host).
			Str("db", dbName).
			Msg("Connect MariaDB database")

		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=UTC",
			username,
			pwd,
			host,
			dbName)

		customLogger := gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  true,
			IgnoreRecordNotFoundError: false,
			LogLevel:                  GormLogLevel,
		})

		if certPath := os.Getenv("CMS_MYSQL_CERT"); certPath != "" {
			rootCertPool := x509.NewCertPool()
			pem, err := ioutil.ReadFile(certPath)
			if err != nil {
				zlogger.Fatal().Msgf("read cert file %s failed: %s", certPath, err)
			}

			if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
				zlogger.Fatal().Msgf("append pem failed: %s", err)
			}

			err = mysql.RegisterTLSConfig("custom", &tls.Config{
				RootCAs: rootCertPool,
			})
			if err != nil {
				zlogger.Fatal().Msgf("register tls config failed: %s", err)
			}

			dsn += "&tls=custom"
		}

		db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{
			Logger: customLogger,
		})

		// Check the server version.
		var version string
		err = db.Raw("SELECT VERSION();").Row().Scan(&version)

		zlogger.Info().Msgf("%s version: %s\n",
			db.Dialector.Name(),
			version)

		sqlDBInstance = db
	})

	return sqlDBInstance
}
