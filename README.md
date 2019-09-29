# mysql 查询简化工具
    只针对 mysql, queryRowContext,queryContext查询，自动转发为目标结构体或切片   
 
```sql 
CREATE TABLE `users` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(20) NOT NULL DEFAULT '',
  `phone` varchar(11) DEFAULT '',
  `front_cover` varchar(255) DEFAULT '',
  `address` varchar(255) DEFAULT '',
  `balance` int(11) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=latin1;
```



## 查询

```go
    var db *sql.DB
    query := "SELECT * FROM users WHERE ID < ?"
    list := make([]UserInfo, 0)
    err := QueryContext(ctx, db, &list, query, 10)
    if err != nil {
    	panic(err)
    }
    for _, v := range list {
    	fmt.Printf("info=%#v\n", v)
    }
    
    list2 := make([]*UserInfo, 0)
    err = QueryContext(ctx, db, &list2, query, 10)
    if err != nil {
    	panic(err)
    }
    for _, v := range list2 {
    	fmt.Printf("info2=%#v\n", v)
    }
```