up: 
	docker-compose up
	
down: 
	docker-compose down

reset: 
	docker-compose down
	docker-compose up

redis: 
	make down
	docker-compose up redis

db: 
	make down
	docker-compose up datastore

rmq: 
	make down
	docker-compose up rabbitmq
