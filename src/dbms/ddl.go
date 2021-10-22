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

// UpdateTuple 更新指定关系的元组, 如果sql语句含有where查询, 则只需要更新满足条件的元组
// where 语句中只允许包含一个条件语句
// eg. UPDATE employee SET salary = 3000  where id = '1';
// 正则表达式将捕获上述SQL语句中的如下分组:
// 1: 关系名称 employee
// 2: 设置属性名 salary
// 3. 设置属性值 3000
// 4. 条件属性名 id
// 5. 条件属性值 '1'
func UpdateTuple(udaRe *regexp.Regexp, str string) (err error) {
	// 从关系中查询设置属性名和条件属性名的位置索引
	file, err := os.Open("table.txt")
	if err != nil {
		return
	}
	defer file.Close()
	setIndex, whIndex := -1, -1
	reader := bufio.NewReader(file)
	tn := udaRe.FindStringSubmatch(str)[1]
	for {
		tJson, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		t := table{}
		err = json.Unmarshal(tJson, &t)
		if err != nil {
			return err
		}
		if t.Name != tn {
			continue
		}
		for i, v := range t.Attr {
			// 找出需要设置的属性列索引
			if v["name"] == udaRe.FindStringSubmatch(str)[2] {
				setIndex = i
			}
			// 找出条件属性索引
			if v["name"] == udaRe.FindStringSubmatch(str)[4] {
				whIndex = i
			}
		}
	}
	file.Close()
	if setIndex == -1 {
		err = fmt.Errorf("cannot update, no haveing such attr: %v\n", udaRe.FindStringSubmatch(str)[2])
		return
	}

	// 修改元素
	err = modify(tn, func(tJson []byte, writer *bufio.Writer) (err error) {
		var attrs []string
		err = json.Unmarshal(tJson, &attrs)
		if err != nil {
			return
		}
		// 对指定关系的所有元组, 只有当条件存在且条件属性值不满足时才能跳过修改
		// 逆否命题 : 条件不存在 或 与条件属性匹配的的元组
		if whIndex == -1 || attrs[whIndex] == udaRe.FindStringSubmatch(str)[5] {
			attrs[setIndex] = udaRe.FindStringSubmatch(str)[3]
		}
		tJson, _ = json.Marshal(attrs)
		writer.Write(tJson)
		writer.WriteByte('\n')
		return
	})
	return
}

// DeleteTuple 删除指定关系中满足where条件子句的元组
// e.g. delete from employee where id = '230101198009081234';
// 正则表达式将捕获上述SQL语句中的如下分组:
// 1: 关系名称 employee
// 2: 条件属性名 id
// 3. 条件属性值 '230101198009081234'
func DeleteTuple(delRe *regexp.Regexp, str string) (err error) {
	// 从table表中找出条件属性的索引
	file, err := os.Open("table.txt")
	if err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	tn, index := delRe.FindStringSubmatch(str)[1], -1

	// 从关系中查询条件属性名的位置索引, 与更新类似
	for {
		tJson, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		t := table{}
		err = json.Unmarshal(tJson, &t)
		if err != nil {
			return err
		}
		if t.Name != tn {
			continue
		}
		for i, v := range t.Attr {
			// 找出需要设置的属性列索引
			if v["name"] == delRe.FindStringSubmatch(str)[2] {
				index = i
			}
		}
	}
	file.Close()
	if index == -1 {
		return
	}

	err = modify(tn, func(tJson []byte, writer *bufio.Writer) (err error) {
		var attrs []string
		err = json.Unmarshal(tJson, &attrs)
		if err != nil {
			return
		}
		if attrs[index] != delRe.FindStringSubmatch(str)[3] {
			writer.Write(tJson)
			writer.WriteByte('\n')
		}
		return
	})
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

// SelectTable 显示指定的关系
func SelectTable(selRe *regexp.Regexp, str string) (err error) {
	file, _ := os.Open("table.txt")
	defer file.Close()
	reader := bufio.NewReader(file)
	t := table{}
	for {
		tJson, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		json.Unmarshal(tJson, &t)
		if t.Name == selRe.FindStringSubmatch(str)[1] {
			break
		}
	}

	if t.Name == "" {
		err = fmt.Errorf("cannot drop %v ,because of not existed", selRe.FindStringSubmatch(str)[1])
		return
	}
	for _, v := range t.Attr {
		fmt.Printf("%v(%v)\t", v["name"], v["type"])
	}
	fmt.Print("\n---------------------------------------------------\n")
	file.Close()
	file, err = os.Open(selRe.FindStringSubmatch(str)[1] + ".txt")
	if err != nil {
		return
	}
	defer file.Close()
	reader = bufio.NewReader(file)
	for {
		tJson, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		var attrs []string
		json.Unmarshal(tJson, &attrs)
		for _, attr := range attrs {
			fmt.Printf("%v\t", attr)
		}
		fmt.Println()
	}
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
