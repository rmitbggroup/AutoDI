select concat('drop table ',table_name,';') from information_schema.TABLES where table_schema='test';

load data local infile 'path/imdb/aka_name.csv' into table aka_name fields terminated by ',' enclosed by '"' lines terminated by '\n';

SELECT a.TABLE_NAME, a.index_name, GROUP_CONCAT(column_name ORDER BY seq_in_index) AS 'Columns' FROM information_schema.statistics a where a.TABLE_SCHEMA='test' GROUP BY a.TABLE_NAME,a.index_name;

drop table aka_name;
drop table aka_title;
drop table cast_info;
drop table char_name;
drop table comp_cast_type;
drop table company_name;
drop table company_type;
drop table complete_cast;
drop table info_type;
drop table keyword;
drop table kind_type;
drop table link_type;
drop table movie_companies;
drop table movie_info;
drop table movie_info_idx;
drop table movie_keyword;
drop table movie_link;
drop table name;
drop table person_info;
drop table role_type;
drop table title;

load data local infile 'path/imdb/aka_name.csv' into table aka_name fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/aka_title.csv' into table aka_title fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/cast_info.csv' into table cast_info fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/char_name.csv' into table char_name fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/comp_cast_type.csv' into table comp_cast_type fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/company_name.csv' into table company_name fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/company_type.csv' into table company_type fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/complete_cast.csv' into table complete_cast fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/info_type.csv' into table info_type fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/keyword.csv' into table keyword fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/kind_type.csv' into table kind_type fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/link_type.csv' into table link_type fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/movie_companies.csv' into table movie_companies fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/movie_info.csv' into table movie_info fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/movie_info_idx.csv' into table movie_info_idx fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/movie_keyword.csv' into table movie_keyword fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/movie_link.csv' into table movie_link fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/name.csv' into table name fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/person_info.csv' into table person_info fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/role_type.csv' into table role_type fields terminated by ',' enclosed by '"' lines terminated by '\n';
load data local infile 'path/imdb/title.csv' into table title fields terminated by ',' enclosed by '"' lines terminated by '\n';

LOAD DATA LOCAL INFILE 'path/nation.tbl' INTO TABLE nation FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';
LOAD DATA LOCAL INFILE 'path/region.tbl' INTO TABLE region FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';
LOAD DATA LOCAL INFILE 'path/part.tbl' INTO TABLE part FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';
LOAD DATA LOCAL INFILE 'path/supplier.tbl' INTO TABLE supplier FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';
LOAD DATA LOCAL INFILE 'path/partsupp.tbl' INTO TABLE partsupp FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';
LOAD DATA LOCAL INFILE 'path/customer.tbl' INTO TABLE customer FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';
LOAD DATA LOCAL INFILE 'path/orders.tbl' INTO TABLE orders FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';
LOAD DATA LOCAL INFILE 'path/lineitem.tbl' INTO TABLE lineitem FIELDS TERMINATED BY '|' LINES TERMINATED BY '|\n';

analyze table nation;
analyze table orders;
analyze table lineitem;
analyze table supplier;
analyze table partsupp;
analyze table region;
analyze table customer;
analyze table part;

delete from aka_name;
delete from aka_title;
delete from cast_info;
delete from char_name;
delete from comp_cast_type;
delete from company_name;
delete from company_type;
delete from complete_cast;
delete from info_type;
delete from keyword;
delete from kind_type;
delete from link_type;
delete from movie_companies;
delete from movie_info;
delete from movie_info_idx;
delete from movie_keyword;
delete from movie_link;
delete from name;
delete from person_info;
delete from role_type;
delete from title;

analyze table aka_name;
analyze table aka_title;
analyze table cast_info;
analyze table char_name;
analyze table comp_cast_type;
analyze table company_name;
analyze table company_type;
analyze table complete_cast;
analyze table info_type;
analyze table keyword;
analyze table kind_type;
analyze table link_type;
analyze table movie_companies;
analyze table movie_info;
analyze table movie_keyword;
analyze table movie_link;
analyze table name;
analyze table person_info;
analyze table role_type;
analyze table title;

./dumpling \
  -u root \
  -P 4000 \
  -h 127.0.0.1 \
  --filetype sql \
  -t 8 \
  -o path/imdbdump \
  -r 200000 \
  -F 256MiB