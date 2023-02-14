package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/crux25/go-compmgt/company-service/data"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/crux25/go-compmgt/helpers"
)

const webPort = "80"

var counts int64

type Config struct {
	DB         *sql.DB
	Models     data.Models
	jsonHelper *helpers.JSONHelper
}

func main() {
	log.Println("Starting company service")

	// connect to DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to Postgres!")
	}
	err := createTables(conn)
	if err != nil {
		panic(err)
	}

	// set up config
	app := &Config{
		DB:         conn,
		Models:     data.New(conn),
		jsonHelper: new(helpers.JSONHelper),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}

func createTables(db *sql.DB) error {
	var exists int64
	db.QueryRow("select 1 from information_schema.tables where table_name=$1", "companies").Scan(&exists)

	if exists != 1 {
		stmt := `CREATE SEQUENCE IF NOT EXISTS public.company_id_seq
            START WITH 1
            INCREMENT BY 1
            NO MINVALUE
            NO MAXVALUE
            CACHE 1;`

		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = `ALTER TABLE public.company_id_seq OWNER TO postgres;`
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = `SET default_tablespace = '';`
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = `SET default_table_access_method = heap;`
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = `CREATE TABLE IF NOT EXISTS public.companies (
		id integer DEFAULT nextval('public.company_id_seq'::regclass) NOT NULL,
		name character varying(15) NOT NULL,
		description character varying(3000),
		num_of_employees integer NOT NULL,
		registered boolean NOT NULL,
		type character varying(25) NOT NULL,
		created_at timestamp without time zone,
		updated_at timestamp without time zone
	);`
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = `ALTER TABLE public.companies OWNER TO postgres;`
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = `SELECT pg_catalog.setval('public.company_id_seq', 1, true);`
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = `ALTER TABLE ONLY public.companies ADD CONSTRAINT companies_pkey PRIMARY KEY (id);`
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		// Insert A dummycompany.
		stmt = `INSERT INTO "public"."companies"("name","description","num_of_employees","registered","type","created_at","updated_at")
	            VALUES($1, $2, $3, $4, $5, $6, $7)`
		_, err = db.Exec(stmt, "Dummy Company", "Dumm company description", 2, true, "Sole Proprietorship", time.Now(), time.Now())
		if err != nil {
			return err
		}

	}

	return nil
}
