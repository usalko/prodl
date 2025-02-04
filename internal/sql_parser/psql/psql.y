/*

Copyright 2024 Vanya Usalko <ivict@rambler.ru>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

%{
package psql

import (
    "github.com/usalko/prodl/internal/sql_parser/ast"
	  "github.com/usalko/prodl/internal/sql_parser/tokenizer"
    "github.com/usalko/prodl/internal/sql_types"
)

func setParseTree(psqlex psqLexer, stmt ast.Statement) {
  psqlex.(tokenizer.Tokenizer).SetParseTree(stmt)
}

func setAllowComments(psqlex psqLexer, allow bool) {
  psqlex.(tokenizer.Tokenizer).SetAllowComments(allow)
}

func setIgnoreCommentKeyword(psqlex psqLexer, ignore bool) {
  psqlex.(tokenizer.Tokenizer).SetIgnoreCommentKeyword(ignore)
}

func setDDL(psqlex psqLexer, node ast.Statement) {
  psqlex.(tokenizer.Tokenizer).SetPartialDDL(node)
}

func incNesting(psqlex psqLexer) bool {
  psqlex.(tokenizer.Tokenizer).IncNesting()
  if psqlex.(tokenizer.Tokenizer).GetNesting() == 200 {
    return true
  }
  return false
}

func decNesting(psqlex psqLexer) {
  psqlex.(tokenizer.Tokenizer).DecNesting()
}

// skipToEnd forces the lexer to end prematurely. Not all SQL statements
// are supported by the Parser, thus calling skipToEnd will make the lexer
// return EOF early.
func skipToEnd(psqlex psqLexer) {
  psqlex.(tokenizer.Tokenizer).SetSkipToEnd(true)
}

func bindVariable(psqlex psqLexer, bvar string) {
  psqlex.(tokenizer.Tokenizer).BindVar(bvar, struct{}{})
}

%}

%struct {
  empty         struct{}
  LengthScaleOption ast.LengthScaleOption
  tableName     ast.TableName
  tableIdent    ast.TableIdent
  str           string
  strs          []string
  vindexParam   ast.VindexParam
  jsonObjectParam *ast.JSONObjectParam
  colIdent      ast.ColIdent
  joinCondition *ast.JoinCondition
  databaseOption ast.DatabaseOption
  columnType    ast.ColumnType
  columnCharset ast.ColumnCharset
  jsonPathParam ast.JSONPathParam
  schemaIdent   ast.SchemaIdent
  schemaName    ast.SchemaName
  sequenceIdent ast.SequenceIdent
  sequenceName  ast.SequenceName
  copyFromSource ast.CopyFromSource
  copyToTarget   ast.CopyToTarget
  copyOptions    ast.CopyOptions
  copyOption     ast.CopyOption
}

%union {
  statement     ast.Statement
  selStmt       ast.SelectStatement
  tableExpr     ast.TableExpr
  expr          ast.Expr
  colTuple      ast.ColTuple
  optVal        ast.Expr
  constraintInfo ast.ConstraintInfo
  alterOption      ast.AlterOption
  characteristic ast.Characteristic

  ins           *ast.Insert
  colName       *ast.ColName
  indexHint    *ast.IndexHint
  indexHints    ast.IndexHints
  indexHintForType ast.IndexHintForType
  literal        *ast.Literal
  subquery      *ast.Subquery
  derivedTable  *ast.DerivedTable
  when          *ast.When
  with          *ast.With
  cte           *ast.CommonTableExpr
  ctes          []*ast.CommonTableExpr
  order         *ast.Order
  limit         *ast.Limit

  updateExpr    *ast.UpdateExpr
  setExpr       *ast.SetExpr
  commentExpr   *ast.CommentOnSchema
  convertType   *ast.ConvertType
  aliasedTableName *ast.AliasedTableExpr
  tableSpec     *ast.TableSpec
  columnDefinition *ast.ColumnDefinition
  indexDefinition *ast.IndexDefinition
  indexInfo     *ast.IndexInfo
  indexOption   *ast.IndexOption
  indexColumn   *ast.IndexColumn
  sequenceSpec  *ast.SequenceSpec
  showFilter    *ast.ShowFilter
  optLike       *ast.OptLike
  selectInto	  *ast.SelectInto
  createDatabase  *ast.CreateDatabase
  alterDatabase  *ast.AlterDatabase
  createTable      *ast.CreateTable
  tableAndLockType *ast.TableAndLockType
  alterTable       *ast.AlterTable
  tableOption      *ast.TableOption
  columnTypeOptions *ast.ColumnTypeOptions
  createSequence    *ast.CreateSequence
  alterSequence     *ast.AlterSequence
  constraintDefinition *ast.ConstraintDefinition
  revertMigration   *ast.RevertMigration
  alterMigration    *ast.AlterMigration
  alterSchema       *ast.AlterSchema
  trimType          ast.TrimType

  whens         []*ast.When
  columnDefinitions []*ast.ColumnDefinition
  indexOptions  []*ast.IndexOption
  indexColumns  []*ast.IndexColumn
  databaseOptions []ast.DatabaseOption
  tableAndLockTypes ast.TableAndLockTypes
  renameTablePairs []*ast.RenameTablePair
  alterOptions	   []ast.AlterOption
  vindexParams  []ast.VindexParam
  jsonPathParams []ast.JSONPathParam
  jsonObjectParams []*ast.JSONObjectParam
  characteristics []ast.Characteristic
  selectExpr    ast.SelectExpr
  columns       ast.Columns
  tableExprs    ast.TableExprs
  tableNames    ast.TableNames
  exprs         ast.Exprs
  values        ast.Values
  valTuple      ast.ValTuple
  orderBy       ast.OrderBy
  updateExprs   ast.UpdateExprs
  setExprs      ast.SetExprs
  selectExprs   ast.SelectExprs
  tableOptions     ast.TableOptions

  colKeyOpt     ast.ColumnKeyOption
  referenceAction ast.ReferenceAction
  matchAction ast.MatchAction
  isolationLevel ast.IsolationLevel
  insertAction ast.InsertAction
  scope 	ast.Scope
  lock 		ast.Lock
  joinType  	ast.JoinType
  comparisonExprOperator ast.ComparisonExprOperator
  isExprOperator ast.IsExprOperator
  matchExprOption ast.MatchExprOption
  orderDirection  ast.OrderDirection
  explainType 	  ast.ExplainType
  intervalType	  ast.IntervalTypes
  lockType ast.LockType
  referenceDefinition *ast.ReferenceDefinition

  columnStorage ast.ColumnStorage
  columnFormat ast.ColumnFormat

  boolean bool
  boolVal ast.BoolVal
  ignore ast.Ignore
  definer 	*ast.Definer
  integer 	int

  JSONTableExpr	*ast.JSONTableExpr
  jtColumnDefinition *ast.JtColumnDefinition
  jtColumnList	[]*ast.JtColumnDefinition
  jtOnResponse	*ast.JtOnResponse
}

// These precedence rules are there to handle shift-reduce conflicts.
%nonassoc <str> MEMBER
// FUNCTION_CALL_NON_KEYWORD is used to resolve shift-reduce conflicts occuring due to function_call_generic symbol and
// having special parsing for functions whose names are non-reserved keywords. The shift-reduce conflict occurrs because
// after seeing a non-reserved keyword, if we see '(', then we can either shift to use the special parsing grammar rule or
// reduce the non-reserved keyword into sql_id and eventually use a rule from function_call_generic.
// The way to fix this conflict is to give shifting higher precedence than reducing.
// Adding no precedence also works, since shifting is the default, but it reports a large number of conflicts
// Shifting on '(' already has an assigned precedence.
// All we need to add is a lower precedence to reducing the grammar symbol to non-reserved keywords.
// In order to ensure lower precedence of reduction, this rule has to come before the precedence declaration of '('.
// This precedence should not be used anywhere else other than with function names that are non-reserved-keywords.
%nonassoc <str> FUNCTION_CALL_NON_KEYWORD

%token LEX_ERROR
%left <str> UNION
%token <str> SELECT STREAM VSTREAM INSERT UPDATE DELETE FROM WHERE GROUP HAVING ORDER BY LIMIT OFFSET FOR
%token <str> ALL DISTINCT AS EXISTS ASC DESC INTO DUPLICATE DEFAULT SET LOCK UNLOCK KEYS DO CALL COMMENT
%token <str> DISTINCTROW PARSER GENERATED ALWAYS
// Not implemented:
%token <str> ANY ASYMMETRIC AUTHORIZATION CONCURRENTLY CURRENT_CATALOG CURRENT_ROLE CURRENT_SCHEMA DEFERRABLE
%token <str> FETCH FREEZE GRANT ILIKE INITIALLY INTERSECT ISNULL NOTNULL OVERLAPS PLACING SESSION_USER
%token <str> SIMILAR SOME SYMMETRIC SYSTEM_USER TABLESAMPLE VARIADIC VERBOSE ABORT ABSENT ABSOLUTE
%token <str> ACCESS AGGREGATE ALSO ASENSITIVE ASSERTION ASSIGNMENT AT ATOMIC ATTACH ATTRIBUTE
%token <str> BACKWARD BEFORE BREADTH CACHE CALLED CATALOG CHAIN CHARACTERISTICS
%token <str> CHECKPOINT CLASS CLOSE CLUSTER COMMENTS CONDITIONAL
%token <str> CONFIGURATION CONFLICT CONSTRAINTS CONTENT
%token <str> CONTINUE CONVERSION COST OPTIONALLY
%token <str> ESCAPED ENCLOSED TERMINATED
%token <str> STARTING LINES OVERWRITE
%token <str> MANIFEST HEADER CSV
%token <str> CUBE CURRENT CURSOR
%token <str> CYCLE DATA DEC
%token <str> DECLARE DEFAULTS
%token <str> DEFERRED DELIMITER
%token <str> DELIMITERS DEPENDS DEPTH
%token <str> DETACH DICTIONARY DOCUMENT DOMAIN
%token <str> EACH ENCODING ENCRYPTED EXCLUDING EXPRESSION
%token <str> EXTENSION EXTERNAL FAMILY FILTER FINALIZE FORWARD
%token <str> FUNCTIONS GRANTED GREATEST HANDLER HOLD IDENTITY IMMEDIATE
%token <str> IMMUTABLE IMPLICIT INCLUDE INCLUDING INCREMENT INDENT INHERIT INHERITS
%token <str> INLINE INOUT INPUT INSENSITIVE INSTEAD JSON_ARRAYAGG JSON_EXISTS JSON_OBJECTAGG
%token <str> JSON_QUERY JSON_SCALAR JSON_SERIALIZE KEEP LABEL LARGE LEAKPROOF LEAST LISTEN LOAD LOCATION
%token <str> LOGGED MAPPING MATCHED MATERIALIZED MERGE_ACTION METHOD MINVALUE MOVE NATIONAL NEW NFC NFD NFKC
%token <str> NFKD NORMALIZE NORMALIZED NOTHING NOTIFY NULLIF OBJECT OIDS OMIT OPERATOR OPTIONS OUT OVERLAY
%token <str> OVERRIDING OWNED OWNER PARALLEL PARAMETER PASSING PLAN PLANS POLICY POSITION PRECISION
%token <str> PREPARED PRESERVE PRIOR PROCEDURAL PROCEDURES PROGRAM PUBLICATION QUOTE QUOTES
%token <str> RANGE REASSIGN RECHECK REF REFERENCING REFRESH REINDEX RELATIVE REPLICA
%token <str> RESET RETURN RETURNS REVOKE ROLLUP ROUTINE ROUTINES ROW ROWS
%token <str> RULE SCALAR SCROLL SEARCH SEQUENCES SERVER SETOF SETS
%token <str> SNAPSHOT SOURCE STABLE STANDALONE STATEMENT
%token <str> STATISTICS STDIN STDOUT STRICT STRIP
%token <str> SUBSCRIPTION SUPPORT SYSID
%token <str> TARGET TEMP TEMPLATE
%token <str> TRANSFORM TREAT
%token <str> TRUSTED
%token <str> TYPE TYPES
%token <str> UESCAPE UNCONDITIONAL
%token <str> UNENCRYPTED UNKNOWN UNLISTEN
%token <str> UNLOGGED UNTIL VACUUM VALID VALIDATE
%token <str> VALIDATOR VARYING VERSION VIEWS VOLATILE WHITESPACE
%token <str> WITHIN WRAPPER XML XMLATTRIBUTES XMLCONCAT XMLELEMENT XMLEXISTS
%token <str> XMLFOREST XMLNAMESPACES XMLPARSE XMLPI XMLROOT XMLSERIALIZE XMLTABLE YES
%token <str> ZONE STOP LOG_VERBOSITY ON_ERROR FORCE_NULL FORCE_NOT_NULL FORCE_QUOTE
// Non reserved but sql2023
%token <str> ARRAY_MAX_CARDINALITY CHARACTER_SET_CATALOG COMMAND_FUNCTION_CODE CURRENT_DEFAULT_TRANSFORM_GROUP
%token <str> CURRENT_TRANSFORM_GROUP_FOR_TYPE DATETIME_INTERVAL_CODE DATETIME_INTERVAL_PRECISION
%token <str> DYNAMIC_FUNCTION_CODE END_EXEC PARAMETER_ORDINAL_POSITION PARAMETER_SPECIFIC_CATALOG
%token <str> PARAMETER_SPECIFIC_NAME PARAMETER_SPECIFIC_SCHEMA RETURNED_OCTET_LENGTH TRANSACTIONS_COMMITTED
%token <str> TRANSACTIONS_ROLLED_BACK USER_DEFINED_TYPE_CATALOG USER_DEFINED_TYPE_CODE USER_DEFINED_TYPE_NAME
%token <str> USER_DEFINED_TYPE_SCHEMA
// <<<<
%token <str> VALUES LAST_INSERT_ID
%token <str> NEXT VALUE SHARE MODE
%token <str> SQL_NO_CACHE SQL_CACHE SQL_CALC_FOUND_ROWS
%left <str> JOIN STRAIGHT_JOIN LEFT RIGHT INNER OUTER CROSS NATURAL USE FORCE
%left <str> ON USING INPLACE COPY INSTANT NONE SHARED EXCLUSIVE
%left <str> SUBQUERY_AS_EXPR
%left <str> '(' ',' ')'
%token <str> ID AT_ID AT_AT_ID HEX STRING NCHAR_STRING INTEGRAL FLOAT DECIMAL HEXNUM VALUE_ARG LIST_ARG COMMENT_KEYWORD BIT_LITERAL COMPRESSION
%token <str> JSON_PRETTY JSON_STORAGE_SIZE JSON_STORAGE_FREE JSON_CONTAINS JSON_CONTAINS_PATH JSON_EXTRACT JSON_KEYS JSON_OVERLAPS JSON_SEARCH JSON_VALUE
%token <str> EXTRACT
%token <str> NULL TRUE FALSE OFF
%token <str> DISCARD IMPORT ENABLE DISABLE TABLESPACE
%token <str> VIRTUAL STORED
%token <str> BOTH LEADING TRAILING

%left EMPTY_FROM_CLAUSE
%right INTO

// Precedence dictated by psql. But the vitess grammar is simplified.
// Some of these operators don\'t conflict in our situation. Nevertheless,
// it\'s better to have these listed in the correct order. Also, we don\'t
// support all operators yet.
// * NOTE: ast.If you change anything here, update precedence.go as well *
%nonassoc <str> LOWER_THAN_CHARSET
%nonassoc <str> CHARSET
// Resolve column attribute ambiguity.
%right <str> UNIQUE KEY
%left <str> EXPRESSION_PREC_SETTER
%left <str> OR '|'
%left <str> AND
%right <str> NOT '!'
%left <str> BETWEEN CASE WHEN THEN ELSE END
%left <str> '=' '<' '>' LE GE NE NULL_SAFE_EQUAL IS LIKE REGEXP IN
%left <str> '&'
%left <str> SHIFT_LEFT SHIFT_RIGHT
%left <str> '+' '-'
%left <str> '*' '/' DIV '%' MOD
%left <str> '^'
%right <str> '~' UNARY
%left <str> COLLATE
%right <str> BINARY UNDERSCORE_ARMSCII8 UNDERSCORE_ASCII UNDERSCORE_BIG5 UNDERSCORE_BINARY UNDERSCORE_CP1250 UNDERSCORE_CP1251
%right <str> UNDERSCORE_CP1256 UNDERSCORE_CP1257 UNDERSCORE_CP850 UNDERSCORE_CP852 UNDERSCORE_CP866 UNDERSCORE_CP932
%right <str> UNDERSCORE_DEC8 UNDERSCORE_EUCJPMS UNDERSCORE_EUCKR UNDERSCORE_GB18030 UNDERSCORE_GB2312 UNDERSCORE_GBK UNDERSCORE_GEOSTD8
%right <str> UNDERSCORE_GREEK UNDERSCORE_HEBREW UNDERSCORE_HP8 UNDERSCORE_KEYBCS2 UNDERSCORE_KOI8R UNDERSCORE_KOI8U UNDERSCORE_LATIN1 UNDERSCORE_LATIN2 UNDERSCORE_LATIN5
%right <str> UNDERSCORE_LATIN7 UNDERSCORE_MACCE UNDERSCORE_MACROMAN UNDERSCORE_SJIS UNDERSCORE_SWE7 UNDERSCORE_TIS620 UNDERSCORE_UCS2 UNDERSCORE_UJIS UNDERSCORE_UTF16
%right <str> UNDERSCORE_UTF16LE UNDERSCORE_UTF32 UNDERSCORE_UTF8 UNDERSCORE_UTF8MB4 UNDERSCORE_UTF8MB3
%nonassoc <str> '.'

// There is no need to define precedence for the JSON
// operators because the syntax is restricted enough that
// they don\'t cause conflicts.
%token <empty> JSON_EXTRACT_OP JSON_UNQUOTE_EXTRACT_OP

// DDL Tokens
%token <str> CREATE ALTER DROP RENAME ANALYZE ANALYSE ADD FLUSH CHANGE MODIFY DEALLOCATE
%token <str> REVERT
%token <str> SCHEMA TABLE INDEX VIEW TO IGNORE IF PRIMARY COLUMN SPATIAL FULLTEXT KEY_BLOCK_SIZE CHECK INDEXES
%token <str> ACTION CASCADE CONSTRAINT FOREIGN NO REFERENCES RESTRICT
%token <str> SHOW DESCRIBE EXPLAIN ESCAPE REPAIR OPTIMIZE TRUNCATE COALESCE EXCHANGE REBUILD PARTITIONING REMOVE PREPARE EXECUTE
%token <str> MAXVALUE PARTITION REORGANIZE LESS THAN PROCEDURE TRIGGER
%token <str> VINDEX VINDEXES DIRECTORY NAME UPGRADE
%token <str> STATUS VARIABLES WARNINGS CASCADED DEFINER OPTION SQL UNDEFINED
%token <str> SEQUENCE MERGE TEMPORARY TEMPTABLE INVOKER SECURITY FIRST AFTER LAST

// Migration tokens
%token <str> CANCEL RETRY COMPLETE CLEANUP THROTTLE UNTHROTTLE EXPIRE RATIO

// Transaction Tokens
%token <str> BEGIN START TRANSACTION COMMIT ROLLBACK SAVEPOINT RELEASE WORK

// Type Tokens
%token <str> BIT TINYINT SMALLINT MEDIUMINT INT INTEGER BIGINT INTNUM
%token <str> REAL DOUBLE FLOAT_TYPE DECIMAL_TYPE NUMERIC
%token <str> DATE TIME TIMESTAMP INTERVAL
%token <str> CHAR VARCHAR BOOL CHARACTER VARBINARY NCHAR
%token <str> TEXT
%token <str> JSON JSON_SCHEMA_VALID JSON_SCHEMA_VALIDATION_REPORT ENUM
%token <str> GEOMETRY POINT LINESTRING POLYGON GEOMETRYCOLLECTION MULTIPOINT MULTILINESTRING MULTIPOLYGON
%token <str> ASCII UNICODE // used in CONVERT/CAST types

// Type Modifiers
%token <str> NULLX AUTO_INCREMENT APPROXNUM SIGNED UNSIGNED ZEROFILL

// SHOW tokens
%token <str> CODE COLLATION COLUMNS DATABASES ENGINES EVENT EXTENDED FIELDS FULL FUNCTION GTID_EXECUTED
%token <str> KEYSPACES OPEN PLUGINS PRIVILEGES PROCESSLIST SCHEMAS TABLES TRIGGERS USER
%token <str> VGTID_EXECUTED VSCHEMA

// SET tokens
%token <str> NAMES GLOBAL SESSION ISOLATION LEVEL READ WRITE ONLY REPEATABLE COMMITTED UNCOMMITTED SERIALIZABLE

// Functions
%token <str> CURRENT_TIMESTAMP DATABASE CURRENT_DATE NOW
%token <str> CURRENT_TIME LOCALTIME LOCALTIMESTAMP CURRENT_USER
%token <str> UTC_DATE UTC_TIME UTC_TIMESTAMP
%token <str> DAY DAY_HOUR DAY_MICROSECOND DAY_MINUTE DAY_SECOND HOUR HOUR_MICROSECOND HOUR_MINUTE HOUR_SECOND MICROSECOND MINUTE MINUTE_MICROSECOND MINUTE_SECOND MONTH QUARTER SECOND SECOND_MICROSECOND YEAR_MONTH WEEK YEAR
%token <str> REPLACE
%token <str> CONVERT CAST
%token <str> SUBSTR SUBSTRING
%token <str> GROUP_CONCAT SEPARATOR
%token <str> TIMESTAMPADD TIMESTAMPDIFF
%token <str> WEIGHT_STRING
%token <str> LTRIM RTRIM TRIM
%token <str> JSON_ARRAY JSON_OBJECT JSON_QUOTE
%token <str> JSON_DEPTH JSON_TYPE JSON_LENGTH JSON_VALID
%token <str> JSON_ARRAY_APPEND JSON_ARRAY_INSERT JSON_INSERT JSON_MERGE JSON_MERGE_PATCH JSON_MERGE_PRESERVE JSON_REMOVE JSON_REPLACE JSON_SET JSON_UNQUOTE

// Match
%token <str> MATCH AGAINST BOOLEAN LANGUAGE WITH QUERY EXPANSION WITHOUT VALIDATION

// PostgresQL reserved words that are unused by this grammar will map to this token.
%token <str> UNUSED ARRAY BYTEA BYTE CUME_DIST DESCRIPTION DENSE_RANK EMPTY EXCEPT FIRST_VALUE GROUPING GROUPS JSON_TABLE LAG LAST_VALUE LATERAL LEAD
%token <str> NTH_VALUE NTILE OF OVER PERCENT_RANK RANK RECURSIVE ROW_NUMBER SYSTEM WINDOW
%token <str> ACTIVE ADMIN AUTOEXTEND_SIZE BUCKETS CLONE COLUMN_FORMAT COMPONENT DEFINITION ENFORCED ENGINE_ATTRIBUTE EXCLUDE FOLLOWING GEOMCOLLECTION GET_MASTER_PUBLIC_KEY HISTOGRAM HISTORY
%token <str> INACTIVE INVISIBLE LOCKED MASTER_COMPRESSION_ALGORITHMS MASTER_PUBLIC_KEY_PATH MASTER_TLS_CIPHERSUITES MASTER_ZSTD_COMPRESSION_LEVEL
%token <str> NESTED NETWORK_NAMESPACE NOWAIT NULLS OJ OLD OPTIONAL ORDINALITY ORGANIZATION OTHERS PARTIAL PATH PERSIST PERSIST_ONLY PRECEDING PRIVILEGE_CHECKS_USER PROCESS
%token <str> RANDOM REFERENCE REQUIRE_ROW_FORMAT RESOURCE RESPECT RESTART RETAIN REUSE ROLE SECONDARY SECONDARY_ENGINE SECONDARY_ENGINE_ATTRIBUTE SECONDARY_LOAD SECONDARY_UNLOAD SIMPLE SKIP SRID
%token <str> THREAD_PRIORITY TIES UNBOUNDED VCPU VISIBLE RETURNING

// Explain tokens
%token <str> FORMAT TREE TRADITIONAL

// Lock type tokens
%token <str> LOCAL LOW_PRIORITY

// Flush tokens
%token <str> NO_WRITE_TO_BINLOG LOGS ERROR GENERAL HOSTS OPTIMIZER_COSTS USER_RESOURCES SLOW CHANNEL RELAY EXPORT

// ast.TableOptions tokens
%token <str> AVG_ROW_LENGTH CONNECTION CHECKSUM DELAY_KEY_WRITE ENCRYPTION INSERT_METHOD MAX_ROWS MIN_ROWS PACK_KEYS PASSWORD
%token <str> FIXED DYNAMIC COMPRESSED REDUNDANT COMPACT ROW_FORMAT STATS_AUTO_RECALC STATS_PERSISTENT STATS_SAMPLE_PAGES STORAGE MEMORY DISK

%type <statement> command
%type <selStmt> query_expression_parens query_expression query_expression_body select_statement query_primary select_stmt_with_into
%type <statement> explain_statement explainable_statement
%type <statement> prepare_statement
%type <statement> execute_statement deallocate_statement
%type <statement> stream_statement vstream_statement insert_statement update_statement delete_statement set_statement set_transaction_statement
%type <statement> create_statement alter_statement rename_statement drop_statement truncate_statement flush_statement do_statement comment_statement
%type <statement> copy_statement
%type <with> with_clause_opt with_clause
%type <cte> common_table_expr
%type <ctes> with_list
%type <schemaName> schema_name
%type <alterSchema> alter_schema_prefix
%type <createSequence> create_sequence_prefix
%type <alterSequence> alter_sequence_prefix
%type <renameTablePairs> rename_list
%type <alterOptions> alter_schema_commands_list
%type <sequenceName> sequence_name
%type <createTable> create_table_prefix
%type <alterTable> alter_table_prefix
%type <alterOption> alter_option alter_commands_modifier
%type <alterOptions> alter_options alter_commands_list alter_commands_modifier_list
%type <alterTable> create_index_prefix
%type <createDatabase> create_database_prefix
%type <alterDatabase> alter_database_prefix
%type <databaseOption> collate character_set encryption
%type <databaseOptions> create_options create_options_opt
%type <boolean> default_optional first_opt jt_exists_opt jt_path_opt
%type <statement> analyze_statement show_statement use_statement other_statement
%type <statement> begin_statement commit_statement rollback_statement savepoint_statement release_statement load_statement
%type <statement> lock_statement unlock_statement call_statement
%type <strs> comment_opt comment_list
%type <str> wild_opt check_option_opt cascade_or_local_opt restrict_or_cascade_opt
%type <explainType> explain_format_opt
%type <trimType> trim_type
%type <insertAction> insert_or_replace
%type <str> explain_synonyms
%type <intervalType> interval_time_stamp interval
%type <str> cache_opt separator_opt flush_option for_channel_opt
%type <matchExprOption> match_option
%type <boolean> distinct_opt union_op replace_opt local_opt only_opt
%type <selectExprs> select_expression_list select_expression_list_opt
%type <selectExpr> select_expression
%type <strs> select_options flush_option_list
%type <str> select_option security_view security_view_opt
%type <str> generated_always_opt user_username address_opt
%type <definer> definer_opt user
%type <expr> expression signed_literal signed_literal_or_null null_as_literal now_or_signed_literal signed_literal bit_expr simple_expr literal NUM_literal text_literal text_literal_or_arg bool_pri literal_or_null now predicate tuple_expression
%type <tableExprs> from_opt table_references from_clause
%type <tableExpr> table_reference table_factor join_table json_table_function
%type <jtColumnDefinition> jt_column
%type <jtColumnList> jt_columns_clause columns_list
%type <jtOnResponse> on_error on_empty json_on_response
%type <joinCondition> join_condition join_condition_opt on_expression_opt
%type <tableNames> table_name_list delete_table_list view_name_list
%type <joinType> inner_join outer_join straight_join natural_join
%type <tableName> table_name into_table_name delete_table_name
%type <aliasedTableName> aliased_table_name
%type <indexHint> index_hint
%type <indexHintForType> index_hint_for_opt
%type <indexHints> index_hint_list index_hint_list_opt
%type <expr> where_expression_opt
%type <boolVal> boolean_value
%type <comparisonExprOperator> compare
%type <ins> insert_data
%type <expr> num_val
%type <expr> function_call_keyword function_call_nonkeyword function_call_generic function_call_conflict
%type <isExprOperator> is_suffix
%type <colTuple> col_tuple
%type <exprs> expression_list expression_list_opt
%type <values> tuple_list
%type <valTuple> row_tuple tuple_or_empty
%type <subquery> subquery
%type <derivedTable> derived_table
%type <colName> column_name after_opt
%type <whens> when_expression_list
%type <when> when_expression
%type <expr> expression_opt else_expression_opt
%type <exprs> group_by_opt
%type <expr> having_opt
%type <orderBy> order_by_opt order_list order_by_clause
%type <order> order
%type <orderDirection> asc_desc_opt
%type <limit> limit_opt limit_clause
%type <columnTypeOptions> column_attribute_list_opt generated_column_attribute_list_opt
%type <lock> locking_clause
%type <selectInto> into_clause
%type <columns> ins_column_list column_list at_id_list column_list_opt index_list execute_statement_list_opt
%type <updateExprs> on_dup_opt
%type <updateExprs> update_list
%type <setExprs> set_list
%type <str> charset_or_character_set charset_or_character_set_or_names
%type <updateExpr> update_expression
%type <setExpr> set_expression
%type <characteristic> transaction_char
%type <characteristics> transaction_chars
%type <isolationLevel> isolation_level
%type <str> for_from from_or_on
%type <str> default_opt
%type <ignore> ignore_opt
%type <str> columns_or_fields extended_opt storage_opt
%type <showFilter> like_or_where_opt
%type <boolean> exists_opt not_exists_opt enforced enforced_opt temp_opt full_opt
%type <empty> to_opt
%type <str> reserved_keyword non_reserved_keyword non_reserved_keyword_sql2023
%type <colIdent> sql_id reserved_sql_id col_alias as_ci_opt
%type <schemaIdent> schema_id
%type <sequenceIdent> sequence_id
%type <expr> charset_value
%type <tableIdent> table_id reserved_table_id table_alias as_opt_id table_id_opt from_database_opt
%type <empty> as_opt work_opt savepoint_opt
%type <empty> skip_to_end ddl_skip_to_end
%type <str> charset
%type <sequenceIdent> reserved_sequence_id
%type <scope> set_session_or_global
%type <convertType> convert_type returning_type_opt convert_type_weight_string
%type <columnType> column_type
%type <columnType> int_type decimal_type numeric_type time_type char_type spatial_type
%type <literal> length_opt varying_opt with_timezone_opt
%type <expr> func_datetime_precision
%type <columnCharset> charset_opt
%type <str> collate_opt
%type <boolean> binary_opt
%type <LengthScaleOption> float_length_opt decimal_length_opt
%type <boolean> unsigned_opt zero_fill_opt
%type <strs> enum_values
%type <columnDefinition> column_definition
%type <columnDefinitions> column_definition_list
%type <indexDefinition> index_definition
%type <constraintDefinition> constraint_definition check_constraint_definition
%type <str> index_or_key index_symbols from_or_in index_or_key_opt
%type <str> name_opt constraint_name_opt
%type <str> equal_opt
%type <tableSpec> table_spec table_column_list
%type <optLike> create_like
%type <str> table_opt_value
%type <tableOption> table_option
%type <tableOptions> table_option_list table_option_list_opt space_separated_table_option_list
%type <indexInfo> index_info
%type <indexColumn> index_column
%type <indexColumns> index_column_list
%type <indexOption> index_option using_index_type
%type <indexOptions> index_option_list index_option_list_opt using_opt
%type <constraintInfo> constraint_info check_constraint_info
%type <vindexParam> vindex_param
%type <sequenceSpec> sequence_spec
%type <vindexParams> vindex_param_list vindex_params_opt
%type <jsonPathParam> json_path_param
%type <jsonPathParams> json_path_param_list json_path_param_list_opt
%type <jsonObjectParam> json_object_param
%type <jsonObjectParams> json_object_param_list json_object_param_opt
%type <colIdent> id_or_var vindex_type vindex_type_opt id_or_var_opt
%type <str> database column_opt insert_method_options row_format_options
%type <referenceAction> fk_reference_action fk_on_delete fk_on_update
%type <matchAction> fk_match fk_match_opt fk_match_action
%type <tableAndLockTypes> lock_table_list
%type <tableAndLockType> lock_table
%type <lockType> lock_type
%type <empty> session_or_local_opt
%type <columnStorage> column_storage
%type <columnFormat> column_format
%type <colKeyOpt> keys
%type <referenceDefinition> reference_definition reference_definition_opt
%type <str> underscore_charsets
%type <copyFromSource> copy_from
%type <copyToTarget> copy_to
%type <copyOptions> copy_option_list copy_option_list_opt
%type <copyOption> copy_option
%type <str> error_action
%start any_command

%%

any_command:
  command semicolon_opt
  {
    setParseTree(psqlex, $1)
  }

semicolon_opt:
/*empty*/ {}
| ';' {}

