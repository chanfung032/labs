// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

%{
package main
%}

%union {
  bytes       []byte
}

%token LEX_ERROR
%left <bytes> UNION
%token <bytes> SELECT INSERT UPDATE DELETE FROM WHERE GROUP HAVING ORDER BY LIMIT OFFSET FOR
%token <bytes> ALL DISTINCT AS EXISTS ASC DESC INTO DUPLICATE KEY DEFAULT SET LOCK
%token <bytes> VALUES LAST_INSERT_ID
%token <bytes> NEXT VALUE SHARE MODE
%token <bytes> SQL_NO_CACHE SQL_CACHE
%left <bytes> JOIN STRAIGHT_JOIN LEFT RIGHT INNER OUTER CROSS NATURAL USE FORCE
%left <bytes> ON
%token <empty> '(' ',' ')'
%token <bytes> ID HEX STRING INTEGRAL FLOAT HEXNUM VALUE_ARG LIST_ARG COMMENT
%token <bytes> NULL TRUE FALSE

// Precedence dictated by mysql. But the vitess grammar is simplified.
// Some of these operators don't conflict in our situation. Nevertheless,
// it's better to have these listed in the correct order. Also, we don't
// support all operators yet.
%left <bytes> OR
%left <bytes> AND
%right <bytes> NOT '!'
%left <bytes> BETWEEN CASE WHEN THEN ELSE END
%left <bytes> '=' '<' '>' LE GE NE NULL_SAFE_EQUAL IS LIKE REGEXP IN
%left <bytes> '|'
%left <bytes> '&'
%left <bytes> SHIFT_LEFT SHIFT_RIGHT
%left <bytes> '+' '-'
%left <bytes> '*' '/' DIV '%' MOD
%left <bytes> '^'
%right <bytes> '~' UNARY
%left <bytes> COLLATE
%right <bytes> BINARY
%right <bytes> INTERVAL
%nonassoc <bytes> '.'

// There is no need to define precedence for the JSON
// operators because the syntax is restricted enough that
// they don't cause conflicts.
%token <empty> JSON_EXTRACT_OP JSON_UNQUOTE_EXTRACT_OP

// DDL Tokens
%token <bytes> CREATE ALTER DROP RENAME ANALYZE
%token <bytes> TABLE INDEX VIEW TO IGNORE IF UNIQUE USING
%token <bytes> SHOW DESCRIBE EXPLAIN DATE ESCAPE REPAIR OPTIMIZE TRUNCATE

// Supported SHOW tokens
%token <bytes> DATABASES TABLES VITESS_KEYSPACES VITESS_SHARDS VSCHEMA_TABLES

// Convert Type Tokens
%token <bytes> INTEGER CHARACTER

// Functions
%token <bytes> CURRENT_TIMESTAMP DATABASE CURRENT_DATE
%token <bytes> CURRENT_TIME LOCALTIME LOCALTIMESTAMP
%token <bytes> UTC_DATE UTC_TIME UTC_TIMESTAMP
%token <bytes> REPLACE
%token <bytes> CONVERT CAST
%token <bytes> GROUP_CONCAT SEPARATOR

// Match
%token <bytes> MATCH AGAINST BOOLEAN LANGUAGE WITH QUERY EXPANSION

// MySQL reserved words that are unused by this grammar will map to this token.
%token <bytes> UNUSED

