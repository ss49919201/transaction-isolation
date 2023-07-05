DROP TABLE IF EXISTS "tbl";

CREATE TABLE tbl (
     id MEDIUMINT NOT NULL AUTO_INCREMENT,
     name CHAR(30) NOT NULL,
	 counter TINYINT NOT NULL,
     PRIMARY KEY (id)
);

INSERT INTO tbl (name, counter)
VALUES (1, 'A', 1);