command:
  select_statement
  {
    $$ = $1
  }
| comment_statement
  {
    $$ = $1
  }
| stream_statement
| vstream_statement
| insert_statement
| update_statement
| delete_statement
| set_statement
| set_transaction_statement
| create_statement
| alter_statement
| rename_statement
| drop_statement
| truncate_statement
| analyze_statement
| show_statement
| use_statement
| begin_statement
| commit_statement
| rollback_statement
| savepoint_statement
| release_statement
| explain_statement
| other_statement
| flush_statement
| do_statement
| load_statement
| lock_statement
| unlock_statement
| call_statement
| prepare_statement
| execute_statement
| deallocate_statement
| copy_statement
| /*empty*/
  {
    setParseTree(psqlex, nil)
  }

id_or_var:
  ID
  {
    $$ = ast.NewColIdentWithAt(string($1), ast.NoAt)
  }
| AT_ID
  {
    $$ = ast.NewColIdentWithAt(string($1), ast.SingleAt)
  }
| AT_AT_ID
  {
    $$ = ast.NewColIdentWithAt(string($1), ast.DoubleAt)
  }

id_or_var_opt:
  {
    $$ = ast.NewColIdentWithAt("", ast.NoAt)
  }
| id_or_var
  {
    $$ = $1
  }

do_statement:
  DO expression_list
  {
    $$ = &ast.OtherAdmin{}
  }

load_statement:
  LOAD DATA skip_to_end
  {
    $$ = &ast.Load{}
  }

with_clause:
  WITH with_list
  {
	$$ = &ast.With{Ctes: $2, Recursive: false}
  }
| WITH RECURSIVE with_list
  {
	$$ = &ast.With{Ctes: $3, Recursive: true}
  }

with_clause_opt:
  {
    $$ = nil
  }
 | with_clause
 {
 	$$ = $1
 }

with_list:
  with_list ',' common_table_expr
  {
	$$ = append($1, $3)
  }
| common_table_expr
  {
	$$ = []*ast.CommonTableExpr{$1}
  }

common_table_expr:
  table_id column_list_opt AS subquery
  {
	$$ = &ast.CommonTableExpr{TableID: $1, Columns: $2, Subquery: $4}
  }

query_expression_parens:
  openb query_expression_parens closeb
  {
  	$$ = $2
  }
| openb query_expression closeb
  {
     $$ = $2
  }
| openb query_expression locking_clause closeb
  {
    ast.SetLockInSelect($2, $3)
    $$ = $2
  }

// TODO; (Manan, Ritwiz) : ast.Use this in create, insert statements
//query_expression_or_parens:
//	query_expression
//	{
//		$$ = $1
//	}
//	| query_expression locking_clause
//	{
//		ast.SetLockInSelect($1, $2)
//		$$ = $1
//	}
//	| query_expression_parens
//	{
//		$$ = $1
//	}

query_expression:
 query_expression_body order_by_opt limit_opt
  {
	$1.SetOrderBy($2)
	$1.SetLimit($3)
	$$ = $1
  }
| query_expression_parens limit_clause
  {
	$1.SetLimit($2)
	$$ = $1
  }
| query_expression_parens order_by_clause limit_opt
  {
	$1.SetOrderBy($2)
	$1.SetLimit($3)
	$$ = $1
  }
| with_clause query_expression_body order_by_opt limit_opt
  {
  	$2.SetWith($1)
		$2.SetOrderBy($3)
		$2.SetLimit($4)
		$$ = $2
  }
| with_clause query_expression_parens limit_clause
  {
  	$2.SetWith($1)
		$2.SetLimit($3)
		$$ = $2
  }
| with_clause query_expression_parens order_by_clause limit_opt
  {
  	$2.SetWith($1)
		$2.SetOrderBy($3)
		$2.SetLimit($4)
		$$ = $2
  }
| with_clause query_expression_parens
  {
	$2.SetWith($1)
  }
| SELECT comment_opt cache_opt NEXT num_val for_from table_name
  {
	$$ = ast.NewSelect(ast.Comments($2), ast.SelectExprs{&ast.Nextval{Expr: $5}}, []string{$3}/*options*/, nil, ast.TableExprs{&ast.AliasedTableExpr{Expr: $7}}, nil/*where*/, nil/*groupBy*/, nil/*having*/)
  }

query_expression_body:
 query_primary
  {
	$$ = $1
  }
| query_expression_body union_op query_primary
  {
 	$$ = &ast.Union{Left: $1, Distinct: $2, Right: $3}
  }
| query_expression_parens union_op query_primary
  {
	$$ = &ast.Union{Left: $1, Distinct: $2, Right: $3}
  }
| query_expression_body union_op query_expression_parens
  {
  	$$ = &ast.Union{Left: $1, Distinct: $2, Right: $3}
  }
| query_expression_parens union_op query_expression_parens
  {
	$$ = &ast.Union{Left: $1, Distinct: $2, Right: $3}
  }

select_statement:
query_expression
  {
	$$ = $1
  }
| query_expression locking_clause
  {
	ast.SetLockInSelect($1, $2)
	$$ = $1
  }
| query_expression_parens
  {
	$$ = $1
  }
| select_stmt_with_into
  {
	$$ = $1
  }

select_stmt_with_into:
  openb select_stmt_with_into closeb
  {
	$$ = $2;
  }
| query_expression into_clause
  {
	$1.SetInto($2)
	$$ = $1
  }
| query_expression_parens into_clause
  {
 	$1.SetInto($2)
	$$ = $1
  }

stream_statement:
  STREAM comment_opt select_expression FROM table_name
  {
    $$ = &ast.Stream{Comments: ast.Comments($2).Parsed(), SelectExpr: $3, Table: $5}
  }

vstream_statement:
  VSTREAM comment_opt select_expression FROM table_name where_expression_opt limit_opt
  {
    $$ = &ast.VStream{Comments: ast.Comments($2).Parsed(), SelectExpr: $3, Table: $5, Where: ast.NewWhere(ast.WhereClause, $6), Limit: $7}
  }

// query_primary is an unparenthesized SELECT with no order by clause or beyond.
query_primary:
//  1         2            3              4                    5             6                7           8
  SELECT comment_opt select_options select_expression_list from_opt where_expression_opt group_by_opt having_opt
  {
    $$ = ast.NewSelect(ast.Comments($2), $4/*SelectExprs*/, $3/*options*/, nil, $5/*from*/, ast.NewWhere(ast.WhereClause, $6), ast.GroupBy($7), ast.NewWhere(ast.HavingClause, $8))
  }

copy_statement:
  COPY comment_opt table_name column_list_opt FROM copy_from copy_option_list_opt where_expression_opt
  {
    $$ = &ast.CopyFrom{Comments: ast.Comments($2).Parsed(), Table: $3, Columns: $4, From: $6, With: $7, Where: $8}
  }
| COPY comment_opt table_name column_list_opt TO copy_to copy_option_list_opt
  {
    $$ = &ast.CopyTo{Comments: ast.Comments($2).Parsed(), Table: $3, Columns: $4, To: $6, With: $7}
  }
| COPY comment_opt query_expression TO copy_to copy_option_list_opt
  {
    $$ = &ast.CopyTo{Comments: ast.Comments($2).Parsed(), Query: $3, To: $5, With: $6}
  }

copy_from:
  STRING
  {
    $$ = ast.CopyFromSource{Type: ast.CopyFromFile, V: $1}
  }
| PROGRAM STRING
  {
    $$ = ast.CopyFromSource{Type: ast.CopyFromProgram, V: $1}
  }
| STDIN
  {
    $$ = ast.CopyFromSource{Type: ast.CopyFromStdin}
  }

copy_to:
  STRING
  {
    $$ = ast.CopyToTarget{Type: ast.CopyToFile, V: $1}
  }
| PROGRAM STRING
  {
    $$ = ast.CopyToTarget{Type: ast.CopyToProgram, V: $1}
  }
| STDOUT
  {
    $$ = ast.CopyToTarget{Type: ast.CopyToStdout}
  }

copy_option_list_opt:
  {
    $$ = nil
  }
| WITH '(' copy_option_list ')'
  {
    $$ = $3
  }
| '(' copy_option_list ')'
  {
    $$ = $2
  }

copy_option_list:
  copy_option
  {
    $$ = []ast.CopyOption{$1}
  }
| copy_option_list ',' copy_option
  {
    $$ = append($$, $3)
  }

copy_option:
  FORMAT ID
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionFormat, Value: $2}
  }
| FREEZE boolean_value
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionFreeze, Value: $2.String()}
  }
| DELIMITER STRING
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionDelimiter, Value: $2}
  }
| NULL STRING
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionNull, Value: $2}
  }
| DEFAULT STRING
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionDefault, Value: $2}
  }
| HEADER boolean_value
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionHeader, Value: $2.String()}
  }
| HEADER MATCH
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionHeaderMatch}
  }
| QUOTE STRING
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionQuote, Value: $2}
  }
| ESCAPE STRING
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionEscape, Value: $2}
  }
| FORCE_QUOTE column_list_opt
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionForceQuote}
  }
| FORCE_NOT_NULL column_list_opt
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionForceNotNull}
  }
| FORCE_NULL column_list_opt
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionForceNull}
  }
| ON_ERROR error_action
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionOnError, Value: $2}
  }
| ENCODING STRING
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionEncoding, Value: $2}
  }
| LOG_VERBOSITY INTEGRAL
  {
    $$ = ast.CopyOption{Type: ast.CopyOptionHeaderLogVerbosity, Value: $2}
  }

error_action:
  {
    $$ = "stop"
  }
| STOP
  {
    $$ = "stop"
  }
| IGNORE
  {
    $$ = "ignore"
  }

insert_statement:
  insert_or_replace comment_opt ignore_opt into_table_name insert_data on_dup_opt
  {
    // insert_data returns a *ast.Insert pre-filled with Columns & Values
    ins := $5
    ins.Action = $1
    ins.Comments = ast.Comments($2).Parsed()
    ins.Ignore = $3
    ins.Table = $4
    ins.OnDup = ast.OnDup($6)
    $$ = ins
  }
| insert_or_replace comment_opt ignore_opt into_table_name SET update_list on_dup_opt
  {
    cols := make(ast.Columns, 0, len($6))
    vals := make(ast.ValTuple, 0, len($7))
    for _, updateList := range $6 {
      cols = append(cols, updateList.Name.Name)
      vals = append(vals, updateList.Expr)
    }
    $$ = &ast.Insert{Action: $1, Comments: ast.Comments($2).Parsed(), Ignore: $3, Table: $4, Columns: cols, Rows: ast.Values{vals}, OnDup: ast.OnDup($7)}
  }

