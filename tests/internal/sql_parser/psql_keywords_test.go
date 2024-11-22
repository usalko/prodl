package sql_parser

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/usalko/sent/internal/sql_parser"
	"github.com/usalko/sent/internal/sql_parser/cache"
	"github.com/usalko/sent/internal/sql_parser/psql"
)

func TestKeywordTable(t *testing.T) {
	for _, kw := range psql.GetKeywords() {
		lookup, ok := cache.KeywordLookup(kw.Name)
		require.Truef(t, ok, "keyword %q failed to match", kw.Name)
		require.Equalf(t, lookup, kw.Id, "keyword %q matched to %d (expected %d)", kw.Name, lookup, kw.Id)
	}
}

func TestCompatibility(t *testing.T) {
	file, err := os.Open(path.Join("test_data", "psql_keywords.txt"))
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	skipStep := 2
	for scanner.Scan() {
		if skipStep != 0 {
			skipStep--
			continue
		}

		afterSplit := strings.SplitN(scanner.Text(), "\t", 3)
		word, reserved := afterSplit[0], strings.Trim(afterSplit[1], " ") != ""
		if reserved {
			word = "\"" + word + "\""
		}
		sql := fmt.Sprintf("create table %s(c1 int)", word)
		_, err := sql_parser.ParseStrictDDL(sql)
		if err != nil {
			t.Errorf("%s is not compatible with psql", word)
		}
	}
}

func TestMultiExpressionsParsing(t *testing.T) {
	expressions := `
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
);
	`
	_, err := sql_parser.SplitStatementToPieces(expressions)
	if err != nil {
		t.Errorf("%q", err)
	}
}
