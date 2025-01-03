package cmd

import (
	"fmt"
	"io"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/usalko/prodl/cmd/graph_templates"
	"github.com/usalko/prodl/internal/archive_stream"
	"github.com/usalko/prodl/internal/sql_parser"
	"github.com/usalko/prodl/internal/sql_parser/ast"
	"github.com/usalko/prodl/internal/sql_parser/dialect"
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "The 'graph' subcommand will graph tables in dump or in sql connection or in both.",
	Long: `The 'graph' subcommand graphs tables for a sql dump or for a database connection. For example:

'<cmd> graph dump-file-name.tar.gz',
'<cmd> graph -c sqlite3://./local.sqlite3',
'<cmd> graph dump-file-name.tar.gz -c sqlite3://./local.sqlite3'.
`,
	Args: cobra.RangeArgs(1, MAX_COUNT_FOR_PROCESSING_FILES),
	Run: func(cmd *cobra.Command, args []string) {
		debugLevel, _ := cmd.Flags().GetInt("debug-level")
		// 1. Open file and detect dialect
		// 2. If connection specified extract tables structures to dot file
		// 3. If dump specified extract tables structures to dot file

		dumpSqlDialect := dialect.PSQL

		// Open reader and do StatementStream
		for _, fileName := range args {
			rootCmd.Printf("process file %v", fileName)
			graph, err := processFileForGraph(fileName, dumpSqlDialect, debugLevel)
			if err != nil {
				rootCmd.Println(" - fail")
				rootCmd.Println()
				rootCmd.PrintErrf("Error is %v", err)
				rootCmd.PrintErrln()
			} else {
				// Save graphviz .dot file

				dotFileName := getComprehensiveDotFileName(fileName)
				err = saveGraphToDotFile(graph, dotFileName)
				if err != nil {
					rootCmd.PrintErrf("Error is %v\n", err)
				} else {
					rootCmd.Printf("Successfully save .dot file %v, for more details @see information from https://graphviz.org/ \n", dotFileName)
				}
			}
		}
	},
}

func init() {
	graphCmd.Flags().IntP("debug-level", "d", 0, `
Debug level:

    0 no debug messages
	1 show debug messages
	2 show advanced debug messages

`)
	rootCmd.AddCommand(graphCmd)
}

type RankDir string

const (
	TB RankDir = "TB"
	BT RankDir = "BT"
	LR RankDir = "LR"
	RL RankDir = "RL"
)

type Relation struct {
	NeedsNode    bool
	TargetSchema string
	Target       string
	SchemaName   string
	Name         string
	Label        string
	Arrows       []string
}

type Field struct {
	DisableAbstractFields bool
	Abstract              bool
	PrimaryKey            bool
	Blank                 bool
	Relation              Relation
	Label                 string
	Type                  string
}

type Table struct {
	SchemaName    string
	Name          string
	Label         string
	Abstracts     []string
	DisableFields bool
	Fields        []Field
	Relations     []Relation
}

type Graph struct {
	UseSubgraph bool
	SchemaName  string
	Tables      []*Table
}

type DiGraph struct {
	CreatedAt  time.Time
	CliOptions string
	Rankdir    RankDir
	Graphs     []*Graph
}

func newDigraph() *DiGraph {
	return &DiGraph{
		CreatedAt:  time.Now(),
		CliOptions: "",
		Rankdir:    BT,
		Graphs:     make([]*Graph, 0, 1),
	}
}

func (dg *DiGraph) addTable(tableName string, schemaName string) {
	if len(dg.Graphs) == 0 {
		dg.Graphs = append(dg.Graphs, &Graph{
			UseSubgraph: false,
			SchemaName:  schemaName,
			Tables:      make([]*Table, 0, 100),
		})
	}
	dg.Graphs[0].Tables = append(dg.Graphs[0].Tables, &Table{
		SchemaName: schemaName,
		Name:       tableName,
		Label:      tableName,
	})
}

func getComprehensiveDotFileName(dumpFileName string) string {
	return dumpFileName + ".dot"
}

func saveGraphToDotFile(digraph *DiGraph, fileName string) error {
	tmpl, err := template.New("digraph").Parse(graph_templates.GetTemplate(graph_templates.DIGRAPH))
	if err != nil {
		return err
	}
	fileWriter, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fileWriter.Close()

	return tmpl.Execute(fileWriter, digraph)
}

func processFileForGraph(
	fileName string,
	sqlDialect dialect.SqlDialect,
	debugLevel int,
) (*DiGraph, error) {
	respBody, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("file %s open error (%v)", fileName, err)
	}
	defer respBody.Close()

	reader := archive_stream.NewReader(respBody)
	dumpGraph := newDigraph()
	for {
		entry, err := reader.GetNextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to get next entry (%v)", err)
		}

		if !entry.IsDir() {
			rc, err := entry.Open()
			defer func() {
				if err := rc.Close(); err != nil {
					rootCmd.PrintErrf("close entry reader fail: %s", err)
				}
			}()

			if err != nil {
				return nil, fmt.Errorf("unable to open file: %s", err)
			}

			statementsCount := 0
			lastTime := time.Now()
			sql_parser.StatementStream(rc, sqlDialect,
				func(statementText string, statement ast.Statement, parseError error) {
					if parseError != nil {
						if debugLevel >= 1 {
							rootCmd.PrintErrf("parse sql statement:\n %s \n\nfail: %s\n", statementText, parseError)
						} else {
							rootCmd.PrintErrf("%s\n", parseError)
						}
					}
					createStatement, ok := statement.(*ast.CreateTable)
					if ok {
						dumpGraph.addTable(createStatement.Table.Name.V, createStatement.Table.Qualifier.V)
					}
					statementsCount++
					if debugLevel >= 2 {
						rootCmd.Printf("[%v] processed statements: %v\n", time.Since(lastTime), statementsCount)
					}
					lastTime = time.Now()
				})
		}
	}
	return dumpGraph, nil
}