insert_or_replace:
  INSERT
  {
    $$ = ast.InsertAct
  }
| REPLACE
  {
    $$ = ast.ReplaceAct
  }

update_statement:
  with_clause_opt UPDATE comment_opt ignore_opt table_references SET update_list where_expression_opt order_by_opt limit_opt
  {
    $$ = &ast.Update{With: $1, Comments: ast.Comments($3).Parsed(), Ignore: $4, TableExprs: $5, Exprs: $7, Where: ast.NewWhere(ast.WhereClause, $8), OrderBy: $9, Limit: $10}
  }

delete_statement:
  with_clause_opt DELETE comment_opt ignore_opt FROM table_name as_opt_id where_expression_opt order_by_opt limit_opt
  {
    $$ = &ast.Delete{With: $1, Comments: ast.Comments($3).Parsed(), Ignore: $4, TableExprs: ast.TableExprs{&ast.AliasedTableExpr{Expr:$6, As: $7}}, Where: ast.NewWhere(ast.WhereClause, $8), OrderBy: $9, Limit: $10}
  }
| with_clause_opt DELETE comment_opt ignore_opt FROM table_name_list USING table_references where_expression_opt
  {
    $$ = &ast.Delete{With: $1, Comments: ast.Comments($3).Parsed(), Ignore: $4, Targets: $6, TableExprs: $8, Where: ast.NewWhere(ast.WhereClause, $9)}
  }
| with_clause_opt DELETE comment_opt ignore_opt table_name_list from_or_using table_references where_expression_opt
  {
    $$ = &ast.Delete{With: $1, Comments: ast.Comments($3).Parsed(), Ignore: $4, Targets: $5, TableExprs: $7, Where: ast.NewWhere(ast.WhereClause, $8)}
  }
| with_clause_opt DELETE comment_opt ignore_opt delete_table_list from_or_using table_references where_expression_opt
  {
    $$ = &ast.Delete{With: $1, Comments: ast.Comments($3).Parsed(), Ignore: $4, Targets: $5, TableExprs: $7, Where: ast.NewWhere(ast.WhereClause, $8)}
  }

from_or_using:
  FROM {}
| USING {}

view_name_list:
  table_name
  {
    $$ = ast.TableNames{$1.ToViewName()}
  }
| view_name_list ',' table_name
  {
    $$ = append($$, $3.ToViewName())
  }

table_name_list:
  table_name
  {
    $$ = ast.TableNames{$1}
  }
| table_name_list ',' table_name
  {
    $$ = append($$, $3)
  }

delete_table_list:
  delete_table_name
  {
    $$ = ast.TableNames{$1}
  }
| delete_table_list ',' delete_table_name
  {
    $$ = append($$, $3)
  }

set_statement:
  SET comment_opt set_list
  {
    $$ = &ast.Set{Comments: ast.Comments($2).Parsed(), Exprs: $3}
  }

set_transaction_statement:
  SET comment_opt set_session_or_global TRANSACTION transaction_chars
  {
    $$ = &ast.SetTransaction{Comments: ast.Comments($2).Parsed(), Scope: $3, Characteristics: $5}
  }
| SET comment_opt TRANSACTION transaction_chars
  {
    $$ = &ast.SetTransaction{Comments: ast.Comments($2).Parsed(), Characteristics: $4, Scope: ast.ImplicitScope}
  }

transaction_chars:
  transaction_char
  {
    $$ = []ast.Characteristic{$1}
  }
| transaction_chars ',' transaction_char
  {
    $$ = append($$, $3)
  }

transaction_char:
  ISOLATION LEVEL isolation_level
  {
    $$ = $3
  }
| READ WRITE
  {
    $$ = ast.ReadWrite
  }
| READ ONLY
  {
    $$ = ast.ReadOnly
  }

isolation_level:
  REPEATABLE READ
  {
    $$ = ast.RepeatableRead
  }
| READ COMMITTED
  {
    $$ = ast.ReadCommitted
  }
| READ UNCOMMITTED
  {
    $$ = ast.ReadUncommitted
  }
| SERIALIZABLE
  {
    $$ = ast.Serializable
  }

set_session_or_global:
  SESSION
  {
    $$ = ast.SessionScope
  }
| GLOBAL
  {
    $$ = ast.GlobalScope
  }

create_statement:
  create_table_prefix table_spec
  {
    $1.TableSpec = $2
    $1.FullyParsed = true
    $$ = $1
  }
| create_table_prefix create_like
  {
    // Create table [name] like [name]
    $1.OptLike = $2
    $1.FullyParsed = true
    $$ = $1
  }
| create_index_prefix '(' index_column_list ')' index_option_list_opt
  {
    indexDef := $1.AlterOptions[0].(*ast.AddIndexDefinition).IndexDefinition
    indexDef.Columns = $3
    indexDef.Options = append(indexDef.Options,$5...)
    $1.FullyParsed = true
    $$ = $1
  }
| CREATE comment_opt replace_opt definer_opt security_view_opt VIEW table_name column_list_opt AS select_statement check_option_opt
  {
    $$ = &ast.CreateView{ViewName: $7.ToViewName(), Comments: ast.Comments($2).Parsed(), IsReplace:$3, Definer: $4 ,Security:$5, Columns:$8, Select: $10, CheckOption: $11 }
  }
| create_database_prefix create_options_opt
  {
    $1.FullyParsed = true
    $1.CreateOptions = $2
    $$ = $1
  }
| create_sequence_prefix sequence_spec
  {
    $1.SequenceSpec = $2
    $1.FullyParsed = true
    $$ = $1
  }

replace_opt:
  {
    $$ = false
  }
| OR REPLACE
  {
    $$ = true
  }

vindex_type_opt:
  {
    $$ = ast.NewColIdent("")
  }
| USING vindex_type
  {
    $$ = $2
  }

vindex_type:
  sql_id
  {
    $$ = $1
  }

vindex_params_opt:
  {
    var v []ast.VindexParam
    $$ = v
  }
| WITH vindex_param_list
  {
    $$ = $2
  }

vindex_param_list:
  vindex_param
  {
    $$ = make([]ast.VindexParam, 0, 4)
    $$ = append($$, $1)
  }
| vindex_param_list ',' vindex_param
  {
    $$ = append($$, $3)
  }

vindex_param:
  reserved_sql_id '=' table_opt_value
  {
    $$ = ast.VindexParam{Key: $1, Val: $3}
  }

json_object_param_opt:
  {
    $$ = nil
  }
| json_object_param_list
  {
    $$ = $1
  }

json_object_param_list:
  json_object_param
  {
    $$ = []*ast.JSONObjectParam{$1}
  }
| json_object_param_list ',' json_object_param
  {
    $$ = append($$, $3)
  }

json_object_param:
  expression ',' expression
  {
    $$ = &ast.JSONObjectParam{Key:$1, Value:$3}
  }

alter_schema_prefix:
  ALTER comment_opt schema_name
  {
    $$ = &ast.AlterSchema{Comments: ast.Comments($2).Parsed(), Schema: $3}
    setDDL(psqlex, $$)
  }

create_table_prefix:
  CREATE comment_opt temp_opt TABLE not_exists_opt table_name
  {
    $$ = &ast.CreateTable{Comments: ast.Comments($2).Parsed(), Table: $6, IfNotExists: $5, Temp: $3}
    setDDL(psqlex, $$)
  }

alter_table_prefix:
  ALTER comment_opt TABLE only_opt table_name
  {
    $$ = &ast.AlterTable{Comments: ast.Comments($2).Parsed(), Only: $4, Table: $5}
    setDDL(psqlex, $$)
  }

create_index_prefix:
  CREATE comment_opt INDEX id_or_var using_opt ON table_name
  {
    $$ = &ast.AlterTable{Table: $7, AlterOptions: []ast.AlterOption{&ast.AddIndexDefinition{IndexDefinition:&ast.IndexDefinition{Info: &ast.IndexInfo{Name:$4, Type:string($3)}, Options:$5}}}}
    setDDL(psqlex, $$)
  }
| CREATE comment_opt FULLTEXT INDEX id_or_var using_opt ON table_name
  {
    $$ = &ast.AlterTable{Table: $8, AlterOptions: []ast.AlterOption{&ast.AddIndexDefinition{IndexDefinition:&ast.IndexDefinition{Info: &ast.IndexInfo{Name:$5, Type:string($3)+" "+string($4), Fulltext:true}, Options:$6}}}}
    setDDL(psqlex, $$)
  }
| CREATE comment_opt SPATIAL INDEX id_or_var using_opt ON table_name
  {
    $$ = &ast.AlterTable{Table: $8, AlterOptions: []ast.AlterOption{&ast.AddIndexDefinition{IndexDefinition:&ast.IndexDefinition{Info: &ast.IndexInfo{Name:$5, Type:string($3)+" "+string($4), Spatial:true}, Options:$6}}}}
    setDDL(psqlex, $$)
  }
| CREATE comment_opt UNIQUE INDEX id_or_var using_opt ON table_name
  {
    $$ = &ast.AlterTable{Table: $8, AlterOptions: []ast.AlterOption{&ast.AddIndexDefinition{IndexDefinition:&ast.IndexDefinition{Info: &ast.IndexInfo{Name:$5, Type:string($3)+" "+string($4), Unique:true}, Options:$6}}}}
    setDDL(psqlex, $$)
  }

create_sequence_prefix:
  CREATE comment_opt sequence_name
  {
    $$ = &ast.CreateSequence{Comments: ast.Comments($2).Parsed(), Sequence: $3}
    setDDL(psqlex, $$)
  }

alter_sequence_prefix:
  ALTER comment_opt sequence_name
  {
    $$ = &ast.AlterSequence{Comments: ast.Comments($2).Parsed(), Sequence: $3}
    setDDL(psqlex, $$)
  }

create_database_prefix:
  CREATE comment_opt database comment_opt not_exists_opt table_id
  {
    $$ = &ast.CreateDatabase{Comments: ast.Comments($4).Parsed(), DBName: $6, IfNotExists: $5}
    setDDL(psqlex,$$)
  }

alter_database_prefix:
  ALTER comment_opt database
  {
    $$ = &ast.AlterDatabase{}
    setDDL(psqlex,$$)
  }

database:
  DATABASE

table_spec:
  '(' table_column_list ')' table_option_list_opt
  {
    $$ = $2
    $$.Options = $4
  }

create_options_opt:
  {
    $$ = nil
  }
| create_options
  {
    $$ = $1
  }

create_options:
  character_set
  {
    $$ = []ast.DatabaseOption{$1}
  }
| collate
  {
    $$ = []ast.DatabaseOption{$1}
  }
| encryption
  {
    $$ = []ast.DatabaseOption{$1}
  }
| create_options collate
  {
    $$ = append($1,$2)
  }
| create_options character_set
  {
    $$ = append($1,$2)
  }
| create_options encryption
  {
    $$ = append($1,$2)
  }

default_optional:
  /* empty */ %prec LOWER_THAN_CHARSET
  {
    $$ = false
  }
| DEFAULT
  {
    $$ = true
  }

character_set:
  default_optional charset_or_character_set equal_opt id_or_var
  {
    $$ = ast.DatabaseOption{Type:ast.CharacterSetType, Value:($4.String()), IsDefault:$1}
  }
| default_optional charset_or_character_set equal_opt STRING
  {
    $$ = ast.DatabaseOption{Type:ast.CharacterSetType, Value:(sql_types.EncodeStringSQL($4)), IsDefault:$1}
  }

collate:
  default_optional COLLATE equal_opt id_or_var
  {
    $$ = ast.DatabaseOption{Type:ast.CollateType, Value:($4.String()), IsDefault:$1}
  }
| default_optional COLLATE equal_opt STRING
  {
    $$ = ast.DatabaseOption{Type:ast.CollateType, Value:(sql_types.EncodeStringSQL($4)), IsDefault:$1}
  }

encryption:
  default_optional ENCRYPTION equal_opt id_or_var
  {
    $$ = ast.DatabaseOption{Type:ast.EncryptionType, Value:($4.String()), IsDefault:$1}
  }
| default_optional ENCRYPTION equal_opt STRING
  {
    $$ = ast.DatabaseOption{Type:ast.EncryptionType, Value:(sql_types.EncodeStringSQL($4)), IsDefault:$1}
  }

create_like:
  LIKE table_name
  {
    $$ = &ast.OptLike{LikeTable: $2}
  }
| '(' LIKE table_name ')'
  {
    $$ = &ast.OptLike{LikeTable: $3}
  }

column_definition_list:
  column_definition
  {
    $$ = []*ast.ColumnDefinition{$1}
  }
| column_definition_list ',' column_definition
  {
    $$ = append($1,$3)
  }

table_column_list:
  column_definition
  {
    $$ = &ast.TableSpec{}
    $$.AddColumn($1)
  }
| check_constraint_definition
  {
    $$ = &ast.TableSpec{}
    $$.AddConstraint($1)
  }
| table_column_list ',' column_definition
  {
    $$.AddColumn($3)
  }
| table_column_list ',' column_definition check_constraint_definition
  {
    $$.AddColumn($3)
    $$.AddConstraint($4)
  }
| table_column_list ',' index_definition
  {
    $$.AddIndex($3)
  }
| table_column_list ',' constraint_definition
  {
    $$.AddConstraint($3)
  }
| table_column_list ',' check_constraint_definition
  {
    $$.AddConstraint($3)
  }

// collate_opt has to be in the first rule so that we don\'t have a shift reduce conflict when seeing a COLLATE
// with column_attribute_list_opt. Always shifting there would have meant that we would have always ended up using the
// second rule in the grammar whenever COLLATE was specified.
// We now have a shift reduce conflict between COLLATE and collate_opt. Shifting there is fine. Essentially, we have
// postponed the decision of which rule to use until we have consumed the COLLATE id/string tokens.
column_definition:
  sql_id column_type collate_opt column_attribute_list_opt reference_definition_opt
  {
    $2.Options = $4
    if $2.Options.Collate == "" {
    	$2.Options.Collate = $3
    }
    $2.Options.Reference = $5
    $$ = &ast.ColumnDefinition{Name: $1, Type: $2}
  }
| sql_id column_type collate_opt generated_always_opt AS '(' expression ')' generated_column_attribute_list_opt reference_definition_opt
  {
    $2.Options = $9
    $2.Options.As = $7
    $2.Options.Reference = $10
    $2.Options.Collate = $3
    $$ = &ast.ColumnDefinition{Name: $1, Type: $2}
  }

generated_always_opt:
  {
    $$ = ""
  }
|  GENERATED ALWAYS
  {
    $$ = ""
  }

// There is a shift reduce conflict that arises here because UNIQUE and KEY are column_type_option and so is UNIQUE KEY.
// So in the state "column_type_options UNIQUE. KEY" there is a shift-reduce conflict(resovled by "%rigth <str> UNIQUE KEY").
// This has been added to emulate what PostgresQL does. The previous architecture was such that the order of the column options
// was specific (as stated in the PostgresQL guide) and did not accept arbitrary order options. For example NOT NULL DEFAULT 1 and not DEFAULT 1 NOT NULL
column_attribute_list_opt:
  {
    $$ = &ast.ColumnTypeOptions{Null: nil, Default: nil, OnUpdate: nil, Autoincrement: false, KeyOpt: ast.ColKeyNone, Comment: nil, As: nil, Invisible: nil, Format: ast.UnspecifiedFormat, EngineAttribute: nil, SecondaryEngineAttribute: nil }
  }
| column_attribute_list_opt NULL
  {
    val := true
    $1.Null = &val
    $$ = $1
  }
| column_attribute_list_opt NOT NULL
  {
    val := false
    $1.Null = &val
    $$ = $1
  }
| column_attribute_list_opt DEFAULT openb expression closeb
  {
	$1.Default = $4
	$$ = $1
  }
| column_attribute_list_opt DEFAULT now_or_signed_literal
  {
    $1.Default = $3
    $$ = $1
  }
| column_attribute_list_opt ON UPDATE function_call_nonkeyword
  {
    $1.OnUpdate = $4
    $$ = $1
  }
| column_attribute_list_opt AUTO_INCREMENT
  {
    $1.Autoincrement = true
    $$ = $1
  }
| column_attribute_list_opt COMMENT_KEYWORD STRING
  {
    $1.Comment = ast.NewStrLiteral($3)
    $$ = $1
  }
| column_attribute_list_opt keys
  {
    $1.KeyOpt = $2
    $$ = $1
  }
| column_attribute_list_opt COLLATE STRING
  {
    $1.Collate = sql_types.EncodeStringSQL($3)
  }
| column_attribute_list_opt COLLATE id_or_var
  {
    $1.Collate = string($3.String())
    $$ = $1
  }
| column_attribute_list_opt COLUMN_FORMAT column_format
  {
    $1.Format = $3
  }
| column_attribute_list_opt SRID INTEGRAL
  {
    $1.SRID = ast.NewIntLiteral($3)
    $$ = $1
  }
| column_attribute_list_opt VISIBLE
  {
    val := false
    $1.Invisible = &val
    $$ = $1
  }
| column_attribute_list_opt INVISIBLE
  {
    val := true
    $1.Invisible = &val
    $$ = $1
  }

column_format:
  FIXED
{
  $$ = ast.FixedFormat
}
| DYNAMIC
{
  $$ = ast.DynamicFormat
}
| DEFAULT
{
  $$ = ast.DefaultFormat
}

column_storage:
  VIRTUAL
{
  $$ = ast.VirtualStorage
}
| STORED
{
  $$ = ast.StoredStorage
}

generated_column_attribute_list_opt:
  {
    $$ = &ast.ColumnTypeOptions{}
  }
| generated_column_attribute_list_opt column_storage
  {
    $1.Storage = $2
    $$ = $1
  }
| generated_column_attribute_list_opt NULL
  {
    val := true
    $1.Null = &val
    $$ = $1
  }
| generated_column_attribute_list_opt NOT NULL
  {
    val := false
    $1.Null = &val
    $$ = $1
  }
| generated_column_attribute_list_opt COMMENT_KEYWORD STRING
  {
    $1.Comment = ast.NewStrLiteral($3)
    $$ = $1
  }
| generated_column_attribute_list_opt keys
  {
    $1.KeyOpt = $2
    $$ = $1
  }
| generated_column_attribute_list_opt VISIBLE
  {
    val := false
    $1.Invisible = &val
    $$ = $1
  }
| generated_column_attribute_list_opt INVISIBLE
  {
    val := true
    $1.Invisible = &val
    $$ = $1
  }

now_or_signed_literal:
now
  {
  	$$ = $1
  }
| signed_literal_or_null

now:
CURRENT_TIMESTAMP func_datetime_precision
  {
    $$ = &ast.CurTimeFuncExpr{Name:ast.NewColIdent("current_timestamp"), Fsp: $2}
  }
| LOCALTIME func_datetime_precision
  {
    $$ = &ast.CurTimeFuncExpr{Name:ast.NewColIdent("localtime"), Fsp: $2}
  }
| LOCALTIMESTAMP func_datetime_precision
  {
    $$ = &ast.CurTimeFuncExpr{Name:ast.NewColIdent("localtimestamp"), Fsp: $2}
  }
| UTC_TIMESTAMP func_datetime_precision
  {
    $$ = &ast.CurTimeFuncExpr{Name:ast.NewColIdent("utc_timestamp"), Fsp:$2}
  }
| NOW func_datetime_precision
  {
    $$ = &ast.CurTimeFuncExpr{Name:ast.NewColIdent("now"), Fsp: $2}
  }

sequence_spec:
  START WITH INTEGRAL INCREMENT BY INTEGRAL NO MINVALUE NO MAXVALUE CACHE INTEGRAL
  {
    $$ = &ast.SequenceSpec{StartWith: ast.IntRef($3), IncrementBy: ast.IntRef($6)}
  }

signed_literal_or_null:
signed_literal
| null_as_literal

 null_as_literal:
NULL
 {
    $$ = &ast.NullVal{}
 }

 signed_literal:
 literal
| '+' NUM_literal
   {
 	$$= $2
   }
| '-' NUM_literal
   {
   	$$ = &ast.UnaryExpr{Operator: ast.UMinusOp, Expr: $2}
   }

literal:
text_literal
  {
   $$= $1
  }
| NUM_literal
  {
  	$$= $1
  }
| boolean_value
  {
  	$$ = $1
  }
| HEX
  {
	$$ = ast.NewHexLiteral($1)
  }
| HEXNUM
  {
  	$$ = ast.NewHexNumLiteral($1)
  }
| BIT_LITERAL
  {
	$$ = ast.NewBitLiteral($1)
  }
| VALUE_ARG
  {
    $$ = ast.NewArgument($1[1:])
    bindVariable(psqlex, $1[1:])
  }
| underscore_charsets  BIT_LITERAL %prec UNARY
  {
  	$$ = &ast.IntroducerExpr{CharacterSet: $1, Expr: ast.NewBitLiteral($2)}
  }
| underscore_charsets HEXNUM %prec UNARY
  {
  	$$ = &ast.IntroducerExpr{CharacterSet: $1, Expr: ast.NewHexNumLiteral($2)}
  }
| underscore_charsets HEX %prec UNARY
  {
   	$$ = &ast.IntroducerExpr{CharacterSet: $1, Expr: ast.NewHexLiteral($2)}
  }
| underscore_charsets column_name %prec UNARY
  {
    $$ = &ast.IntroducerExpr{CharacterSet: $1, Expr: $2}
  }
| underscore_charsets VALUE_ARG %prec UNARY
  {
    bindVariable(psqlex, $2[1:])
    $$ = &ast.IntroducerExpr{CharacterSet: $1, Expr: ast.NewArgument($2[1:])}
  }

underscore_charsets:
  UNDERSCORE_ARMSCII8
  {
    $$ = ast.Armscii8Str
  }
| UNDERSCORE_ASCII
  {
    $$ = ast.ASCIIStr
  }
