SELECT * FROM t1, t2 WHERE t1.a = t2.a AND t1.b in (1, 7, 9, 10) AND t2.b in (11, 14, 20);