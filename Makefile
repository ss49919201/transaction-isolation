.PHONY: dc-up exec exec-mysql init-db

dc-up:
	docker compose up -d

exec:dc-up
	docker exec -it mydb /bin/bash

exec-mysql:dc-up
	docker exec -it mydb /bin/bash -c "mysql -u root -ppassword mydb"

init-db:dc-up
	docker exec -it mydb mysql -u root -ppassword mydb -e "$(shell cat ./script/init.sql)"