| UNDERSCORE_BIG5
  {
    $$ = ast.Big5Str
  }
| UNDERSCORE_BINARY
  {
    $$ = ast.UBinaryStr
  }
| UNDERSCORE_CP1250
  {
    $$ = ast.Cp1250Str
  }
| UNDERSCORE_CP1251
  {
    $$ = ast.Cp1251Str
  }
| UNDERSCORE_CP1256
  {
    $$ = ast.Cp1256Str
  }
| UNDERSCORE_CP1257
  {
    $$ = ast.Cp1257Str
  }
| UNDERSCORE_CP850
  {
    $$ = ast.Cp850Str
  }
| UNDERSCORE_CP852
  {
    $$ = ast.Cp852Str
  }
| UNDERSCORE_CP866
  {
    $$ = ast.Cp866Str
  }
| UNDERSCORE_CP932
  {
    $$ = ast.Cp932Str
  }
| UNDERSCORE_DEC8
  {
    $$ = ast.Dec8Str
  }
| UNDERSCORE_EUCJPMS
  {
    $$ = ast.EucjpmsStr
  }
| UNDERSCORE_EUCKR
  {
    $$ = ast.EuckrStr
  }
| UNDERSCORE_GB18030
  {
    $$ = ast.Gb18030Str
  }
| UNDERSCORE_GB2312
  {
    $$ = ast.Gb2312Str
  }
| UNDERSCORE_GBK
  {
    $$ = ast.GbkStr
  }
| UNDERSCORE_GEOSTD8
  {
    $$ = ast.Geostd8Str
  }
| UNDERSCORE_GREEK
  {
    $$ = ast.GreekStr
  }
| UNDERSCORE_HEBREW
  {
    $$ = ast.HebrewStr
  }
| UNDERSCORE_HP8
  {
    $$ = ast.Hp8Str
  }
| UNDERSCORE_KEYBCS2
  {
    $$ = ast.Keybcs2Str
  }
| UNDERSCORE_KOI8R
  {
    $$ = ast.Koi8rStr
  }
| UNDERSCORE_KOI8U
  {
    $$ = ast.Koi8uStr
  }
| UNDERSCORE_LATIN1
  {
    $$ = ast.Latin1Str
  }
| UNDERSCORE_LATIN2
  {
    $$ = ast.Latin2Str
  }
| UNDERSCORE_LATIN5
  {
    $$ = ast.Latin5Str
  }
| UNDERSCORE_LATIN7
  {
    $$ = ast.Latin7Str
  }
| UNDERSCORE_MACCE
  {
    $$ = ast.MacceStr
  }
| UNDERSCORE_MACROMAN
  {
    $$ = ast.MacromanStr
  }
| UNDERSCORE_SJIS
  {
    $$ = ast.SjisStr
  }
| UNDERSCORE_SWE7
  {
    $$ = ast.Swe7Str
  }
| UNDERSCORE_TIS620
  {
    $$ = ast.Tis620Str
  }
| UNDERSCORE_UCS2
  {
    $$ = ast.Ucs2Str
  }
| UNDERSCORE_UJIS
  {
    $$ = ast.UjisStr
  }
| UNDERSCORE_UTF16
  {
    $$ = ast.Utf16Str
  }
| UNDERSCORE_UTF16LE
  {
    $$ = ast.Utf16leStr
  }
| UNDERSCORE_UTF32
  {
    $$ = ast.Utf32Str
  }
| UNDERSCORE_UTF8
  {
    $$ = ast.Utf8Str
  }
| UNDERSCORE_UTF8MB4
  {
    $$ = ast.Utf8mb4Str
  }
| UNDERSCORE_UTF8MB3
  {
    $$ = ast.Utf8Str
  }

literal_or_null:
literal
| null_as_literal

NUM_literal:
INTEGRAL
  {
    $$ = ast.NewIntLiteral($1)
  }
| FLOAT
  {
    $$ = ast.NewFloatLiteral($1)
  }
| DECIMAL
  {
    $$ = ast.NewDecimalLiteral($1)
  }

text_literal:
STRING
  {
	$$ = ast.NewStrLiteral($1)
  }
| NCHAR_STRING
  {
	$$ = &ast.UnaryExpr{Operator: ast.NStringOp, Expr: ast.NewStrLiteral($1)}
  }
 | underscore_charsets STRING %prec UNARY
   {
   	$$ = &ast.IntroducerExpr{CharacterSet: $1, Expr: ast.NewStrLiteral($2)}
   }

text_literal_or_arg:
  text_literal
  {
    $$ = $1
  }
| VALUE_ARG
  {
    $$ = ast.NewArgument($1[1:])
    bindVariable(psqlex, $1[1:])
  }

keys:
  PRIMARY KEY
  {
    $$ = ast.ColKeyPrimary
  }
| UNIQUE
  {
    $$ = ast.ColKeyUnique
  }
| UNIQUE KEY
  {
    $$ = ast.ColKeyUniqueKey
  }
| KEY
  {
    $$ = ast.ColKey
  }

column_type:
  numeric_type unsigned_opt zero_fill_opt
  {
    $$ = $1
    $$.Unsigned = $2
    $$.Zerofill = $3
  }
| char_type
| time_type
| spatial_type

numeric_type:
  int_type length_opt
  {
    $$ = $1
    $$.Length = $2
  }
| decimal_type
  {
    $$ = $1
  }

int_type:
  BIT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| BOOL
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| BOOLEAN
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| TINYINT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| SMALLINT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| MEDIUMINT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| INT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| INTEGER
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| BIGINT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }

decimal_type:
REAL float_length_opt
  {
    $$ = ast.ColumnType{Type: string($1)}
    $$.Length = $2.Length
    $$.Scale = $2.Scale
  }
| DOUBLE float_length_opt
  {
    $$ = ast.ColumnType{Type: string($1)}
    $$.Length = $2.Length
    $$.Scale = $2.Scale
  }
| DOUBLE PRECISION
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| FLOAT_TYPE float_length_opt
  {
    $$ = ast.ColumnType{Type: string($1)}
    $$.Length = $2.Length
    $$.Scale = $2.Scale
  }
| DECIMAL_TYPE decimal_length_opt
  {
    $$ = ast.ColumnType{Type: string($1)}
    $$.Length = $2.Length
    $$.Scale = $2.Scale
  }
| NUMERIC decimal_length_opt
  {
    $$ = ast.ColumnType{Type: string($1)}
    $$.Length = $2.Length
    $$.Scale = $2.Scale
  }
| DECIMAL decimal_length_opt
  {
    $$ = ast.ColumnType{Type: string($1)}
    $$.Length = $2.Length
    $$.Scale = $2.Scale
  }

time_type:
  DATE
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| TIME length_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $2}
  }
| TIMESTAMP with_timezone_opt length_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $3}
  }
| INTERVAL length_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $2}
  }

char_type:
  CHAR length_opt charset_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $2, Charset: $3}
  }
| CHAR length_opt BYTE
  {
    // CHAR BYTE is an alias for binary. See also:
    // https://dev.psql.com/doc/refman/8.0/en/string-type-syntax.html
    $$ = ast.ColumnType{Type: "binary", Length: $2}
  }
| CHARACTER varying_opt length_opt charset_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $3, Charset: $4}
  }
| VARCHAR length_opt charset_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $2, Charset: $3}
  }
| BINARY length_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $2}
  }
| BYTEA length_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $2}
  }
| VARBINARY length_opt
  {
    $$ = ast.ColumnType{Type: string($1), Length: $2}
  }
| TEXT charset_opt
  {
    $$ = ast.ColumnType{Type: string($1), Charset: $2}
  }
| JSON
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| ENUM '(' enum_values ')' charset_opt
  {
    $$ = ast.ColumnType{Type: string($1), EnumValues: $3, Charset: $5}
  }
// need set_values / SetValues ?
| SET '(' enum_values ')' charset_opt
  {
    $$ = ast.ColumnType{Type: string($1), EnumValues: $3, Charset: $5}
  }

spatial_type:
  GEOMETRY
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| POINT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| LINESTRING
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| POLYGON
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| GEOMETRYCOLLECTION
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| MULTIPOINT
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| MULTILINESTRING
  {
    $$ = ast.ColumnType{Type: string($1)}
  }
| MULTIPOLYGON
  {
    $$ = ast.ColumnType{Type: string($1)}
  }

enum_values:
  STRING
  {
    $$ = make([]string, 0, 4)
    $$ = append($$, sql_types.EncodeStringSQL($1))
  }
| enum_values ',' STRING
  {
    $$ = append($1, sql_types.EncodeStringSQL($3))
  }

length_opt:
  {
    $$ = nil
  }
| '(' INTEGRAL ')'
  {
    $$ = ast.NewIntLiteral($2)
  }

varying_opt:
  {
    $$ = nil
  }
| VARYING
  {
    $$ = ast.NewStrLiteral($1)
  }

with_timezone_opt:
  {
    $$ = nil
  }
| WITH TIME ZONE
  {
    $$ = ast.NewStrLiteral($1)
  }

float_length_opt:
  {
    $$ = ast.LengthScaleOption{}
  }

decimal_length_opt:
  {
    $$ = ast.LengthScaleOption{}
  }
| '(' INTEGRAL ')'
  {
    $$ = ast.LengthScaleOption{
        Length: ast.NewIntLiteral($2),
    }
  }
| '(' INTEGRAL ',' INTEGRAL ')'
  {
    $$ = ast.LengthScaleOption{
        Length: ast.NewIntLiteral($2),
        Scale: ast.NewIntLiteral($4),
    }
  }

unsigned_opt:
  {
    $$ = false
  }
| UNSIGNED
  {
    $$ = true
  }
| SIGNED
  {
    $$ = false
  }

zero_fill_opt:
  {
    $$ = false
  }
| ZEROFILL
  {
    $$ = true
  }

charset_opt:
  {
    $$ = ast.ColumnCharset{}
  }
| charset_or_character_set sql_id binary_opt
  {
    $$ = ast.ColumnCharset{Name: string($2.String()), Binary: $3}
  }
| charset_or_character_set STRING binary_opt
  {
    $$ = ast.ColumnCharset{Name: sql_types.EncodeStringSQL($2), Binary: $3}
  }
| charset_or_character_set BINARY
  {
    $$ = ast.ColumnCharset{Name: string($2)}
  }
| ASCII binary_opt
  {
    // ASCII: ast.Shorthand for CHARACTER SET latin1.
    $$ = ast.ColumnCharset{Name: "latin1", Binary: $2}
  }
| UNICODE binary_opt
  {
    // UNICODE: ast.Shorthand for CHARACTER SET ucs2.
    $$ = ast.ColumnCharset{Name: "ucs2", Binary: $2}
  }
| BINARY
  {
    // BINARY: ast.Shorthand for default CHARACTER SET but with binary collation
    $$ = ast.ColumnCharset{Name: "", Binary: true}
  }
| BINARY ASCII
  {
    // BINARY ASCII: ast.Shorthand for CHARACTER SET latin1 with binary collation
    $$ = ast.ColumnCharset{Name: "latin1", Binary: true}
  }
| BINARY UNICODE
  {
    // BINARY UNICODE: ast.Shorthand for CHARACTER SET ucs2 with binary collation
    $$ = ast.ColumnCharset{Name: "ucs2", Binary: true}
  }

binary_opt:
  {
    $$ = false
  }
| BINARY
  {
    $$ = true
  }

collate_opt:
  {
    $$ = ""
  }
| COLLATE id_or_var
  {
    $$ = string($2.String())
  }
| COLLATE STRING
  {
    $$ = sql_types.EncodeStringSQL($2)
  }


index_definition:
  index_info '(' index_column_list ')' index_option_list_opt
  {
    $$ = &ast.IndexDefinition{Info: $1, Columns: $3, Options: $5}
  }

index_option_list_opt:
  {
    $$ = nil
  }
| index_option_list
  {
    $$ = $1
  }

index_option_list:
  index_option
  {
    $$ = []*ast.IndexOption{$1}
  }
| index_option_list index_option
  {
    $$ = append($$, $2)
  }

index_option:
  using_index_type
  {
    $$ = $1
  }
| COMMENT_KEYWORD STRING
  {
    $$ = &ast.IndexOption{Name: string($1), Value: ast.NewStrLiteral($2)}
  }
| VISIBLE
  {
    $$ = &ast.IndexOption{Name: string($1) }
  }
| INVISIBLE
  {
    $$ = &ast.IndexOption{Name: string($1) }
  }
| WITH PARSER id_or_var
  {
    $$ = &ast.IndexOption{Name: string($1) + " " + string($2), String: $3.String()}
  }

equal_opt:
  /* empty */
  {
    $$ = ""
  }
| '='
  {
    $$ = string($1)
  }

index_info:
  constraint_name_opt PRIMARY KEY name_opt
  {
    $$ = &ast.IndexInfo{Type: string($2) + " " + string($3), ConstraintName: ast.NewColIdent($1), Name: ast.NewColIdent("PRIMARY"), Primary: true, Unique: true}
  }
| SPATIAL index_or_key_opt name_opt
  {
    $$ = &ast.IndexInfo{Type: string($1) + " " + string($2), Name: ast.NewColIdent($3), Spatial: true, Unique: false}
  }
| FULLTEXT index_or_key_opt name_opt
  {
    $$ = &ast.IndexInfo{Type: string($1) + " " + string($2), Name: ast.NewColIdent($3), Fulltext: true, Unique: false}
  }
| constraint_name_opt UNIQUE index_or_key_opt name_opt
  {
    $$ = &ast.IndexInfo{Type: string($2) + " " + string($3), ConstraintName: ast.NewColIdent($1), Name: ast.NewColIdent($4), Unique: true}
  }
| index_or_key name_opt
  {
    $$ = &ast.IndexInfo{Type: string($1), Name: ast.NewColIdent($2), Unique: false}
  }

constraint_name_opt:
  {
    $$ = ""
  }
| CONSTRAINT name_opt
  {
    $$ = $2
  }

index_symbols:
  INDEX
  {
    $$ = string($1)
  }
| KEYS
  {
    $$ = string($1)
  }
| INDEXES
  {
    $$ = string($1)
  }


from_or_in:
  FROM
  {
    $$ = string($1)
  }
| IN
  {
    $$ = string($1)
  }

index_or_key_opt:
  {
    $$ = "key"
  }
| index_or_key
  {
    $$ = $1
  }

index_or_key:
  INDEX
  {
    $$ = string($1)
  }
  | KEY
  {
    $$ = string($1)
  }

name_opt:
  {
    $$ = ""
  }
| id_or_var
  {
    $$ = string($1.String())
  }

index_column_list:
  index_column
  {
    $$ = []*ast.IndexColumn{$1}
  }
| index_column_list ',' index_column
  {
    $$ = append($$, $3)
  }

index_column:
  sql_id length_opt asc_desc_opt
  {
    $$ = &ast.IndexColumn{Column: $1, Length: $2, Direction: $3}
  }
| openb expression closeb asc_desc_opt
  {
    $$ = &ast.IndexColumn{Expression: $2, Direction: $4}
  }

constraint_definition:
  CONSTRAINT id_or_var_opt constraint_info
  {
    $$ = &ast.ConstraintDefinition{Name: $2, Details: $3}
  }
|  constraint_info
  {
    $$ = &ast.ConstraintDefinition{Details: $1}
  }

check_constraint_definition:
  CONSTRAINT id_or_var_opt check_constraint_info
  {
    $$ = &ast.ConstraintDefinition{Name: $2, Details: $3}
  }
|  check_constraint_info
  {
    $$ = &ast.ConstraintDefinition{Details: $1}
  }

constraint_info:
  FOREIGN KEY name_opt '(' column_list ')' reference_definition
  {
    $$ = &ast.ForeignKeyDefinition{IndexName: ast.NewColIdent($3), Source: $5, ReferenceDefinition: $7}
  }

reference_definition:
  REFERENCES table_name '(' column_list ')' fk_match_opt
  {
    $$ = &ast.ReferenceDefinition{ReferencedTable: $2, ReferencedColumns: $4, Match: $6}
  }
| REFERENCES table_name '(' column_list ')' fk_match_opt fk_on_delete
  {
    $$ = &ast.ReferenceDefinition{ReferencedTable: $2, ReferencedColumns: $4, Match: $6, OnDelete: $7}
  }
| REFERENCES table_name '(' column_list ')' fk_match_opt fk_on_update
  {
    $$ = &ast.ReferenceDefinition{ReferencedTable: $2, ReferencedColumns: $4, Match: $6, OnUpdate: $7}
  }
| REFERENCES table_name '(' column_list ')' fk_match_opt fk_on_delete fk_on_update
  {
    $$ = &ast.ReferenceDefinition{ReferencedTable: $2, ReferencedColumns: $4, Match: $6, OnDelete: $7, OnUpdate: $8}
  }
| REFERENCES table_name '(' column_list ')' fk_match_opt fk_on_update fk_on_delete
  {
    $$ = &ast.ReferenceDefinition{ReferencedTable: $2, ReferencedColumns: $4, Match: $6, OnUpdate: $7, OnDelete: $8}
  }

reference_definition_opt:
  {
    $$ = nil
  }
| reference_definition
  {
    $$ = $1
  }

check_constraint_info:
  CHECK '(' expression ')' enforced_opt
  {
    $$ = &ast.CheckConstraintDefinition{Expr: $3, Enforced: $5}
  }

fk_match:
  MATCH fk_match_action
  {
    $$ = $2
  }

fk_match_action:
  FULL
  {
    $$ = ast.Full
  }
| PARTIAL
  {
    $$ = ast.Partial
  }
| SIMPLE
  {
    $$ = ast.Simple
  }

fk_match_opt:
  {
    $$ = ast.DefaultMatch
  }
| fk_match
  {
    $$ = $1
  }

fk_on_delete:
  ON DELETE fk_reference_action
  {
    $$ = $3
  }

fk_on_update:
  ON UPDATE fk_reference_action
  {
    $$ = $3
  }

fk_reference_action:
  RESTRICT
  {
    $$ = ast.Restrict
  }
| CASCADE
  {
    $$ = ast.Cascade
  }
| NO ACTION
  {
    $$ = ast.NoAction
  }
| SET DEFAULT
  {
    $$ = ast.SetDefault
  }
| SET NULL
  {
    $$ = ast.SetNull
  }

restrict_or_cascade_opt:
  {
    $$ = ""
  }
| RESTRICT
  {
    $$ = string($1)
  }
| CASCADE
  {
    $$ = string($1)
  }

enforced:
  ENFORCED
  {
    $$ = true
  }
| NOT ENFORCED
  {
    $$ = false
  }

enforced_opt:
  {
    $$ = true
  }
| enforced
  {
    $$ = $1
  }

table_option_list_opt:
  {
    $$ = nil
  }
| table_option_list
  {
    $$ = $1
  }

table_option_list:
  table_option
  {
    $$ = ast.TableOptions{$1}
  }
| table_option_list ',' table_option
  {
    $$ = append($1,$3)
  }
| table_option_list table_option
  {
    $$ = append($1,$2)
  }

space_separated_table_option_list:
  table_option
  {
    $$ = ast.TableOptions{$1}
  }
| space_separated_table_option_list table_option
  {
    $$ = append($1,$2)
  }

table_option:
  default_optional charset_or_character_set equal_opt charset
  {
    $$ = &ast.TableOption{Name:(string($2)), String:$4, CaseSensitive: true}
  }
| default_optional COLLATE equal_opt charset
  {
    $$ = &ast.TableOption{Name:string($2), String:$4, CaseSensitive: true}
  }
| COMMENT_KEYWORD equal_opt STRING
  {
    $$ = &ast.TableOption{Name:string($1), Value:ast.NewStrLiteral($3)}
  }
| COMPRESSION equal_opt STRING
  {
    $$ = &ast.TableOption{Name:string($1), Value:ast.NewStrLiteral($3)}
  }
| CONNECTION equal_opt STRING
  {
    $$ = &ast.TableOption{Name:string($1), Value:ast.NewStrLiteral($3)}
  }
| INDEX DIRECTORY equal_opt STRING
  {
    $$ = &ast.TableOption{Name:(string($1)+" "+string($2)), Value:ast.NewStrLiteral($4)}
  }
| ENCRYPTION equal_opt STRING
  {
    $$ = &ast.TableOption{Name:string($1), Value:ast.NewStrLiteral($3)}
  }
| INSERT_METHOD equal_opt insert_method_options
  {
    $$ = &ast.TableOption{Name:string($1), String:string($3)}
  }
| PACK_KEYS equal_opt DEFAULT
  {
    $$ = &ast.TableOption{Name:string($1), String:string($3)}
  }
| PASSWORD equal_opt STRING
  {
    $$ = &ast.TableOption{Name:string($1), Value:ast.NewStrLiteral($3)}
  }
| ROW_FORMAT equal_opt row_format_options
  {
    $$ = &ast.TableOption{Name:string($1), String:string($3)}
  }
| STATS_AUTO_RECALC equal_opt DEFAULT
  {
    $$ = &ast.TableOption{Name:string($1), String:string($3)}
  }
| STATS_PERSISTENT equal_opt DEFAULT
  {
    $$ = &ast.TableOption{Name:string($1), String:string($3)}
  }
| TABLESPACE equal_opt sql_id storage_opt
  {
    $$ = &ast.TableOption{Name:string($1), String: ($3.String() + $4)}
  }
| UNION equal_opt '(' table_name_list ')'
  {
    $$ = &ast.TableOption{Name:string($1), Tables: $4}
  }

storage_opt:
  {
    $$ = ""
  }
| STORAGE DISK
  {
    $$ = " " + string($1) + " " + string($2)
  }
| STORAGE MEMORY
  {
    $$ = " " + string($1) + " " + string($2)
  }

row_format_options:
  DEFAULT
| DYNAMIC
| FIXED
| COMPRESSED
| REDUNDANT
| COMPACT

insert_method_options:
  NO
| FIRST
| LAST

