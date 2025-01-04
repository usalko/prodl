package sql_parser

import (
	"fmt"
	"testing"

	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
	"github.com/usalko/prodl/internal/sql_parser/psql"
)

func TestPsqlStatements(t *testing.T) {
	testcases := []struct {
		in string
		id []int
	}{
		{
			in: "-- comment\nCOMMENT ON SCHEMA public IS '1'",
			id: []int{psql.COMMENT, psql.ON, psql.SCHEMA, 0},
		},
		{
			in: "-- comment\n\nSET statement_timeout = 0",
			id: []int{psql.SET, 0},
		},
		{
			in: "\n\n--\n-- Name: public; Type: SCHEMA; Schema: -; Owner: phytonyms.dev\n--\n\n-- *not* creating schema, since initdb creates it\n\n\nALTER SCHEMA public OWNER TO \"phytonyms.dev\"",
			id: []int{psql.ALTER, 0},
		},
		{
			in: "\n\n--\n-- Name: articles_article_id_seq; Type: SEQUENCE; Schema: public; Owner: phytonyms.dev\n--\n\nCREATE SEQUENCE public.articles_article_id_seq\n    START WITH 1\n    INCREMENT BY 1\n    NO MINVALUE\n    NO MAXVALUE\n    CACHE 1",
			id: []int{psql.CREATE, 0},
		},
		{
			in: "ALTER SEQUENCE serial RESTART WITH 105",
			id: []int{psql.ALTER, 0},
		},
		{
			in: "ALTER TABLE ONLY public.articles_article ALTER COLUMN id SET DEFAULT nextval('public.articles_article_id_seq'::regclass)",
			id: []int{psql.ALTER, 0},
		},
		{
			in: "COPY public.articles_article (id, title) FROM stdin;\n11\tArticle 1\n12\tArticle 2\n\\.",
			id: []int{psql.COPY, 0},
		},
		{
			in: "COPY public.feedback_feedback (id, name, comment, contacts, created) FROM stdin;\n1\tBilly\tcomment\tbiden@wash.gov\t2023-06-07 13:37:49.001783+00\n\\.",
			id: []int{psql.COPY, 0},
		},
		{
			in: "ALTER TABLE ONLY public.articles_article_tag ADD CONSTRAINT articles_article_tag_article_id_d7f5235a_fk_articles_article_id FOREIGN KEY (article_id) REFERENCES public.articles_article(id) DEFERRABLE INITIALLY DEFERRED",
			id: []int{psql.FOREIGN, 0},
		},
	}

	for _, tcase := range testcases {
		t.Run(tcase.in, func(t *testing.T) {
			tok, err := sql_parser.Parse(tcase.in, dialect.PSQL)
			if err != nil {
				t.Fatalf("%v", err)
			}
			fmt.Printf("tok: %v\n", tok)
		})
	}
}

func TestCreateStatement(t *testing.T) {
	text := `CREATE TABLE public.articles_article (
    id bigint NOT NULL,
    title character varying(300) NOT NULL,
    text text NOT NULL,
    youtube_link character varying(200) NOT NULL,
    author character varying(200) NOT NULL,
    pub_date date NOT NULL,
    preview character varying(300) NOT NULL,
    published boolean NOT NULL
)`

	tok, err := sql_parser.Parse(text, dialect.PSQL)
	if err != nil {
		t.Fatalf("%v", err)
		return
	}
	createTable, ok := tok.(*ast.CreateTable)
	if !ok {
		t.Fatalf("%v", fmt.Errorf("not a create table statement: %v", text))
		return
	}
	if createTable.TableSpec == nil {
		t.Fatalf("%v", fmt.Errorf("doesn't recognize fields for create table statement: %v", text))
		return
	}
	fmt.Printf("tok: %v\n", tok)
}