%type <statement> command
%type <selStmt> select_statement base_select union_lhs union_rhs
%type <statement> insert_statement update_statement delete_statement set_statement
%type <statement> create_statement alter_statement rename_statement drop_statement
%type <statement> analyze_statement show_statement other_statement
%type <bytes2> comment_opt comment_list
%type <str> union_op
%type <str> distinct_opt straight_join_opt cache_opt match_option separator_opt
%type <expr> like_escape_opt
%type <selectExprs> select_expression_list select_expression_list_opt
%type <selectExpr> select_expression
%type <expr> expression
%type <tableExprs> from_opt table_references
%type <tableExpr> table_reference table_factor join_table
%type <str> inner_join outer_join natural_join
%type <tableName> table_name into_table_name
%type <aliasedTableName> aliased_table_name
%type <indexHints> index_hint_list
%type <colIdents> index_list
%type <expr> where_expression_opt
%type <expr> condition
%type <boolVal> boolean_value
%type <str> compare
%type <ins> insert_data
%type <expr> value value_expression num_val
%type <expr> function_call_keyword function_call_nonkeyword function_call_generic function_call_conflict
%type <str> is_suffix
%type <colTuple> col_tuple
%type <exprs> expression_list
%type <values> tuple_list
%type <valTuple> row_tuple tuple_or_empty
%type <expr> tuple_expression
%type <subquery> subquery
%type <colName> column_name
%type <whens> when_expression_list
%type <when> when_expression
%type <expr> expression_opt else_expression_opt
%type <exprs> group_by_opt
%type <expr> having_opt
%type <orderBy> order_by_opt order_list
%type <order> order
%type <str> asc_desc_opt
%type <limit> limit_opt
%type <str> lock_opt
%type <columns> ins_column_list
%type <updateExprs> on_dup_opt
%type <updateExprs> update_list
%type <updateExpr> update_expression
%type <bytes> for_from
%type <str> ignore_opt
%type <byt> exists_opt
%type <empty> not_exists_opt non_rename_operation to_opt index_opt constraint_opt using_opt
%type <bytes> reserved_keyword non_reserved_keyword
%type <colIdent> sql_id reserved_sql_id col_alias as_ci_opt
%type <tableIdent> table_id reserved_table_id table_alias as_opt_id
%type <empty> as_opt
%type <str> charset
%type <convertType> convert_type
%type <str> show_statement_type
%start any_command

%%

any_command:
  command semicolon_opt

semicolon_opt:
/*empty*/
| ';'

command:
  select_statement
| insert_statement
| update_statement
| delete_statement
| set_statement
| create_statement
| alter_statement
| rename_statement
| drop_statement
| analyze_statement
| show_statement
| other_statement

select_statement:
  base_select order_by_opt limit_opt lock_opt
| union_lhs union_op union_rhs order_by_opt limit_opt lock_opt
| SELECT comment_opt cache_opt NEXT num_val for_from table_name

// base_select is an unparenthesized SELECT with no order by clause or beyond.
base_select:
  SELECT comment_opt cache_opt distinct_opt straight_join_opt select_expression_list from_opt where_expression_opt group_by_opt having_opt

union_lhs:
  select_statement
| openb select_statement closeb

union_rhs:
  base_select
| openb select_statement closeb


insert_statement:
  INSERT comment_opt ignore_opt into_table_name insert_data on_dup_opt
| INSERT comment_opt ignore_opt into_table_name SET update_list on_dup_opt

update_statement:
  UPDATE comment_opt aliased_table_name SET update_list where_expression_opt order_by_opt limit_opt

delete_statement:
  DELETE comment_opt FROM table_name where_expression_opt order_by_opt limit_opt

set_statement:
  SET comment_opt update_list

create_statement:
  CREATE TABLE not_exists_opt table_name
| CREATE constraint_opt INDEX ID using_opt ON table_name
| CREATE VIEW table_name
| CREATE OR REPLACE VIEW table_name

alter_statement:
  ALTER ignore_opt TABLE table_name non_rename_operation
| ALTER ignore_opt TABLE table_name RENAME to_opt table_name
| ALTER ignore_opt TABLE table_name RENAME index_opt
| ALTER VIEW table_name

rename_statement:
  RENAME TABLE table_name TO table_name

drop_statement:
  DROP TABLE exists_opt table_name
| DROP INDEX ID ON table_name
| DROP VIEW exists_opt table_name

analyze_statement:
  ANALYZE TABLE table_name

show_statement_type:
  ID
| reserved_keyword
| non_reserved_keyword

show_statement:
  SHOW show_statement_type

other_statement:
  DESCRIBE
| EXPLAIN
| REPAIR
| OPTIMIZE
| TRUNCATE

comment_opt:
  comment_list

comment_list:
| comment_list COMMENT

union_op:
  UNION
| UNION ALL
| UNION DISTINCT

cache_opt:
| SQL_NO_CACHE
| SQL_CACHE