table_opt_value:
  reserved_sql_id
  {
    $$ = $1.String()
  }
| STRING
  {
    $$ = sql_types.EncodeStringSQL($1)
  }
| INTEGRAL
  {
    $$ = string($1)
  }

column_opt:
  {
    $$ = ""
  }
| COLUMN

first_opt:
  {
    $$ = false
  }
| FIRST
  {
    $$ = true
  }

after_opt:
  {
    $$ = nil
  }
| AFTER column_name
  {
    $$ = $2
  }

alter_schema_commands_list:
  OWNER TO ID
  {
    $$ = []ast.AlterOption{&ast.AlterOwner{Owner: &ast.RoleName{Name: ast.RoleIdent{V: $3}}}}
  }

alter_commands_list:
  {
    $$ = nil
  }
| alter_options
  {
    $$ = $1
  }
| alter_options ',' ORDER BY column_list
  {
    $$ = append($1,&ast.OrderByOption{Cols:$5})
  }
| alter_commands_modifier_list
  {
    $$ = $1
  }
| alter_commands_modifier_list ',' alter_options
  {
    $$ = append($1,$3...)
  }
| alter_commands_modifier_list ',' alter_options ',' ORDER BY column_list
  {
    $$ = append(append($1,$3...),&ast.OrderByOption{Cols:$7})
  }

alter_options:
  alter_option
  {
    $$ = []ast.AlterOption{$1}
  }
| alter_options ',' alter_option
  {
    $$ = append($1,$3)
  }
| alter_options ',' alter_commands_modifier
  {
    $$ = append($1,$3)
  }

alter_option:
  space_separated_table_option_list
  {
    $$ = $1
  }
| ADD check_constraint_definition
  {
    $$ = &ast.AddConstraintDefinition{ConstraintDefinition: $2}
  }
| ADD constraint_definition
  {
    $$ = &ast.AddConstraintDefinition{ConstraintDefinition: $2}
  }
| ADD index_definition
  {
    $$ = &ast.AddIndexDefinition{IndexDefinition: $2}
  }
| ADD column_opt '(' column_definition_list ')'
  {
    $$ = &ast.AddColumns{Columns: $4}
  }
| ADD column_opt column_definition first_opt after_opt
  {
    $$ = &ast.AddColumns{Columns: []*ast.ColumnDefinition{$3}, First:$4, After:$5}
  }
| ALTER column_opt column_name DROP DEFAULT
  {
    $$ = &ast.AlterColumn{Column: $3, DropDefault:true}
  }
| ALTER column_opt column_name SET DEFAULT signed_literal_or_null
  {
    $$ = &ast.AlterColumn{Column: $3, DropDefault:false, DefaultVal:$6}
  }
| ALTER column_opt column_name SET DEFAULT openb expression closeb
  {
	$$ = &ast.AlterColumn{Column: $3, DropDefault:false, DefaultVal:$7}
  }
| ALTER column_opt column_name SET VISIBLE
  {
    val := false
    $$ = &ast.AlterColumn{Column: $3, Invisible:&val}
  }
| ALTER column_opt column_name SET INVISIBLE
  {
    val := true
    $$ = &ast.AlterColumn{Column: $3, Invisible:&val}
  }
| ALTER CHECK id_or_var enforced
  {
    $$ = &ast.AlterCheck{Name: $3, Enforced: $4}
  }
| ALTER INDEX id_or_var VISIBLE
  {
    $$ = &ast.AlterIndex{Name: $3, Invisible: false}
  }
| ALTER INDEX id_or_var INVISIBLE
  {
    $$ = &ast.AlterIndex{Name: $3, Invisible: true}
  }
| CHANGE column_opt column_name column_definition first_opt after_opt
  {
    $$ = &ast.ChangeColumn{OldColumn:$3, NewColDefinition:$4, First:$5, After:$6}
  }
| MODIFY column_opt column_definition first_opt after_opt
  {
    $$ = &ast.ModifyColumn{NewColDefinition:$3, First:$4, After:$5}
  }
| CONVERT TO charset_or_character_set charset collate_opt
  {
    $$ = &ast.AlterCharset{CharacterSet:$4, Collate:$5}
  }
| DISABLE KEYS
  {
    $$ = &ast.KeyState{Enable:false}
  }
| ENABLE KEYS
  {
    $$ = &ast.KeyState{Enable:true}
  }
| DISCARD TABLESPACE
  {
    $$ = &ast.TablespaceOperation{Import:false}
  }
| IMPORT TABLESPACE
  {
    $$ = &ast.TablespaceOperation{Import:true}
  }
| DROP column_opt column_name
  {
    $$ = &ast.DropColumn{Name:$3}
  }
| DROP index_or_key id_or_var
  {
    $$ = &ast.DropKey{Type: ast.NormalKeyType, Name:$3}
  }
| DROP PRIMARY KEY
  {
    $$ = &ast.DropKey{Type: ast.PrimaryKeyType}
  }
| DROP FOREIGN KEY id_or_var
  {
    $$ = &ast.DropKey{Type: ast.ForeignKeyType, Name:$4}
  }
| DROP CHECK id_or_var
  {
    $$ = &ast.DropKey{Type: ast.CheckKeyType, Name:$3}
  }
| DROP CONSTRAINT id_or_var
  {
    $$ = &ast.DropKey{Type: ast.CheckKeyType, Name:$3}
  }
| FORCE
  {
    $$ = &ast.Force{}
  }
| RENAME to_opt table_name
  {
    $$ = &ast.RenameTableName{Table:$3}
  }
| RENAME index_or_key id_or_var TO id_or_var
  {
    $$ = &ast.RenameIndex{OldName:$3, NewName:$5}
  }

alter_commands_modifier_list:
  alter_commands_modifier
  {
    $$ = []ast.AlterOption{$1}
  }
| alter_commands_modifier_list ',' alter_commands_modifier
  {
    $$ = append($1,$3)
  }

alter_commands_modifier:
  LOCK equal_opt DEFAULT
    {
      $$ = &ast.LockOption{Type: ast.DefaultType}
    }
  | LOCK equal_opt NONE
    {
      $$ = &ast.LockOption{Type: ast.NoneType}
    }
  | LOCK equal_opt SHARED
    {
      $$ = &ast.LockOption{Type: ast.SharedType}
    }
  | LOCK equal_opt EXCLUSIVE
    {
      $$ = &ast.LockOption{Type: ast.ExclusiveType}
    }
  | WITH VALIDATION
    {
      $$ = &ast.Validation{With:true}
    }
  | WITHOUT VALIDATION
    {
      $$ = &ast.Validation{With:false}
    }

alter_statement:
  alter_table_prefix alter_commands_list
  {
    $1.FullyParsed = true
    $1.AlterOptions = $2
    $$ = $1
  }
| alter_schema_prefix alter_schema_commands_list
  {
    $1.FullyParsed = true
    $1.AlterOptions = $2
    $$ = $1
  }
| alter_sequence_prefix sequence_spec
  {
    $1.SequenceSpec = $2
    $1.FullyParsed = true
    $$ = $1
  }
| ALTER comment_opt definer_opt security_view_opt VIEW table_name column_list_opt AS select_statement check_option_opt
  {
    $$ = &ast.AlterView{ViewName: $6.ToViewName(), Comments: ast.Comments($2).Parsed(), Definer: $3 ,Security:$4, Columns:$7, Select: $9, CheckOption: $10 }
  }
// The syntax here causes a shift / reduce issue, because ENCRYPTION is a non reserved keyword
// and the database identifier is optional. When no identifier is given, the current database
// is used. This means though that `alter database encryption` is ambiguous whether it means
// the encryption keyword, or the encryption database name, resulting in the conflict.
// The preference here is to shift, so it is treated as a database name. This matches the PostgresQL
// behavior as well.
| alter_database_prefix table_id_opt create_options
  {
    $1.FullyParsed = true
    $1.DBName = $2
    $1.AlterOptions = $3
    $$ = $1
  }
| ALTER comment_opt VSCHEMA CREATE VINDEX table_name vindex_type_opt vindex_params_opt
  {
    $$ = &ast.AlterVschema{
        Action: ast.CreateVindexDDLAction,
        Table: $6,
        VindexSpec: &ast.VindexSpec{
          Name: ast.NewColIdent($6.Name.String()),
          Type: $7,
          Params: $8,
        },
      }
  }
| ALTER comment_opt VSCHEMA DROP VINDEX table_name
  {
    $$ = &ast.AlterVschema{
        Action: ast.DropVindexDDLAction,
        Table: $6,
        VindexSpec: &ast.VindexSpec{
          Name: ast.NewColIdent($6.Name.String()),
        },
      }
  }
| ALTER comment_opt VSCHEMA ADD TABLE table_name
  {
    $$ = &ast.AlterVschema{Action: ast.AddVschemaTableDDLAction, Table: $6}
  }
| ALTER comment_opt VSCHEMA DROP TABLE table_name
  {
    $$ = &ast.AlterVschema{Action: ast.DropVschemaTableDDLAction, Table: $6}
  }
| ALTER comment_opt VSCHEMA ON table_name ADD VINDEX sql_id '(' column_list ')' vindex_type_opt vindex_params_opt
  {
    $$ = &ast.AlterVschema{
        Action: ast.AddColVindexDDLAction,
        Table: $5,
        VindexSpec: &ast.VindexSpec{
            Name: $8,
            Type: $12,
            Params: $13,
        },
        VindexCols: $10,
      }
  }
| ALTER comment_opt VSCHEMA ON table_name DROP VINDEX sql_id
  {
    $$ = &ast.AlterVschema{
        Action: ast.DropColVindexDDLAction,
        Table: $5,
        VindexSpec: &ast.VindexSpec{
            Name: $8,
        },
      }
  }
| ALTER comment_opt VSCHEMA ADD SEQUENCE table_name
  {
    $$ = &ast.AlterVschema{Action: ast.AddSequenceDDLAction, Table: $6}
  }
| ALTER comment_opt VSCHEMA ON table_name ADD AUTO_INCREMENT sql_id USING table_name
  {
    $$ = &ast.AlterVschema{
        Action: ast.AddAutoIncDDLAction,
        Table: $5,
        AutoIncSpec: &ast.AutoIncSpec{
            Column: $8,
            Sequence: $10,
        },
    }
  }

json_table_function:
  JSON_TABLE openb expression ',' text_literal_or_arg jt_columns_clause closeb as_opt_id
  {
    $$ = &ast.JSONTableExpr{Expr: $3, Filter: $5, Columns: $6, Alias: $8}
  }

jt_columns_clause:
  COLUMNS openb columns_list closeb
  {
    $$= $3
  }

columns_list:
  jt_column
  {
    $$= []*ast.JtColumnDefinition{$1}
  }
| columns_list ',' jt_column
  {
    $$ = append($1, $3)
  }

jt_column:
 sql_id FOR ORDINALITY
  {
    $$ = &ast.JtColumnDefinition{JtOrdinal: &ast.JtOrdinalColDef{Name: $1}}
  }
| sql_id column_type collate_opt jt_exists_opt PATH text_literal_or_arg
  {
    $2.Options= &ast.ColumnTypeOptions{Collate:$3}
    jtPath := &ast.JtPathColDef{Name: $1, Type: $2, JtColExists: $4, Path: $6}
    $$ = &ast.JtColumnDefinition{JtPath: jtPath}
  }
| sql_id column_type collate_opt jt_exists_opt PATH text_literal_or_arg on_empty
  {
    $2.Options= &ast.ColumnTypeOptions{Collate:$3}
    jtPath := &ast.JtPathColDef{Name: $1, Type: $2, JtColExists: $4, Path: $6, EmptyOnResponse: $7}
    $$ = &ast.JtColumnDefinition{JtPath: jtPath}
  }
| sql_id column_type collate_opt jt_exists_opt PATH text_literal_or_arg on_error
  {
    $2.Options= &ast.ColumnTypeOptions{Collate:$3}
    jtPath := &ast.JtPathColDef{Name: $1, Type: $2, JtColExists: $4, Path: $6, ErrorOnResponse: $7}
    $$ = &ast.JtColumnDefinition{JtPath: jtPath}
  }
| sql_id column_type collate_opt jt_exists_opt PATH text_literal_or_arg on_empty on_error
  {
    $2.Options= &ast.ColumnTypeOptions{Collate:$3}
    jtPath := &ast.JtPathColDef{Name: $1, Type: $2, JtColExists: $4, Path: $6, EmptyOnResponse: $7, ErrorOnResponse: $8}
    $$ = &ast.JtColumnDefinition{JtPath: jtPath}
  }
| NESTED jt_path_opt text_literal_or_arg jt_columns_clause
  {
    jtNestedPath := &ast.JtNestedPathColDef{Path: $3, Columns: $4}
    $$ = &ast.JtColumnDefinition{JtNestedPath: jtNestedPath}
  }

jt_path_opt:
  {
    $$ = false
  }
| PATH
  {
    $$ = true
  }
jt_exists_opt:
  {
    $$=false
  }
| EXISTS
  {
    $$=true
  }

on_empty:
  json_on_response ON EMPTY
  {
    $$= $1
  }

on_error:
  json_on_response ON ERROR
  {
    $$= $1
  }

json_on_response:
  ERROR
  {
    $$ = &ast.JtOnResponse{ResponseType: ast.ErrorJSONType}
  }
| NULL
  {
    $$ = &ast.JtOnResponse{ResponseType: ast.NullJSONType}
  }
| DEFAULT text_literal_or_arg
  {
    $$ = &ast.JtOnResponse{ResponseType: ast.DefaultJSONType, Expr: $2}
  }

rename_statement:
  RENAME TABLE rename_list
  {
    $$ = &ast.RenameTable{TablePairs: $3}
  }

rename_list:
  table_name TO table_name
  {
    $$ = []*ast.RenameTablePair{{FromTable: $1, ToTable: $3}}
  }
| rename_list ',' table_name TO table_name
  {
    $$ = append($1, &ast.RenameTablePair{FromTable: $3, ToTable: $5})
  }

drop_statement:
  DROP comment_opt temp_opt TABLE exists_opt table_name_list restrict_or_cascade_opt
  {
    $$ = &ast.DropTable{FromTables: $6, IfExists: $5, Comments: ast.Comments($2).Parsed(), Temp: $3}
  }
| DROP comment_opt INDEX id_or_var ON table_name
  {
    // Change this to an alter statement
    if $4.Lowered() == "primary" {
      $$ = &ast.AlterTable{FullyParsed:true, Table: $6,AlterOptions: append([]ast.AlterOption{&ast.DropKey{Type: ast.PrimaryKeyType}})}
    } else {
      $$ = &ast.AlterTable{FullyParsed: true, Table: $6,AlterOptions: append([]ast.AlterOption{&ast.DropKey{Type: ast.NormalKeyType, Name:$4}})}
    }
  }
| DROP comment_opt VIEW exists_opt view_name_list restrict_or_cascade_opt
  {
    $$ = &ast.DropView{FromTables: $5, Comments: ast.Comments($2).Parsed(), IfExists: $4}
  }
| DROP comment_opt database exists_opt table_id
  {
    $$ = &ast.DropDatabase{Comments: ast.Comments($2).Parsed(), DBName: $5, IfExists: $4}
  }

truncate_statement:
  TRUNCATE TABLE table_name
  {
    $$ = &ast.TruncateTable{Table: $3}
  }
| TRUNCATE table_name
  {
    $$ = &ast.TruncateTable{Table: $2}
  }
analyze_statement:
  ANALYZE TABLE table_name
  {
    $$ = &ast.OtherRead{}
  }

show_statement:
  SHOW charset_or_character_set like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Charset, Filter: $3}}
  }
| SHOW COLLATION like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Collation, Filter: $3}}
  }
| SHOW full_opt columns_or_fields from_or_in table_name from_database_opt like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Full: $2, Command: ast.Column, Tbl: $5, DbName: $6, Filter: $7}}
  }
| SHOW DATABASES like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Database, Filter: $3}}
  }
| SHOW SCHEMAS like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Database, Filter: $3}}
  }
| SHOW KEYSPACES like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Keyspace, Filter: $3}}
  }
| SHOW FUNCTION STATUS like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Function, Filter: $4}}
  }
| SHOW extended_opt index_symbols from_or_in table_name from_database_opt like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Index, Tbl: $5, DbName: $6, Filter: $7}}
  }
| SHOW OPEN TABLES from_database_opt like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.OpenTable, DbName:$4, Filter: $5}}
  }
| SHOW PRIVILEGES
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Privilege}}
  }
| SHOW PROCEDURE STATUS like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Procedure, Filter: $4}}
  }
| SHOW session_or_local_opt STATUS like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.StatusSession, Filter: $4}}
  }
| SHOW GLOBAL STATUS like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.StatusGlobal, Filter: $4}}
  }
| SHOW session_or_local_opt VARIABLES like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.VariableSession, Filter: $4}}
  }
| SHOW GLOBAL VARIABLES like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.VariableGlobal, Filter: $4}}
  }
| SHOW TABLE STATUS from_database_opt like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.TableStatus, DbName:$4, Filter: $5}}
  }
| SHOW full_opt TABLES from_database_opt like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Table, Full: $2, DbName:$4, Filter: $5}}
  }
| SHOW TRIGGERS from_database_opt like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Trigger, DbName:$3, Filter: $4}}
  }
| SHOW CREATE DATABASE table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowCreate{Command: ast.CreateDb, Op: $4}}
  }
| SHOW CREATE EVENT table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowCreate{Command: ast.CreateE, Op: $4}}
  }
| SHOW CREATE FUNCTION table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowCreate{Command: ast.CreateF, Op: $4}}
  }
| SHOW CREATE PROCEDURE table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowCreate{Command: ast.CreateProc, Op: $4}}
  }
| SHOW CREATE TABLE table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowCreate{Command: ast.CreateTbl, Op: $4}}
  }
| SHOW CREATE TRIGGER table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowCreate{Command: ast.CreateTr, Op: $4}}
  }
| SHOW CREATE VIEW table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowCreate{Command: ast.CreateV, Op: $4}}
  }
| SHOW PLUGINS
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Plugins}}
  }
| SHOW GLOBAL GTID_EXECUTED from_database_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.GtidExecGlobal, DbName: $4}}
  }
| SHOW GLOBAL VGTID_EXECUTED from_database_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.VGtidExecGlobal, DbName: $4}}
  }
| SHOW VSCHEMA TABLES
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.VschemaTables}}
  }
| SHOW VSCHEMA VINDEXES
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.VschemaVindexes}}
  }
| SHOW VSCHEMA VINDEXES from_or_on table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.VschemaVindexes, Tbl: $5}}
  }
| SHOW WARNINGS
  {
    $$ = &ast.Show{Internal: &ast.ShowBasic{Command: ast.Warnings}}
  }
/*
 * Catch-all for show statements without vitess keywords:
 */
| SHOW id_or_var ddl_skip_to_end
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($2.String())}}
  }
| SHOW CREATE USER ddl_skip_to_end
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($2) + " " + string($3)}}
   }
| SHOW BINARY id_or_var ddl_skip_to_end /* SHOW BINARY ... */
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($2) + " " + $3.String()}}
  }
| SHOW BINARY LOGS ddl_skip_to_end /* SHOW BINARY LOGS */
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($2) + " " + string($3)}}
  }
| SHOW FUNCTION CODE table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($2) + " " + string($3) + " " + ast.String($4)}}
  }
| SHOW PROCEDURE CODE table_name
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($2) + " " + string($3) + " " + ast.String($4)}}
  }
| SHOW full_opt PROCESSLIST from_database_opt like_or_where_opt
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($3)}}
  }
| SHOW STORAGE ddl_skip_to_end
  {
    $$ = &ast.Show{Internal: &ast.ShowOther{Command: string($2)}}
  }

extended_opt:
  /* empty */
  {
    $$ = ""
  }
  | EXTENDED
  {
    $$ = "extended "
  }

full_opt:
  /* empty */
  {
    $$ = false
  }
| FULL
  {
    $$ = true
  }

columns_or_fields:
  COLUMNS
  {
      $$ = string($1)
  }
| FIELDS
  {
      $$ = string($1)
  }

from_database_opt:
  /* empty */
  {
    $$ = ast.NewTableIdent("")
  }
| FROM table_id
  {
    $$ = $2
  }
| IN table_id
  {
    $$ = $2
  }

like_or_where_opt:
  /* empty */
  {
    $$ = nil
  }
| LIKE STRING
  {
    $$ = &ast.ShowFilter{Like:string($2)}
  }
| WHERE expression
  {
    $$ = &ast.ShowFilter{Filter:$2}
  }

session_or_local_opt:
  /* empty */
  {
    $$ = struct{}{}
  }
| SESSION
  {
    $$ = struct{}{}
  }
| LOCAL
  {
    $$ = struct{}{}
  }

from_or_on:
  FROM
  {
    $$ = string($1)
  }
| ON
  {
    $$ = string($1)
  }

use_statement:
  USE table_id
  {
    $$ = &ast.Use{DBName: $2}
  }
| USE
  {
    $$ = &ast.Use{DBName:ast.TableIdent{V:""}}
  }
| USE table_id AT_ID
  {
    $$ = &ast.Use{DBName:ast.NewTableIdent($2.String()+"@"+string($3))}
  }

begin_statement:
  BEGIN
  {
    $$ = &ast.Begin{}
  }
| START TRANSACTION
  {
    $$ = &ast.Begin{}
  }

commit_statement:
  COMMIT
  {
    $$ = &ast.Commit{}
  }

rollback_statement:
  ROLLBACK
  {
    $$ = &ast.Rollback{}
  }
| ROLLBACK work_opt TO savepoint_opt sql_id
  {
    $$ = &ast.SRollback{Name: $5}
  }

work_opt:
  { $$ = struct{}{} }
| WORK
  { $$ = struct{}{} }

savepoint_opt:
  { $$ = struct{}{} }
| SAVEPOINT
  { $$ = struct{}{} }


