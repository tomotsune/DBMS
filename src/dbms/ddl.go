package dbms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type table struct {
	Name string              `json:"name"`
	Attr []map[string]string `json:"attr"`
}

// CreateTable 创建一个关系结构
// e.g. CREATE TABLE employee (id CHAR,superior_id CHAR,department_id CHAR,name CHAr,birth_date CHAR,address CHAR,sex int,salary int);
// 正则表达式将捕获上述SQL语句中的如下分组:
// 1: 关系名称 employee
// 2: 属性向量 id char, superior_id char .... salary int
func CreateTable(cteRe *regexp.Regexp, str string) (err error) {
	file, err := os.OpenFile("table.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer file.Close()
	reader, tn, t := bufio.NewReader(file), cteRe.FindStringSubmatch(str)[1], table{}
	for {
		tJson, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}

		err = json.Unmarshal(tJson, &t)
		if err != nil {
			return err
		}
		if t.Name == tn {
			break
		}
	}
	if t.Name != "" {
		err = fmt.Errorf("%v already being", tn)
		return
	}
	// 将属性向量解析为 "id char" 形式的顺序结构
	attrs := regexp.MustCompile(`(?m)\w+\s+\w+`).FindAllString(cteRe.FindStringSubmatch(str)[2], -1)

	// 包装为 {tableName, [{AttrName, AttrType}..]} 形式的json对象
	var attrMapList []map[string]string
	for _, attr := range attrs {
		row := strings.Split(attr, " ")
		attrMapList = append(attrMapList, map[string]string{
			"name": strings.TrimSpace(row[0]),
			"type": strings.ToLower(strings.TrimSpace(row[1])),
		})
	}
	tJson, _ := json.Marshal(table{Name: cteRe.FindStringSubmatch(str)[1], Attr: attrMapList})

	// 将json写table文件, 并创建指定的关系文件
	os.Create(tn + ".txt")
	writer := bufio.NewWriter(file)
	writer.Write(tJson)
	writer.WriteByte('\n')
	writer.Flush()
	return
}

// InsertTable 先关系中插入一条或多条元组
// e.g. insert into employee values ('230101198009081234','23010119751201312X','d1','张三','1980-09-08','哈尔滨道里区十二道街',1,3125),
//('230101198107023736','23010119751201312X','d1','李四','1980-09-08','哈尔滨道外区三道街',1,2980);
// 正则表达式将捕获上述SQL语句中的如下分组:
// 1: 关系名称 employee
// 2: 属性值向量组 ('230101198009081234','23010119751201312X',...,3125), ('230101198107023736','23010119751201312X',..,2980)
func InsertTable(insRe *regexp.Regexp, str string) (err error) {
	// 读取数据字典中的关系定义, 获得关系模式以完成类型检查
	tableFile, err := os.Open("table.txt")
	if err != nil {
		return err
	}
	defer tableFile.Close()
	tableReader, t := bufio.NewReader(tableFile), table{}
	for {
		tJson, err := tableReader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		err = json.Unmarshal(tJson, &t)
		if err != nil {
			return err
		}
		if t.Name == insRe.FindStringSubmatch(str)[1] {
			break
		}
	}
	if t.Name == "" {
		return
	}

	tupleFile, err := os.OpenFile(t.Name+".txt", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer tupleFile.Close()
	tupleWriter := bufio.NewWriter(tupleFile)

	// 将属性值向量组解析为 "'230101198009081234','23010119751201312X',...,3125" 形式的顺序表结构
	cenRe := regexp.MustCompile(`(?:\s*(?:\d+|'.+')\s*,)*\s*(?:\d+|'.+')\s*`)
	for i, tuple := range cenRe.FindAllString(insRe.FindStringSubmatch(str)[2], -1) {
		var attrs []string
		// 逐个进行类型检查并封装成顺序表结构
		for j, attr := range strings.Split(tuple, ",") {
			attrType := t.Attr[j]["type"]
			if (attr[0] != '\'' && attrType == "char") || (attr[0] == '\'' && attrType == "int") {
				err = fmt.Errorf("data type not in line at (%v,%v)", i+1, j+1)
				return
			}
			attrs = append(attrs, strings.TrimSpace(attr))
		}
		tJson, _ := json.Marshal(attrs)
		tupleWriter.Write(tJson)
		tupleWriter.WriteByte('\n')
	}
	tupleWriter.Flush()
	return
}

// DropTable 从数据字典中删除指定的关系模式
// e.g. drop table employee;
// 正则表达式将捕获上述SQL语句中的如下分组:
// 1: 关系名称 employee
func DropTable(droRe *regexp.Regexp, str string) (err error) {
	err = modify("table", func(tJson []byte, writer *bufio.Writer) (err error) {
		t := table{}
		err = json.Unmarshal(tJson, &t)
		if err != nil {
			return
		}
		if t.Name != droRe.FindStringSubmatch(str)[1] {
			writer.Write(tJson)
			writer.WriteByte('\n')
		}
		return
	})
	err = os.Remove(droRe.FindStringSubmatch(str)[1] + ".txt")
	return
}

// modify 创建一个临时文件用于保存修改后的文件, 该临时文件将覆盖原文件以完成修改
// tn 木匾文件名
// fun 写入文件
func modify(tn string, fun func([]byte, *bufio.Writer) error) (err error) {
	// 删除操作逻辑
	file, err := os.Open(tn + ".txt")
	if err != nil {
		return
	}
	defer file.Close()
	tempFile, err := os.OpenFile("~"+tn+".txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	defer tempFile.Close()

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(tempFile)
	for {
		tJson, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		err = fun(tJson, writer)
		if err != nil {
			break
		}
	}
	writer.Flush()
	file.Close()
	tempFile.Close()
	os.Remove(tn + ".txt")
	os.Rename("~"+tn+".txt", tn+".txt")
	return
}
