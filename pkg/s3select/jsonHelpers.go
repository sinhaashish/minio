/*
 * Minio Cloud Storage, (C) 2018 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package s3select

import (
	"fmt"
	"io"
	"log"

	"github.com/tidwall/gjson"
	"github.com/xwb1989/sqlparser"
)

//
func (reader *JSONInput) jsonRead() map[string]interface{} {
	dec := reader.reader
	var m interface{}
	for dec.More() {
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		return m.(map[string]interface{})
	}
	return nil
}

func jsonValue(input string, row string) string {
	value := gjson.Get(row, input)
	return value.String()

}

// It evaluates the where clause for JSON and return true if condition suffices
func jsonWhereClause(row string, columnNames map[string]int, alias string, whereClause interface{}) (bool, error) {

	var operator string
	var operand interface{}
	if fmt.Sprintf("%v", whereClause) == "false" {
		return false, nil
	}

	switch expr := whereClause.(type) {
	case *sqlparser.IsExpr:
		// To be Implemented
	case *sqlparser.RangeCond:
		operator = expr.Operator
		if operator != "between" && operator != "not between" {
			return false, ErrUnsupportedSQLOperation
		}
		if operator == "not between" {
			myResult, err := jsonEvaluateBetween(expr, alias, row, columnNames)
			if err != nil {
				return false, err
			}
			return !myResult, nil
		}
		myResult, err := jsonEvaluateBetween(expr, alias, row, columnNames)
		if err != nil {
			return false, err
		}
		return myResult, nil
	case *sqlparser.ComparisonExpr:
		operator = expr.Operator
		switch right := expr.Right.(type) {
		// case *sqlparser.FuncExpr:
		// 	fmt.Println(" In FuncExpr")
		// 	operand = evaluateFuncExpr(right, "", row, columnNames)
		case *sqlparser.SQLVal:

			var err error
			operand, err = evaluateParserType(right)
			if err != nil {
				return false, err
			}
		}

		//	evaluateOperator()

		switch left := expr.Left.(type) {
		// case *sqlparser.FuncExpr:
		// 	myVal = evaluateFuncExpr(left, "", row, columnNames)
		// 	fmt.Println(" In FuncExpr in Right")
		// 	conversionColumn = ""
		case *sqlparser.ColName:
			return evaluateOperator(jsonValue((left.Name.CompliantName()), row), operator, operand)

		}

	case *sqlparser.AndExpr:
		var leftVal bool
		var rightVal bool
		switch left := expr.Left.(type) {
		case *sqlparser.ComparisonExpr:
			temp, err := jsonWhereClause(row, columnNames, alias, left)
			if err != nil {
				return false, err
			}
			leftVal = temp
		}
		switch right := expr.Right.(type) {
		case *sqlparser.ComparisonExpr:
			temp, err := jsonWhereClause(row, columnNames, alias, right)
			if err != nil {
				return false, err
			}
			rightVal = temp
		}
		return (rightVal && leftVal), nil

	case *sqlparser.OrExpr:
		var leftVal bool
		var rightVal bool
		switch left := expr.Left.(type) {
		// var colToVal interface{}
		// var colFromVal interface{}
		// var conversionColumn string
		// var funcName string
		// switch colTo := betweenExpr.To.(type) {
		// case sqlparser.Expr:
		// 	switch colToMyVal := colTo.(type) {
		// 	case *sqlparser.FuncExpr:
		// 		var temp string
		// 		temp = stringOps(colToMyVal, record, "", columnNames)
		// 		colToVal = []byte(temp)
		// 	case *sqlparser.SQLVal:
		// 		var err error
		// 		colToVal, err = evaluateParserType(colToMyVal)
		// 		if err != nil {
		// 			return false, err
		// 		}
		// 	}
		// }
		// switch colFrom := betweenExpr.From.(type) {
		// case sqlparser.Expr:
		// 	switch colFromMyVal := colFrom.(type) {
		// 	case *sqlparser.FuncExpr:
		// 		colFromVal = stringOps(colFromMyVal, record, "", columnNames)
		// 	case *sqlparser.SQLVal:
		// 		var err error
		// 		colFromVal, err = evaluateParserType(colFromMyVal)
		// 		if err != nil {
		// 			return false, err
		// 		}
		// 	}
		// }
		// var myFuncVal string
		// myFuncVal = ""
		// switch left := betweenExpr.Left.(type) {
		// case *sqlparser.FuncExpr:
		// 	myFuncVal = evaluateFuncExpr(left, "", record, columnNames)
		// 	conversionColumn = ""
		// case *sqlparser.ColName:
		// 	conversionColumn = cleanCol(left.Name.CompliantName(), alias)
		// }

		// toGreater, err := evaluateOperator(fmt.Sprintf("%v", colToVal), ">", colFromVal)
		// if err != nil {
		// 	return false, err
		// }
		// if toGreater {
		// 	return evalBetweenGreater(conversionColumn, record, funcName, columnNames, colFromVal, colToVal, myFuncVal)
		// }
		// return evalBetweenLess(conversionColumn, record, funcName, columnNames, colFromVal, colToVal, myFuncVal)
		case *sqlparser.ComparisonExpr:
			leftVal, _ = jsonWhereClause(row, columnNames, alias, left)
		}
		switch right := expr.Right.(type) {
		case *sqlparser.ComparisonExpr:
			rightVal, _ = jsonWhereClause(row, columnNames, alias, right)
		}
		return (rightVal || leftVal), nil

	}

	return true, nil
}

// jsonEvaluateBetween is a function which evaluates a Between Clause.
func jsonEvaluateBetween(betweenExpr *sqlparser.RangeCond, alias string, record string, columnNames map[string]int) (bool, error) {
	fmt.Println(" In jsonEvaluateBetween  alias %#v \n record %#v \n   columnNames  %#v", alias, record, columnNames)

	// var colToVal interface{}
	// var colFromVal interface{}
	// var conversionColumn string
	// var funcName string
	// switch colTo := betweenExpr.To.(type) {
	// case sqlparser.Expr:
	// 	switch colToMyVal := colTo.(type) {
	// 	case *sqlparser.FuncExpr:
	// 		var temp string
	// 		temp = stringOps(colToMyVal, record, "", columnNames)
	// 		colToVal = []byte(temp)
	// 	case *sqlparser.SQLVal:
	// 		var err error
	// 		colToVal, err = evaluateParserType(colToMyVal)
	// 		if err != nil {
	// 			return false, err
	// 		}
	// 	}
	// }
	// switch colFrom := betweenExpr.From.(type) {
	// case sqlparser.Expr:
	// 	switch colFromMyVal := colFrom.(type) {
	// 	case *sqlparser.FuncExpr:
	// 		colFromVal = stringOps(colFromMyVal, record, "", columnNames)
	// 	case *sqlparser.SQLVal:
	// 		var err error
	// 		colFromVal, err = evaluateParserType(colFromMyVal)
	// 		if err != nil {
	// 			return false, err
	// 		}
	// 	}
	// }
	// var myFuncVal string
	// myFuncVal = ""
	// switch left := betweenExpr.Left.(type) {
	// case *sqlparser.FuncExpr:
	// 	myFuncVal = evaluateFuncExpr(left, "", record, columnNames)
	// 	conversionColumn = ""
	// case *sqlparser.ColName:
	// 	conversionColumn = cleanCol(left.Name.CompliantName(), alias)
	// }

	// toGreater, err := evaluateOperator(fmt.Sprintf("%v", colToVal), ">", colFromVal)
	// if err != nil {
	// 	return false, err
	// }
	// if toGreater {
	// 	return evalBetweenGreater(conversionColumn, record, funcName, columnNames, colFromVal, colToVal, myFuncVal)
	// }
	// return evalBetweenLess(conversionColumn, record, funcName, columnNames, colFromVal, colToVal, myFuncVal)
	return false, nil
}