savepoint_statement:
  SAVEPOINT sql_id
  {
    $$ = &ast.Savepoint{Name: $2}
  }

release_statement:
  RELEASE SAVEPOINT sql_id
  {
    $$ = &ast.Release{Name: $3}
  }

explain_format_opt:
  {
    $$ = ast.EmptyType
  }
| FORMAT '=' JSON
  {
    $$ = ast.JSONType
  }
| FORMAT '=' TREE
  {
    $$ = ast.TreeType
  }
| FORMAT '=' TRADITIONAL
  {
    $$ = ast.TraditionalType
  }
| ANALYSE
  {
    $$ = ast.AnalyzeType
  }
| ANALYZE
  {
    $$ = ast.AnalyzeType
  }

explain_synonyms:
  EXPLAIN
  {
    $$ = $1
  }
| DESCRIBE
  {
    $$ = $1
  }
| DESC
  {
    $$ = $1
  }

explainable_statement:
  select_statement
  {
    $$ = $1
  }
| update_statement
  {
    $$ = $1
  }
| insert_statement
  {
    $$ = $1
  }
| delete_statement
  {
    $$ = $1
  }

wild_opt:
  {
    $$ = ""
  }
| sql_id
  {
    $$ = $1.Val
  }
| STRING
  {
    $$ = sql_types.EncodeStringSQL($1)
  }

explain_statement:
  explain_synonyms table_name wild_opt
  {
    $$ = &ast.ExplainTab{Table: $2, Wild: $3}
  }
| explain_synonyms explain_format_opt explainable_statement
  {
    $$ = &ast.ExplainStmt{Type: $2, Statement: $3}
  }

other_statement:
  REPAIR skip_to_end
  {
    $$ = &ast.OtherAdmin{}
  }
| OPTIMIZE skip_to_end
  {
    $$ = &ast.OtherAdmin{}
  }

lock_statement:
  LOCK TABLES lock_table_list
  {
    $$ = &ast.LockTables{Tables: $3}
  }

lock_table_list:
  lock_table
  {
    $$ = ast.TableAndLockTypes{$1}
  }
| lock_table_list ',' lock_table
  {
    $$ = append($1, $3)
  }

lock_table:
  aliased_table_name lock_type
  {
    $$ = &ast.TableAndLockType{Table:$1, Lock:$2}
  }

lock_type:
  READ
  {
    $$ = ast.Read
  }
| READ LOCAL
  {
    $$ = ast.ReadLocal
  }
| WRITE
  {
    $$ = ast.Write
  }
| LOW_PRIORITY WRITE
  {
    $$ = ast.LowPriorityWrite
  }

unlock_statement:
  UNLOCK TABLES
  {
    $$ = &ast.UnlockTables{}
  }

flush_statement:
  FLUSH local_opt flush_option_list
  {
    $$ = &ast.Flush{IsLocal: $2, FlushOptions:$3}
  }
| FLUSH local_opt TABLES
  {
    $$ = &ast.Flush{IsLocal: $2}
  }
| FLUSH local_opt TABLES WITH READ LOCK
  {
    $$ = &ast.Flush{IsLocal: $2, WithLock:true}
  }
| FLUSH local_opt TABLES table_name_list
  {
    $$ = &ast.Flush{IsLocal: $2, TableNames:$4}
  }
| FLUSH local_opt TABLES table_name_list WITH READ LOCK
  {
    $$ = &ast.Flush{IsLocal: $2, TableNames:$4, WithLock:true}
  }
| FLUSH local_opt TABLES table_name_list FOR EXPORT
  {
    $$ = &ast.Flush{IsLocal: $2, TableNames:$4, ForExport:true}
  }

flush_option_list:
  flush_option
  {
    $$ = []string{$1}
  }
| flush_option_list ',' flush_option
  {
    $$ = append($1,$3)
  }

flush_option:
  BINARY LOGS
  {
    $$ = string($1) + " " + string($2)
  }
| ERROR LOGS
  {
    $$ = string($1) + " " + string($2)
  }
| GENERAL LOGS
  {
    $$ = string($1) + " " + string($2)
  }
| HOSTS
  {
    $$ = string($1)
  }
| LOGS
  {
    $$ = string($1)
  }
| PRIVILEGES
  {
    $$ = string($1)
  }
| RELAY LOGS for_channel_opt
  {
    $$ = string($1) + " " + string($2) + $3
  }
| SLOW LOGS
  {
    $$ = string($1) + " " + string($2)
  }
| OPTIMIZER_COSTS
  {
    $$ = string($1)
  }
| STATUS
  {
    $$ = string($1)
  }
| USER_RESOURCES
  {
    $$ = string($1)
  }

local_opt:
  {
    $$ = false
  }
| LOCAL
  {
    $$ = true
  }
| NO_WRITE_TO_BINLOG
  {
    $$ = true
  }

for_channel_opt:
  {
    $$ = ""
  }
| FOR CHANNEL id_or_var
  {
    $$ = " " + string($1) + " " + string($2) + " " + $3.String()
  }

comment_statement:
  {
    setAllowComments(psqlex, true)
  }
  COMMENT comment_list ON schema_name IS text_literal_or_arg
  {
    // Strange argumets shift
    $$ = &ast.CommentOnSchema{Comments: ast.Comments{$2}.Parsed(), Schema: $5.Name, Value: $7}
    setAllowComments(psqlex, false)
  }


comment_opt:
  {
    setAllowComments(psqlex, true)
  }
  comment_list
  {
    $$ = $2
    setAllowComments(psqlex, false)
  }

comment_list:
  {
    $$ = nil
  }
| comment_list COMMENT
  {
    $$ = append($1, $2)
  }

only_opt:
  {
    $$ = false
  }
| ONLY
  {
    $$ = true
  }

union_op:
  UNION
  {
    $$ = true
  }
| UNION ALL
  {
    $$ = false
  }
| UNION DISTINCT
  {
    $$ = true
  }

cache_opt:
{
  $$ = ""
}
| SQL_NO_CACHE
{
  $$ = ast.SQLNoCacheStr
}
| SQL_CACHE
{
  $$ = ast.SQLCacheStr
}

distinct_opt:
  {
    $$ = false
  }
| DISTINCT
  {
    $$ = true
  }
| DISTINCTROW
  {
    $$ = true
  }

prepare_statement:
  PREPARE comment_opt sql_id FROM text_literal_or_arg
  {
    $$ = &ast.PrepareStmt{Name:$3, Comments: ast.Comments($2).Parsed(), Statement:$5}
  }
| PREPARE comment_opt sql_id FROM AT_ID
  {
    $$ = &ast.PrepareStmt{
    	Name:$3,
    	Comments: ast.Comments($2).Parsed(),
    	Statement: &ast.ColName{
    		Name: ast.NewColIdentWithAt(string($5), ast.SingleAt),
    	},
    }
  }

execute_statement:
  EXECUTE comment_opt sql_id execute_statement_list_opt
  {
    $$ = &ast.ExecuteStmt{Name:$3, Comments: ast.Comments($2).Parsed(), Arguments: $4}
  }

execute_statement_list_opt:
  {
    $$ = nil
  }
| USING at_id_list
  {
    $$ = $2
  }

deallocate_statement:
  DEALLOCATE comment_opt PREPARE sql_id
  {
    $$ = &ast.DeallocateStmt{Type: ast.DeallocateType, Comments: ast.Comments($2).Parsed(), Name:$4}
  }
| DROP comment_opt PREPARE sql_id
  {
    $$ = &ast.DeallocateStmt{Type: ast.DropType, Comments: ast.Comments($2).Parsed(), Name: $4}
  }

select_expression_list_opt:
  {
    $$ = nil
  }
| select_expression_list
  {
    $$ = $1
  }

select_options:
  {
    $$ = nil
  }
| select_option
  {
    $$ = []string{$1}
  }
| select_option select_option // TODO: figure out a way to do this recursively instead.
  {                           // TODO: ast.This is a hack since I couldn't get it to work in a nicer way. I got 'conflicts: 8 shift/reduce'
    $$ = []string{$1, $2}
  }
| select_option select_option select_option
  {
    $$ = []string{$1, $2, $3}
  }
| select_option select_option select_option select_option
  {
    $$ = []string{$1, $2, $3, $4}
  }

select_option:
  SQL_NO_CACHE
  {
    $$ = ast.SQLNoCacheStr
  }
| SQL_CACHE
  {
    $$ = ast.SQLCacheStr
  }
| DISTINCT
  {
    $$ = ast.DistinctStr
  }
| DISTINCTROW
  {
    $$ = ast.DistinctStr
  }
| STRAIGHT_JOIN
  {
    $$ = ast.StraightJoinHint
  }
| SQL_CALC_FOUND_ROWS
  {
    $$ = ast.SQLCalcFoundRowsStr
  }
| ALL
  {
    $$ = ast.AllStr // These are not picked up by NewSelect, and so ALL will be dropped. But this is OK, since it's redundant anyway
  }

select_expression_list:
  select_expression
  {
    $$ = ast.SelectExprs{$1}
  }
| select_expression_list ',' select_expression
  {
    $$ = append($$, $3)
  }

select_expression:
  '*'
  {
    $$ = &ast.StarExpr{}
  }
| expression as_ci_opt
  {
    $$ = &ast.AliasedExpr{Expr: $1, As: $2}
  }
| table_id '.' '*'
  {
    $$ = &ast.StarExpr{TableName: ast.TableName{Name: $1}}
  }
| table_id '.' reserved_table_id '.' '*'
  {
    $$ = &ast.StarExpr{TableName: ast.TableName{Qualifier: $1, Name: $3}}
  }

as_ci_opt:
  {
    $$ = ast.ColIdent{}
  }
| col_alias
  {
    $$ = $1
  }
| AS col_alias
  {
    $$ = $2
  }

col_alias:
  sql_id
| STRING
  {
    $$ = ast.NewColIdent(string($1))
  }

from_opt:
  %prec EMPTY_FROM_CLAUSE {
    $$ = ast.TableExprs{&ast.AliasedTableExpr{Expr:ast.TableName{Name: ast.NewTableIdent("dual")}}}
  }
  | from_clause
  {
  	$$ = $1
  }

from_clause:
FROM table_references
  {
    $$ = $2
  }

table_references:
  table_reference
  {
    $$ = ast.TableExprs{$1}
  }
| table_references ',' table_reference
  {
    $$ = append($$, $3)
  }

table_reference:
  table_factor
| join_table

table_factor:
  aliased_table_name
  {
    $$ = $1
  }
| derived_table as_opt table_id column_list_opt
  {
    $$ = &ast.AliasedTableExpr{Expr:$1, As: $3, Columns: $4}
  }
| openb table_references closeb
  {
    $$ = &ast.ParenTableExpr{Exprs: $2}
  }
| json_table_function
  {
    $$ = $1
  }

derived_table:
  openb query_expression closeb
  {
    $$ = &ast.DerivedTable{Lateral: false, Select: $2}
  }
| LATERAL openb query_expression closeb
  {
    $$ = &ast.DerivedTable{Lateral: true, Select: $3}
  }

aliased_table_name:
table_name as_opt_id index_hint_list_opt
  {
    $$ = &ast.AliasedTableExpr{Expr:$1, As: $2, Hints: $3}
  }

column_list_opt:
  {
    $$ = nil
  }
| '(' column_list ')'
  {
    setIgnoreCommentKeyword(psqlex, true)
    $$ = $2
    setIgnoreCommentKeyword(psqlex, false)
  }
| '*'
  {
    $$ = nil
  }

column_list:
  sql_id
  {
    $$ = ast.Columns{$1}
  }
| column_list ',' sql_id
  {
    $$ = append($$, $3)
  }

at_id_list:
  AT_ID
  {
    $$ = ast.Columns{ast.NewColIdentWithAt(string($1), ast.SingleAt)}
  }
| column_list ',' AT_ID
  {
    $$ = append($$, ast.NewColIdentWithAt(string($3), ast.SingleAt))
  }

index_list:
  sql_id
  {
    $$ = ast.Columns{$1}
  }
| PRIMARY
  {
    $$ = ast.Columns{ast.NewColIdent(string($1))}
  }
| index_list ',' sql_id
  {
    $$ = append($$, $3)
  }
| index_list ',' PRIMARY
  {
    $$ = append($$, ast.NewColIdent(string($3)))
  }

// There is a grammar conflict here:
// 1: ast.INSERT INTO a SELECT * FROM b JOIN c ON b.i = c.i
// 2: ast.INSERT INTO a SELECT * FROM b JOIN c ON DUPLICATE KEY UPDATE a.i = 1
// When yacc encounters the ON clause, it cannot determine which way to
// resolve. The %prec override below makes the parser choose the
// first construct, which automatically makes the second construct a
// syntax error. This is the same behavior as PostgresQL.
join_table:
  table_reference inner_join table_factor join_condition_opt
  {
    $$ = &ast.JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3, Condition: $4}
  }
| table_reference straight_join table_factor on_expression_opt
  {
    $$ = &ast.JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3, Condition: $4}
  }
| table_reference outer_join table_reference join_condition
  {
    $$ = &ast.JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3, Condition: $4}
  }
| table_reference natural_join table_factor
  {
    $$ = &ast.JoinTableExpr{LeftExpr: $1, Join: $2, RightExpr: $3}
  }

join_condition:
  ON expression
  { $$ = &ast.JoinCondition{On: $2} }
| USING '(' column_list ')'
  { $$ = &ast.JoinCondition{Using: $3} }

join_condition_opt:
%prec JOIN
  { $$ = &ast.JoinCondition{} }
| join_condition
  { $$ = $1 }

on_expression_opt:
%prec JOIN
  { $$ = &ast.JoinCondition{} }
| ON expression
  { $$ = &ast.JoinCondition{On: $2} }

schema_name:
  SCHEMA schema_id
  {
    $$ = ast.SchemaName{Name: $2}
  }

as_opt:
  { $$ = struct{}{} }
| AS
  { $$ = struct{}{} }

as_opt_id:
  {
    $$ = ast.NewTableIdent("")
  }
| table_alias
  {
    $$ = $1
  }
| AS table_alias
  {
    $$ = $2
  }

table_alias:
  table_id
| STRING
  {
    $$ = ast.NewTableIdent(string($1))
  }

inner_join:
  JOIN
  {
    $$ = ast.NormalJoinType
  }
| INNER JOIN
  {
    $$ = ast.NormalJoinType
  }
| CROSS JOIN
  {
    $$ = ast.NormalJoinType
  }

straight_join:
  STRAIGHT_JOIN
  {
    $$ = ast.StraightJoinType
  }

outer_join:
  LEFT JOIN
  {
    $$ = ast.LeftJoinType
  }
| LEFT OUTER JOIN
  {
    $$ = ast.LeftJoinType
  }
| RIGHT JOIN
  {
    $$ = ast.RightJoinType
  }
| RIGHT OUTER JOIN
  {
    $$ = ast.RightJoinType
  }

natural_join:
 NATURAL JOIN
  {
    $$ = ast.NaturalJoinType
  }
| NATURAL outer_join
  {
    if $2 == ast.LeftJoinType {
      $$ = ast.NaturalLeftJoinType
    } else {
      $$ = ast.NaturalRightJoinType
    }
  }

into_table_name:
  INTO table_name
  {
    $$ = $2
  }
| table_name
  {
    $$ = $1
  }

table_name:
  table_id
  {
    $$ = ast.TableName{Name: $1}
  }
| table_id '.' reserved_table_id
  {
    $$ = ast.TableName{Qualifier: $1, Name: $3}
  }

delete_table_name:
table_id '.' '*'
  {
    $$ = ast.TableName{Name: $1}
  }

index_hint_list_opt:
  {
    $$ = nil
  }
| index_hint_list
  {
    $$ = $1
  }

index_hint_list:
index_hint
  {
    $$ = ast.IndexHints{$1}
  }
| index_hint_list index_hint
  {
    $$ = append($1,$2)
  }

index_hint:
  USE index_or_key index_hint_for_opt openb index_list closeb
  {
    $$ = &ast.IndexHint{Type: ast.UseOp, ForType:$3, Indexes: $5}
  }
| USE index_or_key index_hint_for_opt openb closeb
  {
    $$ = &ast.IndexHint{Type: ast.UseOp, ForType: $3}
  }
| IGNORE index_or_key index_hint_for_opt openb index_list closeb
  {
    $$ = &ast.IndexHint{Type: ast.IgnoreOp, ForType: $3, Indexes: $5}
  }
| FORCE index_or_key index_hint_for_opt openb index_list closeb
  {
    $$ = &ast.IndexHint{Type: ast.ForceOp, ForType: $3, Indexes: $5}
  }

index_hint_for_opt:
  {
    $$ = ast.NoForType
  }
| FOR JOIN
  {
    $$ = ast.JoinForType
  }
| FOR ORDER BY
  {
    $$ = ast.OrderByForType
  }
| FOR GROUP BY
  {
    $$ = ast.GroupByForType
  }

sequence_name:
  SEQUENCE sequence_id
  {
    $$ = ast.SequenceName{Name: $2}
  }
| SEQUENCE sequence_id '.' reserved_sequence_id
  {
    $$ = ast.SequenceName{Qualifier: $2, Name: $4}
  }


where_expression_opt:
  {
    $$ = nil
  }
| WHERE expression
  {
    $$ = $2
  }

/* all possible expressions */
expression:
  expression OR expression %prec OR
  {
	$$ = &ast.OrExpr{Left: $1, Right: $3}
  }
| expression AND expression %prec AND
  {
	$$ = &ast.AndExpr{Left: $1, Right: $3}
  }
| NOT expression %prec NOT
  {
	  $$ = &ast.NotExpr{Expr: $2}
  }
| bool_pri IS is_suffix %prec IS
  {
	 $$ = &ast.IsExpr{Left: $1, Right: $3}
  }
| bool_pri %prec EXPRESSION_PREC_SETTER
  {
	$$ = $1
  }
| expression MEMBER OF openb expression closeb
  {
    $$ = &ast.MemberOfExpr{Value: $1, JSONArr:$5 }
  }


bool_pri:
bool_pri IS NULL %prec IS
  {
	 $$ = &ast.IsExpr{Left: $1, Right: ast.IsNullOp}
  }
| bool_pri IS NOT NULL %prec IS
  {
  	$$ = &ast.IsExpr{Left: $1, Right: ast.IsNotNullOp}
  }
| bool_pri compare predicate
  {
	$$ = &ast.ComparisonExpr{Left: $1, Operator: $2, Right: $3}
  }
| predicate %prec EXPRESSION_PREC_SETTER
  {
	$$ = $1
  }

predicate:
bit_expr IN col_tuple
  {
	$$ = &ast.ComparisonExpr{Left: $1, Operator: ast.InOp, Right: $3}
  }
| bit_expr NOT IN col_tuple
  {
	$$ = &ast.ComparisonExpr{Left: $1, Operator: ast.NotInOp, Right: $4}
  }
| bit_expr BETWEEN bit_expr AND predicate
  {
	 $$ = &ast.BetweenExpr{Left: $1, IsBetween: true, From: $3, To: $5}
  }
| bit_expr NOT BETWEEN bit_expr AND predicate
  {
	$$ = &ast.BetweenExpr{Left: $1, IsBetween: false, From: $4, To: $6}
  }
| bit_expr LIKE simple_expr
  {
	  $$ = &ast.ComparisonExpr{Left: $1, Operator: ast.LikeOp, Right: $3}
  }
| bit_expr NOT LIKE simple_expr
  {
	$$ = &ast.ComparisonExpr{Left: $1, Operator: ast.NotLikeOp, Right: $4}
  }
| bit_expr LIKE simple_expr ESCAPE simple_expr %prec LIKE
  {
	  $$ = &ast.ComparisonExpr{Left: $1, Operator: ast.LikeOp, Right: $3, Escape: $5}
  }
| bit_expr NOT LIKE simple_expr ESCAPE simple_expr %prec LIKE
  {
	$$ = &ast.ComparisonExpr{Left: $1, Operator: ast.NotLikeOp, Right: $4, Escape: $6}
  }
| bit_expr REGEXP bit_expr
  {
	$$ = &ast.ComparisonExpr{Left: $1, Operator: ast.RegexpOp, Right: $3}
  }
| bit_expr NOT REGEXP bit_expr
  {
	 $$ = &ast.ComparisonExpr{Left: $1, Operator: ast.NotRegexpOp, Right: $4}
  }
| bit_expr %prec EXPRESSION_PREC_SETTER
 {
	$$ = $1
 }
| bit_expr '=' ANY col_tuple
  {
	$$ = &ast.ComparisonExpr{Left: $1, Operator: ast.InOp, Right: $4}
  }

bit_expr:
bit_expr '|' bit_expr %prec '|'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.BitOrOp, Right: $3}
  }
| bit_expr '&' bit_expr %prec '&'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.BitAndOp, Right: $3}
  }
| bit_expr SHIFT_LEFT bit_expr %prec SHIFT_LEFT
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.ShiftLeftOp, Right: $3}
  }
| bit_expr SHIFT_RIGHT bit_expr %prec SHIFT_RIGHT
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.ShiftRightOp, Right: $3}
  }
| bit_expr '+' bit_expr %prec '+'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.PlusOp, Right: $3}
  }
| bit_expr '-' bit_expr %prec '-'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.MinusOp, Right: $3}
  }
| bit_expr '*' bit_expr %prec '*'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.MultOp, Right: $3}
  }
| bit_expr '/' bit_expr %prec '/'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.DivOp, Right: $3}
  }
| bit_expr '%' bit_expr %prec '%'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.ModOp, Right: $3}
  }
| bit_expr DIV bit_expr %prec DIV
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.IntDivOp, Right: $3}
  }
| bit_expr MOD bit_expr %prec MOD
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.ModOp, Right: $3}
  }
| bit_expr '^' bit_expr %prec '^'
  {
	  $$ = &ast.BinaryExpr{Left: $1, Operator: ast.BitXorOp, Right: $3}
  }
