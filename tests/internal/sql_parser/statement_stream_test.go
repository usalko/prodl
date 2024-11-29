package sql_parser

import (
	"strings"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

func TestStatementStream(t *testing.T) {
	stringForStream := `
--
-- PostgreSQL database dump
--

-- Dumped from database version 15.6
-- Dumped by pg_dump version 15.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: phytonyms.dev
--

-- *not* creating schema, since initdb creates it

ALTER SCHEMA public OWNER TO "phytonyms.dev";

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: phytonyms.dev
--

COMMENT ON SCHEMA public IS '';

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: articles_article; Type: TABLE; Schema: public; Owner: phytonyms.dev
--

CREATE TABLE public.articles_article (
    id bigint NOT NULL,
    title character varying(300) NOT NULL,
    text text NOT NULL,
    youtube_link character varying(200) NOT NULL,
    author character varying(200) NOT NULL,
    pub_date date NOT NULL,
    preview character varying(300) NOT NULL,
    published boolean NOT NULL
);`
	var textPieces []string = make([]string, 0)
	var parsedStatements []ast.Statement = make([]ast.Statement, 0)
	err := sql_parser.StatementStream(
		strings.NewReader(stringForStream),
		dialect.PSQL,
		func(statement_text string, statement ast.Statement, parse_error error) {
			textPieces = append(textPieces, statement_text)
			if statement != nil {
				parsedStatements = append(parsedStatements, statement)
			}
		},
	)
	if err != nil {
		t.Errorf("%q", err)
	}
	if len(textPieces) != 15 {
		t.Errorf("count of text pieces is %v but expected %v", len(textPieces), 15)
	}
	if len(parsedStatements) != 15 {
		t.Errorf("count of statements is %v but expected %v", len(parsedStatements), 15)
	}
}