distinct_opt:
| DISTINCT

straight_join_opt:
| STRAIGHT_JOIN

select_expression_list_opt:
| select_expression_list

select_expression_list:
  select_expression
| select_expression_list ',' select_expression

select_expression:
  '*'
| expression as_ci_opt
| table_id '.' '*'
| table_id '.' reserved_table_id '.' '*'

as_ci_opt:
| col_alias
| AS col_alias

col_alias:
  sql_id
| STRING

from_opt:
| FROM table_references

table_references:
  table_reference
| table_references ',' table_reference

table_reference:
  table_factor
| join_table

table_factor:
  aliased_table_name
| subquery as_opt table_id
| openb table_references closeb

aliased_table_name:
table_name as_opt_id index_hint_list

// There is a grammar conflict here:
// 1: INSERT INTO a SELECT * FROM b JOIN c ON b.i = c.i
// 2: INSERT INTO a SELECT * FROM b JOIN c ON DUPLICATE KEY UPDATE a.i = 1
// When yacc encounters the ON clause, it cannot determine which way to
// resolve. The %prec override below makes the parser choose the
// first construct, which automatically makes the second construct a
// syntax error. This is the same behavior as MySQL.
join_table:
  table_reference inner_join table_factor %prec JOIN
| table_reference inner_join table_factor ON expression
| table_reference outer_join table_reference ON expression
| table_reference natural_join table_factor

as_opt:
| AS

as_opt_id:
| table_alias
| AS table_alias

table_alias:
  table_id
| STRING

inner_join:
  JOIN
| INNER JOIN
| CROSS JOIN
| STRAIGHT_JOIN

outer_join:
  LEFT JOIN
| LEFT OUTER JOIN
| RIGHT JOIN
| RIGHT OUTER JOIN

natural_join:
 NATURAL JOIN
| NATURAL outer_join

into_table_name:
  INTO table_name
| table_name

table_name:
  table_id
| table_id '.' reserved_table_id

index_hint_list:
| USE INDEX openb index_list closeb
| IGNORE INDEX openb index_list closeb
| FORCE INDEX openb index_list closeb

index_list:
  sql_id
| index_list ',' sql_id

where_expression_opt:
| WHERE expression

expression:
  condition
| expression AND expression
| expression OR expression
| NOT expression
| expression IS is_suffix
| value_expression

boolean_value:
  TRUE
| FALSE

condition:
  boolean_value
| value_expression compare boolean_value
| value_expression compare value_expression
| value_expression IN col_tuple
| value_expression NOT IN col_tuple
| value_expression LIKE value_expression like_escape_opt
| value_expression NOT LIKE value_expression like_escape_opt
| value_expression REGEXP value_expression
| value_expression NOT REGEXP value_expression
| value_expression BETWEEN value_expression AND value_expression
| value_expression NOT BETWEEN value_expression AND value_expression
| EXISTS subquery

is_suffix:
  NULL
| NOT NULL
| TRUE
| NOT TRUE
| FALSE
| NOT FALSE

compare:
  '='
| '<'
| '>'
| LE
| GE
| NE
| NULL_SAFE_EQUAL

like_escape_opt:
| ESCAPE value_expression

col_tuple:
  row_tuple
| subquery
| LIST_ARG

subquery:
  openb select_statement closeb

expression_list:
  expression
| expression_list ',' expression

charset:
  ID
| STRING
| BINARY
| DATE

value_expression:
  value
| column_name
| tuple_expression
| subquery
| value_expression '&' value_expression
| value_expression '|' value_expression
| value_expression '^' value_expression
| value_expression '+' value_expression
| value_expression '-' value_expression
| value_expression '*' value_expression
| value_expression '/' value_expression
| value_expression DIV value_expression
| value_expression '%' value_expression
| value_expression MOD value_expression
| value_expression SHIFT_LEFT value_expression
| value_expression SHIFT_RIGHT value_expression
| column_name JSON_EXTRACT_OP value
| column_name JSON_UNQUOTE_EXTRACT_OP value
| value_expression COLLATE charset
| BINARY value_expression %prec UNARY
| '+'  value_expression %prec UNARY
| '-'  value_expression %prec UNARY
| '~'  value_expression
| '!' value_expression %prec UNARY
| INTERVAL value_expression sql_id
| function_call_generic
| function_call_keyword
| function_call_nonkeyword
| function_call_conflict