| simple_expr %prec EXPRESSION_PREC_SETTER
  {
	$$ = $1
  }

simple_expr:
function_call_keyword
  {
  	$$ = $1
  }
| function_call_nonkeyword
  {
  	$$ = $1
  }
| function_call_generic
  {
  	$$ = $1
  }
| function_call_conflict
  {
  	$$ = $1
  }
| simple_expr COLLATE charset %prec UNARY
  {
	$$ = &ast.CollateExpr{Expr: $1, Collation: $3}
  }
| literal_or_null
  {
  	$$ = $1
  }
| column_name
  {
  	$$ = $1
  }
| '+' simple_expr %prec UNARY
  {
	$$= $2; // TODO: do we really want to ignore unary '+' before any kind of literals?
  }
| '-' simple_expr %prec UNARY
  {
	$$ = &ast.UnaryExpr{Operator: ast.UMinusOp, Expr: $2}
  }
| '~' simple_expr %prec UNARY
  {
	$$ = &ast.UnaryExpr{Operator: ast.TildaOp, Expr: $2}
  }
| '!' simple_expr %prec UNARY
  {
    $$ = &ast.UnaryExpr{Operator: ast.BangOp, Expr: $2}
  }
| subquery
  {
	$$= $1
  }
| tuple_expression
  {
	$$ = $1
  }
| EXISTS subquery
  {
	$$ = &ast.ExistsExpr{Subquery: $2}
  }
| MATCH openb select_expression_list closeb AGAINST openb bit_expr match_option closeb
  {
  $$ = &ast.MatchExpr{Columns: $3, Expr: $7, Option: $8}
  }
| CAST openb expression AS convert_type closeb
  {
    $$ = &ast.ConvertExpr{Expr: $3, Type: $5}
  }
| CONVERT openb expression ',' convert_type closeb
  {
    $$ = &ast.ConvertExpr{Expr: $3, Type: $5}
  }
| CONVERT openb expression USING charset closeb
  {
    $$ = &ast.ConvertUsingExpr{Expr: $3, Type: $5}
  }
| BINARY simple_expr %prec UNARY
  {
    // From: https://dev.psql.com/doc/refman/8.0/en/cast-functions.html#operator_binary
    // To convert a string expression to a binary string, these constructs are equivalent:
    //    CAST(expr AS BINARY)
    //    BINARY expr
    $$ = &ast.ConvertExpr{Expr: $2, Type: &ast.ConvertType{Type: $1}}
  }
| DEFAULT default_opt
  {
	 $$ = &ast.Default{ColName: $2}
  }
| INTERVAL simple_expr sql_id
  {
	// This rule prevents the usage of INTERVAL
	// as a function. If support is needed for that,
	// we'll need to revisit this. The solution
	// will be non-trivial because of grammar conflicts.
	$$ = &ast.IntervalExpr{Expr: $2, Unit: $3.String()}
  }
| column_name JSON_EXTRACT_OP text_literal_or_arg
  {
	$$ = &ast.BinaryExpr{Left: $1, Operator: ast.JSONExtractOp, Right: $3}
  }
| column_name JSON_UNQUOTE_EXTRACT_OP text_literal_or_arg
  {
	$$ = &ast.BinaryExpr{Left: $1, Operator: ast.JSONUnquoteExtractOp, Right: $3}
  }

trim_type:
  BOTH
  {
    $$ = ast.BothTrimType
  }
| LEADING
  {
    $$ = ast.LeadingTrimType
  }
| TRAILING
  {
    $$ = ast.TrailingTrimType
  }

default_opt:
  /* empty */
  {
    $$ = ""
  }
| openb id_or_var closeb
  {
    $$ = string($2.String())
  }

boolean_value:
  TRUE
  {
    $$ = ast.BoolVal(true)
  }
| FALSE
  {
    $$ = ast.BoolVal(false)
  }


is_suffix:
 TRUE
  {
    $$ = ast.IsTrueOp
  }
| NOT TRUE
  {
    $$ = ast.IsNotTrueOp
  }
| FALSE
  {
    $$ = ast.IsFalseOp
  }
| NOT FALSE
  {
    $$ = ast.IsNotFalseOp
  }

compare:
  '='
  {
    $$ = ast.EqualOp
  }
| '<'
  {
    $$ = ast.LessThanOp
  }
| '>'
  {
    $$ = ast.GreaterThanOp
  }
| LE
  {
    $$ = ast.LessEqualOp
  }
| GE
  {
    $$ = ast.GreaterEqualOp
  }
| NE
  {
    $$ = ast.NotEqualOp
  }
| NULL_SAFE_EQUAL
  {
    $$ = ast.NullSafeEqualOp
  }

col_tuple:
  row_tuple
  {
    $$ = $1
  }
| subquery
  {
    $$ = $1
  }
| LIST_ARG
  {
    $$ = ast.ListArg($1[2:])
    bindVariable(psqlex, $1[2:])
  }

subquery:
  query_expression_parens %prec SUBQUERY_AS_EXPR
  {
  	$$ = &ast.Subquery{$1}
  }

expression_list:
  expression
  {
    $$ = ast.Exprs{$1}
  }
| expression_list ',' expression
  {
    $$ = append($1, $3)
  }

/*
  Regular function calls without special token or syntax, guaranteed to not
  introduce side effects due to being a simple identifier
*/
function_call_generic:
  sql_id openb select_expression_list_opt closeb
  {
    $$ = &ast.FuncExpr{Name: $1, Exprs: $3}
  }
| sql_id openb DISTINCT select_expression_list closeb
  {
    $$ = &ast.FuncExpr{Name: $1, Distinct: true, Exprs: $4}
  }
| sql_id openb DISTINCTROW select_expression_list closeb
  {
    $$ = &ast.FuncExpr{Name: $1, Distinct: true, Exprs: $4}
  }
| table_id '.' reserved_sql_id openb select_expression_list_opt closeb
  {
    $$ = &ast.FuncExpr{Qualifier: $1, Name: $3, Exprs: $5}
  }

/*
  Function calls using reserved keywords, with dedicated grammar rules
  as a result
*/
function_call_keyword:
  LEFT openb select_expression_list closeb
  {
    $$ = &ast.FuncExpr{Name: ast.NewColIdent("left"), Exprs: $3}
  }
| RIGHT openb select_expression_list closeb
  {
    $$ = &ast.FuncExpr{Name: ast.NewColIdent("right"), Exprs: $3}
  }
| SUBSTRING openb expression ',' expression ',' expression closeb
  {
    $$ = &ast.SubstrExpr{Name: $3, From: $5, To: $7}
  }
| SUBSTRING openb expression ',' expression closeb
  {
    $$ = &ast.SubstrExpr{Name: $3, From: $5}
  }
| SUBSTRING openb expression FROM expression FOR expression closeb
  {
  	$$ = &ast.SubstrExpr{Name: $3, From: $5, To: $7}
  }
| SUBSTRING openb expression FROM expression closeb
  {
  	$$ = &ast.SubstrExpr{Name: $3, From: $5}
  }
| GROUP_CONCAT openb distinct_opt select_expression_list order_by_opt separator_opt limit_opt closeb
  {
    $$ = &ast.GroupConcatExpr{Distinct: $3, Exprs: $4, OrderBy: $5, Separator: $6, Limit: $7}
  }
| CASE expression_opt when_expression_list else_expression_opt END
  {
    $$ = &ast.CaseExpr{Expr: $2, Whens: $3, Else: $4}
  }
| VALUES openb column_name closeb
  {
    $$ = &ast.ValuesFuncExpr{Name: $3}
  }
| CURRENT_USER func_paren_opt
  {
    $$ =  &ast.FuncExpr{Name: ast.NewColIdent($1)}
  }

/*
  Function calls using non reserved keywords but with special syntax forms.
  Dedicated grammar rules are needed because of the special syntax
*/
function_call_nonkeyword:
/* doesn't support fsp */
UTC_DATE func_paren_opt
  {
    $$ = &ast.FuncExpr{Name:ast.NewColIdent("utc_date")}
  }
| now
  {
  	$$ = $1
  }
  // curdate
/* doesn't support fsp */
| CURRENT_DATE func_paren_opt
  {
    $$ = &ast.FuncExpr{Name:ast.NewColIdent("current_date")}
  }
| UTC_TIME func_datetime_precision
  {
    $$ = &ast.CurTimeFuncExpr{Name:ast.NewColIdent("utc_time"), Fsp: $2}
  }
  // curtime
| CURRENT_TIME func_datetime_precision
  {
    $$ = &ast.CurTimeFuncExpr{Name:ast.NewColIdent("current_time"), Fsp: $2}
  }
| TIMESTAMPADD openb sql_id ',' expression ',' expression closeb
  {
    $$ = &ast.TimestampFuncExpr{Name:string("timestampadd"), Unit:$3.String(), Expr1:$5, Expr2:$7}
  }
| TIMESTAMPDIFF openb sql_id ',' expression ',' expression closeb
  {
    $$ = &ast.TimestampFuncExpr{Name:string("timestampdiff"), Unit:$3.String(), Expr1:$5, Expr2:$7}
  }
| EXTRACT openb interval FROM expression closeb
  {
	$$ = &ast.ExtractFuncExpr{IntervalTypes: $3, Expr: $5}
  }
| WEIGHT_STRING openb expression convert_type_weight_string closeb
  {
    $$ = &ast.WeightStringFuncExpr{Expr: $3, As: $4}
  }
| JSON_PRETTY openb expression closeb
  {
    $$ = &ast.JSONPrettyExpr{JSONVal: $3}
  }
| JSON_STORAGE_FREE openb expression closeb
  {
    $$ = &ast.JSONStorageFreeExpr{ JSONVal: $3}
  }
| JSON_STORAGE_SIZE openb expression closeb
  {
    $$ = &ast.JSONStorageSizeExpr{ JSONVal: $3}
  }
| LTRIM openb expression closeb
  {
    $$ = &ast.TrimFuncExpr{TrimFuncType:ast.LTrimType, StringArg: $3}
  }
| RTRIM openb expression closeb
  {
    $$ = &ast.TrimFuncExpr{TrimFuncType:ast.RTrimType, StringArg: $3}
  }
| TRIM openb trim_type expression_opt FROM expression closeb
  {
    $$ = &ast.TrimFuncExpr{Type:$3, TrimArg:$4, StringArg: $6}
  }
| TRIM openb expression closeb
  {
    $$ = &ast.TrimFuncExpr{StringArg: $3}
  }
| TRIM openb expression FROM expression closeb
  {
    $$ = &ast.TrimFuncExpr{TrimArg:$3, StringArg: $5}
  }
| JSON_SCHEMA_VALID openb expression ',' expression closeb
  {
    $$ = &ast.JSONSchemaValidFuncExpr{ Schema: $3, Document: $5}
  }
| JSON_SCHEMA_VALIDATION_REPORT openb expression ',' expression closeb
  {
    $$ = &ast.JSONSchemaValidationReportFuncExpr{ Schema: $3, Document: $5}
  }
| JSON_ARRAY openb expression_list_opt closeb
  {
    $$ = &ast.JSONArrayExpr{ Params:$3 }
  }
| JSON_OBJECT openb json_object_param_opt closeb
  {
    $$ = &ast.JSONObjectExpr{ Params:$3 }
  }
| JSON_QUOTE openb expression closeb
  {
    $$ = &ast.JSONQuoteExpr{ StringArg:$3 }
  }
| JSON_CONTAINS openb expression ',' expression json_path_param_list_opt closeb
  {
    $$ = &ast.JSONContainsExpr{Target: $3, Candidate: $5, PathList: $6}
  }
| JSON_CONTAINS_PATH openb expression ',' expression ',' json_path_param_list closeb
  {
    $$ = &ast.JSONContainsPathExpr{JSONDoc: $3, OneOrAll: $5, PathList: $7}
  }
| JSON_EXTRACT openb expression ',' json_path_param_list closeb
  {
    $$ = &ast.JSONExtractExpr{JSONDoc: $3, PathList: $5}
  }
| JSON_KEYS openb expression json_path_param_list_opt closeb
  {
    $$ = &ast.JSONKeysExpr{JSONDoc: $3, PathList: $4}
  }
| JSON_OVERLAPS openb expression ',' expression closeb
  {
    $$ = &ast.JSONOverlapsExpr{JSONDoc1:$3, JSONDoc2:$5}
  }
| JSON_SEARCH openb expression ',' expression ',' expression closeb
  {
    $$ = &ast.JSONSearchExpr{JSONDoc: $3, OneOrAll: $5, SearchStr: $7 }
  }
| JSON_SEARCH openb expression ',' expression ',' expression ',' expression json_path_param_list_opt closeb
  {
    $$ = &ast.JSONSearchExpr{JSONDoc: $3, OneOrAll: $5, SearchStr: $7, EscapeChar: $9, PathList:$10 }
  }
| JSON_VALUE openb expression ',' json_path_param returning_type_opt closeb
  {
    $$ = &ast.JSONValueExpr{JSONDoc: $3, Path: $5, ReturningType: $6}
  }
| JSON_VALUE openb expression ',' json_path_param returning_type_opt on_empty closeb
  {
    $$ = &ast.JSONValueExpr{JSONDoc: $3, Path: $5, ReturningType: $6, EmptyOnResponse: $7}
  }
| JSON_VALUE openb expression ',' json_path_param returning_type_opt on_error closeb
  {
    $$ = &ast.JSONValueExpr{JSONDoc: $3, Path: $5, ReturningType: $6, ErrorOnResponse: $7}
  }
| JSON_VALUE openb expression ',' json_path_param returning_type_opt on_empty on_error closeb
  {
    $$ = &ast.JSONValueExpr{JSONDoc: $3, Path: $5, ReturningType: $6, EmptyOnResponse: $7, ErrorOnResponse: $8}
  }
| JSON_DEPTH openb expression closeb
  {
    $$ = &ast.JSONAttributesExpr{Type:ast.DepthAttributeType, JSONDoc:$3}
  }
| JSON_VALID openb expression closeb
  {
    $$ = &ast.JSONAttributesExpr{Type:ast.ValidAttributeType, JSONDoc:$3}
  }
| JSON_TYPE openb expression closeb
  {
    $$ = &ast.JSONAttributesExpr{Type:ast.TypeAttributeType, JSONDoc:$3}
  }
| JSON_LENGTH openb expression closeb
  {
    $$ = &ast.JSONAttributesExpr{Type:ast.LengthAttributeType, JSONDoc:$3 }
  }
| JSON_LENGTH openb expression ',' json_path_param closeb
  {
    $$ = &ast.JSONAttributesExpr{Type:ast.LengthAttributeType, JSONDoc:$3, Path: $5 }
  }
| JSON_ARRAY_APPEND openb expression ',' json_object_param_list closeb
  {
    $$ = &ast.JSONValueModifierExpr{Type:ast.JSONArrayAppendType ,JSONDoc:$3, Params:$5}
  }
| JSON_ARRAY_INSERT openb expression ',' json_object_param_list closeb
  {
    $$ = &ast.JSONValueModifierExpr{Type:ast.JSONArrayInsertType ,JSONDoc:$3, Params:$5}
  }
| JSON_INSERT openb expression ',' json_object_param_list closeb
  {
    $$ = &ast.JSONValueModifierExpr{Type:ast.JSONInsertType ,JSONDoc:$3, Params:$5}
  }
| JSON_REPLACE openb expression ',' json_object_param_list closeb
  {
    $$ = &ast.JSONValueModifierExpr{Type:ast.JSONReplaceType ,JSONDoc:$3, Params:$5}
  }
| JSON_SET openb expression ',' json_object_param_list closeb
  {
    $$ = &ast.JSONValueModifierExpr{Type:ast.JSONSetType ,JSONDoc:$3, Params:$5}
  }
| JSON_MERGE openb expression ',' expression_list closeb
  {
    $$ = &ast.JSONValueMergeExpr{Type: ast.JSONMergeType, JSONDoc: $3, JSONDocList: $5}
  }
| JSON_MERGE_PATCH openb expression ',' expression_list closeb
  {
    $$ = &ast.JSONValueMergeExpr{Type: ast.JSONMergePatchType, JSONDoc: $3, JSONDocList: $5}
  }
| JSON_MERGE_PRESERVE openb expression ',' expression_list closeb
  {
    $$ = &ast.JSONValueMergeExpr{Type: ast.JSONMergePreserveType, JSONDoc: $3, JSONDocList: $5}
  }
| JSON_REMOVE openb expression ',' expression_list closeb
  {
    $$ = &ast.JSONRemoveExpr{JSONDoc:$3, PathList: $5}
  }
| JSON_UNQUOTE openb expression closeb
  {
    $$ = &ast.JSONUnquoteExpr{JSONValue:$3}
  }

returning_type_opt:
  {
    $$ = nil
  }
| RETURNING convert_type
  {
    $$ = $2
  }

json_path_param_list_opt:
  {
    $$ = nil
  }
| ',' json_path_param_list
  {
    $$ = $2
  }

json_path_param_list:
  json_path_param
  {
    $$ = []ast.JSONPathParam{$1}
  }
| json_path_param_list ',' json_path_param
  {
    $$ = append($$, $3)
  }

json_path_param:
  text_literal_or_arg
  {
    $$ = ast.JSONPathParam($1)
  }
| column_name
  {
    $$ = ast.JSONPathParam($1)
  }

interval:
 interval_time_stamp
 {}
| DAY_HOUR
  {
	$$=ast.IntervalDayHour
  }
| DAY_MICROSECOND
  {
	$$=ast.IntervalDayMicrosecond
  }
| DAY_MINUTE
  {
	$$=ast.IntervalDayMinute
  }
| DAY_SECOND
  {
	$$=ast.IntervalDaySecond
  }
| HOUR_MICROSECOND
  {
	$$=ast.IntervalHourMicrosecond
  }
| HOUR_MINUTE
  {
	$$=ast.IntervalHourMinute
  }
| HOUR_SECOND
  {
	$$=ast.IntervalHourSecond
  }
| MINUTE_MICROSECOND
  {
	$$=ast.IntervalMinuteMicrosecond
  }
| MINUTE_SECOND
  {
	$$=ast.IntervalMinuteSecond
  }
| SECOND_MICROSECOND
  {
	$$=ast.IntervalSecondMicrosecond
  }
| YEAR_MONTH
  {
	$$=ast.IntervalYearMonth
  }

interval_time_stamp:
 DAY
  {
 	$$=ast.IntervalDay
  }
| WEEK
  {
  	$$=ast.IntervalWeek
  }
| HOUR
  {
 	$$=ast.IntervalHour
  }
| MINUTE
  {
 	$$=ast.IntervalMinute
  }
| MONTH
  {
	$$=ast.IntervalMonth
  }
| QUARTER
  {
	$$=ast.IntervalQuarter
  }
| SECOND
  {
	$$=ast.IntervalSecond
  }
| MICROSECOND
  {
	$$=ast.IntervalMicrosecond
  }
| YEAR
  {
	$$=ast.IntervalYear
  }

func_paren_opt:
  /* empty */
| openb closeb

func_datetime_precision:
  /* empty */
  {
  	$$ = nil
  }
| openb closeb
  {
    $$ = nil
  }
| openb VALUE_ARG closeb
  {
    $$ = ast.NewArgument($2[1:])
    bindVariable(psqlex, $2[1:])
  }

/*
  Function calls using non reserved keywords with *normal* syntax forms. Because
  the names are non-reserved, they need a dedicated rule so as not to conflict
*/
function_call_conflict:
  IF openb select_expression_list closeb
  {
    $$ = &ast.FuncExpr{Name: ast.NewColIdent("if"), Exprs: $3}
  }
| DATABASE openb select_expression_list_opt closeb
  {
    $$ = &ast.FuncExpr{Name: ast.NewColIdent("database"), Exprs: $3}
  }
| SCHEMA openb select_expression_list_opt closeb
  {
    $$ = &ast.FuncExpr{Name: ast.NewColIdent("schema"), Exprs: $3}
  }
| MOD openb select_expression_list closeb
  {
    $$ = &ast.FuncExpr{Name: ast.NewColIdent("mod"), Exprs: $3}
  }
| REPLACE openb select_expression_list closeb
  {
    $$ = &ast.FuncExpr{Name: ast.NewColIdent("replace"), Exprs: $3}
  }

match_option:
/*empty*/
  {
    $$ = ast.NoOption
  }
| IN BOOLEAN MODE
  {
    $$ = ast.BooleanModeOpt
  }
| IN NATURAL LANGUAGE MODE
 {
    $$ = ast.NaturalLanguageModeOpt
 }
| IN NATURAL LANGUAGE MODE WITH QUERY EXPANSION
 {
    $$ = ast.NaturalLanguageModeWithQueryExpansionOpt
 }
| WITH QUERY EXPANSION
 {
    $$ = ast.QueryExpansionOpt
 }

charset:
  sql_id
  {
    $$ = string($1.String())
  }
| STRING
  {
    $$ = string($1)
  }
| BINARY
  {
    $$ = string($1)
  }

convert_type_weight_string:
  /* empty */
  {
    $$ = nil
  }

convert_type:
  BINARY length_opt
  {
    $$ = &ast.ConvertType{Type: string($1), Length: $2}
  }
| CHAR length_opt charset_opt
  {
    $$ = &ast.ConvertType{Type: string($1), Length: $2, Charset: $3}
  }
| DATE
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }
| DECIMAL_TYPE decimal_length_opt
  {
    $$ = &ast.ConvertType{Type: string($1)}
    $$.Length = $2.Length
    $$.Scale = $2.Scale
  }
| JSON
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }
| NCHAR length_opt
  {
    $$ = &ast.ConvertType{Type: string($1), Length: $2}
  }
| SIGNED
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }
| SIGNED INTEGER
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }
| TIME length_opt
  {
    $$ = &ast.ConvertType{Type: string($1), Length: $2}
  }
