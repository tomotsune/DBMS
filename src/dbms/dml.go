// @Title  dbms
// @Description
// @Author  haipinHu  25/10/2021 18:14
// @Update  haipinHu  25/10/2021 18:14
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

// SelectTable 删除指定关系中满足where条件子句的元组
// e.g. select emloyee.name, emloyee.id
//FROM works_on,employee
//WHERE works_on.project_id = 'p2'
//AND works_on.employee_id = employee.id;
// 正则表达式将捕获上述SQL语句中的如下分组:
// 1 employee.name, emloyee.id 查询属性向量
// 2 works_on, employee 查询表向量
// 3 works_on.project_id = 'p2'  连接
// AND works_on.employee_id = employee.id
func SelectTable(selRe *regexp.Regexp, str string) (err error) {
	selField, schemaField, condField := selRe.FindStringSubmatch(str)[1], selRe.FindStringSubmatch(str)[2], selRe.FindStringSubmatch(str)[3]
	condMap := map[string]string{}
	if condField != "" {
		for _, row := range strings.Split(strings.ToLower(condField), "and") {
			var tempToken []string
			for _, v := range strings.Split(row, "=") {
				tempToken = append(tempToken, strings.TrimSpace(v))
			}
			condMap[strings.Split(tempToken[0], ".")[1]] = tempToken[1]
		}
	}

	// 1. 获取表头
	file, _ := os.Open("table.txt")
	defer file.Close()
	reader := bufio.NewReader(file)
	t := table{}
	for {
		tJson, err := reader.ReadBytes('\n')
		if err == io.EOF {
			file.Close()
			break
		}
		json.Unmarshal(tJson, &t)
		if t.Name == schemaField {
			break
		}
	}
	if t.Name == "" {
		err = fmt.Errorf("cannot open %v ,because of not existed", schemaField)
		return
	}

	// 3. 输出表头并记录选择属性索引
	var selIdxs, condIdxs []int
	for i, v := range t.Attr {
		if selField == "*" {
			fmt.Printf("%v(%v)\t", v["name"], v["type"])
			selIdxs = append(selIdxs, i)
		} else {
			var selAttrs []string
			for _, v := range strings.Split(selField, ",") {
				selAttrs = append(selAttrs, strings.TrimSpace(v))
			}
			for _, attr := range selAttrs {
				// 记录投影属性索引
				if v["name"] == strings.Split(attr, ".")[1] {
					fmt.Printf("%v(%v)\t", v["name"], v["type"])
					selIdxs = append(selIdxs, i)
				}
				// 记录选择属性索引
				if condMap[v["name"]] != "" {
					condIdxs = append(condIdxs, i)
				}
			}
		}
	}
	fmt.Print("\n---------------------------------------------------\n")

	file, err = os.Open(schemaField + ".txt")
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

		// where条件
		flag := true
		for _, idx := range condIdxs {
			if attrs[idx] != condMap[t.Attr[idx]["name"]] {
				flag = false
			}
		}
		for _, idx := range selIdxs {
			if flag {
				fmt.Printf("%v\t", attrs[idx])
			}
		}
		fmt.Println()
	}
	return
}