/*
  Regular function calls without special token or syntax, guaranteed to not
  introduce side effects due to being a simple identifier
*/
function_call_generic:
  sql_id openb select_expression_list_opt closeb
| sql_id openb DISTINCT select_expression_list closeb
| table_id '.' reserved_sql_id openb select_expression_list_opt closeb

/*
  Function calls using reserved keywords, with dedicated grammar rules
  as a result
*/
function_call_keyword:
  LEFT openb select_expression_list closeb
| RIGHT openb select_expression_list closeb
| CONVERT openb expression ',' convert_type closeb
| CONVERT openb expression USING convert_type closeb
| CAST openb expression AS convert_type closeb
| MATCH openb select_expression_list closeb AGAINST openb value_expression match_option closeb
| GROUP_CONCAT openb distinct_opt select_expression_list order_by_opt separator_opt closeb
| CASE expression_opt when_expression_list else_expression_opt END
| VALUES openb sql_id closeb

/*
  Function calls using non reserved keywords but with special syntax forms.
  Dedicated grammar rules are needed because of the special syntax
*/
function_call_nonkeyword:
  CURRENT_TIMESTAMP func_datetime_precision_opt
| UTC_TIMESTAMP func_datetime_precision_opt
| UTC_TIME func_datetime_precision_opt
| UTC_DATE func_datetime_precision_opt
  // now
| LOCALTIME func_datetime_precision_opt
  // now
| LOCALTIMESTAMP func_datetime_precision_opt
  // curdate
| CURRENT_DATE func_datetime_precision_opt
  // curtime
| CURRENT_TIME func_datetime_precision_opt

func_datetime_precision_opt:
  /* empty */
| openb closeb

/*
  Function calls using non reserved keywords with *normal* syntax forms. Because
  the names are non-reserved, they need a dedicated rule so as not to conflict
*/
function_call_conflict:
  IF openb select_expression_list closeb
| DATABASE openb select_expression_list_opt closeb
| MOD openb select_expression_list closeb
| REPLACE openb select_expression_list closeb

match_option:
/*empty*/
| IN BOOLEAN MODE
| IN NATURAL LANGUAGE MODE
| IN NATURAL LANGUAGE MODE WITH QUERY EXPANSION
| WITH QUERY EXPANSION


convert_type:
  charset
|  charset INTEGER
| charset openb INTEGRAL closeb
| charset openb INTEGRAL closeb charset
| charset openb INTEGRAL closeb CHARACTER SET charset
| charset charset
| charset CHARACTER SET charset
| charset openb INTEGRAL ',' INTEGRAL closeb

expression_opt:
| expression

separator_opt:
| SEPARATOR STRING

when_expression_list:
  when_expression
| when_expression_list when_expression

when_expression:
  WHEN expression THEN expression

else_expression_opt:
| ELSE expression

column_name:
  sql_id
| table_id '.' reserved_sql_id
| table_id '.' reserved_table_id '.' reserved_sql_id

value:
  STRING
| HEX
| INTEGRAL
| FLOAT
| HEXNUM
| VALUE_ARG
| NULL

num_val:
  sql_id
| INTEGRAL VALUES
| VALUE_ARG VALUES

group_by_opt:
| GROUP BY expression_list

having_opt:
| HAVING expression

order_by_opt:
| ORDER BY order_list

order_list:
  order
| order_list ',' order

order:
  expression asc_desc_opt

asc_desc_opt:
| ASC
| DESC

limit_opt:
| LIMIT expression
| LIMIT expression ',' expression
| LIMIT expression OFFSET expression

lock_opt:
| FOR UPDATE
| LOCK IN SHARE MODE

// insert_data expands all combinations into a single rule.
// This avoids a shift/reduce conflict while encountering the
// following two possible constructs:
// insert into t1(a, b) (select * from t2)
// insert into t1(select * from t2)
// Because the rules are together, the parser can keep shifting
// the tokens until it disambiguates a as sql_id and select as keyword.
insert_data:
  VALUES tuple_list
