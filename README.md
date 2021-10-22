# 第三部分：创建数据库及数据操作功能实践

## 实验目的

熟练掌握数据库管理系统中创建数据库、关系模式维护以及数据维护操作的实现技术。

## 实验内容

**注：以下所有功能的实现，要进行语法和语义检查，并注意维护相应的数据*字典文件。**

- 用高级语言建立数据库表。

  ~~~sql
  # ..,..
  (?P<att>(?:\s*\w+\s+(?:int|char)\s*,)*(?:\s*\w+\s+(?:int|char)\s*))
  # result: name type, name type
  create\s+table\s+(?P<name>\w+)\s*\((?P<att>(?:\s*\w+\s+(?:int|char)\s*,)*(?:\s*\w+\s+(?:int|char)\s*)+)\);?
  # //从... ...,... ...中匹配出... ...的集合
  \w+\s+\w+
  ~~~

  ~~~sql
  CREATE TABLE employee (
    id` CHAR,
    superior_id CHAR,
    department_id CHAR,
    name CHAr,
    birth_date CHAR,
    address CHAR,
    sex int,
    salary int
  );
  ~~~

  - 设计文件存储结构和存取方法。
  - 属性的个数任意，属性的类型至少包括整数和字符串。
  - 把表的相关信息存入数据字典。

- 用高级语言为关系表插入元组。

  ~~~sql
  # (),..();
  (?P<rows>(?:\s*\(\)\s*,\s*)*\(\));?
  # ..,..
  (?:\s*(?:\d+|'.+')\s*,)*\s*(?:\d+|'.+')\s*
  # result
  insert\s+into\s+(?P<name>\w+)\s+values(?P<rows>(?:\s*\((?:\s*(?:\d+|'[^']+')\s*,)*\s*(?:\d+|'[^']+')\s*\)\s*,)*\s*\((?:\s*(?:\d+|'[^']+')\s*,)*\s*(?:\d+|'[^']+')\s*\));?
  ~~~

  ~~~sql
  insert into employee values
  ('230101198009081234','23010119751201312X','d1','张三','1980-09-08','哈尔滨道里区十二道街',1,3125),
  ('230101198107023736','23010119751201312X','d1','李四','1980-09-08','哈尔滨道外区三道街',1,2980);
  ~~~

  - 用VALUES子句为新建立的关系插入元组。
  - 用VALUES子句在关系模式修改之后按照新的模式插入元组。（选做）

- 用高级语言实现属性的添加和删除功能。(选做)

  - 为关系表添加属性并维护数据字典。
  - 为关系表删除属性并维护数据字典。

- 用高级语言实现表中元组的删除和修改功能。

  ~~~sql
  (?i)update\s+(?P<name>\w+)\s+set\s+(?P<setAttr>\w+)\s*=\s*(?P<setVal>\d+|'[^']+')(?:\s+where\s+(?P<whAttr>\w+)\s*=\s*(?P<whVal>\d+|'[^']+'))?\s*;?
  (?i)delete\s+from\s+(?P<name>\w+)\s+where\s+(?P<attr>\w+)\s*=\s*(?P<val>\d+|'\w+')\s*;?                                                                     
  ~~~

  ~~~sql
  UPDATE employee
  SET salary = 3000
  where id = '1';
  
  delete from employee
  where id = '230101198009081234';
  ~~~

  - 实现删除关系表元组的功能，包括如下两种情况：

    - 没有WHERE条件，删除关系中的所有元组。
    - 指定WHERE条件，删除满足条件的元组。

  - 实现修改关系表元组的功能，包括如下两种情况（选做）：

    - 没有WHERE条件，修改所有元组的指定属性的值。

    - 指定WHERE条件，修改满足条件的元组的指定属性的值。

- 用高级语言实现表的删除功能

  ~~~sql
  (?i)drop\s+table\s+(?P<name>\w+);?
  ~~~

  ~~~sql
  drop table employee;
  ~~~

  - 删除表并维护数据字典。

- 用高级语言实现显示数据库表的功能，用于对上面的操作结果进行测试。

  ~~~sql
  select\s+\*\s+from\s+(?P<name>\w+);?
  ~~~

  ~~~sql
  select * from employee
  ~~~

  - 实现“SELECT * FROM 表名”。
  - 显示表的结构和内容。