| UNSIGNED
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }
| UNSIGNED INTEGER
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }
| FLOAT_TYPE length_opt
  {
    $$ = &ast.ConvertType{Type: string($1), Length: $2}
  }
| DOUBLE PRECISION
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }
| REAL
  {
    $$ = &ast.ConvertType{Type: string($1)}
  }


expression_opt:
  {
    $$ = nil
  }
| expression
  {
    $$ = $1
  }

separator_opt:
  {
    $$ = string("")
  }
| SEPARATOR STRING
  {
    $$ = " separator "+sql_types.EncodeStringSQL($2)
  }

when_expression_list:
  when_expression
  {
    $$ = []*ast.When{$1}
  }
| when_expression_list when_expression
  {
    $$ = append($1, $2)
  }

when_expression:
  WHEN expression THEN expression
  {
    $$ = &ast.When{Cond: $2, Val: $4}
  }

else_expression_opt:
  {
    $$ = nil
  }
| ELSE expression
  {
    $$ = $2
  }

column_name:
  sql_id
  {
    $$ = &ast.ColName{Name: $1}
  }
| table_id '.' reserved_sql_id
  {
    $$ = &ast.ColName{Qualifier: ast.TableName{Name: $1}, Name: $3}
  }
| table_id '.' reserved_table_id '.' reserved_sql_id
  {
    $$ = &ast.ColName{Qualifier: ast.TableName{Qualifier: $1, Name: $3}, Name: $5}
  }

num_val:
  sql_id
  {
    // TODO(sougou): ast.Deprecate this construct.
    if $1.Lowered() != "value" {
      psqlex.Error("expecting value after next")
      return 1
    }
    $$ = ast.NewIntLiteral("1")
  }
| VALUE_ARG VALUES
  {
    $$ = ast.NewArgument($1[1:])
    bindVariable(psqlex, $1[1:])
  }

group_by_opt:
  {
    $$ = nil
  }
| GROUP BY expression_list
  {
    $$ = $3
  }

having_opt:
  {
    $$ = nil
  }
| HAVING expression
  {
    $$ = $2
  }

order_by_opt:
  {
    $$ = nil
  }
 | order_by_clause
 {
 	$$ = $1
 }

order_by_clause:
ORDER BY order_list
  {
    $$ = $3
  }

order_list:
  order
  {
    $$ = ast.OrderBy{$1}
  }
| order_list ',' order
  {
    $$ = append($1, $3)
  }

order:
  expression asc_desc_opt
  {
    $$ = &ast.Order{Expr: $1, Direction: $2}
  }

asc_desc_opt:
  {
    $$ = ast.AscOrder
  }
| ASC
  {
    $$ = ast.AscOrder
  }
| DESC
  {
    $$ = ast.DescOrder
  }

limit_opt:
  {
    $$ = nil
  }
 | limit_clause
 {
 	$$ = $1
 }

limit_clause:
LIMIT expression
  {
    $$ = &ast.Limit{Rowcount: $2}
  }
| LIMIT expression ',' expression
  {
    $$ = &ast.Limit{Offset: $2, Rowcount: $4}
  }
| LIMIT expression OFFSET expression
  {
    $$ = &ast.Limit{Offset: $4, Rowcount: $2}
  }

security_view_opt:
  {
    $$ = ""
  }
| SQL SECURITY security_view
  {
    $$ = $3
  }

security_view:
  DEFINER
  {
    $$ = string($1)
  }
| INVOKER
  {
    $$ = string($1)
  }

check_option_opt:
  {
    $$ = ""
  }
| WITH cascade_or_local_opt CHECK OPTION
  {
    $$ = $2
  }

cascade_or_local_opt:
  {
    $$ = "cascaded"
  }
| CASCADED
  {
    $$ = string($1)
  }
| LOCAL
  {
    $$ = string($1)
  }

definer_opt:
  {
    $$ = nil
  }
| DEFINER '=' user
  {
    $$ = $3
  }

user:
CURRENT_USER
  {
    $$ = &ast.Definer{
    	Name: string($1),
    }
  }
| CURRENT_USER '(' ')'
  {
    $$ = &ast.Definer{
        Name: string($1),
    }
  }
| user_username address_opt
  {
    $$ = &ast.Definer{
        Name: $1,
        Address: $2,
    }
  }

user_username:
  STRING
  {
    $$ = sql_types.EncodeStringSQL($1)
  }
| ID
  {
    $$ = ast.FormatIdentifier($1)
  }

address_opt:
  {
    $$ = ""
  }
| AT_ID
  {
    $$ = ast.FormatAddress($1)
  }

locking_clause:
FOR UPDATE
  {
    $$ = ast.ForUpdateLock
  }
| LOCK IN SHARE MODE
  {
    $$ = ast.ShareModeLock
  }

into_clause:
INTO table_name
{
$$ = &ast.SelectInto{ExportOption:sql_types.EncodeStringSQL($2.Name.V)}
}

// insert_data expands all combinations into a single rule.
// This avoids a shift/reduce conflict while encountering the
// following two possible constructs:
// insert into t1(a, b) (select * from t2)
// insert into t1(select * from t2)
// Because the rules are together, the parser can keep shifting
// the tokens until it disambiguates a as sql_id and select as keyword.
insert_data:
  VALUES tuple_list
  {
    $$ = &ast.Insert{Rows: $2}
  }
| select_statement
  {
    $$ = &ast.Insert{Rows: $1}
  }
| openb ins_column_list closeb VALUES tuple_list
  {
    $$ = &ast.Insert{Columns: $2, Rows: $5}
  }
| openb closeb VALUES tuple_list
  {
    $$ = &ast.Insert{Rows: $4}
  }
| openb ins_column_list closeb select_statement
  {
    $$ = &ast.Insert{Columns: $2, Rows: $4}
  }

ins_column_list:
  sql_id
  {
    $$ = ast.Columns{$1}
  }
| sql_id '.' sql_id
  {
    $$ = ast.Columns{$3}
  }
| ins_column_list ',' sql_id
  {
    $$ = append($$, $3)
  }
| ins_column_list ',' sql_id '.' sql_id
  {
    $$ = append($$, $5)
  }

on_dup_opt:
  {
    $$ = nil
  }
| ON DUPLICATE KEY UPDATE update_list
  {
    $$ = $5
  }

tuple_list:
  tuple_or_empty
  {
    $$ = ast.Values{$1}
  }
| tuple_list ',' tuple_or_empty
  {
    $$ = append($1, $3)
  }

tuple_or_empty:
  row_tuple
  {
    $$ = $1
  }
| openb closeb
  {
    $$ = ast.ValTuple{}
  }

row_tuple:
  openb expression_list closeb
  {
    $$ = ast.ValTuple($2)
  }
tuple_expression:
 row_tuple
  {
    if len($1) == 1 {
      $$ = $1[0]
    } else {
      $$ = $1
    }
  }

update_list:
  update_expression
  {
    $$ = ast.UpdateExprs{$1}
  }
| update_list ',' update_expression
  {
    $$ = append($1, $3)
  }

update_expression:
  column_name '=' expression
  {
    $$ = &ast.UpdateExpr{Name: $1, Expr: $3}
  }

set_list:
  set_expression
  {
    $$ = ast.SetExprs{$1}
  }
| set_list ',' set_expression
  {
    $$ = append($1, $3)
  }

set_expression:
  reserved_sql_id '=' ON
  {
    $$ = &ast.SetExpr{Name: $1, Scope: ast.ImplicitScope, Expr: ast.NewStrLiteral("on")}
  }
| reserved_sql_id '=' OFF
  {
    $$ = &ast.SetExpr{Name: $1, Scope: ast.ImplicitScope, Expr: ast.NewStrLiteral("off")}
  }
| reserved_sql_id '=' expression
  {
    $$ = &ast.SetExpr{Name: $1, Scope: ast.ImplicitScope, Expr: $3}
  }
| charset_or_character_set_or_names charset_value collate_opt
  {
    $$ = &ast.SetExpr{Name: ast.NewColIdent(string($1)), Scope: ast.ImplicitScope, Expr: $2}
  }
|  set_session_or_global set_expression
  {
    $2.Scope = $1
    $$ = $2
  }

charset_or_character_set:
  CHARSET
| CHARACTER SET
  {
    $$ = "charset"
  }

charset_or_character_set_or_names:
  charset_or_character_set
| NAMES

charset_value:
  sql_id
  {
    $$ = ast.NewStrLiteral($1.String())
  }
| STRING
  {
    $$ = ast.NewStrLiteral($1)
  }
| DEFAULT
  {
    $$ = &ast.Default{}
  }

for_from:
  FOR
| FROM

temp_opt:
  { $$ = false }
| TEMPORARY
  { $$ = true }

exists_opt:
  { $$ = false }
| IF EXISTS
  { $$ = true }

not_exists_opt:
  { $$ = false }
| IF NOT EXISTS
  { $$ = true }

ignore_opt:
  { $$ = false }
| IGNORE
  { $$ = true }

to_opt:
  { $$ = struct{}{} }
| TO
  { $$ = struct{}{} }
| AS
  { $$ = struct{}{} }

call_statement:
  CALL table_name openb expression_list_opt closeb
  {
    $$ = &ast.CallProc{Name: $2, Params: $4}
  }

expression_list_opt:
  {
    $$ = nil
  }
| expression_list
  {
    $$ = $1
  }

using_opt:
  { $$ = nil }
| using_index_type
  { $$ = []*ast.IndexOption{$1} }

using_index_type:
  USING sql_id
  {
    $$ = &ast.IndexOption{Name: string($1), String: string($2.String())}
  }

sql_id:
  id_or_var
  {
    $$ = $1
  }
| non_reserved_keyword
  {
    $$ = ast.NewColIdent(string($1))
  }
| non_reserved_keyword_sql2023
  {
    $$ = ast.NewColIdent(string($1))
  }

reserved_sql_id:
  sql_id
| reserved_keyword
  {
    $$ = ast.NewColIdent(string($1))
  }

schema_id:
  id_or_var
  {
    $$ = ast.NewSchemaIdent(string($1.String()))
  }

sequence_id:
  id_or_var
  {
    $$ = ast.NewSequenceIdent(string($1.String()))
  }

table_id:
  id_or_var
  {
    $$ = ast.NewTableIdent(string($1.String()))
  }
| non_reserved_keyword
  {
    $$ = ast.NewTableIdent(string($1))
  }
| non_reserved_keyword_sql2023
  {
    $$ = ast.NewTableIdent(string($1))
  }

table_id_opt:
  /* empty */ %prec LOWER_THAN_CHARSET
  {
    $$ = ast.NewTableIdent("")
  }
| table_id
  {
    $$ = $1
  }

reserved_table_id:
  table_id
| reserved_keyword
  {
    $$ = ast.NewTableIdent(string($1))
  }

reserved_sequence_id:
  sequence_id
| reserved_keyword
  {
    $$ = ast.NewSequenceIdent(string($1))
  }

/*
  These are not all necessarily reserved in PostgresQL, but some are.

  These are more importantly reserved because they may conflict with our grammar.
  If you want to move one that is not reserved in PostgresQL (i.e. ESCAPE) to the
  non_reserved_keywords, you\'ll need to deal with any conflicts.

  Sorted alphabetically
*/
reserved_keyword:
  ALL
| ANALYSE
| ANALYZE
| AND
| ANY
| ARRAY
| AS
| ASC
| ASYMMETRIC
| AUTHORIZATION
| BINARY
| BOTH
| BYTEA
| CASE
| CAST
| CHECK
| COLLATE
| COLLATION
| COLUMN
| COMMENT
| CONCURRENTLY
| CONSTRAINT
| CREATE
| CROSS
| CURRENT_CATALOG
| CURRENT_DATE
| CURRENT_ROLE
| CURRENT_SCHEMA
| CURRENT_TIME
| CURRENT_TIMESTAMP
| CURRENT_USER
| DEFAULT
| DEFERRABLE
| DESC
| DISTINCT
| DO
| ELSE
| END
| EXCEPT
| FALSE
| FETCH
| FORCE_QUOTE
| FORCE_NOT_NULL
| FORCE_NULL
| FOREIGN
| FREEZE
| FROM
| FULL
| GRANT
| GROUP
| HAVING
| ILIKE
| IN
| INITIALLY
| INNER
| INTERSECT
| INTO
| IS
| ISNULL
| JOIN
| LATERAL
| LEADING
| LEFT
| LIKE
| LIMIT
| LOCALTIME
| LOCALTIMESTAMP
| LOG_VERBOSITY
| NATURAL
| NOT
| NOTNULL
| NULL
| OFFSET
| ON
| ON_ERROR
| ONLY
| OR
| ORDER
| OUTER
| OVERLAPS
| PLACING
| PRIMARY
| REFERENCES
| RETURNING
| RIGHT
| SELECT
| SESSION_USER
| SIMILAR
| SOME
| STOP
| SYMMETRIC
| SYSTEM_USER
| TABLE
| TABLESAMPLE
| THEN
| TO
| TRAILING
| TRUE
| UNION
| UNIQUE
| USER
| USING
| VARIADIC
| VERBOSE
| WHEN
| WHERE
| WINDOW
| WITH

/*
  These are non-reserved, because they don\'t cause conflicts in the grammar.
  Some of them may be reserved in PostgresQL. The good news is we use \" quote them
  when we rewrite the query, so no issue should arise.

  Sorted alphabetically
*/
non_reserved_keyword:
  ABORT
| ABSENT
| ABSOLUTE
| ACCESS
| ACTION
| ADD
| ADMIN
| AFTER
| AGGREGATE
| ALSO
| ALTER
| ALWAYS
| ASENSITIVE
| ASSERTION
| ASSIGNMENT
| AT
| ATOMIC
| ATTACH
| ATTRIBUTE
| BACKWARD
| BEFORE
| BEGIN
| BETWEEN
| BIGINT
| BIT
| BOOLEAN
| BREADTH
| BY
| CACHE
| CALL
| CALLED
| CASCADE
| CASCADED
| CATALOG
| CHAIN
| CHAR
| CHARACTER
| CHARACTERISTICS
| CHECKPOINT
| CLASS
| CLOSE
| CLUSTER
| COALESCE
| COLUMNS
| COMMENTS
| COMMIT
| COMMITTED
| COMPRESSION
| CONDITIONAL
| CONFIGURATION
| CONFLICT
| CONNECTION
| CONSTRAINTS
| CONTENT
| CONTINUE
| CONVERSION
| COPY
| COST
| CSV
| CUBE
| CURRENT
| CURSOR
| CYCLE
| DATA
| DATABASE
| DAY
| DEALLOCATE
| DEC
| DECIMAL
| DECLARE
| DEFAULTS
| DEFERRED
| DEFINER
| DELETE
| DELIMITER
| DELIMITERS
| DEPENDS
| DEPTH
| DETACH
| DICTIONARY
| DISABLE
| DISCARD
| DOCUMENT
| DOMAIN
| DOUBLE
| DROP
| EACH
| EMPTY
| ENABLE
| ENCODING
| ENCRYPTED
| ENUM
| ERROR
| ESCAPE
| EVENT
| EXCLUDE
| EXCLUDING
| EXCLUSIVE
| EXECUTE
| EXISTS
| EXPLAIN
| EXPRESSION
| EXTENSION
| EXTERNAL
| EXTRACT
| FAMILY
| FILTER
| FINALIZE
| FIRST
| FLOAT
| FOLLOWING
| FORCE
| FORMAT
| FORWARD
| FUNCTION
| FUNCTIONS
| GENERATED
| GLOBAL
| GRANTED
| GREATEST
| GROUPING
| GROUPS
| HANDLER
| HEADER
| HOLD
| HOUR
| IDENTITY
| IF
| IMMEDIATE
| IMMUTABLE
| IMPLICIT
| IMPORT
| INCLUDE
| INCLUDING
| INCREMENT
| INDENT
| INDEX
| INDEXES
| INHERIT
| INHERITS
| INLINE
| INOUT
| INPUT
| INSENSITIVE
| INSERT
| INSTEAD
| INT
| INTEGER
| INTERVAL
| INVOKER
| ISOLATION
| JSON
| JSON_ARRAY
| JSON_ARRAYAGG
| JSON_EXISTS
| JSON_OBJECT
| JSON_OBJECTAGG
| JSON_QUERY
| JSON_SCALAR
| JSON_SERIALIZE
| JSON_TABLE
| JSON_VALUE
| KEEP
| KEY
| KEYS
| LABEL
| LANGUAGE
| LARGE
| LAST
| LEAKPROOF
| LEAST
| LEVEL
| LISTEN
| LOAD
| LOCAL
| LOCATION
| LOCK
| LOCKED
| LOGGED
| MAPPING
| MATCH
| MATCHED
| MATERIALIZED
| MAXVALUE
| MERGE
| MERGE_ACTION
| METHOD
| MINUTE
| MINVALUE
| MODE
| MONTH
| MOVE
| NAME
| NAMES
| NATIONAL
| NCHAR
| NESTED
| NEW
| NEXT
| NFC
| NFD
| NFKC
| NFKD
| NO
| NONE
| NORMALIZE
| NORMALIZED
| NOTHING
| NOTIFY
| NOWAIT
| NULLIF
| NULLS
| NUMERIC
| OBJECT
| OF
| OFF
| OIDS
| OLD
| OMIT
| OPERATOR
| OPTION
| OPTIONS
| ORDINALITY
| OTHERS
| OUT
| OVER
| OVERLAY
| OVERRIDING
| OWNED
| OWNER
| PARALLEL
| PARAMETER
| PARSER
| PARTIAL
| PARTITION
| PASSING
| PASSWORD
| PATH
| PLAN
| PLANS
| POLICY
| POSITION
| PRECEDING
| PRECISION
| PREPARE
| PREPARED
| PRESERVE
| PRIOR
| PRIVILEGES
| PROCEDURAL
| PROCEDURE
| PROCEDURES
| PROGRAM
| PUBLICATION
| QUOTE
| QUOTES
| RANGE
| READ
| REAL
| REASSIGN
| RECHECK
| RECURSIVE
| REF
| REFERENCING
| REFRESH
| REINDEX
| RELATIVE
| RELEASE
| RENAME
| REPEATABLE
| REPLACE
| REPLICA
| RESET
| RESTART
| RESTRICT
| RETURN
| RETURNS
| REVOKE
| ROLE
| ROLLBACK
| ROLLUP
| ROUTINE
| ROUTINES
| ROW
| ROWS
| RULE
| SAVEPOINT
| SCALAR
| SCHEMA
| SCHEMAS
| SCROLL
| SEARCH
| SECOND
| SECURITY
| SEQUENCE
| SEQUENCES
| SERIALIZABLE
| SERVER
| SESSION
| SET
| SETOF
| SETS
| SHARE
| SHOW
| SIMPLE
| SKIP
| SMALLINT
| SNAPSHOT
| SOURCE
| SQL
| STABLE
| STANDALONE
| START
| STATEMENT
| STATISTICS
| STDIN
| STDOUT
| STORAGE
| STORED
| STRICT
| STRING
| STRIP
| SUBSCRIPTION
| SUBSTRING
| SUPPORT
| SYSID
| SYSTEM
| TABLES
| TABLESPACE
| TARGET
| TEMP
| TEMPLATE
| TEMPORARY
| TEXT
| TIES
| DATE
| TIME
| TIMESTAMP
| TRANSACTION
| TRANSFORM
| TREAT
| TRIGGER
| TRIM
| TRUNCATE
| TRUSTED
| TYPE
| TYPES
| UESCAPE
| UNBOUNDED
| UNCOMMITTED
| UNCONDITIONAL
| UNENCRYPTED
| UNKNOWN
| UNLISTEN
| UNLOGGED
| UNTIL
| UPDATE
| VACUUM
| VALID
| VALIDATE
| VALIDATOR
| VALUE
| VALUES
| VARCHAR
| VARYING
| VERSION
| VIEW
| VIEWS
| VOLATILE
| WHITESPACE
| WITHIN
| WITHOUT
| WORK
| WRAPPER
| WRITE
| XML
| XMLATTRIBUTES
| XMLCONCAT
| XMLELEMENT
| XMLEXISTS
| XMLFOREST
| XMLNAMESPACES
| XMLPARSE
| XMLPI
| XMLROOT
| XMLSERIALIZE
| XMLTABLE
| YEAR
| YES
| ZONE


/*
  These are non-reserved, in PostgresQL, but reserved in sql2023. The good news is we use \" quote them
  when we rewrite the query, so no issue should arise.

  Sorted alphabetically
*/
non_reserved_keyword_sql2023:
  ARRAY_MAX_CARDINALITY
| CHARACTER_SET_CATALOG
| COMMAND_FUNCTION_CODE
| CURRENT_DEFAULT_TRANSFORM_GROUP
| CURRENT_TRANSFORM_GROUP_FOR_TYPE
| DATETIME_INTERVAL_CODE
| DATETIME_INTERVAL_PRECISION
| DYNAMIC_FUNCTION_CODE
| END_EXEC
| PARAMETER_ORDINAL_POSITION
| PARAMETER_SPECIFIC_CATALOG
| PARAMETER_SPECIFIC_NAME
| PARAMETER_SPECIFIC_SCHEMA
| RETURNED_OCTET_LENGTH 
| TRANSACTIONS_COMMITTED
| TRANSACTIONS_ROLLED_BACK
| USER_DEFINED_TYPE_CATALOG
| USER_DEFINED_TYPE_CODE
| USER_DEFINED_TYPE_NAME
| USER_DEFINED_TYPE_SCHEMA

openb:
  '('
  {
    if incNesting(psqlex) {
      psqlex.Error("max nesting level reached")
      return 1
    }
  }

closeb:
  ')'
  {
    decNesting(psqlex)
  }

skip_to_end:
{
  skipToEnd(psqlex)
}

ddl_skip_to_end:
  {
    skipToEnd(psqlex)
  }
| openb
  {
    skipToEnd(psqlex)
  }
| reserved_sql_id
  {
    skipToEnd(psqlex)
  }
