package main

import (
	"fmt"
	"for_test/src/dbms"
	"regexp"
)

func main() {
	str := `CREATE table employee (
	id char,
	superior_id char,
	department_id char,
	name char,
	birth_date char,
	address char,
	sex int,
	salary int
	);`

	//str := `insert into employee values
	//('230101198009081234','23010119751201312X','d1','张三','1980-09-08','哈尔滨道里区十二道街',1,3125),
	//('230101198107023736','23010119751201312X','d1','李四','1980-09-08','哈尔滨道外区三道街',1,2980);`

	//str := `Select * from  employee;`

	//str := `UPDATE employee
	//SET salary = 3000
	//where id = '230101198009081234';`

	//str := `delete from employee
	//where id = '230101198009081234';`

	// str := `drop table employee;`

	err := process(str)
	if err != nil {
		fmt.Println(err)
	}
}
func process(str string) (err error) {
	cteRe := regexp.MustCompile(`(?i)create\s+table\s+(?P<name>\w+)\s*\((?P<att>(?:\s*\w+\s+(?:int|char)\s*,)*(?:\s*\w+\s+(?:int|char)\s*)+)\);?`)
	droRe := regexp.MustCompile(`(?i)drop\s+table\s+(?P<name>\w+);?`)
	insRe := regexp.MustCompile(`(?i)insert\s+into\s+(?P<name>\w+)\s+values(?P<rows>(?:\s*\((?:\s*(?:\d+|'[^']+')\s*,)*\s*(?:\d+|'[^']+')\s*\)\s*,)*\s*\((?:\s*(?:\d+|'[^']+')\s*,)*\s*(?:\d+|'[^']+')\s*\));?`)
	selRe := regexp.MustCompile(`(?i)select\s+\*\s+from\s+(?P<name>\w+);?`)
	udaRe := regexp.MustCompile(`(?i)update\s+(?P<name>\w+)\s+set\s+(?P<setAttr>\w+)\s*=\s*(?P<setVal>\d+|'[^']+')(?:\s+where\s+(?P<whAttr>\w+)\s*=\s*(?P<whVal>\d+|'[^']+'))?\s*;?`)
	delRe := regexp.MustCompile(`(?i)delete\s+from\s+(?P<name>\w+)\s+where\s+(?P<attr>\w+)\s*=\s*(?P<val>\d+|'\w+')\s*;?`)
	if cteRe.MatchString(str) {
		err = dbms.CreateTable(cteRe, str)
	} else if droRe.MatchString(str) {
		err = dbms.DropTable(droRe, str)
	} else if insRe.MatchString(str) {
		err = dbms.InsertTable(insRe, str)
	} else if selRe.MatchString(str) {
		err = dbms.SelectTable(selRe, str)
	} else if udaRe.MatchString(str) {
		err = dbms.UpdateTuple(udaRe, str)
	} else if delRe.MatchString(str) {
		err = dbms.DeleteTuple(delRe, str)
	} else {
		fmt.Println("error instruction")
	}
	return
}