| select_statement
| openb select_statement closeb
| openb ins_column_list closeb VALUES tuple_list
| openb ins_column_list closeb select_statement
| openb ins_column_list closeb openb select_statement closeb

ins_column_list:
  sql_id
| sql_id '.' sql_id
| ins_column_list ',' sql_id
| ins_column_list ',' sql_id '.' sql_id

on_dup_opt:
| ON DUPLICATE KEY UPDATE update_list

tuple_list:
  tuple_or_empty
| tuple_list ',' tuple_or_empty

tuple_or_empty:
  row_tuple
| openb closeb

row_tuple:
  openb expression_list closeb

tuple_expression:
  row_tuple

update_list:
  update_expression
| update_list ',' update_expression

update_expression:
  column_name '=' expression

for_from:
  FOR
| FROM

exists_opt:
| IF EXISTS

not_exists_opt:
| IF NOT EXISTS

ignore_opt:
| IGNORE

non_rename_operation:
  ALTER
| DEFAULT
| DROP
| ORDER
| CONVERT
| UNUSED
| ID

to_opt:
| TO
| AS

index_opt:
  INDEX
| KEY

constraint_opt:
| UNIQUE
| sql_id

using_opt:
| USING sql_id

sql_id:
  ID
| non_reserved_keyword

reserved_sql_id:
  sql_id
| reserved_keyword

table_id:
  ID
| non_reserved_keyword

reserved_table_id:
  table_id
| reserved_keyword

/*
  These are not all necessarily reserved in MySQL, but some are.
  These are more importantly reserved because they may conflict with our grammar.
  If you want to move one that is not reserved in MySQL (i.e. ESCAPE) to the
  non_reserved_keywords, you'll need to deal with any conflicts.

  Sorted alphabetically
*/
reserved_keyword:
  AND
| AS
| ASC
| BETWEEN
| BINARY
| BY
| CASE
| COLLATE
| CONVERT
| CREATE
| CROSS
| CURRENT_DATE
| CURRENT_TIME
| CURRENT_TIMESTAMP
| DATABASE
| DATABASES
| DEFAULT
| DELETE
| DESC
| DESCRIBE
| DISTINCT
| DIV
| DROP
| ELSE
| END
| ESCAPE
| EXISTS
| EXPLAIN
| FALSE
| FOR
| FORCE
| FROM
| GROUP
| HAVING
| IF
| IGNORE
| IN
| INDEX
| INNER
| INSERT
| INTERVAL
| INTO
| IS
| JOIN
| KEY
| LEFT
| LIKE
| LIMIT
| LOCALTIME
| LOCALTIMESTAMP
| LOCK
| MATCH
| MOD
| NATURAL
| NEXT // next should be doable as non-reserved, but is not due to the special `select next num_val` query that vitess supports
| NOT
| NULL
| ON
| OR
| ORDER
| OUTER
| REGEXP
| RENAME
| REPLACE
| RIGHT
| SELECT
| SEPARATOR
| SET
| VITESS_KEYSPACES
| VITESS_SHARDS
| VSCHEMA_TABLES
| SHOW
| STRAIGHT_JOIN
| TABLE
| TABLES
| THEN
| TO
| TRUE
| UNION
| UNIQUE
| UPDATE
| USE
| USING
| UTC_DATE
| UTC_TIME
| UTC_TIMESTAMP
| VALUES
| WHEN
| WHERE

/*
  These are non-reserved Vitess, because they don't cause conflicts in the grammar.
  Some of them may be reserved in MySQL. The good news is we backtick quote them
  when we rewrite the query, so no issue should arise.

  Sorted alphabetically
*/
non_reserved_keyword:
  AGAINST
| DATE
| DUPLICATE
| EXPANSION
| INTEGER
| LANGUAGE
| MODE
| OFFSET
| OPTIMIZE
| QUERY
| REPAIR
| SHARE
| TRUNCATE
| UNUSED
| VIEW
| WITH

openb:
  '('

closeb:
  ')'
