CREATE TABLE t1(a int, b int, c int, d int);
CREATE TABLE t2(a int, b int, c int, d int);

CREATE INDEX t1a ON t1(a);
CREATE INDEX t1b ON t1(b);

CREATE INDEX t2a ON t2(a);
CREATE INDEX t2b ON t2(b);